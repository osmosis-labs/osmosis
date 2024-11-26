package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// BenchmarkEpochHook benchmarks the epoch hook. In particular, it benchmarks the UpdatePools function.
func BenchmarkEpochHook(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	// Setup the test suite
	s := new(KeeperTestSuite)
	s.SetT(&testing.T{})
	s.SetupPoolsTest()

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err := s.App.ProtoRevKeeper.UpdatePools(s.Ctx)
		if err != nil {
			panic(fmt.Sprintf("error updating pools in protorev epoch hook benchmark: %s", err))
		}
		b.StopTimer()
	}
}

// TestEpochHook tests that the epoch hook is correctly setting the pool IDs for all base denoms. Base denoms are the denoms that will
// be used for cyclic arbitrage and must be stored in the keeper. The epoch hook is run after the pools are set and committed in keeper_test.go.
// All of the pools are initialized in the setup function in keeper_test.go and are available in the s.pools variable. In this test
// function, the pools are filtered to only include the pools that have at least one base denom as an asset. The pools are then filtered
// again to only include the pools that have the highest liquidity. The pools are then checked to see if the pool IDs are correctly set in the
// DenomPairToPool stores.
func (s *KeeperTestSuite) TestEpochHook() {
	s.SetupPoolsTest()
	// All of the pools initialized in the setup function are available in keeper_test.go
	// akash <-> types.OsmosisDenomination
	// juno <-> types.OsmosisDenomination
	// ethereum <-> types.OsmosisDenomination
	// bitcoin <-> types.OsmosisDenomination
	// canto <-> types.OsmosisDenomination
	// and so on...

	totalNumberExpected := 0
	expectedToSee := make(map[string]Pool)
	baseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	for _, pool := range s.pools {
		// Module currently limited to two asset pools
		// Instantiate asset and amounts for the pool
		if len(pool.PoolAssets) == 2 {
			pool.Asset1 = pool.PoolAssets[0].Token.Denom
			pool.Amount1 = pool.PoolAssets[0].Token.Amount
			pool.Asset2 = pool.PoolAssets[1].Token.Denom
			pool.Amount2 = pool.PoolAssets[1].Token.Amount
		}

		if contains(baseDenoms, pool.Asset1) || contains(baseDenoms, pool.Asset2) {
			// create a key that is a combination of asset1 and asset2 in alphabetical order
			key := fmt.Sprintf("%s-%s", pool.Asset1, pool.Asset2)
			if pool.Asset1 > pool.Asset2 {
				key = fmt.Sprintf("%s-%s", pool.Asset2, pool.Asset1)
			}

			if storedPool, ok := expectedToSee[key]; !ok {
				expectedToSee[key] = pool
				totalNumberExpected++
			} else {
				liquidity := pool.Amount1.Mul(pool.Amount2)
				if liquidity.GT(storedPool.Amount1.Mul(storedPool.Amount2)) {
					expectedToSee[key] = pool
				}
			}
		}
	}

	// Iterate and ensure that the keeper has the correct pool IDs for the base denoms
	totalActuallySeen := 0
	for key, pool := range expectedToSee {
		poolVisited := false

		// split the key and check if it contains a base denom
		denoms := strings.Split(key, "-")
		if contains(baseDenoms, denoms[0]) {
			poolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, denoms[0], denoms[1])
			s.Require().NoError(err)
			s.Require().Equal(pool.PoolId, poolId)
			poolVisited = true
		}

		if contains(baseDenoms, denoms[1]) {
			poolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, denoms[1], denoms[0])
			s.Require().NoError(err)
			s.Require().Equal(pool.PoolId, poolId)
			poolVisited = true
		}

		// In the case where the pool contains two base denoms, make sure that they both store the same pool ID
		if contains(baseDenoms, denoms[0]) && contains(baseDenoms, denoms[1]) {
			poolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, denoms[0], denoms[1])
			s.Require().NoError(err)
			s.Require().Equal(pool.PoolId, poolId)

			otherPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, denoms[1], denoms[0])
			s.Require().NoError(err)
			s.Require().Equal(pool.PoolId, otherPoolId)

			s.Require().Equal(poolId, otherPoolId)
		}

		if poolVisited {
			totalActuallySeen++
		}
	}

	s.Require().Equal(totalNumberExpected, totalActuallySeen)
}

