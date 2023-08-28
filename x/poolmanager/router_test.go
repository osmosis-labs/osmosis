package poolmanager_test

import (
	"errors"
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/tests/mocks"
	cl "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	cwpool "github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool"
	cwmodel "github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/model"
	gamm "github.com/osmosis-labs/osmosis/v19/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

type poolSetup struct {
	poolType         types.PoolType
	initialLiquidity sdk.Coins
}

const (
	foo   = "foo"
	bar   = "bar"
	baz   = "baz"
	uosmo = "uosmo"
)

var (
	defaultInitPoolAmount   = sdk.NewInt(1000000000000)
	defaultPoolSpreadFactor = sdk.NewDecWithPrec(1, 3) // 0.1% pool spread factor default
	defaultSwapAmount       = sdk.NewInt(1000000)
	gammKeeperType          = reflect.TypeOf(&gamm.Keeper{})
	concentratedKeeperType  = reflect.TypeOf(&cl.Keeper{})
	cosmwasmKeeperType      = reflect.TypeOf(&cwpool.Keeper{})

	defaultPoolInitAmount     = sdk.NewInt(10_000_000_000)
	twentyFiveBaseUnitsAmount = sdk.NewInt(25_000_000)

	fooCoin   = sdk.NewCoin(foo, defaultPoolInitAmount)
	barCoin   = sdk.NewCoin(bar, defaultPoolInitAmount)
	bazCoin   = sdk.NewCoin(baz, defaultPoolInitAmount)
	uosmoCoin = sdk.NewCoin(uosmo, defaultPoolInitAmount)

	// Note: These are initialized in such a way as it makes
	// it easier to reason about the test cases.
	fooBarCoins    = sdk.NewCoins(fooCoin, barCoin)
	fooBarPoolId   = uint64(1)
	fooBazCoins    = sdk.NewCoins(fooCoin, bazCoin)
	fooBazPoolId   = fooBarPoolId + 1
	fooUosmoCoins  = sdk.NewCoins(fooCoin, uosmoCoin)
	fooUosmoPoolId = fooBazPoolId + 1
	barBazCoins    = sdk.NewCoins(barCoin, bazCoin)
	barBazPoolId   = fooUosmoPoolId + 1
	barUosmoCoins  = sdk.NewCoins(barCoin, uosmoCoin)
	barUosmoPoolId = barBazPoolId + 1
	bazUosmoCoins  = sdk.NewCoins(bazCoin, uosmoCoin)
	bazUosmoPoolId = barUosmoPoolId + 1

	defaultValidPools = []poolSetup{
		{
			poolType:         types.Balancer,
			initialLiquidity: fooBarCoins,
		},
		{
			poolType:         types.Concentrated,
			initialLiquidity: fooBazCoins,
		},
		{
			poolType:         types.Balancer,
			initialLiquidity: fooUosmoCoins,
		},
		{
			poolType:         types.Concentrated,
			initialLiquidity: barBazCoins,
		},
		{
			poolType:         types.Balancer,
			initialLiquidity: barUosmoCoins,
		},
		{
			poolType:         types.Concentrated,
			initialLiquidity: bazUosmoCoins,
		},
	}
)

// TestGetPoolModule tests that the correct pool module is returned for a given pool id.
// Additionally, validates that the expected errors are produced when expected.
func (s *KeeperTestSuite) TestGetPoolModule() {
	tests := map[string]struct {
		poolId            uint64
		preCreatePoolType types.PoolType
		routesOverwrite   map[types.PoolType]types.PoolModuleI

		expectedModule reflect.Type
		expectError    error
	}{
		"valid balancer pool": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			expectedModule:    gammKeeperType,
		},
		"valid stableswap pool": {
			preCreatePoolType: types.Stableswap,
			poolId:            1,
			expectedModule:    gammKeeperType,
		},
		"valid concentrated liquidity pool": {
			preCreatePoolType: types.Concentrated,
			poolId:            1,
			expectedModule:    concentratedKeeperType,
		},
		"valid cosmwasm pool": {
			preCreatePoolType: types.CosmWasm,
			poolId:            1,
			expectedModule:    cosmwasmKeeperType,
		},
		"non-existent pool": {
			preCreatePoolType: types.Balancer,
			poolId:            2,
			expectedModule:    gammKeeperType,

			expectError: types.FailedToFindRouteError{PoolId: 2},
		},
		"undefined route": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			routesOverwrite: map[types.PoolType]types.PoolModuleI{
				types.Stableswap: &gamm.Keeper{}, // undefined for balancer.
			},

			expectError: types.UndefinedRouteError{PoolId: 1, PoolType: types.Balancer},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.CreatePoolFromType(tc.preCreatePoolType)

			if len(tc.routesOverwrite) > 0 {
				poolmanagerKeeper.SetPoolRoutesUnsafe(tc.routesOverwrite)
			}

			swapModule, err := poolmanagerKeeper.GetPoolModule(s.Ctx, tc.poolId)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				s.Require().Nil(swapModule)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(swapModule)

			s.Require().Equal(tc.expectedModule, reflect.TypeOf(swapModule))
		})
	}
}

func (s *KeeperTestSuite) TestRouteGetPoolDenoms() {
	tests := map[string]struct {
		poolId            uint64
		preCreatePoolType types.PoolType
		routesOverwrite   map[types.PoolType]types.PoolModuleI

		expectedDenoms []string
		expectError    error
	}{
		"valid balancer pool": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			expectedDenoms:    []string{"bar", "baz", "foo", "uosmo"},
		},
		"valid stableswap pool": {
			preCreatePoolType: types.Stableswap,
			poolId:            1,
			expectedDenoms:    []string{"bar", "baz", "foo"},
		},
		"valid concentrated liquidity pool": {
			preCreatePoolType: types.Concentrated,
			poolId:            1,
			expectedDenoms:    []string{"eth", "usdc"},
		},
		"valid cosmwasm pool": {
			preCreatePoolType: types.CosmWasm,
			poolId:            1,
			expectedDenoms:    []string{apptesting.DefaultTransmuterDenomA, apptesting.DefaultTransmuterDenomB},
		},

		"non-existent pool": {
			preCreatePoolType: types.Balancer,
			poolId:            2,

			expectError: types.FailedToFindRouteError{PoolId: 2},
		},
		"undefined route": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			routesOverwrite: map[types.PoolType]types.PoolModuleI{
				types.Stableswap: &gamm.Keeper{}, // undefined for balancer.
			},

			expectError: types.UndefinedRouteError{PoolId: 1, PoolType: types.Balancer},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.CreatePoolFromType(tc.preCreatePoolType)

			if len(tc.routesOverwrite) > 0 {
				poolmanagerKeeper.SetPoolRoutesUnsafe(tc.routesOverwrite)
			}

			denoms, err := poolmanagerKeeper.RouteGetPoolDenoms(s.Ctx, tc.poolId)
			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedDenoms, denoms)
		})
	}
}

