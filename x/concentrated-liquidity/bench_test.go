package concentrated_liquidity_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type BenchTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s BenchTestSuite) createPosition(accountIndex int, poolId uint64, coin0, coin1 sdk.Coin, lowerTick, upperTick int64) {
	tokensDesired := sdk.NewCoins(coin0, coin1)

	_, _, _, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[accountIndex], tokensDesired, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
	if err != nil {
		// This can happen for ticks that map to the very small prices
		// e.g 2 * 10^(-18) ends up mapping to the same sqrt price
		fmt.Println("error creating position", err)
	}
}

func BenchmarkSwapExactAmountIn(b *testing.B) {
	// Notice we stop the timer to skip setup code.
	b.StopTimer()

	// We cannot use s.Require().NoError() becuase the suite context
	// is defined on the testing.T and not testing.B
	noError := func(err error) {
		require.NoError(b, err)
	}

	const (
		numberOfPositions = 10000

		// max amount of each token deposited per position.
		maxAmountDeposited = int64(1_000_000_000_000)

		// amount swapped in.
		amountIn = "9999999999999999999"

		// flag controlling whether to create additional numberOfPositions full
		// range positions for deeper liquidity.
		shouldCreateFullRangePositions = true

		// flag controlling whether to create positions concentrated around current tick to mimic
		// realistic scenario and deeper liquidity.
		// if true,
		// creates numberOfPositions positions within 10 ticks of the current tick.
		// creates numberOfPositions positions within 100 ticks of the current tick.
		shouldConcentrate = true

		// tickSpacing is the spacing between ticks.
		tickSpacing = 1
	)

	var (
		// denoms of the pool.
		denom0 = DefaultCoin0.Denom
		denom1 = DefaultCoin1.Denom

		// denom of the token to swap in.
		denomIn = denom0

		numberOfPositionsInt = sdk.NewInt(numberOfPositions)
		maxAmountOfEachToken = sdk.NewInt(maxAmountDeposited).Mul(numberOfPositionsInt)

		// Seed controlling determinism of the randomized positions.
		seed = int64(1)
	)
	rand.Seed(seed)

	for i := 0; i < b.N; i++ {
		s := BenchTestSuite{}
		s.Setup()

		// Fund all accounts with max amounts they would need to consume.
		for _, acc := range s.TestAccs {
			simapp.FundAccount(s.App.BankKeeper, s.Ctx, acc, sdk.NewCoins(sdk.NewCoin(denom0, maxAmountOfEachToken), sdk.NewCoin(denom1, maxAmountOfEachToken), sdk.NewCoin("uosmo", maxAmountOfEachToken)))
		}

		// Create a pool
		poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(s.TestAccs[0], denom0, denom1, tickSpacing, sdk.MustNewDecFromStr("0.001")))
		noError(err)

		clKeeper := s.App.ConcentratedLiquidityKeeper

		// Create first position to set a price of 1 and tick of zero.
		tokenDesired0 := sdk.NewCoin(denom0, sdk.NewInt(100))
		tokenDesired1 := sdk.NewCoin(denom1, sdk.NewInt(100))
		tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)
		_, _, _, _, _, _, _, err = clKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], tokensDesired, sdk.ZeroInt(), sdk.ZeroInt(), types.MinTick, types.MaxTick)

		pool, err := clKeeper.GetPoolById(s.Ctx, poolId)
		noError(err)

		// Zero by default, can configure by setting a specific position.
		currentTick := pool.GetCurrentTick()

		// Setup numberOfPositions positions at random ranges
		for i := 0; i < numberOfPositions; i++ {

			var (
				lowerTick int64
				upperTick int64
			)

			if denomIn == denom0 {
				// Decreasing price so want to be below current tick

				// minTick <= lowerTick <= currentTick
				lowerTick = rand.Int63n(currentTick-types.MinTick+1) + types.MinTick
				// lowerTick <= upperTick <= currentTick
				upperTick = currentTick - rand.Int63n(int64(math.Abs(float64(currentTick-lowerTick))))
			} else {
				// Increasing price so want to be above current tick

				// currentTick <= lowerTick <= maxTick
				lowerTick := rand.Int63n(types.MaxTick-currentTick+1) + currentTick
				// lowerTick <= upperTick <= maxTick
				upperTick = types.MaxTick - rand.Int63n(int64(math.Abs(float64(types.MaxTick-lowerTick))))
			}
			// Normalize lowerTick to be a multiple of tickSpacing
			lowerTick = lowerTick + (tickSpacing - lowerTick%tickSpacing)
			// Normalize upperTick to be a multiple of tickSpacing
			upperTick = upperTick - upperTick%tickSpacing

			tokenDesired0 := sdk.NewCoin(denom0, sdk.NewInt(rand.Int63n(maxAmountDeposited)))
			tokenDesired1 := sdk.NewCoin(denom1, sdk.NewInt(rand.Int63n(maxAmountDeposited)))

			accountIndex := rand.Intn(len(s.TestAccs))

			s.createPosition(accountIndex, poolId, tokenDesired0, tokenDesired1, lowerTick, upperTick)
		}

		// Setup numberOfPositions full range positions for deeper liquidity.
		if shouldCreateFullRangePositions {
			for i := 0; i < numberOfPositions; i++ {
				lowerTick := types.MinTick
				upperTick := types.MaxTick

				maxAmountDepositedFullRange := sdk.NewInt(maxAmountDeposited).MulRaw(5)
				tokenDesired0 := sdk.NewCoin(denom0, maxAmountDepositedFullRange)
				tokenDesired1 := sdk.NewCoin(denom1, maxAmountDepositedFullRange)
				tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)

				accountIndex := rand.Intn(len(s.TestAccs))

				account := s.TestAccs[accountIndex]

				simapp.FundAccount(s.App.BankKeeper, s.Ctx, account, tokensDesired)

				s.createPosition(accountIndex, poolId, tokenDesired0, tokenDesired1, lowerTick, upperTick)
			}
		}

		// Setup numberOfPositions * 2 positions at random ranges around the current tick for deeper
		// liquidity.
		if shouldConcentrate {
			// Within 10 ticks of the current
			if tickSpacing <= 10 {
				for i := 0; i < numberOfPositions; i++ {
					lowerTick := currentTick - 10
					upperTick := currentTick + 10

					tokenDesired0 := sdk.NewCoin(denom0, sdk.NewInt(maxAmountDeposited).MulRaw(5))
					tokenDesired1 := sdk.NewCoin(denom1, sdk.NewInt(maxAmountDeposited).MulRaw(5))
					tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)

					accountIndex := rand.Intn(len(s.TestAccs))

					account := s.TestAccs[accountIndex]

					simapp.FundAccount(s.App.BankKeeper, s.Ctx, account, tokensDesired)

					s.createPosition(accountIndex, poolId, tokenDesired0, tokenDesired1, lowerTick, upperTick)
				}
			}

			// Within 100 ticks of the current
			for i := 0; i < numberOfPositions; i++ {
				lowerTick := currentTick - 100
				upperTick := currentTick + 100
				// Normalize lowerTick to be a multiple of tickSpacing
				lowerTick = lowerTick + (tickSpacing - lowerTick%tickSpacing)
				// Normalize upperTick to be a multiple of tickSpacing
				upperTick = upperTick - upperTick%tickSpacing

				tokenDesired0 := sdk.NewCoin(denom0, sdk.NewInt(maxAmountDeposited).MulRaw(5))
				tokenDesired1 := sdk.NewCoin(denom1, sdk.NewInt(maxAmountDeposited).MulRaw(5))
				tokensDesired := sdk.NewCoins(tokenDesired0, tokenDesired1)

				accountIndex := rand.Intn(len(s.TestAccs))

				account := s.TestAccs[accountIndex]

				simapp.FundAccount(s.App.BankKeeper, s.Ctx, account, tokensDesired)

				s.createPosition(accountIndex, poolId, tokenDesired0, tokenDesired1, lowerTick, upperTick)
			}
		}

		swapAmountIn := sdk.MustNewDecFromStr(amountIn).TruncateInt()
		largeSwapInCoin := sdk.NewCoin(denomIn, swapAmountIn)

		liquidityNet, err := clKeeper.GetTickLiquidityNetInDirection(s.Ctx, pool.GetId(), largeSwapInCoin.Denom, sdk.NewInt(currentTick), sdk.Int{})
		noError(err)

		fmt.Println("num_ticks_traversed", len(liquidityNet))
		fmt.Println("current_tick", currentTick)

		// Commit the block so that position updates are propagated to IAVL.
		s.Commit()

		// Fund swap amount.
		simapp.FundAccount(s.App.BankKeeper, s.Ctx, s.TestAccs[0], sdk.NewCoins(largeSwapInCoin))

		// Notice that we start the timer as this is the system under test
		b.StartTimer()

		// System under test
		_, err = clKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, largeSwapInCoin, denom1, sdk.NewInt(1), pool.GetSpreadFactor(s.Ctx))
		noError(err)

		// Notice that we stop the timer again in case there are multiple iterations.
		b.StopTimer()
	}
}
