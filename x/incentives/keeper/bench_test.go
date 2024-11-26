package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// genRewardCoins takes coins and returns a randomized coin struct used as rewards for the distribution benchmark.
func genRewardCoins(r *rand.Rand, coins sdk.Coins) (res sdk.Coins) {
	numCoins := 1 + r.Intn(Min(coins.Len(), 1))
	denomIndices := r.Perm(numCoins)
	for i := 0; i < numCoins; i++ {
		denom := coins[denomIndices[i]].Denom
		amt, _ := simtypes.RandPositiveInt(r, coins[i].Amount)
		res = append(res, sdk.Coin{Denom: denom, Amount: amt})
	}

	return
}

// genQueryCondition takes coins and durations and returns a QueryConditon struct.
func genQueryCondition(
	r *rand.Rand,
	coins sdk.Coins,
	durationOptions []time.Duration,
) lockuptypes.QueryCondition {
	// only use lockQueryType ByDuration (0) since ByTime (1) is deprecated
	lockQueryType := 0
	denom := coins[r.Intn(len(coins))].Denom
	durationOption := r.Intn(len(durationOptions))
	duration := durationOptions[durationOption]
	timestamp := time.Time{}

	return lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.LockQueryType(lockQueryType),
		Denom:         denom,
		Duration:      duration,
		Timestamp:     timestamp,
	}
}

// benchmarkDistributionLogic creates gauges with lockups that get distributed to. Benchmarks the performance of the distribution process.
func benchmarkDistributionLogic(b *testing.B, numAccts, numDenoms, numGauges, numLockups, numDistrs int) {
	b.Helper()
	b.StopTimer()

	blockStartTime := time.Now().UTC()
	app, cleanupFn := app.SetupTestingAppWithLevelDb(false)
	defer cleanupFn()
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: blockStartTime})

	r := rand.New(rand.NewSource(10))

	// setup accounts with balances
	addrs := []sdk.AccAddress{}
	for i := 0; i < numAccts; i++ {
		addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		coins := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000)}
		for j := 0; j < numDenoms; j++ {
			coins = coins.Add(sdk.NewInt64Coin(fmt.Sprintf("token%d", j), r.Int63n(100000000)))
		}
		_ = testutil.FundAccount(ctx, app.BankKeeper, addr, coins)
		app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
		addrs = append(addrs, addr)
	}

	distrEpoch := app.EpochsKeeper.GetEpochInfo(ctx, app.IncentivesKeeper.GetParams(ctx).DistrEpochIdentifier)
	durationOptions := app.IncentivesKeeper.GetLockableDurations(ctx)

	// setup gauges
	gaugeIds := []uint64{}
	for i := 0; i < numGauges; i++ {
		addr := addrs[r.Int()%numAccts]
		simCoins := app.BankKeeper.SpendableCoins(ctx, addr)

		// isPerpetual := r.Int()%2 == 0
		isPerpetual := true
		distributeTo := genQueryCondition(r, simCoins, durationOptions)
		rewards := genRewardCoins(r, simCoins)
		startTime := ctx.BlockTime().Add(time.Duration(-1) * time.Second)
		durationMillisecs := distributeTo.Duration.Milliseconds()
		numEpochsPaidOver := uint64(1)
		if !isPerpetual {
			millisecsPerEpoch := distrEpoch.Duration.Milliseconds()
			numEpochsPaidOver = uint64(r.Int63n(durationMillisecs/millisecsPerEpoch)) + 1
		}

		gaugeId, err := app.IncentivesKeeper.CreateGauge(ctx, isPerpetual, addr, rewards, distributeTo, startTime, numEpochsPaidOver, 0)
		if err != nil {
			fmt.Printf("Create Gauge, %v\n", err)
			b.FailNow()
		} else {
			gaugeIds = append(gaugeIds, gaugeId)
		}
	}

	// jump time to the future
	futureSecs := r.Intn(1 * 60 * 60 * 24 * 7)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(futureSecs) * time.Second))

	lockSecs := r.Intn(1 * 60 * 60 * 8)
	// setup lockups
	for i := 0; i < numLockups; i++ {
		addr := addrs[i%numAccts]
		simCoins := app.BankKeeper.SpendableCoins(ctx, addr)

		if i%10 == 0 {
			lockSecs = r.Intn(1 * 60 * 60 * 8)
		}
		duration := time.Duration(lockSecs) * time.Second
		_, err := app.LockupKeeper.CreateLock(ctx, addr, simCoins, duration)
		if err != nil {
			fmt.Printf("Lock tokens, %v\n", err)
			b.FailNow()
		}
	}
	fmt.Println("created all lockups")

	// begin distribution for all gauges
	for _, gaugeId := range gaugeIds {
		gauge, _ := app.IncentivesKeeper.GetGaugeByID(ctx, gaugeId)
		err := app.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(ctx, *gauge)
		if err != nil {
			fmt.Printf("Begin distribution, %v\n", err)
			b.FailNow()
		}
	}

	b.StartTimer()
	// distribute coins from gauges to lockup owners
	for i := 0; i < numDistrs; i++ {
		gauges := []types.Gauge{}
		for _, gaugeId := range gaugeIds {
			gauge, _ := app.IncentivesKeeper.GetGaugeByID(ctx, gaugeId)
			gauges = append(gauges, *gauge)
		}
		_, err := app.IncentivesKeeper.Distribute(ctx, gauges)
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkDistributionLogicTiny(b *testing.B) {
	numAccts := 1
	numDenoms := 1
	numGauges := 1
	numLockups := 1
	numDistrs := 1
	benchmarkDistributionLogic(b, numAccts, numDenoms, numGauges, numLockups, numDistrs)
}

func BenchmarkDistributionLogicSmall(b *testing.B) {
	numAccts := 10
	numDenoms := 1
	numGauges := 10
	numLockups := 1000
	numDistrs := 100
	benchmarkDistributionLogic(b, numAccts, numDenoms, numGauges, numLockups, numDistrs)
}

func BenchmarkDistributionLogicMedium(b *testing.B) {
	numAccts := 1000
	numDenoms := 8
	numGauges := 30
	numLockups := 20000
	numDistrs := 1

	benchmarkDistributionLogic(b, numAccts, numDenoms, numGauges, numLockups, numDistrs)
}

func BenchmarkDistributionLogicLarge(b *testing.B) {
	numAccts := 50000
	numDenoms := 10
	numGauges := 60
	numLockups := 100000
	numDistrs := 1

	benchmarkDistributionLogic(b, numAccts, numDenoms, numGauges, numLockups, numDistrs)
}

func BenchmarkDistributionLogicHuge(b *testing.B) {
	numAccts := 1000
	numDenoms := 100
	numGauges := 1000
	numLockups := 1000
	numDistrs := 30000
	benchmarkDistributionLogic(b, numAccts, numDenoms, numGauges, numLockups, numDistrs)
}