func (s *KeeperTestSuite) TestRouteCalculateSpotPrice() {
	tests := map[string]struct {
		poolId               uint64
		preCreatePoolType    types.PoolType
		quoteAssetDenom      string
		baseAssetDenom       string
		setPositionForCLPool bool

		routesOverwrite   map[types.PoolType]types.PoolModuleI
		expectedSpotPrice sdk.Dec

		expectError error
	}{
		"valid balancer pool": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			quoteAssetDenom:   "bar",
			baseAssetDenom:    "baz",
			expectedSpotPrice: sdk.MustNewDecFromStr("1.5"),
		},
		"valid stableswap pool": {
			preCreatePoolType: types.Stableswap,
			poolId:            1,
			quoteAssetDenom:   "bar",
			baseAssetDenom:    "baz",
			expectedSpotPrice: sdk.MustNewDecFromStr("0.99999998"),
		},
		"valid concentrated liquidity pool with position": {
			preCreatePoolType:    types.Concentrated,
			poolId:               1,
			quoteAssetDenom:      "usdc",
			baseAssetDenom:       "eth",
			setPositionForCLPool: true,
			// We generate this value using the scripts in x/concentrated-liquidity/python
			// Exact output: 5000.000000000000000129480272834995458481
			// SDK Bankers rounded output: 5000.000000000000000129
			expectedSpotPrice: sdk.MustNewDecFromStr("5000.000000000000000129"),
		},
		"valid concentrated liquidity pool without position": {
			preCreatePoolType: types.Concentrated,
			poolId:            1,
			quoteAssetDenom:   "usdc",
			baseAssetDenom:    "eth",

			expectError: cltypes.NoSpotPriceWhenNoLiquidityError{
				PoolId: 1,
			},
		},
		"valid cosmwasm pool with LP": {
			preCreatePoolType: types.CosmWasm,
			poolId:            1,
			quoteAssetDenom:   apptesting.DefaultTransmuterDenomA,
			baseAssetDenom:    apptesting.DefaultTransmuterDenomB,
			// For transmuter, the spot price is always 1. (hard-coded even if no liquidity)
			expectedSpotPrice: sdk.OneDec(),
		},
		"non-existent pool": {
			preCreatePoolType: types.Balancer,
			poolId:            2,

			expectError: types.FailedToFindRouteError{PoolId: 2},
		},
		"undefined route": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			routesOverwrite: map[types.PoolType]types.PoolModuleI{
				types.Stableswap: &gamm.Keeper{}, // undefined for balancer.
			},

			expectError: types.UndefinedRouteError{PoolId: 1, PoolType: types.Balancer},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.CreatePoolFromType(tc.preCreatePoolType)

			// we manually set position for CL to set spot price to correct value
			if tc.setPositionForCLPool {
				coins := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000)), sdk.NewCoin("usdc", sdk.NewInt(5000000000)))
				s.FundAcc(s.TestAccs[0], coins)

				clMsgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)
				_, err := clMsgServer.CreatePosition(sdk.WrapSDKContext(s.Ctx), &cltypes.MsgCreatePosition{
					PoolId:          1,
					Sender:          s.TestAccs[0].String(),
					LowerTick:       int64(30545000),
					UpperTick:       int64(31500000),
					TokensProvided:  coins,
					TokenMinAmount0: sdk.ZeroInt(),
					TokenMinAmount1: sdk.ZeroInt(),
				})
				s.Require().NoError(err)
			}

			if len(tc.routesOverwrite) > 0 {
				poolmanagerKeeper.SetPoolRoutesUnsafe(tc.routesOverwrite)
			}

			spotPrice, err := poolmanagerKeeper.RouteCalculateSpotPrice(s.Ctx, tc.poolId, tc.quoteAssetDenom, tc.baseAssetDenom)
			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectError.Error())
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedSpotPrice, spotPrice)
		})
	}
}

// TestMultihopSwapExactAmountIn tests that the swaps are routed correctly.
// That is:
// - to the correct module (concentrated-liquidity or gamm)
// - over the right routes (hops)
// - fee reduction is applied correctly
func (s *KeeperTestSuite) TestMultihopSwapExactAmountIn() {
	tests := []struct {
		name                    string
		poolCoins               []sdk.Coins
		poolSpreadFactor        []sdk.Dec
		poolType                []types.PoolType
		routes                  []types.SwapAmountInRoute
		incentivizedGauges      []uint64
		tokenIn                 sdk.Coin
		tokenOutMinAmount       sdk.Int
		spreadFactor            sdk.Dec
		expectError             bool
		expectReducedFeeApplied bool
	}{
		{
			name:             "One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins:        []sdk.Coins{sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount))},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor},
			poolType:         []types.PoolType{types.Balancer},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
			},
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
		},
		{
			name: "Two routes: Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges: []uint64{},
			tokenIn:            sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:  sdk.NewInt(1),
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin("foo", sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: true,
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, (pool 1) 1 percent fee, (pool 2) 10 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, sdk.NewDecWithPrec(1, 1)},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)),   // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
				{
					PoolId:        3,
					TokenOutDenom: bar,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 3.                                                                                      // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
				{
					PoolId:        3,
					TokenOutDenom: bar,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: false,
		},
		{
			name: "[Concentrated] One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, apptesting.DefaultCoinAmount), sdk.NewCoin(bar, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.Concentrated},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
			},
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
		},
		{
			name: "[Concentrated[ Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, apptesting.DefaultCoinAmount), sdk.NewCoin(uosmo, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(baz, apptesting.DefaultCoinAmount), sdk.NewCoin(uosmo, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(bar, apptesting.DefaultCoinAmount), sdk.NewCoin(baz, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.Concentrated, types.Concentrated, types.Concentrated},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
				{
					PoolId:        3,
					TokenOutDenom: bar,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:            sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:  sdk.NewInt(1),
		},
		{
			name: "[Cosmwasm] One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, apptesting.DefaultCoinAmount), sdk.NewCoin(bar, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm},
			poolSpreadFactor: []sdk.Dec{sdk.OneDec()},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
			},
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
		},
		{
			name: "[Cosmwasm -> Concentrated] One route: Swap - [foo -> bar] -> [bar -> baz], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, apptesting.DefaultCoinAmount), sdk.NewCoin(bar, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm, types.Concentrated},
			poolSpreadFactor: []sdk.Dec{sdk.OneDec(), defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
		},
		//TODO:
		//change values in and out to be different with each swap module type
		//tests for stable-swap pools
		//edge cases:
		//  * invalid route length
		//  * pool does not exist
		//  * swap errors
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			for i := range tc.poolType {
				s.FundAcc(s.TestAccs[0], tc.poolCoins[i])
				s.CreatePoolFromTypeWithCoinsAndSpreadFactor(tc.poolType[i], tc.poolCoins[i], tc.poolSpreadFactor[i])
			}

			// if test specifies incentivized gauges, set them here
			if len(tc.incentivizedGauges) > 0 {
				s.makeGaugesIncentivized(tc.incentivizedGauges)
			}

			if tc.expectError {
				// execute the swap
				_, err := poolmanagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], tc.routes, tc.tokenIn, tc.tokenOutMinAmount)
				s.Require().Error(err)
			} else {
				// calculate the swap as separate swaps with either the reduced swap fee or normal fee
				expectedMultihopTokenOutAmount := s.calcOutGivenInAmountAsSeparatePoolSwaps(tc.expectReducedFeeApplied, tc.routes, tc.tokenIn)

				// execute the swap
				multihopTokenOutAmount, err := poolmanagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], tc.routes, tc.tokenIn, tc.tokenOutMinAmount)
				// compare the expected tokenOut to the actual tokenOut
				s.Require().NoError(err)
				s.Require().Equal(expectedMultihopTokenOutAmount.Amount.String(), multihopTokenOutAmount.String())
			}
		})
	}
}

// TestMultihopSwapExactAmountOut tests that the swaps are routed correctly.
// That is:
// - to the correct module (concentrated-liquidity or gamm)
// - over the right routes (hops)
// - fee reduction is applied correctly
func (s *KeeperTestSuite) TestMultihopSwapExactAmountOut() {
	tests := []struct {
		name                    string
		poolCoins               []sdk.Coins
		poolSpreadFactor        []sdk.Dec
		poolType                []types.PoolType
		routes                  []types.SwapAmountOutRoute
		incentivizedGauges      []uint64
		tokenOut                sdk.Coin
		tokenInMaxAmount        sdk.Int
		spreadFactor            sdk.Dec
		expectError             bool
		expectReducedFeeApplied bool
	}{
		{
			name:             "One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins:        []sdk.Coins{sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount))},
			poolType:         []types.PoolType{types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: bar,
				},
			},
			tokenInMaxAmount: sdk.NewInt(90000000),
			tokenOut:         sdk.NewCoin(foo, defaultSwapAmount),
		},
		{
			name: "Two routes: Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: bar,
				},
			},
			incentivizedGauges: []uint64{},

			tokenInMaxAmount: sdk.NewInt(90000000),
			tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:        sdk.NewInt(90000000),
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			expectReducedFeeApplied: true,
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, (pool 1) 1 percent fee, (pool 2) 10 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, sdk.NewDecWithPrec(1, 1)},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:        sdk.NewInt(90000000),
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)),   // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
				{
					PoolId:       3,
					TokenInDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:        sdk.NewInt(90000000),
			tokenOut:                sdk.NewCoin(bar, sdk.NewInt(100000)),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: bar,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			tokenInMaxAmount:        sdk.NewInt(90000000),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			tokenInMaxAmount:        sdk.NewInt(90000000),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 3.                                                                                    // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
				{
					PoolId:       3,
					TokenInDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			tokenOut:                sdk.NewCoin(bar, sdk.NewInt(100000)),
			tokenInMaxAmount:        sdk.NewInt(90000000),
			expectReducedFeeApplied: false,
		},
		{
			name: "[Cosmwasm] One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, apptesting.DefaultCoinAmount), sdk.NewCoin(bar, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm},
			poolSpreadFactor: []sdk.Dec{sdk.OneDec()},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
			},
			tokenOut:         sdk.NewCoin(bar, sdk.NewInt(100000)),
			tokenInMaxAmount: sdk.NewInt(90000000),
		},
		{
			name: "[Cosmwasm -> Concentrated] One route: Swap - [foo -> bar] -> [bar -> baz], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, apptesting.DefaultCoinAmount), sdk.NewCoin(bar, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm, types.Concentrated},
			poolSpreadFactor: []sdk.Dec{sdk.OneDec(), defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: bar,
				},
			},
			tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
			tokenInMaxAmount: sdk.NewInt(90000000),
		},
		// TODO:
		// tests for concentrated liquidity
		// tests for stable-swap pools
		// change values in and out to be different with each swap module type
		// edge cases:
		//   * invalid route length
		//   * pool does not exist
		//   * swap errors
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.Require().NotEmpty(tc.poolType)
			for i := range tc.poolType {
				s.FundAcc(s.TestAccs[0], tc.poolCoins[i])
				s.CreatePoolFromTypeWithCoinsAndSpreadFactor(tc.poolType[i], tc.poolCoins[i], tc.poolSpreadFactor[i])
			}

			// if test specifies incentivized gauges, set them here
			if len(tc.incentivizedGauges) > 0 {
				s.makeGaugesIncentivized(tc.incentivizedGauges)
			}

			if tc.expectError {
				// execute the swap
				_, err := poolmanagerKeeper.RouteExactAmountOut(s.Ctx, s.TestAccs[0], tc.routes, tc.tokenInMaxAmount, tc.tokenOut)
				s.Require().Error(err)
			} else {
				// calculate the swap as separate swaps with either the reduced swap fee or normal fee
				expectedMultihopTokenInAmount := s.calcInGivenOutAmountAsSeparateSwaps(tc.expectReducedFeeApplied, tc.routes, tc.tokenOut)
				// execute the swap
				multihopTokenInAmount, err := poolmanagerKeeper.RouteExactAmountOut(s.Ctx, s.TestAccs[0], tc.routes, tc.tokenInMaxAmount, tc.tokenOut)
				// compare the expected tokenOut to the actual tokenOut
				s.Require().NoError(err)
				s.Require().Equal(expectedMultihopTokenInAmount.Amount.String(), multihopTokenInAmount.String())
			}
		})
	}
}

