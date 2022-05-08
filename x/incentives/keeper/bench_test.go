package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/osmosis/v3/app"
	lockuptypes "github.com/osmosis-labs/osmosis/v3/x/lockup/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

func genQueryCondition(
	r *rand.Rand,
	blocktime time.Time,
	coins sdk.Coins,
	durationOptions []time.Duration,
) lockuptypes.QueryCondition {
	lockQueryType := lockuptypes.ByDuration
	denom := coins[r.Intn(len(coins))].Denom
	durationOption := r.Intn(len(durationOptions))
	duration := durationOptions[durationOption]
	timestamp := time.Time{}

	return lockuptypes.QueryCondition{
		LockQueryType: lockQueryType,
		Denom:         denom,
		Duration:      duration,
		Timestamp:     timestamp,
	}
}

func benchmarkDistributionLogic(numAccts, numDenoms, numGauges, numLockups, numDistrs int, b *testing.B) {
	// b.ReportAllocs()
	b.StopTimer()

	blockStartTime := time.Now().UTC()
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: blockStartTime})

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	expected_num_events := numDenoms * (numLockups / numDenoms) * (numGauges / numDenoms) * 3
	ctx.EventManager().IncreaseCapacity(expected_num_events)
	fmt.Println("num events", expected_num_events)

	// setup accounts with balances
	addrs := []sdk.AccAddress{}
	for i := 0; i < numAccts; i++ {
		addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		coins := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000)}
		for j := 0; j < numDenoms; j++ {
			coins = coins.Add(sdk.NewInt64Coin(fmt.Sprintf("token%d", j), r.Int63n(100000000)))
		}
		app.BankKeeper.SetBalances(ctx, addr, coins)
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
		distributeTo := genQueryCondition(r, ctx.BlockTime(), simCoins, durationOptions)
		rewards := genRewardCoins(r, simCoins)
		// startTimeSecs := r.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
		startTime := ctx.BlockTime().Add(time.Duration(-1) * time.Second)
		durationMillisecs := distributeTo.Duration.Milliseconds()
		numEpochsPaidOver := uint64(1)
		if !isPerpetual {
			millisecsPerEpoch := distrEpoch.Duration.Milliseconds()
			numEpochsPaidOver = uint64(r.Int63n(durationMillisecs/millisecsPerEpoch)) + 1
		}

		gaugeId, err := app.IncentivesKeeper.CreateGauge(ctx, isPerpetual, addr, rewards, distributeTo, startTime, numEpochsPaidOver)
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

	// setup lockups
	for i := 0; i < numLockups; i++ {
		addr := addrs[r.Int()%numAccts]
		simCoins := app.BankKeeper.SpendableCoins(ctx, addr)
		duration := time.Second
		_, err := app.LockupKeeper.LockTokens(ctx, addr, simCoins, duration)
		if err != nil {
			fmt.Printf("Lock tokens, %v\n", err)
			b.FailNow()
		}
	}

	// begin distribution for all gauges
	for _, gaugeId := range gaugeIds {
		gauge, _ := app.IncentivesKeeper.GetGaugeByID(ctx, gaugeId)
		err := app.IncentivesKeeper.BeginDistribution(ctx, *gauge)
		if err != nil {
			fmt.Printf("Begin distribution, %v\n", err)
			b.FailNow()
		}
	}

	b.StartTimer()
	// distribute coins from gauges to lockup owners
	for i := 0; i < numDistrs; i++ {
		for _, gaugeId := range gaugeIds {
			gauge, _ := app.IncentivesKeeper.GetGaugeByID(ctx, gaugeId)
			_, err := app.IncentivesKeeper.Distribute(ctx, *gauge)
			if err != nil {
				fmt.Printf("Distribute, %v\n", err)
				b.FailNow()
			}
		}
	}
}

func BenchmarkDistributionLogicTiny(b *testing.B) {
	benchmarkDistributionLogic(1, 1, 1, 1, 1, b)
}

func BenchmarkDistributionLogicSmall(b *testing.B) {
	benchmarkDistributionLogic(10, 1, 10, 1000, 100, b)
}

func BenchmarkDistributionLogicMedium(b *testing.B) {
	numAccts := 1000
	numDenoms := 8
	numGauges := 30
	numLockups := 20000
	numDistrs := 1

	benchmarkDistributionLogic(numAccts, numDenoms, numGauges, numLockups, numDistrs, b)
}

func BenchmarkDistributionLogicLarge(b *testing.B) {
	benchmarkDistributionLogic(100, 10, 100, 100, 5000, b)
}

func BenchmarkDistributionLogicHuge(b *testing.B) {
	benchmarkDistributionLogic(1000, 100, 1000, 1000, 30000, b)
}
