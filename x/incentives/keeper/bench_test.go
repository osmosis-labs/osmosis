package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/osmosis/app"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
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

func genQueryCondition(r *rand.Rand, blocktime time.Time, coins sdk.Coins) lockuptypes.QueryCondition {
	lockQueryType := r.Intn(2)
	denom := coins[r.Intn(len(coins))].Denom
	durationSecs := r.Intn(1*60*60*24*7) + 1*60*60 // range of 1 week, min 1 hour
	duration := time.Duration(durationSecs) * time.Second
	timestampSecs := r.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
	timestamp := blocktime.Add(time.Duration(timestampSecs) * time.Second)

	return lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.LockQueryType(lockQueryType),
		Denom:         denom,
		Duration:      duration,
		Timestamp:     timestamp,
	}
}

func benchmarkDistributionLogic(numAccts int64, numGauges int64, numLockups int64, numDistrs int64, b *testing.B) {
	b.ReportAllocs()

	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// setup accounts with balances
	addrs := []sdk.AccAddress{}
	for i := int64(0); i < numAccts; i++ {
		addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
		app.BankKeeper.SetBalance(ctx, addr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
		addrs = append(addrs, addr)
	}

	// setup gauges
	gaugeIds := []uint64{}
	for i := int64(0); i < numGauges; i++ {
		addr := addrs[0]
		simCoins := app.BankKeeper.SpendableCoins(ctx, addr)

		isPerpetual := r.Int()%2 == 0
		distributeTo := genQueryCondition(r, ctx.BlockTime(), simCoins)
		rewards := genRewardCoins(r, simCoins)
		startTimeSecs := r.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
		startTime := ctx.BlockTime().Add(time.Duration(startTimeSecs) * time.Second)
		durationSecs := r.Intn(1*60*60*24*7) + 1*60*60*24 // range of 1 week, min 1 day
		numEpochsPaidOver := uint64(r.Int63n(int64(durationSecs)/(app.EpochsKeeper.GetEpochInfo(ctx, app.IncentivesKeeper.GetParams(ctx).DistrEpochIdentifier).Duration.Milliseconds()/1000))) + 1
		if isPerpetual {
			numEpochsPaidOver = 1
		}

		gaugeId, err := app.IncentivesKeeper.CreateGauge(ctx, isPerpetual, addr, rewards, distributeTo, startTime, numEpochsPaidOver)
		if err != nil {
			gaugeIds = append(gaugeIds, gaugeId)
		}
	}

	// setup lockups
	for i := int64(0); i < numLockups; i++ {
		addr := addrs[0]
		simCoins := app.BankKeeper.SpendableCoins(ctx, addr)
		duration := time.Second
		app.LockupKeeper.LockTokens(ctx, addr, simCoins, duration)
	}

	// begin distribution for all gauges
	for _, gaugeId := range gaugeIds {
		gauge, _ := app.IncentivesKeeper.GetGaugeByID(ctx, gaugeId)
		app.IncentivesKeeper.BeginDistribution(ctx, *gauge)
	}

	// distribute coins from gauges to lockup owners
	for i := int64(0); i < numDistrs; i++ {
		for _, gaugeId := range gaugeIds {
			gauge, _ := app.IncentivesKeeper.GetGaugeByID(ctx, gaugeId)
			app.IncentivesKeeper.Distribute(ctx, *gauge)
		}
	}
}

func BenchmarkDistributionLogicTiny(b *testing.B) {
	benchmarkDistributionLogic(1, 1, 1, 1, b)
}

func BenchmarkDistributionLogicSmall(b *testing.B) {
	benchmarkDistributionLogic(10, 10, 10, 100, b)
}

func BenchmarkDistributionLogicMedium(b *testing.B) {
	benchmarkDistributionLogic(50, 50, 50, 1000, b)
}

func BenchmarkDistributionLogicLarge(b *testing.B) {
	benchmarkDistributionLogic(100, 100, 100, 5000, b)
}

func BenchmarkDistributionLogicHuge(b *testing.B) {
	benchmarkDistributionLogic(1000, 1000, 1000, 30000, b)
}