// TestEstimateMultihopSwapExactAmountIn tests that the estimation done via `EstimateSwapExactAmountIn`
// results in the same amount of token out as the actual swap.
func (s *KeeperTestSuite) TestEstimateMultihopSwapExactAmountIn() {
	type param struct {
		routes            []types.SwapAmountInRoute
		estimateRoutes    []types.SwapAmountInRoute
		tokenIn           sdk.Coin
		tokenOutMinAmount sdk.Int
	}

	tests := []struct {
		name              string
		param             param
		expectPass        bool
		reducedFeeApplied bool
		poolType          types.PoolType
	}{
		{
			name: "Proper swap - foo -> bar(pool 1) - bar(pool 2) -> baz",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: bar,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: bar,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: uosmo,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: uosmo,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			reducedFeeApplied: true,
			expectPass:        true,
		},
		{
			name: "Proper swap (stableswap pool) - foo -> bar(pool 1) - bar(pool 2) -> baz",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: bar,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: bar,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: true,
			poolType:   types.Stableswap,
		},
		{
			name: "Asserts panic catching in MultihopEstimateOutGivenExactAmountIn works: tokenOut more than pool reserves",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: bar,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: bar,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(9000000000000000000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: false,
			poolType:   types.Stableswap,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		s.SetupTest()

		s.Run(test.name, func() {
			poolmanagerKeeper := s.App.PoolManagerKeeper

			firstEstimatePoolId, secondEstimatePoolId := s.setupPools(test.poolType, defaultPoolSpreadFactor)

			firstEstimatePool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, firstEstimatePoolId)
			s.Require().NoError(err)
			secondEstimatePool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, secondEstimatePoolId)
			s.Require().NoError(err)

			// calculate token out amount using `MultihopSwapExactAmountIn`
			multihopTokenOutAmount, errMultihop := poolmanagerKeeper.RouteExactAmountIn(
				s.Ctx,
				s.TestAccs[0],
				test.param.routes,
				test.param.tokenIn,
				test.param.tokenOutMinAmount)

			// calculate token out amount using `EstimateMultihopSwapExactAmountIn`
			estimateMultihopTokenOutAmount, errEstimate := poolmanagerKeeper.MultihopEstimateOutGivenExactAmountIn(
				s.Ctx,
				test.param.estimateRoutes,
				test.param.tokenIn)

			if test.expectPass {
				s.Require().NoError(errMultihop, "test: %v", test.name)
				s.Require().NoError(errEstimate, "test: %v", test.name)
				s.Require().Equal(multihopTokenOutAmount, estimateMultihopTokenOutAmount)
			} else {
				s.Require().Error(errMultihop, "test: %v", test.name)
				s.Require().Error(errEstimate, "test: %v", test.name)
			}
			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, firstEstimatePoolId)
			s.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, secondEstimatePoolId)
			s.Require().NoError(err)

			s.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			s.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)
		})
	}
}

// TestEstimateMultihopSwapExactAmountOut tests that the estimation done via `EstimateSwapExactAmountOut`
// results in the same amount of token in as the actual swap.
func (s *KeeperTestSuite) TestEstimateMultihopSwapExactAmountOut() {
	type param struct {
		routes           []types.SwapAmountOutRoute
		estimateRoutes   []types.SwapAmountOutRoute
		tokenInMaxAmount sdk.Int
		tokenOut         sdk.Coin
	}

	tests := []struct {
		name              string
		param             param
		expectPass        bool
		reducedFeeApplied bool
		poolType          types.PoolType
	}{
		{
			name: "Proper swap: foo -> bar (pool 1), bar -> baz (pool 2)",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: bar,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: bar,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: uosmo,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: uosmo,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
			},
			expectPass:        true,
			reducedFeeApplied: true,
		},
		{
			name: "Proper swap, stableswap pool: foo -> bar (pool 1), bar -> baz (pool 2)",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: bar,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: bar,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
			},
			expectPass: true,
			poolType:   types.Stableswap,
		},
		{
			name: "Asserts panic catching in MultihopEstimateInGivenExactAmountOut works: tokenOut more than pool reserves",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: bar,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: bar,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(9000000000000000000)),
			},
			expectPass: false,
			poolType:   types.Stableswap,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		s.SetupTest()

		s.Run(test.name, func() {
			poolmanagerKeeper := s.App.PoolManagerKeeper

			firstEstimatePoolId, secondEstimatePoolId := s.setupPools(test.poolType, defaultPoolSpreadFactor)

			firstEstimatePool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, firstEstimatePoolId)
			s.Require().NoError(err)
			secondEstimatePool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, secondEstimatePoolId)
			s.Require().NoError(err)

			multihopTokenInAmount, errMultihop := poolmanagerKeeper.RouteExactAmountOut(
				s.Ctx,
				s.TestAccs[0],
				test.param.routes,
				test.param.tokenInMaxAmount,
				test.param.tokenOut)

			estimateMultihopTokenInAmount, errEstimate := poolmanagerKeeper.MultihopEstimateInGivenExactAmountOut(
				s.Ctx,
				test.param.estimateRoutes,
				test.param.tokenOut)

			if test.expectPass {
				s.Require().NoError(errMultihop, "test: %v", test.name)
				s.Require().NoError(errEstimate, "test: %v", test.name)
				s.Require().Equal(estimateMultihopTokenInAmount.String(), multihopTokenInAmount.String())
			} else {
				s.Require().Error(errMultihop, "test: %v", test.name)
				s.Require().Error(errEstimate, "test: %v", test.name)
			}

			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, firstEstimatePoolId)
			s.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, secondEstimatePoolId)
			s.Require().NoError(err)

			s.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			s.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)
		})
	}
}

func (s *KeeperTestSuite) makeGaugesIncentivized(incentivizedGauges []uint64) {
	var records []poolincentivestypes.DistrRecord
	totalWeight := sdk.NewInt(int64(len(incentivizedGauges)))
	for _, gauge := range incentivizedGauges {
		records = append(records, poolincentivestypes.DistrRecord{GaugeId: gauge, Weight: sdk.OneInt()})
	}
	distInfo := poolincentivestypes.DistrInfo{
		TotalWeight: totalWeight,
		Records:     records,
	}
	s.App.PoolIncentivesKeeper.SetDistrInfo(s.Ctx, distInfo)
}

