package concentrated_liquidity_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v14/app"
	"github.com/osmosis-labs/osmosis/v14/app/apptesting"
	sdkrand "github.com/osmosis-labs/osmosis/v14/simulation/simtypes/random"
	clmodel "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
)

// benchmarkCrossTick performs bench test on the amount of computational resources it takes to cross ticks N times.
// We achieve the goal of doing so initializing N amount of positions each with 1 * denom0, 1 * denom1.
// This way performing a swap of N tokens would cause us to cross N amount of ticks, being able to bench `crossTick`.
func benchmarkCrossTick(numCrossTick int, b *testing.B) {
	b.StopTimer()

	blockStartTime := time.Now().UTC()
	app, cleanupFn := app.SetupTestingAppWithLevelDb(false)
	defer cleanupFn()
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: blockStartTime})
	r := rand.New(rand.NewSource(10))
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// first create pool
	_ = simapp.FundAccount(app.BankKeeper, ctx, addr, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000000))))
	poolID, _ := app.PoolManagerKeeper.CreatePool(ctx, clmodel.NewMsgCreateConcentratedPool(
		addr,
		apptesting.ETH,
		apptesting.USDC,
		apptesting.DefaultTickSpacing,
		apptesting.DefaultExponentAtPriceOne,
		sdk.ZeroDec(),
	))

	// first we need to initialize n amount of ticks via adding position
	// the goal is to initialize ticks with the minimum amount of liquidity(1 coins each) so that we can easily test crossing ticks
	cl, err := app.ConcentratedLiquidityKeeper.GetPoolById(ctx, poolID)
	currentTick := cl.GetCurrentTick()
	startingTick := sdkrand.RandIntBetween(r, int(currentTick.Int64()), 5000)
	spaceBetweenTicks := 100

	// iterate over the amount of ticks we want to create
	for i := 0; i < numCrossTick; i++ {
		// fund 1 eth and 1 usdc
		coins := sdk.NewCoins(sdk.NewCoin(apptesting.ETH, sdk.OneInt()), sdk.NewCoin(apptesting.USDC, sdk.OneInt()))
		simapp.FundAccount(app.BankKeeper, ctx, addr, coins)

		_, _, _, err := app.ConcentratedLiquidityKeeper.CreatePosition(
			ctx,
			poolID,
			addr,
			sdk.OneInt(),
			sdk.OneInt(),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			int64(startingTick+i),
			int64(startingTick+i+spaceBetweenTicks),
		)
		if err != nil {
			fmt.Println("inside error of creating position")
			fmt.Println(err.Error())
			b.FailNow()
		}
	}

	// test swapping and start bench testing
	b.StartTimer()
	tokenIn := sdk.NewCoin("usdc", sdk.NewInt(int64(numCrossTick)))
	simapp.FundAccount(app.BankKeeper, ctx, addr, sdk.NewCoins(tokenIn))
	_, _, currentTick, _, _, err = app.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
		ctx,
		tokenIn,
		"eth",
		sdk.ZeroDec(),
		sdk.ZeroDec(),
		poolID,
	)
	if err != nil {
		fmt.Println(err.Error())
		b.FailNow()
	}

}

func BenchmarkCrossTickTiny(b *testing.B) {
	benchmarkCrossTick(10, b)
}

func BenchmarkCrossTickMedium(b *testing.B) {
	benchmarkCrossTick(100, b)
}

func BenchmarkCrossTickLarge(b *testing.B) {
	benchmarkCrossTick(1000, b)
}