// TestUpdateHighestLiquidityPools tests that UpdateHighestLiquidityPools correctly returns the pools with the highest liquidity
// given specific base denoms as input. The pools this test uses are created in the SetupTest function in keeper_test.go.
// This test uses pools with denoms prefixed with "epoch" which are only used for this test, so that pools created for
// other tests do not change the results of this test.
func (s *KeeperTestSuite) TestUpdateHighestLiquidityPools() {
	testCases := []struct {
		name                   string
		inputBaseDenomPools    map[string]map[string]keeper.LiquidityPoolStruct
		expectedBaseDenomPools map[string]map[string]keeper.LiquidityPoolStruct
	}{
		{
			// There are 2 pools with epochOne and uosmo as denoms, both in the GAMM module.
			// pool with ID 46 has a liquidity value of 1,000,000
			// pool with ID 47 has a liquidity value of 2,000,000
			// pool with ID 47 should be returned as the highest liquidity pool
			// We provide epochOne as the input base denom, to test the method chooses the correct pool
			// within the same pool module
			name: "Get highest liquidity pools for two GAMM pools",
			inputBaseDenomPools: map[string]map[string]keeper.LiquidityPoolStruct{
				"epochOne": {},
			},
			expectedBaseDenomPools: map[string]map[string]keeper.LiquidityPoolStruct{
				"epochOne": {
					"uosmo": {Liquidity: osmomath.NewInt(2000000), PoolId: 47},
				},
			},
		},
		{
			// There are 2 pools with epochTwo and uosmo as denoms,
			// One in the GAMM module and one in the Concentrated Liquidity module.
			// pool with ID 48 has a liquidity value of 1,000,000
			// pool with ID 50 has a liquidity value of 2,000,000
			// pool with ID 50 should be returned as the highest liquidity pool
			// We provide epochTwo as the input base denom, to test the method chooses the correct pool
			// across the GAMM and Concentrated Liquidity modules
			name: "Get highest liquidity pools for one GAMM pool and one Concentrated Liquidity pool",
			inputBaseDenomPools: map[string]map[string]keeper.LiquidityPoolStruct{
				"epochTwo": {},
			},
			expectedBaseDenomPools: map[string]map[string]keeper.LiquidityPoolStruct{
				"epochTwo": {
					"uosmo": {Liquidity: osmomath.Int(osmomath.NewUintFromString("999999000000000001000000000000000000")), PoolId: 50},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			// SetupTest creates all the pools used in the ProtoRev test suite,
			// including the pools with "epoch" prefixed denoms used in this test
			s.SetupPoolsTest()

			err := s.App.ProtoRevKeeper.UpdateHighestLiquidityPools(s.Ctx, tc.inputBaseDenomPools)
			s.Require().NoError(err)
			s.Require().Equal(tc.inputBaseDenomPools, tc.expectedBaseDenomPools)
		})
	}
}

func contains(baseDenoms []types.BaseDenom, denomToMatch string) bool {
	for _, baseDenom := range baseDenoms {
		if baseDenom.Denom == denomToMatch {
			return true
		}
	}
	return false
}

func (s *KeeperTestSuite) TestAfterEpochEnd() {
	tests := []struct {
		name       string
		arbProfits sdk.Coins
	}{
		{
			name:       "osmo denom only",
			arbProfits: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000000))),
		},
		{
			name: "osmo denom and another base denom",
			arbProfits: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000000)),
				sdk.NewCoin("juno", osmomath.NewInt(100000000))),
		},
		{
			name: "osmo denom, another base denom, and a non base denom",
			arbProfits: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000000)),
				sdk.NewCoin("juno", osmomath.NewInt(100000000)),
				sdk.NewCoin("eth", osmomath.NewInt(100000000))),
		},
		{
			name:       "no profits",
			arbProfits: sdk.Coins{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()

			// Set base denoms
			baseDenoms := []types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(1_000_000),
				},
				{
					Denom:    "juno",
					StepSize: osmomath.NewInt(1_000_000),
				},
			}
			s.App.ProtoRevKeeper.SetBaseDenoms(s.Ctx, baseDenoms)

			// Set protorev developer account
			devAccount := apptesting.CreateRandomAccounts(1)[0]
			s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, devAccount)

			err := s.App.BankKeeper.MintCoins(s.Ctx, types.ModuleName, tc.arbProfits)
			s.Require().NoError(err)

			communityPoolBalancePre := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName))

			// System under test
			err = s.App.ProtoRevKeeper.EpochHooks().AfterEpochEnd(s.Ctx, "day", 1)

			expectedDevProfit := sdk.Coins{}
			expectedOsmoBurn := sdk.Coins{}
			arbProfitsBaseDenoms := sdk.Coins{}
			arbProfitsNonBaseDenoms := sdk.Coins{}

			// Split the profits into base and non base denoms
			for _, coin := range tc.arbProfits {
				isBaseDenom := false
				for _, baseDenom := range baseDenoms {
					if coin.Denom == baseDenom.Denom {
						isBaseDenom = true
						break
					}
				}
				if isBaseDenom {
					arbProfitsBaseDenoms = append(arbProfitsBaseDenoms, coin)
				} else {
					arbProfitsNonBaseDenoms = append(arbProfitsNonBaseDenoms, coin)
				}
			}
			profitSplit := types.ProfitSplitPhase1
			for _, arbProfit := range arbProfitsBaseDenoms {
				devProfitAmount := arbProfit.Amount.MulRaw(profitSplit).QuoRaw(100)
				expectedDevProfit = append(expectedDevProfit, sdk.NewCoin(arbProfit.Denom, devProfitAmount))
			}

			// Get the developer account balance
			devAccountBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, devAccount)
			s.Require().Equal(expectedDevProfit, devAccountBalance)

			// Get the burn address balance
			burnAddressBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, types.DefaultNullAddress)
			if arbProfitsBaseDenoms.AmountOf(types.OsmosisDenomination).IsPositive() {
				expectedOsmoBurn = sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, arbProfitsBaseDenoms.AmountOf(types.OsmosisDenomination).Sub(expectedDevProfit.AmountOf(types.OsmosisDenomination))))
				s.Require().Equal(expectedOsmoBurn, burnAddressBalance)
			} else {
				s.Require().Equal(sdk.Coins{}, burnAddressBalance)
			}

			// Get the community pool balance
			communityPoolBalancePost := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName))
			actualCommunityPool := communityPoolBalancePost.Sub(communityPoolBalancePre...)
			expectedCommunityPool := arbProfitsBaseDenoms.Sub(expectedDevProfit...).Sub(expectedOsmoBurn...)
			s.Require().Equal(expectedCommunityPool, actualCommunityPool)

			// The protorev module account should only contain the non base denoms if there are any
			protorevModuleAccount := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
			s.Require().Equal(arbProfitsNonBaseDenoms, protorevModuleAccount)
		})
	}
}