func (s *KeeperTestSuite) calcInGivenOutAmountAsSeparateSwaps(osmoFeeReduced bool, routes []types.SwapAmountOutRoute, tokenOut sdk.Coin) sdk.Coin {
	cacheCtx, _ := s.Ctx.CacheContext()
	if osmoFeeReduced {
		// extract route from swap
		route := types.SwapAmountOutRoutes(routes)
		// utilizing the extracted route, determine the routeSpreadFactor and sumOfspreadFactors
		// these two variables are used to calculate the overall swap fee utilizing the following formula
		// spreadFactor = routeSpreadFactor * ((pool_fee) / (sumOfspreadFactors))
		routeSpreadFactor, sumOfSpreadFactors, err := s.App.PoolManagerKeeper.GetOsmoRoutedMultihopTotalSpreadFactor(s.Ctx, route)
		s.Require().NoError(err)
		nextTokenOut := tokenOut
		for i := len(routes) - 1; i >= 0; i-- {
			hop := routes[i]
			// extract the current pool's swap fee
			hopPool, err := s.App.GAMMKeeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
			s.Require().NoError(err)
			currentPoolSpreadFactor := hopPool.GetSpreadFactor(cacheCtx)
			// utilize the routeSpreadFactor, sumOfSpreadFactors, and current pool swap fee to calculate the new reduced swap fee
			spreadFactor := routeSpreadFactor.Mul((currentPoolSpreadFactor.Quo(sumOfSpreadFactors)))

			takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(cacheCtx, hop.TokenInDenom, nextTokenOut.Denom)
			s.Require().NoError(err)

			swapModule, err := s.App.PoolManagerKeeper.GetPoolModule(cacheCtx, hop.PoolId)
			s.Require().NoError(err)

			// we then do individual swaps until we reach the end of the swap route
			tokenInAmt, err := swapModule.SwapExactAmountOut(cacheCtx, s.TestAccs[0], hopPool, hop.TokenInDenom, sdk.NewInt(100000000), nextTokenOut, spreadFactor)
			s.Require().NoError(err)

			tokenInCoin := sdk.NewCoin(hop.TokenInDenom, tokenInAmt)
			tokenInCoinAfterAddTakerFee, _ := s.App.PoolManagerKeeper.CalcTakerFeeExactOut(tokenInCoin, takerFee)

			nextTokenOut = tokenInCoinAfterAddTakerFee
		}
		return nextTokenOut
	} else {
		nextTokenOut := tokenOut
		for i := len(routes) - 1; i >= 0; i-- {
			hop := routes[i]
			hopPool, err := s.App.PoolManagerKeeper.GetPool(cacheCtx, hop.PoolId)
			s.Require().NoError(err)
			updatedPoolSpreadFactor := hopPool.GetSpreadFactor(cacheCtx)

			takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(cacheCtx, hop.TokenInDenom, nextTokenOut.Denom)
			s.Require().NoError(err)

			swapModule, err := s.App.PoolManagerKeeper.GetPoolModule(cacheCtx, hop.PoolId)
			s.Require().NoError(err)

			tokenInAmt, err := swapModule.SwapExactAmountOut(cacheCtx, s.TestAccs[0], hopPool, hop.TokenInDenom, sdk.NewInt(100000000), nextTokenOut, updatedPoolSpreadFactor)
			s.Require().NoError(err)

			tokenInCoin := sdk.NewCoin(hop.TokenInDenom, tokenInAmt)
			tokenInCoinAfterAddTakerFee, _ := s.App.PoolManagerKeeper.CalcTakerFeeExactOut(tokenInCoin, takerFee)

			nextTokenOut = tokenInCoinAfterAddTakerFee
		}
		return nextTokenOut
	}
}

// calcOutGivenInAmountAsSeparatePoolSwaps calculates the output amount of a series of swaps on PoolManager pools while factoring in reduces swap fee changes.
// If its GAMM pool functions directly to ensure the poolmanager functions route to the correct modules. It it's CL pool functions directly to ensure the
// poolmanager functions route to the correct modules.
func (s *KeeperTestSuite) calcOutGivenInAmountAsSeparatePoolSwaps(osmoFeeReduced bool, routes []types.SwapAmountInRoute, tokenIn sdk.Coin) sdk.Coin {
	cacheCtx, _ := s.Ctx.CacheContext()
	if osmoFeeReduced {
		// extract route from swap
		route := types.SwapAmountInRoutes(routes)
		// utilizing the extracted route, determine the routeSpreadFactor and sumOfSpreadFactors
		// these two variables are used to calculate the overall swap fee utilizing the following formula
		// spreadFactor = routeSpreadFactor * ((pool_fee) / (sumOfSpreadFactors))
		routeSpreadFactor, sumOfSpreadFactors, err := s.App.PoolManagerKeeper.GetOsmoRoutedMultihopTotalSpreadFactor(s.Ctx, route)
		s.Require().NoError(err)
		nextTokenIn := tokenIn

		for _, hop := range routes {
			swapModule, err := s.App.PoolManagerKeeper.GetPoolModule(cacheCtx, hop.PoolId)
			s.Require().NoError(err)

			pool, err := swapModule.GetPool(s.Ctx, hop.PoolId)
			s.Require().NoError(err)

			// utilize the routeSpreadFactor, sumOfSpreadFactors, and current pool swap fee to calculate the new reduced swap fee
			spreadFactor := routeSpreadFactor.Mul(pool.GetSpreadFactor(cacheCtx).Quo(sumOfSpreadFactors))

			takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(cacheCtx, hop.TokenOutDenom, nextTokenIn.Denom)
			s.Require().NoError(err)

			nextTokenInAfterSubTakerFee, _ := s.App.PoolManagerKeeper.CalcTakerFeeExactIn(nextTokenIn, takerFee)

			// we then do individual swaps until we reach the end of the swap route
			tokenOut, err := swapModule.SwapExactAmountIn(cacheCtx, s.TestAccs[0], pool, nextTokenInAfterSubTakerFee, hop.TokenOutDenom, sdk.OneInt(), spreadFactor)
			s.Require().NoError(err)

			nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)
		}
		return nextTokenIn
	} else {
		nextTokenIn := tokenIn
		for _, hop := range routes {
			swapModule, err := s.App.PoolManagerKeeper.GetPoolModule(cacheCtx, hop.PoolId)
			s.Require().NoError(err)

			pool, err := swapModule.GetPool(s.Ctx, hop.PoolId)
			s.Require().NoError(err)

			// utilize the routeSpreadFactor, sumOfSpreadFactors, and current pool swap fee to calculate the new reduced swap fee
			spreadFactor := pool.GetSpreadFactor(cacheCtx)

			takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(cacheCtx, hop.TokenOutDenom, nextTokenIn.Denom)
			s.Require().NoError(err)

			nextTokenInAfterSubTakerFee, _ := s.App.PoolManagerKeeper.CalcTakerFeeExactIn(nextTokenIn, takerFee)

			// we then do individual swaps until we reach the end of the swap route
			tokenOut, err := swapModule.SwapExactAmountIn(cacheCtx, s.TestAccs[0], pool, nextTokenInAfterSubTakerFee, hop.TokenOutDenom, sdk.OneInt(), spreadFactor)
			s.Require().NoError(err)

			nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)

		}
		return nextTokenIn
	}
}

// TODO: abstract SwapAgainstBalancerPool and SwapAgainstConcentratedPool
func (s *KeeperTestSuite) TestSingleSwapExactAmountIn() {
	tests := []struct {
		name                   string
		poolId                 uint64
		poolCoins              sdk.Coins
		poolFee                sdk.Dec
		tokenIn                sdk.Coin
		tokenOutDenom          string
		tokenOutMinAmount      sdk.Int
		expectedTokenOutAmount sdk.Int
		swapWithNoTakerFee     bool
		expectError            bool
	}{
		// Swap with taker fee:
		//  - foo: 1000000000000
		//  - bar: 1000000000000
		//  - spreadFactor: 0.1%
		//  - takerFee: 0.15%
		//  - foo in: 100000
		//  - bar amount out will be calculated according to the formula
		// 		https://www.wolframalpha.com/input?i=solve+%2810%5E12+%2B+10%5E5+x+0.9975%29%2810%5E12+-+x%29+%3D+10%5E24
		{
			name:                   "Swap - [foo -> bar], 0.1 percent fee",
			poolId:                 1,
			poolCoins:              sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:                defaultPoolSpreadFactor,
			tokenIn:                sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:      sdk.NewInt(1),
			tokenOutDenom:          bar,
			expectedTokenOutAmount: sdk.NewInt(99750),
		},
		// Swap with no taker fee:
		//  - foo: 1000000000000
		//  - bar: 1000000000000
		//  - spreadFactor: 0.1%
		//  - foo in: 100000
		//  - bar amount out will be calculated according to the formula
		// 		https://www.wolframalpha.com/input?i=solve+%2810%5E12+%2B+10%5E5+x+0.999%29%2810%5E12+-+x%29+%3D+10%5E24
		{
			name:                   "Swap - [foo -> bar], 0.1 percent fee",
			poolId:                 1,
			poolCoins:              sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:                defaultPoolSpreadFactor,
			tokenIn:                sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:      sdk.NewInt(1),
			tokenOutDenom:          bar,
			swapWithNoTakerFee:     true,
			expectedTokenOutAmount: sdk.NewInt(99899),
		},
		{
			name:              "Wrong pool id",
			poolId:            2,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:           defaultPoolSpreadFactor,
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
			tokenOutDenom:     bar,
			expectError:       true,
		},
		{
			name:              "In denom not exist",
			poolId:            1,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:           defaultPoolSpreadFactor,
			tokenIn:           sdk.NewCoin(baz, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
			tokenOutDenom:     bar,
			expectError:       true,
		},
		{
			name:              "Out denom not exist",
			poolId:            1,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:           defaultPoolSpreadFactor,
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
			tokenOutDenom:     baz,
			expectError:       true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.FundAcc(s.TestAccs[0], tc.poolCoins)
			s.PrepareCustomBalancerPoolFromCoins(tc.poolCoins, balancer.PoolParams{
				SwapFee: tc.poolFee,
				ExitFee: sdk.ZeroDec(),
			})

			// execute the swap
			var multihopTokenOutAmount sdk.Int
			var err error
			if tc.swapWithNoTakerFee {
				multihopTokenOutAmount, err = poolmanagerKeeper.SwapExactAmountInNoTakerFee(s.Ctx, s.TestAccs[0], tc.poolId, tc.tokenIn, tc.tokenOutDenom, tc.tokenOutMinAmount)
			} else {
				multihopTokenOutAmount, err = poolmanagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], tc.poolId, tc.tokenIn, tc.tokenOutDenom, tc.tokenOutMinAmount)
			}
			if tc.expectError {
				s.Require().Error(err)
			} else {
				// compare the expected tokenOut to the actual tokenOut
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedTokenOutAmount.String(), multihopTokenOutAmount.String())
			}
		})
	}
}

type MockPoolModule struct {
	pools []types.PoolI
}

func (m *MockPoolModule) GetPools(ctx sdk.Context) ([]types.PoolI, error) {
	return m.pools, nil
}

// This test suite contains test cases for the AllPools function, which returns a sorted list of all pools from different pool modules.
// The test cases cover various scenarios, including no pool modules, single pool modules, multiple pool modules with varying numbers of pools,
// and overlapping and duplicate pool ids. The expected results and potential errors are defined for each test case.
// The test suite sets up mock pool modules and configures their behavior for the GetPools method, injecting them into the pool manager for testing.
// The actual results of the AllPools function are then compared to the expected results, ensuring the function behaves as intended in each scenario.
// Nots *KeeperTestSuite only test with Balancer Pools, as we're focusing on testing via different modules
func (s *KeeperTestSuite) TestAllPools() {
	s.Setup()

	mockError := errors.New("mock error")

	testCases := []struct {
		name string
		// the return values of each pool module
		// from the call to GetPools()
		poolModuleIds  [][]uint64
		expectedResult []types.PoolI
		expectedError  bool
	}{
		{
			name:           "No pool modules",
			poolModuleIds:  [][]uint64{},
			expectedResult: []types.PoolI{},
		},
		{
			name: "Single pool module",
			poolModuleIds: [][]uint64{
				{1},
			},
			expectedResult: []types.PoolI{
				&balancer.Pool{Id: 1},
			},
		},
		{
			name: "Two pools per module",
			poolModuleIds: [][]uint64{
				{1, 2},
			},
			expectedResult: []types.PoolI{
				&balancer.Pool{Id: 1}, &balancer.Pool{Id: 2},
			},
		},
		{
			name: "Two pools per module, second module with no pools",
			poolModuleIds: [][]uint64{
				{1, 2},
				{},
			},
			expectedResult: []types.PoolI{
				&balancer.Pool{Id: 1}, &balancer.Pool{Id: 2},
			},
		},
		{
			name: "Two pools per module, first module with no pools",
			poolModuleIds: [][]uint64{
				{},
				{1, 2},
			},
			expectedResult: []types.PoolI{
				&balancer.Pool{Id: 1}, &balancer.Pool{Id: 2},
			},
		},
		{
			// This should never happen but added for coverage.
			name: "Two pools per module with repeating ids",
			poolModuleIds: [][]uint64{
				{3, 3},
			},
			expectedResult: []types.PoolI{
				&balancer.Pool{Id: 3}, &balancer.Pool{Id: 3},
			},
		},
		{
			name: "Module with two pools, module with one pool",
			poolModuleIds: [][]uint64{
				{1, 2},
				{4},
			},
			expectedResult: []types.PoolI{
				&balancer.Pool{Id: 1}, &balancer.Pool{Id: 2}, &balancer.Pool{Id: 4},
			},
		},
		{
			name: "Several modules with overlapping and duplicate pool ids",
			poolModuleIds: [][]uint64{
				{1, 32, 77, 1203},
				{1, 4},
				{},
				{4, 88},
			},
			expectedResult: []types.PoolI{
				&balancer.Pool{Id: 1},
				&balancer.Pool{Id: 1},
				&balancer.Pool{Id: 4},
				&balancer.Pool{Id: 4},
				&balancer.Pool{Id: 32},
				&balancer.Pool{Id: 77},
				&balancer.Pool{Id: 88},
				&balancer.Pool{Id: 1203},
			},
		},
		{
			name: "Error case",
			poolModuleIds: [][]uint64{
				{1},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			ctx := s.Ctx
			poolManagerKeeper := s.App.PoolManagerKeeper

			// Configure pool module mocks and inject them into pool manager
			// for testing.
			poolModules := make([]types.PoolModuleI, len(tc.poolModuleIds))
			if tc.expectedError {
				// Configure error case.
				mockPoolModule := mocks.NewMockPoolModuleI(ctrl)
				mockPoolModule.EXPECT().GetPools(ctx).Return(nil, mockError)
				poolModules[0] = mockPoolModule
			} else {
				// Configure success case.
				for i := range tc.poolModuleIds {
					mockPoolModule := mocks.NewMockPoolModuleI(ctrl)
					poolModules[i] = mockPoolModule

					expectedPools := make([]types.PoolI, len(tc.poolModuleIds[i]))
					for j := range tc.poolModuleIds[i] {
						expectedPools[j] = &balancer.Pool{Id: tc.poolModuleIds[i][j]}
					}

					// Set up the expected behavior for the GetPools method
					mockPoolModule.EXPECT().GetPools(ctx).Return(expectedPools, nil)
				}
			}
			poolManagerKeeper.SetPoolModulesUnsafe(poolModules)

			// Call the AllPools function and check if the result matches the expected pools
			actualResult, err := poolManagerKeeper.AllPools(ctx)

			if tc.expectedError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedResult, actualResult)
		})
	}
}

// Tess *KeeperTestSuitests the AllPools function with real pools.
func (s *KeeperTestSuite) TestAllPools_RealPools() {
	s.SetupTest()

	poolManagerKeeper := s.App.PoolManagerKeeper

	expectedResult := []types.PoolI{}

	// Prepare CL pool.
	clPool := s.PrepareConcentratedPool()
	expectedResult = append(expectedResult, clPool)

	// Prepare balancer pool
	balancerId := s.PrepareBalancerPool()
	balancerPool, err := s.App.GAMMKeeper.GetPool(s.Ctx, balancerId)
	s.Require().NoError(err)
	expectedResult = append(expectedResult, balancerPool)

	// Prepare stableswap pool
	stableswapId := s.PrepareBasicStableswapPool()
	stableswapPool, err := s.App.GAMMKeeper.GetPool(s.Ctx, stableswapId)
	s.Require().NoError(err)
	expectedResult = append(expectedResult, stableswapPool)

	// Prepare cosmwasm pool
	cwPool := s.PrepareCosmWasmPool()
	expectedResult = append(expectedResult, cwPool)

	// Call the AllPools function and check if the result matches the expected pools
	actualResult, err := poolManagerKeeper.AllPools(s.Ctx)
	s.Require().NoError(err)

	for i, expectedPool := range expectedResult {
		// N.B. CosmWasm pools cannot be compared directly because part of their declaration
		// is an interface (WasmKeeper). This fails reflection comparison. As a workaround,
		// we type cast and compare the CosmWasmPool field directly.
		if expectedPool.GetType() == types.CosmWasm {
			cwPoolExpected, ok := expectedPool.(*cwmodel.Pool)
			s.Require().True(ok)
			cwPoolActual, ok := actualResult[i].(*cwmodel.Pool)
			s.Require().True(ok)

			s.Require().Equal(cwPoolExpected.CosmWasmPool, cwPoolActual.CosmWasmPool)
		} else {
			s.Require().Equal(expectedPool, actualResult[i])
		}
	}
}

// sets *KeeperTestSuiteof desired type and returns their IDs
func (s *KeeperTestSuite) setupPools(poolType types.PoolType, poolDefaultSpreadFactor sdk.Dec) (firstEstimatePoolId, secondEstimatePoolId uint64) {
	switch poolType {
	case types.Stableswap:
		// Prepare 4 pools,
		// Two pools for calculating `MultihopSwapExactAmountOut`
		// and two pools for calculating `EstimateMultihopSwapExactAmountOut`
		s.PrepareBasicStableswapPool()
		s.PrepareBasicStableswapPool()

		firstEstimatePoolId = s.PrepareBasicStableswapPool()

		secondEstimatePoolId = s.PrepareBasicStableswapPool()
		return
	default:
		// Prepare 4 pools,
		// Two pools for calculating `MultihopSwapExactAmountOut`
		// and two pools for calculating `EstimateMultihopSwapExactAmountOut`
		s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSpreadFactor, // 1%
			ExitFee: sdk.NewDec(0),
		})
		s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSpreadFactor,
			ExitFee: sdk.NewDec(0),
		})

		firstEstimatePoolId = s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSpreadFactor, // 1%
			ExitFee: sdk.NewDec(0),
		})

		secondEstimatePoolId = s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSpreadFactor,
			ExitFee: sdk.NewDec(0),
		})
		return
	}
}

// TestSplitRouteExactAmountIn tests the splitRouteExactAmountIn function.
func (s *KeeperTestSuite) TestSplitRouteExactAmountIn() {
	var (
		defaultSingleRouteOneHop = []types.SwapAmountInSplitRoute{
			{
				Pools: []types.SwapAmountInRoute{
					{
						PoolId:        fooBarPoolId,
						TokenOutDenom: bar,
					},
				},
				TokenInAmount: twentyFiveBaseUnitsAmount,
			},
		}

		defaultTwoHopRoutes = []types.SwapAmountInRoute{
			{
				PoolId:        fooBarPoolId,
				TokenOutDenom: bar,
			},
			{
				PoolId:        barBazPoolId,
				TokenOutDenom: baz,
			},
		}

		defaultSingleRouteTwoHops = types.SwapAmountInSplitRoute{
			Pools:         defaultTwoHopRoutes,
			TokenInAmount: twentyFiveBaseUnitsAmount,
		}

		defaultSingleRouteThreeHops = types.SwapAmountInSplitRoute{
			Pools: []types.SwapAmountInRoute{
				{
					PoolId:        fooBarPoolId,
					TokenOutDenom: bar,
				},
				{
					PoolId:        barUosmoPoolId,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        bazUosmoPoolId,
					TokenOutDenom: baz,
				},
			},
			TokenInAmount: sdk.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
		}

		priceImpactThreshold = sdk.NewInt(97469586)
	)

	tests := map[string]struct {
		isInvalidSender   bool
		routes            []types.SwapAmountInSplitRoute
		tokenInDenom      string
		tokenOutMinAmount sdk.Int

		// This value was taken from the actual result
		// and not manually calculated. This is acceptable
		// for this test because we are not testing the math
		// but the routing logic.
		// The math should be tested per-module.
		// We keep this assertion to make sure that the
		// actual result is within a reasonable range.
		expectedTokenOutEstimate sdk.Int

		expectError error
	}{
		"valid solo route one hop": {
			routes:            defaultSingleRouteOneHop,
			tokenInDenom:      foo,
			tokenOutMinAmount: sdk.OneInt(),

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
		},
		"valid solo route multi hop": {
			routes:            []types.SwapAmountInSplitRoute{defaultSingleRouteTwoHops},
			tokenInDenom:      foo,
			tokenOutMinAmount: sdk.OneInt(),

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
		},
		"valid split route multi hop": {
			routes: []types.SwapAmountInSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenInDenom:      foo,
			tokenOutMinAmount: sdk.OneInt(),

			// 1x from single route two hops and 3x from single route three hops
			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount.MulRaw(4),
		},

		"valid split route multi hop with price impact protection that would fail individual route if given per multihop": {
			routes: []types.SwapAmountInSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenInDenom: foo,
			// equal to the expected amount
			// every route individually would fail, but the split route should succeed
			tokenOutMinAmount: priceImpactThreshold,

			expectedTokenOutEstimate: priceImpactThreshold,
		},

		"error: price impact protection triggered": {
			routes: []types.SwapAmountInSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenInDenom: foo,
			// one greater than expected amount
			tokenOutMinAmount: priceImpactThreshold.Add(sdk.OneInt()),

			expectError: types.PriceImpactProtectionExactInError{Actual: priceImpactThreshold, MinAmount: priceImpactThreshold.Add(sdk.OneInt())},
		},
		"error: duplicate split routes": {
			routes: []types.SwapAmountInSplitRoute{
				defaultSingleRouteTwoHops,
				{
					Pools: defaultSingleRouteTwoHops.Pools,
					// Note that the routes are deemed equal even if the token in amount is different
					// We only care about the pools for comparison.
					TokenInAmount: defaultSingleRouteTwoHops.TokenInAmount.MulRaw(3),
				},
			},
			tokenInDenom:      foo,
			tokenOutMinAmount: sdk.OneInt(),

			expectError: types.ErrDuplicateRoutesNotAllowed,
		},

		"error: invalid pool id": {
			routes: []types.SwapAmountInSplitRoute{
				{
					Pools: []types.SwapAmountInRoute{
						{
							PoolId:        uint64(len(defaultValidPools) + 1),
							TokenOutDenom: bar,
						},
					},
					TokenInAmount: twentyFiveBaseUnitsAmount,
				},
			},
			tokenInDenom:      foo,
			tokenOutMinAmount: sdk.OneInt(),

			expectError: types.FailedToFindRouteError{PoolId: uint64(len(defaultValidPools) + 1)},
		},
	}

	s.PrepareBalancerPool()
	s.PrepareConcentratedPool()

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			k := s.App.PoolManagerKeeper

			sender := s.TestAccs[1]

			for _, pool := range defaultValidPools {
				s.CreatePoolFromTypeWithCoins(pool.poolType, pool.initialLiquidity)

				// Fund sender with initial liqudity
				// If not valid, we don't fund to trigger an error case.
				if !tc.isInvalidSender {
					s.FundAcc(sender, pool.initialLiquidity)
				}
			}

			tokenOut, err := k.SplitRouteExactAmountIn(s.Ctx, sender, tc.routes, tc.tokenInDenom, tc.tokenOutMinAmount)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectError.Error())
				return
			}
			s.Require().NoError(err)

			// Note, we use a 1% error tolerance with rounding down
			// because we initialize the reserves 1:1 so by performing
			// the swap we don't expect the price to change significantly.
			// As a result, we roughly expect the amount out to be the same
			// as the amount in given in another token. However, the actual
			// amount must be stricly less than the given due to price impact.
			errTolerance := osmomath.ErrTolerance{
				RoundingDir:             osmomath.RoundDown,
				MultiplicativeTolerance: sdk.NewDec(1),
			}

			s.Require().Equal(0, errTolerance.Compare(tc.expectedTokenOutEstimate, tokenOut), fmt.Sprintf("expected %s, got %s", tc.expectedTokenOutEstimate, tokenOut))
		})
	}
}

// TestSplitRouteExactAmountOut tests the split route exact amount out functionality
func (s *KeeperTestSuite) TestSplitRouteExactAmountOut() {
	var (
		defaultSingleRouteOneHop = []types.SwapAmountOutSplitRoute{
			{
				Pools: []types.SwapAmountOutRoute{
					{
						PoolId:       fooBarPoolId,
						TokenInDenom: foo,
					},
				},
				TokenOutAmount: twentyFiveBaseUnitsAmount,
			},
		}

		defaultTwoHopRoutes = []types.SwapAmountOutRoute{
			{
				PoolId:       fooBarPoolId,
				TokenInDenom: foo,
			},
			{
				PoolId:       barBazPoolId,
				TokenInDenom: bar,
			},
		}

		defaultSingleRouteTwoHops = types.SwapAmountOutSplitRoute{
			Pools:          defaultTwoHopRoutes,
			TokenOutAmount: twentyFiveBaseUnitsAmount,
		}

		defaultSingleRouteThreeHops = types.SwapAmountOutSplitRoute{
			Pools: []types.SwapAmountOutRoute{
				{
					PoolId:       fooBarPoolId,
					TokenInDenom: foo,
				},
				{
					PoolId:       barUosmoPoolId,
					TokenInDenom: bar,
				},
				{
					PoolId:       bazUosmoPoolId,
					TokenInDenom: uosmo,
				},
			},
			TokenOutAmount: sdk.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
		}

		priceImpactThreshold = sdk.NewInt(102666473)
	)

	tests := map[string]struct {
		isInvalidSender  bool
		routes           []types.SwapAmountOutSplitRoute
		tokenOutDenom    string
		tokenInMaxAmount sdk.Int

		// This value was taken from the actual result
		// and not manually calculated. This is acceptable
		// for this test because we are not testing the math
		// but the routing logic.
		// The math should be tested per-module.
		// We keep this assertion to make sure that the
		// actual result is within a reasonable range.
		expectedTokenOutEstimate sdk.Int

		expectError error
	}{
		"valid solo route one hop": {
			routes:           defaultSingleRouteOneHop,
			tokenOutDenom:    bar,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
		},
		"valid solo route multi hop": {
			routes:           []types.SwapAmountOutSplitRoute{defaultSingleRouteTwoHops},
			tokenOutDenom:    baz,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
		},
		"valid split route multi hop": {
			routes: []types.SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenOutDenom:    baz,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			// 1x from single route two hops and 3x from single route three hops
			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount.MulRaw(4),
		},

		"valid split route multi hop with price impact protection that would fail individual route if given per multihop": {
			routes: []types.SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenOutDenom: baz,
			// equal to the amount calculated.
			// every route individually would fail, but the split route should succeed
			tokenInMaxAmount: priceImpactThreshold,

			expectedTokenOutEstimate: priceImpactThreshold,
		},

		"error: price impact protection triggerred": {
			routes: []types.SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenOutDenom: baz,
			// one less than expected amount
			// every route individually would fail, but the split route should succeed
			tokenInMaxAmount: priceImpactThreshold.Sub(sdk.OneInt()),

			expectError: types.PriceImpactProtectionExactOutError{Actual: priceImpactThreshold, MaxAmount: priceImpactThreshold.Sub(sdk.OneInt())},
		},

		"error: duplicate split routes": {
			routes: []types.SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHops,
				{
					Pools: defaultSingleRouteTwoHops.Pools,
					// Note that the routes are deemed equal even if the token in amount is different
					// We only care about the pools for comparison.
					TokenOutAmount: defaultSingleRouteTwoHops.TokenOutAmount.MulRaw(3),
				},
			},
			tokenOutDenom:    foo,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			expectError: types.ErrDuplicateRoutesNotAllowed,
		},

		"error: invalid pool id": {
			routes: []types.SwapAmountOutSplitRoute{
				{
					Pools: []types.SwapAmountOutRoute{
						{
							PoolId:       uint64(len(defaultValidPools) + 1),
							TokenInDenom: foo,
						},
					},
					TokenOutAmount: twentyFiveBaseUnitsAmount,
				},
			},
			tokenOutDenom:    foo,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			expectError: types.FailedToFindRouteError{PoolId: uint64(len(defaultValidPools) + 1)},
		},
	}

	s.PrepareBalancerPool()
	s.PrepareConcentratedPool()

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			k := s.App.PoolManagerKeeper

			sender := s.TestAccs[1]

			for _, pool := range defaultValidPools {
				s.CreatePoolFromTypeWithCoins(pool.poolType, pool.initialLiquidity)

				// Fund sender with initial liqudity
				// If not valid, we don't fund to trigger an error case.
				if !tc.isInvalidSender {
					s.FundAcc(sender, pool.initialLiquidity)
				}
			}

			tokenOut, err := k.SplitRouteExactAmountOut(s.Ctx, sender, tc.routes, tc.tokenOutDenom, tc.tokenInMaxAmount)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectError.Error())
				return
			}
			s.Require().NoError(err)

			// Note, we use a 1% error tolerance with rounding up
			// because we initialize the reserves 1:1 so by performing
			// the swap we don't expect the price to change significantly.
			// As a result, we roughly expect the amount in to be the same
			// as the amount out given of another token. However, the actual
			// amount must be stricly greater than the given due to price impact.
			errTolerance := osmomath.ErrTolerance{
				RoundingDir:             osmomath.RoundUp,
				MultiplicativeTolerance: sdk.NewDec(1),
			}

			s.Require().Equal(0, errTolerance.Compare(tc.expectedTokenOutEstimate, tokenOut), fmt.Sprintf("expected %s, got %s", tc.expectedTokenOutEstimate, tokenOut))
		})
	}
}

func (s *KeeperTestSuite) TestGetTotalPoolLiquidity() {
	const (
		cosmWasmPoolId = uint64(3)
	)
	var (
		defaultPoolCoinOne = sdk.NewCoin("usdc", sdk.OneInt())
		defaultPoolCoinTwo = sdk.NewCoin("eth", sdk.NewInt(2))
		nonPoolCool        = sdk.NewCoin("uosmo", sdk.NewInt(3))

		defaultCoins = sdk.NewCoins(defaultPoolCoinOne, defaultPoolCoinTwo)
	)

	tests := []struct {
		name           string
		poolId         uint64
		poolLiquidity  sdk.Coins
		expectedResult sdk.Coins
		expectedErr    error
	}{
		{
			name:           "CL Pool: valid with 2 coins",
			poolId:         1,
			poolLiquidity:  defaultCoins,
			expectedResult: defaultCoins,
		},
		{
			name:           "CL Pool: valid with 1 coin",
			poolId:         1,
			poolLiquidity:  sdk.NewCoins(defaultPoolCoinTwo),
			expectedResult: sdk.NewCoins(defaultPoolCoinTwo),
		},
		{
			// can only happen if someone sends extra tokens to pool
			// address. Should not occur in practice.
			name:           "CL Pool: valid with 3 coins",
			poolId:         1,
			poolLiquidity:  sdk.NewCoins(defaultPoolCoinTwo, defaultPoolCoinOne, nonPoolCool),
			expectedResult: defaultCoins,
		},
		{
			// this can happen if someone sends random dust to pool address.
			name:           "CL Pool:only non-pool coin - does not show up in result",
			poolId:         1,
			poolLiquidity:  sdk.NewCoins(nonPoolCool),
			expectedResult: sdk.Coins(nil),
		},
		{
			name:           "Balancer Pool: with default pool assets",
			poolId:         2,
			poolLiquidity:  sdk.NewCoins(apptesting.DefaultPoolAssets[0].Token, apptesting.DefaultPoolAssets[1].Token, apptesting.DefaultPoolAssets[2].Token, apptesting.DefaultPoolAssets[3].Token),
			expectedResult: sdk.NewCoins(apptesting.DefaultPoolAssets[0].Token, apptesting.DefaultPoolAssets[1].Token, apptesting.DefaultPoolAssets[2].Token, apptesting.DefaultPoolAssets[3].Token),
		},
		{
			name:           "Cosmwasm pool",
			poolId:         cosmWasmPoolId,
			poolLiquidity:  sdk.NewCoins(defaultPoolCoinOne, defaultPoolCoinTwo),
			expectedResult: sdk.NewCoins(defaultPoolCoinOne, defaultPoolCoinTwo),
		},
		{
			name:        "round not found because pool id doesnot exist",
			poolId:      cosmWasmPoolId + 1,
			expectedErr: types.FailedToFindRouteError{PoolId: cosmWasmPoolId + 1},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Create default CL pool
			clPool := s.PrepareConcentratedPool()
			// Since CL pool is created with no initial liquidity, we need to fund it.
			s.FundAcc(clPool.GetAddress(), tc.poolLiquidity)
			s.PrepareBalancerPool()
			if tc.poolLiquidity.Len() == 2 {
				s.FundAcc(s.TestAccs[0], tc.poolLiquidity)
				s.CreatePoolFromTypeWithCoinsAndSpreadFactor(types.CosmWasm, tc.poolLiquidity, sdk.ZeroDec())
			}

			// Get pool defined in test case
			actual, err := s.App.PoolManagerKeeper.GetTotalPoolLiquidity(s.Ctx, tc.poolId)
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedErr)
				s.Require().Nil(actual)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedResult, actual)
		})
	}
}

func (s *KeeperTestSuite) TestIsOsmoRoutedMultihop() {
	tests := map[string]struct {
		route                  types.MultihopRoute
		balancerPoolCoins      []sdk.Coins
		concentratedPoolDenoms [][]string
		incentivizedGauges     []uint64
		inDenom                string
		outDenom               string
		expectIsRouted         bool
	}{
		"happy path: osmo routed (balancer)": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			balancerPoolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(uosmo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 2.
			},
			// Note that we incentivize all candidate gauges for the sake of test readability.
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			inDenom:            foo,
			outDenom:           bar,

			expectIsRouted: true,
		},
		"happy path: osmo routed (balancer, only one active gauge for each pool)": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			balancerPoolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(uosmo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 2.
			},
			incentivizedGauges: []uint64{1, 4},
			inDenom:            foo,
			outDenom:           bar,

			expectIsRouted: true,
		},
		"osmo routed (concentrated)": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			concentratedPoolDenoms: [][]string{
				{foo, uosmo}, // pool 1.
				{uosmo, baz}, // pool 2.
			},
			incentivizedGauges: []uint64{1, 2},
			inDenom:            foo,
			outDenom:           bar,

			expectIsRouted: true,
		},
		"osmo routed (mixed concentrated and balancer)": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			concentratedPoolDenoms: [][]string{
				{foo, uosmo}, // pool 1.
			},
			balancerPoolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(uosmo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 2.
			},

			incentivizedGauges: []uint64{1, 2},
			inDenom:            foo,
			outDenom:           bar,

			expectIsRouted: true,
		},
		"not osmo routed (single pool)": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
			}),
			inDenom:  foo,
			outDenom: bar,

			expectIsRouted: false,
		},
		"not osmo routed (two pools)": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			}),
			inDenom:  foo,
			outDenom: baz,

			expectIsRouted: false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolManagerKeeper := s.App.PoolManagerKeeper

			// Create pools to route through
			if tc.concentratedPoolDenoms != nil {
				s.CreateConcentratedPoolsAndFullRangePosition(tc.concentratedPoolDenoms)
			}

			if tc.balancerPoolCoins != nil {
				s.createBalancerPoolsFromCoins(tc.balancerPoolCoins)
			}

			// If test specifies incentivized gauges, set them here
			if len(tc.incentivizedGauges) > 0 {
				s.makeGaugesIncentivized(tc.incentivizedGauges)
			}

			// System under test
			isRouted := poolManagerKeeper.IsOsmoRoutedMultihop(s.Ctx, tc.route, tc.inDenom, tc.outDenom)

			// Check output
			s.Require().Equal(tc.expectIsRouted, isRouted)
		})
	}
}

// TestGetOsmoRoutedMultihopTotalSpreadFactor tests the GetOsmoRoutedMultihopTotalSpreadFactor function
func (s *KeeperTestSuite) TestGetOsmoRoutedMultihopTotalSpreadFactor() {
	tests := map[string]struct {
		route                  types.MultihopRoute
		balancerPoolCoins      []sdk.Coins
		concentratedPoolDenoms [][]string
		poolFees               []sdk.Dec

		expectedRouteFee sdk.Dec
		expectedTotalFee sdk.Dec
		expectedError    error
	}{
		"happy path: balancer route": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			poolFees: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			balancerPoolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(uosmo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 2.
			},

			expectedRouteFee: defaultPoolSpreadFactor,
			expectedTotalFee: defaultPoolSpreadFactor.Add(defaultPoolSpreadFactor),
		},
		"concentrated route": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			poolFees: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			concentratedPoolDenoms: [][]string{
				{foo, uosmo}, // pool 1.
				{uosmo, baz}, // pool 2.
			},

			expectedRouteFee: defaultPoolSpreadFactor,
			expectedTotalFee: defaultPoolSpreadFactor.Add(defaultPoolSpreadFactor),
		},
		"mixed concentrated and balancer route": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			poolFees: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			concentratedPoolDenoms: [][]string{
				{foo, uosmo}, // pool 1.
			},
			balancerPoolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(uosmo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 2.
			},

			expectedRouteFee: defaultPoolSpreadFactor,
			expectedTotalFee: defaultPoolSpreadFactor.Add(defaultPoolSpreadFactor),
		},
		"edge case: average fee is lower than highest pool fee": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			// Note that pool 2 has 5x the swap fee of pool 1
			poolFees: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor.Mul(sdk.NewDec(5))},
			concentratedPoolDenoms: [][]string{
				{foo, uosmo}, // pool 1.
				{uosmo, baz}, // pool 2.
			},

			expectedRouteFee: defaultPoolSpreadFactor.Mul(sdk.NewDec(5)),
			expectedTotalFee: defaultPoolSpreadFactor.Mul(sdk.NewDec(6)),
		},
		"error: pool does not exist": {
			route: types.SwapAmountInRoutes([]types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: bar,
				},
			}),
			poolFees: []sdk.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},

			expectedError: types.FailedToFindRouteError{PoolId: 1},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolManagerKeeper := s.App.PoolManagerKeeper

			// Create pools for test route
			if tc.concentratedPoolDenoms != nil {
				s.CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(tc.concentratedPoolDenoms, tc.poolFees)
			}

			if tc.balancerPoolCoins != nil {
				s.createBalancerPoolsFromCoinsWithSpreadFactor(tc.balancerPoolCoins, tc.poolFees)
			}

			// System under test
			routeFee, totalFee, err := poolManagerKeeper.GetOsmoRoutedMultihopTotalSpreadFactor(s.Ctx, tc.route)

			// Assertions
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
				s.Require().Equal(sdk.Dec{}, routeFee)
				s.Require().Equal(sdk.Dec{}, totalFee)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedRouteFee, routeFee)
			s.Require().Equal(tc.expectedTotalFee, totalFee)
		})
	}
}

func (suite *KeeperTestSuite) TestCreateMultihopExpectedSwapOuts() {
	tests := map[string]struct {
		route                       []types.SwapAmountOutRoute
		tokenOut                    sdk.Coin
		balancerPoolCoins           []sdk.Coins
		concentratedPoolDenoms      [][]string
		poolCoins                   []sdk.Coins
		cumulativeRouteSpreadFactor sdk.Dec
		sumOfSpreadFactors          sdk.Dec

		expectedSwapIns []sdk.Int
		expectedError   bool
	}{
		"happy path: one route": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: bar,
				},
			},
			poolCoins: []sdk.Coins{sdk.NewCoins(sdk.NewCoin(foo, sdk.NewInt(100)), sdk.NewCoin(bar, sdk.NewInt(100)))},

			tokenOut: sdk.NewCoin(foo, sdk.NewInt(10)),
			// expectedSwapIns = (tokenOut * (poolTokenOutBalance / poolPostSwapOutBalance)).ceil()
			// foo token = 10 * (100 / 90) ~ 12
			expectedSwapIns: []sdk.Int{sdk.NewInt(12)},
		},
		"happy path: two route": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: bar,
				},
			},

			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, sdk.NewInt(100)), sdk.NewCoin(bar, sdk.NewInt(100))), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, sdk.NewInt(100)), sdk.NewCoin(baz, sdk.NewInt(100))), // pool 2.
			},
			tokenOut: sdk.NewCoin(baz, sdk.NewInt(10)),
			// expectedSwapIns = (tokenOut * (poolTokenOutBalance / poolPostSwapOutBalance)).ceil()
			// foo token = 10 * (100 / 90) ~ 12
			// bar token = 12 * (100 / 88) ~ 14
			expectedSwapIns: []sdk.Int{sdk.NewInt(14), sdk.NewInt(12)},
		},
		"happy path: one route with swap Fee": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: bar,
				},
			},
			poolCoins:                   []sdk.Coins{sdk.NewCoins(sdk.NewCoin(uosmo, sdk.NewInt(100)), sdk.NewCoin(bar, sdk.NewInt(100)))},
			cumulativeRouteSpreadFactor: sdk.NewDec(100),
			sumOfSpreadFactors:          sdk.NewDec(500),

			tokenOut:        sdk.NewCoin(uosmo, sdk.NewInt(10)),
			expectedSwapIns: []sdk.Int{sdk.NewInt(12)},
		},
		"happy path: two route with swap Fee": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: bar,
				},
			},

			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, sdk.NewInt(100)), sdk.NewCoin(bar, sdk.NewInt(100))),   // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, sdk.NewInt(100)), sdk.NewCoin(uosmo, sdk.NewInt(100))), // pool 2.
			},
			cumulativeRouteSpreadFactor: sdk.NewDec(100),
			sumOfSpreadFactors:          sdk.NewDec(500),

			tokenOut:        sdk.NewCoin(uosmo, sdk.NewInt(10)),
			expectedSwapIns: []sdk.Int{sdk.NewInt(14), sdk.NewInt(12)},
		},
		"error: Invalid Pool": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       100,
					TokenInDenom: foo,
				},
			},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, sdk.NewInt(100)), sdk.NewCoin(bar, sdk.NewInt(100))), // pool 1.
			},
			tokenOut:      sdk.NewCoin(baz, sdk.NewInt(10)),
			expectedError: true,
		},
		"error: calculating in given out": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: uosmo,
				},
			},

			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, sdk.NewInt(100)), sdk.NewCoin(bar, sdk.NewInt(100))), // pool 1.
			},
			tokenOut:        sdk.NewCoin(baz, sdk.NewInt(10)),
			expectedSwapIns: []sdk.Int{},

			expectedError: true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()

			suite.createBalancerPoolsFromCoins(tc.poolCoins)

			var actualSwapOuts []sdk.Int
			var err error

			if !tc.sumOfSpreadFactors.IsNil() && !tc.cumulativeRouteSpreadFactor.IsNil() {
				actualSwapOuts, err = suite.App.PoolManagerKeeper.CreateOsmoMultihopExpectedSwapOuts(suite.Ctx, tc.route, tc.tokenOut, tc.cumulativeRouteSpreadFactor, tc.sumOfSpreadFactors)
			} else {
				actualSwapOuts, err = suite.App.PoolManagerKeeper.CreateMultihopExpectedSwapOuts(suite.Ctx, tc.route, tc.tokenOut)
			}
			if tc.expectedError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedSwapIns, actualSwapOuts)
			}
		})
	}
}
