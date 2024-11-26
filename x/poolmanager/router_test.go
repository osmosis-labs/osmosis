package poolmanager_test

import (
	"errors"
	"reflect"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/tests/mocks"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	cwpool "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool"
	cwmodel "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/model"
	gamm "github.com/osmosis-labs/osmosis/v27/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

type poolSetup struct {
	poolType         types.PoolType
	initialLiquidity sdk.Coins
	takerFee         osmomath.Dec
}

type expectedTakerFees struct {
	communityPoolQuoteAssets    sdk.Coins
	communityPoolNonQuoteAssets sdk.Coins
	stakingRewardAssets         sdk.Coins
}

const (
	FOO   = apptesting.FOO
	BAR   = apptesting.BAR
	BAZ   = apptesting.BAZ
	UOSMO = appparams.BaseCoinUnit

	// Not an authorized quote denom
	// ("abc" ensures its always first lexicographically which simplifies setup)
	abc = "abc"
)

var (
	defaultInitPoolAmount   = osmomath.NewInt(1000000000000)
	defaultPoolSpreadFactor = osmomath.NewDecWithPrec(1, 3) // 0.1% pool spread factor default
	pointOneFivePercent     = osmomath.MustNewDecFromStr("0.0015")
	pointThreePercent       = osmomath.MustNewDecFromStr("0.003")
	pointThreeFivePercent   = osmomath.MustNewDecFromStr("0.0035")
	defaultTakerFee         = osmomath.ZeroDec()
	defaultSwapAmount       = osmomath.NewInt(1000000)
	gammKeeperType          = reflect.TypeOf(&gamm.Keeper{})
	concentratedKeeperType  = reflect.TypeOf(&cl.Keeper{})
	cosmwasmKeeperType      = reflect.TypeOf(&cwpool.Keeper{})
	zeroTakerFeeDistr       = expectedTakerFees{
		communityPoolQuoteAssets:    sdk.NewCoins(),
		communityPoolNonQuoteAssets: sdk.NewCoins(),
		stakingRewardAssets:         sdk.NewCoins(),
	}
	communityPoolAddrName = "distribution"
	txFeesStakingAddrName = txfeestypes.NonNativeTxFeeCollectorName
	nonQuoteCommAddrName  = txfeestypes.TakerFeeCommunityPoolName
	takerFeeAddrName      = txfeestypes.TakerFeeCollectorName

	defaultPoolInitAmount     = osmomath.NewInt(10_000_000_000)
	twentyFiveBaseUnitsAmount = osmomath.NewInt(25_000_000)

	fooCoin   = sdk.NewCoin(FOO, defaultPoolInitAmount)
	barCoin   = sdk.NewCoin(BAR, defaultPoolInitAmount)
	bazCoin   = sdk.NewCoin(BAZ, defaultPoolInitAmount)
	uosmoCoin = sdk.NewCoin(UOSMO, defaultPoolInitAmount)
	abcCoin   = sdk.NewCoin(abc, defaultPoolInitAmount)

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
	fooAbcCoins    = sdk.NewCoins(abcCoin, fooCoin)
	fooAbcPoolId   = bazUosmoPoolId + 1
	bazAbcCoins    = sdk.NewCoins(abcCoin, bazCoin)
	bazAbcPoolId   = fooAbcPoolId + 1
	uosmoAbcCoins  = sdk.NewCoins(abcCoin, uosmoCoin)
	uosmoAbcPoolId = bazAbcPoolId + 1

	defaultValidPools = []poolSetup{
		{
			poolType:         types.Balancer,
			initialLiquidity: fooBarCoins,
			takerFee:         defaultTakerFee,
		},
		{
			poolType:         types.Concentrated,
			initialLiquidity: fooBazCoins,
			takerFee:         defaultTakerFee,
		},
		{
			poolType:         types.Balancer,
			initialLiquidity: fooUosmoCoins,
			takerFee:         defaultTakerFee,
		},
		{
			poolType:         types.Concentrated,
			initialLiquidity: barBazCoins,
			takerFee:         defaultTakerFee,
		},
		{
			poolType:         types.Balancer,
			initialLiquidity: barUosmoCoins,
			takerFee:         defaultTakerFee,
		},
		{
			poolType:         types.Concentrated,
			initialLiquidity: bazUosmoCoins,
			takerFee:         defaultTakerFee,
		},
		// Note that abc is not an authorized quote denom
		{
			poolType:         types.Balancer,
			initialLiquidity: fooAbcCoins,
			takerFee:         defaultTakerFee,
		},
		{
			poolType:         types.Concentrated,
			initialLiquidity: bazAbcCoins,
			takerFee:         defaultTakerFee,
		},
		{
			poolType:         types.Balancer,
			initialLiquidity: uosmoAbcCoins,
			takerFee:         defaultTakerFee,
		},
	}

	emptyCoins = sdk.NewCoins()
)

// withTakerFees overrides the taker fees for the given pool setup info at the given indices and returns the full set of updated pool setup info.
func (s *KeeperTestSuite) withTakerFees(pools []poolSetup, indicesToUpdate []uint64, updatedFees []osmomath.Dec) []poolSetup {
	s.Require().Equal(len(indicesToUpdate), len(updatedFees))

	// Deep copy pools
	copiedPools := make([]poolSetup, len(pools))
	for i, pool := range pools {
		copiedPools[i] = pool
	}

	// Update taker fees on copied pools
	for i, index := range indicesToUpdate {
		copiedPools[index].takerFee = updatedFees[i]
	}

	return copiedPools
}

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

			createdPoolId := s.CreatePoolFromType(tc.preCreatePoolType)

			if len(tc.routesOverwrite) > 0 {
				poolmanagerKeeper.SetPoolRoutesUnsafe(tc.routesOverwrite)
			}

			swapModule, err := poolmanagerKeeper.GetPoolModule(s.Ctx, tc.poolId)

			if tc.expectError != nil {
				s.Require().Error(err, "requested pool ID: %d, created pool ID: %d", tc.poolId, createdPoolId)
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

// TestGetPoolTypeGas tests that the result for GetPoolType charges the
// same gas whether its a cache hit or cache fail.
func (s *KeeperTestSuite) TestGetPoolTypeGas() {
	s.SetupTest()
	poolmanagerKeeper := s.App.PoolManagerKeeper

	createdPoolId := s.CreatePoolFromType(types.Balancer)

	// cache miss
	s.App.PoolManagerKeeper.ResetCaches()
	startGas := s.Ctx.GasMeter().GasConsumed()
	_, err := poolmanagerKeeper.GetPoolType(s.Ctx, createdPoolId)
	s.Require().NoError(err)
	endGas := s.Ctx.GasMeter().GasConsumed()
	cacheMissGas := endGas - startGas

	startGas = s.Ctx.GasMeter().GasConsumed()
	_, err = poolmanagerKeeper.GetPoolType(s.Ctx, createdPoolId)
	s.Require().NoError(err)
	endGas = s.Ctx.GasMeter().GasConsumed()
	cacheHitGas := endGas - startGas
	s.Require().Equal(cacheMissGas, cacheHitGas)
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
			expectedDenoms:    []string{"bar", "baz", "foo", appparams.BaseCoinUnit},
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
		expectedSpotPrice osmomath.BigDec

		expectError error
	}{
		"valid balancer pool": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			quoteAssetDenom:   "bar",
			baseAssetDenom:    "baz",
			expectedSpotPrice: osmomath.MustNewBigDecFromStr("1.5"),
		},
		"valid stableswap pool": {
			preCreatePoolType: types.Stableswap,
			poolId:            1,
			quoteAssetDenom:   "bar",
			baseAssetDenom:    "baz",
			expectedSpotPrice: osmomath.MustNewBigDecFromStr("0.99999998"),
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
			expectedSpotPrice: osmomath.MustNewBigDecFromStr("5000.000000000000000129"),
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
			expectedSpotPrice: osmomath.OneBigDec(),
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
				coins := sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(1000000)), sdk.NewCoin("usdc", osmomath.NewInt(5000000000)))
				s.FundAcc(s.TestAccs[0], coins)

				clMsgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)
				_, err := clMsgServer.CreatePosition(s.Ctx, &cltypes.MsgCreatePosition{
					PoolId:          1,
					Sender:          s.TestAccs[0].String(),
					LowerTick:       int64(30545000),
					UpperTick:       int64(31500000),
					TokensProvided:  coins,
					TokenMinAmount0: osmomath.ZeroInt(),
					TokenMinAmount1: osmomath.ZeroInt(),
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
func (s *KeeperTestSuite) TestMultihopSwapExactAmountIn() {
	tests := []struct {
		name               string
		poolCoins          []sdk.Coins
		poolSpreadFactor   []osmomath.Dec
		poolType           []types.PoolType
		routes             []types.SwapAmountInRoute
		incentivizedGauges []uint64
		tokenIn            sdk.Coin
		tokenOutMinAmount  osmomath.Int
		spreadFactor       osmomath.Dec
		expectError        bool
	}{
		{
			name:             "One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins:        []sdk.Coins{sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount))},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor},
			poolType:         []types.PoolType{types.Balancer},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: BAR,
				},
			},
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount: osmomath.NewInt(1),
		},
		{
			name: "Two routes: Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: BAR,
				},
				{
					PoolId:        2,
					TokenOutDenom: BAZ,
				},
			},
			incentivizedGauges: []uint64{},
			tokenIn:            sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:  osmomath.NewInt(1),
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), both pools 1 percent fee, sanity check no more half fee applied",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAZ, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: UOSMO,
				},
				{
					PoolId:        2,
					TokenOutDenom: BAZ,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:            sdk.NewCoin("foo", osmomath.NewInt(100000)),
			tokenOutMinAmount:  osmomath.NewInt(1),
		},
		{
			name: "Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAZ, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount)),   // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: UOSMO,
				},
				{
					PoolId:        2,
					TokenOutDenom: BAZ,
				},
				{
					PoolId:        3,
					TokenOutDenom: BAR,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:            sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:  osmomath.NewInt(1),
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: BAR,
				},
				{
					PoolId:        2,
					TokenOutDenom: BAZ,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:            sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:  osmomath.NewInt(1),
		},
		{
			name: "Three routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 3.                                                                                      // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: UOSMO,
				},
				{
					PoolId:        2,
					TokenOutDenom: BAZ,
				},
				{
					PoolId:        3,
					TokenOutDenom: BAR,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			tokenIn:            sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:  osmomath.NewInt(1),
		},
		{
			name: "[Concentrated] One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, apptesting.DefaultCoinAmount), sdk.NewCoin(BAR, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.Concentrated},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: BAR,
				},
			},
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount: osmomath.NewInt(1),
		},
		{
			name: "[Concentrated[ Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, apptesting.DefaultCoinAmount), sdk.NewCoin(UOSMO, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(BAZ, apptesting.DefaultCoinAmount), sdk.NewCoin(UOSMO, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(BAR, apptesting.DefaultCoinAmount), sdk.NewCoin(BAZ, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.Concentrated, types.Concentrated, types.Concentrated},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: UOSMO,
				},
				{
					PoolId:        2,
					TokenOutDenom: BAZ,
				},
				{
					PoolId:        3,
					TokenOutDenom: BAR,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:            sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:  osmomath.NewInt(1),
		},
		{
			name: "[Cosmwasm] One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, apptesting.DefaultCoinAmount), sdk.NewCoin(BAR, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm},
			poolSpreadFactor: []osmomath.Dec{osmomath.OneDec()},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: BAR,
				},
			},
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount: osmomath.NewInt(1),
		},
		{
			name: "[Cosmwasm -> Concentrated] One route: Swap - [foo -> bar] -> [bar -> baz], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, apptesting.DefaultCoinAmount), sdk.NewCoin(BAR, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm, types.Concentrated},
			poolSpreadFactor: []osmomath.Dec{osmomath.OneDec(), defaultPoolSpreadFactor},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: BAR,
				},
				{
					PoolId:        2,
					TokenOutDenom: BAZ,
				},
			},
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount: osmomath.NewInt(1),
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
				// calculate the swap as separate swaps
				expectedMultihopTokenOutAmount := s.calcOutGivenInAmountAsSeparatePoolSwaps(tc.routes, tc.tokenIn)

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
func (s *KeeperTestSuite) TestMultihopSwapExactAmountOut() {
	tests := []struct {
		name               string
		poolCoins          []sdk.Coins
		poolSpreadFactor   []osmomath.Dec
		poolType           []types.PoolType
		routes             []types.SwapAmountOutRoute
		incentivizedGauges []uint64
		tokenOut           sdk.Coin
		tokenInMaxAmount   osmomath.Int
		spreadFactor       osmomath.Dec
		expectError        bool
	}{
		{
			name:             "One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins:        []sdk.Coins{sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount))},
			poolType:         []types.PoolType{types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: BAR,
				},
			},
			tokenInMaxAmount: osmomath.NewInt(90000000),
			tokenOut:         sdk.NewCoin(FOO, defaultSwapAmount),
		},
		{
			name: "Two routes: Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
				{
					PoolId:       2,
					TokenInDenom: BAR,
				},
			},
			incentivizedGauges: []uint64{},

			tokenInMaxAmount: osmomath.NewInt(90000000),
			tokenOut:         sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), both pools 1 percent fee, sanity check no more half fee applied",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAZ, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
				{
					PoolId:       2,
					TokenInDenom: UOSMO,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:   osmomath.NewInt(90000000),
			tokenOut:           sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
		},
		{
			name: "Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAZ, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount)),   // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
				{
					PoolId:       2,
					TokenInDenom: UOSMO,
				},
				{
					PoolId:       3,
					TokenInDenom: BAZ,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:   osmomath.NewInt(90000000),
			tokenOut:           sdk.NewCoin(BAR, osmomath.NewInt(100000)),
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
				{
					PoolId:       2,
					TokenInDenom: BAR,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
			tokenOut:           sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
			tokenInMaxAmount:   osmomath.NewInt(90000000),
		},
		{
			name: "Three routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount),
					sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 3.                                                                                    // pool 3.
			},
			poolType:         []types.PoolType{types.Balancer, types.Balancer, types.Balancer},
			poolSpreadFactor: []osmomath.Dec{defaultPoolSpreadFactor, defaultPoolSpreadFactor, defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
				{
					PoolId:       2,
					TokenInDenom: UOSMO,
				},
				{
					PoolId:       3,
					TokenInDenom: BAZ,
				},
			},
			incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			tokenOut:           sdk.NewCoin(BAR, osmomath.NewInt(100000)),
			tokenInMaxAmount:   osmomath.NewInt(90000000),
		},
		{
			name: "[Cosmwasm] One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, apptesting.DefaultCoinAmount), sdk.NewCoin(BAR, apptesting.DefaultCoinAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm},
			poolSpreadFactor: []osmomath.Dec{osmomath.OneDec()},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
			},
			tokenOut:         sdk.NewCoin(BAR, osmomath.NewInt(100000)),
			tokenInMaxAmount: osmomath.NewInt(90000000),
		},
		{
			name: "[Cosmwasm -> Concentrated] One route: Swap - [foo -> bar] -> [bar -> baz], 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, apptesting.DefaultCoinAmount), sdk.NewCoin(BAR, apptesting.DefaultCoinAmount)),
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount)),
			},
			poolType:         []types.PoolType{types.CosmWasm, types.Concentrated},
			poolSpreadFactor: []osmomath.Dec{osmomath.OneDec(), defaultPoolSpreadFactor},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
				{
					PoolId:       2,
					TokenInDenom: BAR,
				},
			},
			tokenOut:         sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
			tokenInMaxAmount: osmomath.NewInt(90000000),
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
				// calculate the swap as separate swaps
				expectedMultihopTokenInAmount := s.calcInGivenOutAmountAsSeparateSwaps(tc.routes, tc.tokenOut)
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
		tokenOutMinAmount osmomath.Int
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
						TokenOutDenom: BAR,
					},
					{
						PoolId:        2,
						TokenOutDenom: BAZ,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: BAR,
					},
					{
						PoolId:        4,
						TokenOutDenom: BAZ,
					},
				},
				tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
				tokenOutMinAmount: osmomath.NewInt(1),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: UOSMO,
					},
					{
						PoolId:        2,
						TokenOutDenom: BAZ,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: UOSMO,
					},
					{
						PoolId:        4,
						TokenOutDenom: BAZ,
					},
				},
				tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
				tokenOutMinAmount: osmomath.NewInt(1),
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
						TokenOutDenom: BAR,
					},
					{
						PoolId:        2,
						TokenOutDenom: BAZ,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: BAR,
					},
					{
						PoolId:        4,
						TokenOutDenom: BAZ,
					},
				},
				tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
				tokenOutMinAmount: osmomath.NewInt(1),
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
						TokenOutDenom: BAR,
					},
					{
						PoolId:        2,
						TokenOutDenom: BAZ,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: BAR,
					},
					{
						PoolId:        4,
						TokenOutDenom: BAZ,
					},
				},
				tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(9000000000000000000)),
				tokenOutMinAmount: osmomath.NewInt(1),
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
		tokenInMaxAmount osmomath.Int
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
						TokenInDenom: FOO,
					},
					{
						PoolId:       2,
						TokenInDenom: BAR,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: FOO,
					},
					{
						PoolId:       4,
						TokenInDenom: BAR,
					},
				},
				tokenInMaxAmount: osmomath.NewInt(90000000),
				tokenOut:         sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: FOO,
					},
					{
						PoolId:       2,
						TokenInDenom: UOSMO,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: FOO,
					},
					{
						PoolId:       4,
						TokenInDenom: UOSMO,
					},
				},
				tokenInMaxAmount: osmomath.NewInt(90000000),
				tokenOut:         sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
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
						TokenInDenom: FOO,
					},
					{
						PoolId:       2,
						TokenInDenom: BAR,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: FOO,
					},
					{
						PoolId:       4,
						TokenInDenom: BAR,
					},
				},
				tokenInMaxAmount: osmomath.NewInt(90000000),
				tokenOut:         sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
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
						TokenInDenom: FOO,
					},
					{
						PoolId:       2,
						TokenInDenom: BAR,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: FOO,
					},
					{
						PoolId:       4,
						TokenInDenom: BAR,
					},
				},
				tokenInMaxAmount: osmomath.NewInt(90000000),
				tokenOut:         sdk.NewCoin(BAZ, osmomath.NewInt(9000000000000000000)),
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
	totalWeight := osmomath.NewInt(int64(len(incentivizedGauges)))
	for _, gauge := range incentivizedGauges {
		records = append(records, poolincentivestypes.DistrRecord{GaugeId: gauge, Weight: osmomath.OneInt()})
	}
	distInfo := poolincentivestypes.DistrInfo{
		TotalWeight: totalWeight,
		Records:     records,
	}
	s.App.PoolIncentivesKeeper.SetDistrInfo(s.Ctx, distInfo)
}

func (s *KeeperTestSuite) calcInGivenOutAmountAsSeparateSwaps(routes []types.SwapAmountOutRoute, tokenOut sdk.Coin) sdk.Coin {
	cacheCtx, _ := s.Ctx.CacheContext()
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

		tokenInAmt, err := swapModule.SwapExactAmountOut(cacheCtx, s.TestAccs[0], hopPool, hop.TokenInDenom, osmomath.NewInt(100000000), nextTokenOut, updatedPoolSpreadFactor)
		s.Require().NoError(err)

		tokenInCoin := sdk.NewCoin(hop.TokenInDenom, tokenInAmt)
		tokenInCoinAfterAddTakerFee, _ := poolmanager.CalcTakerFeeExactOut(tokenInCoin, takerFee)

		nextTokenOut = tokenInCoinAfterAddTakerFee
	}
	return nextTokenOut
}

// calcOutGivenInAmountAsSeparatePoolSwaps calculates the output amount of a series of swaps on PoolManager pools.
// If its GAMM pool functions directly to ensure the poolmanager functions route to the correct modules. It it's CL pool functions directly to ensure the
// poolmanager functions route to the correct modules.
func (s *KeeperTestSuite) calcOutGivenInAmountAsSeparatePoolSwaps(routes []types.SwapAmountInRoute, tokenIn sdk.Coin) sdk.Coin {
	cacheCtx, _ := s.Ctx.CacheContext()
	nextTokenIn := tokenIn
	for _, hop := range routes {
		swapModule, err := s.App.PoolManagerKeeper.GetPoolModule(cacheCtx, hop.PoolId)
		s.Require().NoError(err)

		pool, err := swapModule.GetPool(s.Ctx, hop.PoolId)
		s.Require().NoError(err)

		spreadFactor := pool.GetSpreadFactor(cacheCtx)

		takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(cacheCtx, nextTokenIn.Denom, hop.TokenOutDenom)
		s.Require().NoError(err)

		nextTokenInAfterSubTakerFee, _ := poolmanager.CalcTakerFeeExactIn(nextTokenIn, takerFee)

		// we then do individual swaps until we reach the end of the swap route
		tokenOut, err := swapModule.SwapExactAmountIn(cacheCtx, s.TestAccs[0], pool, nextTokenInAfterSubTakerFee, hop.TokenOutDenom, osmomath.OneInt(), spreadFactor)
		s.Require().NoError(err)

		nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)

	}
	return nextTokenIn
}

// TODO: abstract SwapAgainstBalancerPool and SwapAgainstConcentratedPool
func (s *KeeperTestSuite) TestSingleSwapExactAmountIn() {
	tests := []struct {
		name                   string
		poolId                 uint64
		poolCoins              sdk.Coins
		poolFee                osmomath.Dec
		takerFee               osmomath.Dec
		tokenIn                sdk.Coin
		tokenOutDenom          string
		tokenOutMinAmount      osmomath.Int
		expectedTokenOutAmount osmomath.Int
		swapWithNoTakerFee     bool
		expectError            bool
	}{
		// Swap with taker fee:
		//  - foo: 1000000000000
		//  - bar: 1000000000000
		//  - spreadFactor: 0.1%
		//  - takerFee: 0.25%
		//  - foo in: 100000
		//  - bar amount out will be calculated according to the formula
		// 		https://www.wolframalpha.com/input?i=solve+%2810%5E12+%2B+10%5E5+x+0.9975%29%2810%5E12+-+x%29+%3D+10%5E24
		{
			name:                   "Swap - [foo -> bar], 0.1 percent swap fee, 0.25 percent taker fee",
			poolId:                 1,
			poolCoins:              sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)),
			poolFee:                defaultPoolSpreadFactor,
			takerFee:               osmomath.MustNewDecFromStr("0.0025"), // 0.25%
			tokenIn:                sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:      osmomath.NewInt(1),
			tokenOutDenom:          BAR,
			expectedTokenOutAmount: osmomath.NewInt(99650), // 10000 - 0.35%
		},
		// Swap with taker fee:
		//  - foo: 1000000000000
		//  - bar: 1000000000000
		//  - spreadFactor: 0.1%
		//  - takerFee: 0.25%
		//  - bar in: 100000
		//  - foo amount out: 10000 - 0.35%
		{
			name:                   "Swap - [foo -> bar], 0.1 percent swap fee, 0.25 percent taker fee",
			poolId:                 1,
			poolCoins:              sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)),
			poolFee:                defaultPoolSpreadFactor,
			takerFee:               osmomath.MustNewDecFromStr("0.0025"), // 0.25%
			tokenIn:                sdk.NewCoin(BAR, osmomath.NewInt(100000)),
			tokenOutMinAmount:      osmomath.NewInt(1),
			tokenOutDenom:          FOO,
			expectedTokenOutAmount: osmomath.NewInt(99650), // 100000 - 0.35%
		},
		{
			name:      "Swap - [foo -> bar], 0.1 percent swap fee, 0.33 percent taker fee",
			poolId:    1,
			poolCoins: sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(2000000000000)), sdk.NewCoin(BAR, osmomath.NewInt(1000000000000))),
			poolFee:   defaultPoolSpreadFactor,
			takerFee:  osmomath.MustNewDecFromStr("0.0033"), // 0.33%
			// We capture the expected fees in the input token to simplify the process for calculating output.
			// 100000 / (1 - 0.0043) = 100432
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100432)),
			tokenOutMinAmount: osmomath.NewInt(1),
			tokenOutDenom:     BAR,
			// Since spot price is 2 and the input after fees is 100000, the output should be 50000
			// minus one due to truncation on rounding error.
			expectedTokenOutAmount: osmomath.NewInt(50000 - 1),
		},
		{
			name:                   "Swap - [foo -> bar], 0 percent swap fee, 0 percent taker fee",
			poolId:                 1,
			poolCoins:              sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(1000000000000)), sdk.NewCoin(BAR, osmomath.NewInt(1000000000000))),
			poolFee:                osmomath.ZeroDec(),
			takerFee:               osmomath.ZeroDec(),
			tokenIn:                sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:      osmomath.NewInt(1),
			tokenOutDenom:          BAR,
			expectedTokenOutAmount: osmomath.NewInt(100000 - 1),
		},
		// 99% taker fee 99% swap fee
		{
			name:              "Swap - [foo -> bar], 99 percent swap fee, 99 percent taker fee",
			poolId:            1,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(1000000000000)), sdk.NewCoin(BAR, osmomath.NewInt(1000000000000))),
			poolFee:           osmomath.MustNewDecFromStr("0.99"),
			takerFee:          osmomath.MustNewDecFromStr("0.99"),
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(10000)),
			tokenOutMinAmount: osmomath.NewInt(1),
			tokenOutDenom:     BAR,
			// 10000 * 0.01 * 0.01 = 1 swapped at a spot price of 1
			expectedTokenOutAmount: osmomath.NewInt(1),
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
			poolCoins:              sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)),
			poolFee:                defaultPoolSpreadFactor,
			tokenIn:                sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount:      osmomath.NewInt(1),
			tokenOutDenom:          BAR,
			swapWithNoTakerFee:     true,
			expectedTokenOutAmount: osmomath.NewInt(99899),
		},
		{
			name:              "Wrong pool id",
			poolId:            2,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)),
			poolFee:           defaultPoolSpreadFactor,
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount: osmomath.NewInt(1),
			tokenOutDenom:     BAR,
			expectError:       true,
		},
		{
			name:              "In denom not exist",
			poolId:            1,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)),
			poolFee:           defaultPoolSpreadFactor,
			tokenIn:           sdk.NewCoin(BAZ, osmomath.NewInt(100000)),
			tokenOutMinAmount: osmomath.NewInt(1),
			tokenOutDenom:     BAR,
			expectError:       true,
		},
		{
			name:              "Out denom not exist",
			poolId:            1,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(FOO, defaultInitPoolAmount), sdk.NewCoin(BAR, defaultInitPoolAmount)),
			poolFee:           defaultPoolSpreadFactor,
			tokenIn:           sdk.NewCoin(FOO, osmomath.NewInt(100000)),
			tokenOutMinAmount: osmomath.NewInt(1),
			tokenOutDenom:     BAZ,
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
				ExitFee: osmomath.ZeroDec(),
			})

			var expectedTakerFeeCharged sdk.Coin
			if !tc.takerFee.IsNil() {
				_, expectedTakerFeeCharged = poolmanager.CalcTakerFeeExactIn(tc.tokenIn, tc.takerFee)
			}

			// execute the swap
			var multihopTokenOutAmount osmomath.Int
			var takerFeeCharged sdk.Coin
			var err error
			// TODO: move the denom pair set out and only run SwapExactAmountIn.
			// SwapExactAmountInNoTakerFee should be in a different test.
			if (tc.takerFee != osmomath.Dec{}) {
				// If applicable, set taker fee. Note that denoms are reordered lexicographically before being stored.
				poolmanagerKeeper.SetDenomPairTakerFee(s.Ctx, tc.poolCoins[0].Denom, tc.poolCoins[1].Denom, tc.takerFee)
				poolmanagerKeeper.SetDenomPairTakerFee(s.Ctx, tc.poolCoins[1].Denom, tc.poolCoins[0].Denom, tc.takerFee)

				multihopTokenOutAmount, takerFeeCharged, err = poolmanagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], tc.poolId, tc.tokenIn, tc.tokenOutDenom, tc.tokenOutMinAmount)
			} else {
				multihopTokenOutAmount, err = poolmanagerKeeper.SwapExactAmountInNoTakerFee(s.Ctx, s.TestAccs[0], tc.poolId, tc.tokenIn, tc.tokenOutDenom, tc.tokenOutMinAmount)
			}
			if tc.expectError {
				s.Require().Error(err)
			} else {
				// compare the expected tokenOut to the actual tokenOut
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedTokenOutAmount.String(), multihopTokenOutAmount.String())
				s.Require().Equal(expectedTakerFeeCharged, takerFeeCharged)
			}
		})
	}
}

func (s *KeeperTestSuite) TestEstimateTradeBasedOnPriceImpact() {
	poolId := uint64(1)
	maxPriceImpact := osmomath.MustNewDecFromStr("0.01")        // 1%
	maxPriceImpactHalved := osmomath.MustNewDecFromStr("0.005") // 0.5%
	maxPriceImpactTiny := osmomath.MustNewDecFromStr("0.0005")  // 0.05%

	externalPriceOneBalancer := osmomath.MustNewDecFromStr("0.666666667")            // Spot Price
	externalPriceOneBalancerInv := math.LegacyOneDec().Quo(externalPriceOneBalancer) // Inverse of externalPriceOneBalancer
	externalPriceTwoBalancer := osmomath.MustNewDecFromStr("0.622222222")            // Cheaper than spot price
	externalPriceThreeBalancer := osmomath.MustNewDecFromStr("0.663349917")          // Transform adjusted max price impact by 50%

	externalPriceOneStableSwap := osmomath.MustNewDecFromStr("1.00000002")             // Spot Price
	externalPriceTwoStableSwap := osmomath.MustNewDecFromStr("0.98989903")             // Cheaper than spot price
	externalPriceThreeStableSwap := osmomath.MustNewDecFromStr("0.990589420505200594") // Transform adjusted max price impact by a %

	externalPriceOneConcentrated := osmomath.MustNewDecFromStr("0.0002")                     // Same as spot price 1/5000.000000000000000129
	externalPriceOneConcentratedInv := osmomath.MustNewDecFromStr("5000.000000000000000129") // Inverse of externalPriceOneConcentrated
	externalPriceTwoConcentrated := osmomath.MustNewDecFromStr("0.000198")                   // Cheaper than spot price
	externalPriceThreeConcentrated := osmomath.MustNewDecFromStr("0.000198118")

	assetBaz := "baz"
	assetBar := "bar"
	assetUsdc := "usdc"
	assetEth := "eth"

	clCoinsLiquid := sdk.NewCoins(
		sdk.NewCoin("eth", osmomath.NewInt(1000000)),
		sdk.NewCoin("usdc", osmomath.NewInt(5000000000)),
	)
	clCoinsNotLiquid := sdk.NewCoins(
		sdk.NewCoin("eth", osmomath.NewInt(1)),
		sdk.NewCoin("usdc", osmomath.NewInt(1)),
	)

	// The below values have been tested and hard coded by using the `CalcOutAmtGivenIn` as it is quite hard to
	// mathematically work these out.
	tests := map[string]struct {
		poolId               uint64
		preCreatePoolType    types.PoolType
		setPositionForCLPool bool
		setClTokens          sdk.Coins
		req                  queryproto.EstimateTradeBasedOnPriceImpactRequest
		expectedInputCoin    sdk.Coin
		expectedOutputCoin   sdk.Coin
		expectError          string
	}{
		"valid balancer pool - first estimate works": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneBalancer,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(44_664)),
		},
		"valid balancer pool - multiple estimates work as first exceeds price impact": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(1_000_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneBalancer,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(39_947)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(59_327)),
		},
		"valid balancer pool - estimate trying to trade 1 token": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBar, osmomath.NewInt(1)),
				ToCoinDenom:    assetBaz,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneBalancerInv,
			},
			expectedInputCoin:  sdk.NewCoin(assetBar, osmomath.NewInt(0)),
			expectedOutputCoin: sdk.NewCoin(assetBaz, osmomath.NewInt(0)),
		},
		"valid balancer pool - estimate trying to trade dust": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(20)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneBalancer,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(0)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(0)),
		},
		"valid balancer pool - external price much greater than spot price do not trade": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceTwoBalancer,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(0)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(0)),
		},
		"valid balancer pool - adjusted price impact halved": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpactHalved,
				ExternalPrice:  externalPriceOneBalancer,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(19_936)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(29_755)),
		},
		"valid balancer pool - external price halves adjusted price impact": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceThreeBalancer,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(19_936)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(29_755)),
		},
		"valid balancer pool - adjusted price impact halved - external price not given": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpactHalved,
				ExternalPrice:  osmomath.ZeroDec(),
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(19_936)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(29_755)),
		},
		"valid balancer pool - adjusted price impact zero - external price not given": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.ZeroDec(),
				ExternalPrice:  osmomath.ZeroDec(),
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.ZeroInt()),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.ZeroInt()),
		},
		"valid balancer pool - adjusted price impact negative - external price not given": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.NewDec(-1),
				ExternalPrice:  osmomath.ZeroDec(),
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.ZeroInt()),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.ZeroInt()),
		},
		"valid balancer pool - adjusted price impact zero - external price given - price impact is negative": {
			preCreatePoolType: types.Balancer,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.ZeroDec(),
				ExternalPrice:  externalPriceThreeBalancer,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.ZeroInt()),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.ZeroInt()),
		},
		"valid stableswap pool - first estimate works": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(29_982)),
		},
		"valid stableswap pool - multiple estimates work as first exceeds price impact": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(1_000_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(497_617)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(492_690)),
		},
		"valid stableswap pool - multiple estimates work as first exceeds price impact - panics too large": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(1_000_000_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(497_666)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(492_739)),
		},
		"valid stableswap pool - estimate trying to trade 1 token": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(1)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(0)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(0)),
		},
		"valid stableswap pool - estimate trying to trade dust": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(20)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(0)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(0)),
		},
		"valid stableswap pool - external price value much greater than spot price do not trade": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceTwoStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(0)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(0)),
		},
		"valid stableswap pool - adjusted price impact tiny": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpactTiny,
				ExternalPrice:  externalPriceOneStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(24_501)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(24_488)),
		},
		"valid stableswap pool - external price changes adjusted price impact": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceThreeStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(24_501)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(24_488)),
		},
		"valid stableswap pool - adjusted price impact tiny - external price not given": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpactTiny,
				ExternalPrice:  osmomath.ZeroDec(),
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.NewInt(24_501)),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.NewInt(24_488)),
		},
		"valid stableswap pool - adjusted price impact zero - external price not given": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.ZeroDec(),
				ExternalPrice:  osmomath.ZeroDec(),
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.ZeroInt()),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.ZeroInt()),
		},
		"valid stableswap pool - adjusted price impact negative - external price not given": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.NewDec(-1),
				ExternalPrice:  osmomath.ZeroDec(),
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.ZeroInt()),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.ZeroInt()),
		},
		"valid stableswap pool - adjusted price impact zero - external price given - price impact is negative": {
			preCreatePoolType: types.Stableswap,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetBar,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.ZeroDec(),
				ExternalPrice:  externalPriceThreeStableSwap,
			},
			expectedInputCoin:  sdk.NewCoin(assetBaz, osmomath.ZeroInt()),
			expectedOutputCoin: sdk.NewCoin(assetBar, osmomath.ZeroInt()),
		},
		"valid concentrated pool - first estimate works": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(10)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneConcentrated,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(10)),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.NewInt(49_999)),
		},
		"valid concentrated pool - multiple estimates work as first exceeds price impact": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(1_000_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneConcentrated,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(214_661)),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.NewInt(1_062_678_216)),
		},
		"valid concentrated pool - estimate trying to trade 1 token": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetUsdc, osmomath.NewInt(1)),
				ToCoinDenom:    assetEth,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneConcentratedInv,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetUsdc, osmomath.NewInt(0)),
			expectedOutputCoin:   sdk.NewCoin(assetEth, osmomath.NewInt(0)),
		},
		"valid concentrated pool - estimate trying to trade dust": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetUsdc, osmomath.NewInt(20)),
				ToCoinDenom:    assetEth,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneConcentratedInv,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetUsdc, osmomath.NewInt(0)),
			expectedOutputCoin:   sdk.NewCoin(assetEth, osmomath.NewInt(0)),
		},
		"valid concentrated pool - estimate trying to trade one unit": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetUsdc, math.OneInt()),
				ToCoinDenom:    assetEth,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneConcentratedInv,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetUsdc, osmomath.NewInt(0)),
			expectedOutputCoin:   sdk.NewCoin(assetEth, osmomath.NewInt(0)),
		},
		"valid concentrated pool - external price much greater than spot price do not trade": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceTwoConcentrated,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(0)),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.NewInt(0)),
		},
		"valid concentrated pool - adjusted price impact halved": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpactTiny,
				ExternalPrice:  externalPriceOneConcentrated,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(10_733)),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.NewInt(53_638_181)),
		},
		"valid concentrated pool - external price halves adjusted price impact": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceThreeConcentrated,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(10_746)),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.NewInt(53_703_116)),
		},
		"valid concentrated pool - adjusted price impact halved - external price not given": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpactTiny,
				ExternalPrice:  osmomath.ZeroDec(),
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(10_733)),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.NewInt(53_638_181)),
		},
		"valid concentrated pool - adjusted price impact zero - external price not given": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.ZeroDec(),
				ExternalPrice:  osmomath.ZeroDec(),
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.ZeroInt()),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.ZeroInt()),
		},
		"valid concentrated pool - adjusted price impact negative - external price not given": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.NewDec(-1),
				ExternalPrice:  osmomath.ZeroDec(),
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.ZeroInt()),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.ZeroInt()),
		},
		"valid concentrated pool - adjusted price impact zero - external price given - price impact negative": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: osmomath.ZeroDec(),
				ExternalPrice:  externalPriceThreeConcentrated,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.ZeroInt()),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.ZeroInt()),
		},
		"valid concentrated pool - liquidity too low token out estimation is 0": {
			preCreatePoolType: types.Concentrated,
			poolId:            poolId,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceOneConcentrated,
			},
			setPositionForCLPool: true,
			setClTokens:          clCoinsNotLiquid,
			expectedInputCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(0)),
			expectedOutputCoin:   sdk.NewCoin(assetUsdc, osmomath.NewInt(0)),
		},
		"Invalid Pool ID": {
			preCreatePoolType: types.Balancer,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				PoolId: 0,
			},
			expectError: "Invalid Pool Id",
		},
		"Pool Does not exist": {
			preCreatePoolType: types.Balancer,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				PoolId:      2,
				FromCoin:    sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom: assetUsdc,
			},
			expectError: "failed to find route for pool id (2)",
		},
		"Invalid From Coin Denom": {
			preCreatePoolType: types.Balancer,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				PoolId: 1,
			},
			expectError: "invalid from coin denom",
		},
		"Invalid To Coin Denom": {
			preCreatePoolType: types.Balancer,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				PoolId:   1,
				FromCoin: sdk.NewCoin(assetBaz, osmomath.NewInt(100)),
			},
			expectError: "invalid to coin denom",
		},
		"valid concentrated liquidity pool without position": {
			preCreatePoolType: types.Concentrated,
			poolId:            1,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				FromCoin:       sdk.NewCoin(assetEth, osmomath.NewInt(30_000)),
				ToCoinDenom:    assetUsdc,
				PoolId:         poolId,
				MaxPriceImpact: maxPriceImpact,
				ExternalPrice:  externalPriceThreeConcentrated,
			},
			expectError: "error getting spot price for pool (1), no liquidity in pool",
		},
		"from coin token does not exist": {
			preCreatePoolType: types.Balancer,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				PoolId:      1,
				FromCoin:    sdk.NewCoin("random", osmomath.NewInt(30_000)),
				ToCoinDenom: assetBar,
			},
			expectError: "(random) does not exist in the pool",
		},
		"to coin token does not exist": {
			preCreatePoolType: types.Balancer,
			req: queryproto.EstimateTradeBasedOnPriceImpactRequest{
				PoolId:      1,
				FromCoin:    sdk.NewCoin(assetBaz, osmomath.NewInt(30_000)),
				ToCoinDenom: "random",
			},
			expectError: "(random) does not exist in the pool",
		},
	}
	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper
			poolmanagerQuerier := client.NewQuerier(poolmanagerKeeper)

			s.CreatePoolFromType(tc.preCreatePoolType)

			// we manually set position for CL to set spot price to correct value
			if tc.setPositionForCLPool {

				s.FundAcc(s.TestAccs[0], tc.setClTokens)

				clMsgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)
				_, err := clMsgServer.CreatePosition(s.Ctx, &cltypes.MsgCreatePosition{
					PoolId:          1,
					Sender:          s.TestAccs[0].String(),
					LowerTick:       int64(30545000),
					UpperTick:       int64(31500000),
					TokensProvided:  tc.setClTokens,
					TokenMinAmount0: osmomath.ZeroInt(),
					TokenMinAmount1: osmomath.ZeroInt(),
				})
				s.Require().NoError(err)
			}

			resp, err := poolmanagerQuerier.EstimateTradeBasedOnPriceImpact(s.Ctx, tc.req)
			if len(tc.expectError) > 0 {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectError)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedInputCoin, resp.InputCoin)
			s.Require().Equal(tc.expectedOutputCoin, resp.OutputCoin)
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
func (s *KeeperTestSuite) setupPools(poolType types.PoolType, poolDefaultSpreadFactor osmomath.Dec) (firstEstimatePoolId, secondEstimatePoolId uint64) {
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
			ExitFee: osmomath.NewDec(0),
		})
		s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSpreadFactor,
			ExitFee: osmomath.NewDec(0),
		})

		firstEstimatePoolId = s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSpreadFactor, // 1%
			ExitFee: osmomath.NewDec(0),
		})

		secondEstimatePoolId = s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSpreadFactor,
			ExitFee: osmomath.NewDec(0),
		})
		return
	}
}

// withInput takes in a types.SwapAmountInSplitRoute and returns it with the given input amount as TokenInAmount
func withInputSwapIn(route types.SwapAmountInSplitRoute, input osmomath.Int) types.SwapAmountInSplitRoute {
	route.TokenInAmount = input
	return route
}

// withInput takes in a types.SwapAmountOutSplitRoute and returns it with the given input amount as TokenInAmount
func withInputSwapOut(route types.SwapAmountOutSplitRoute, input osmomath.Int) types.SwapAmountOutSplitRoute {
	route.TokenOutAmount = input
	return route
}

// TestSplitRouteExactAmountIn tests the splitRouteExactAmountIn function.
func (s *KeeperTestSuite) TestSplitRouteExactAmountIn() {
	var (
		// Note: all pools have the default amount of 10_000_000_000 in each asset,
		// meaning in their initial state they have a spot price of 1.
		defaultSingleRouteOneHop = []types.SwapAmountInSplitRoute{
			{
				Pools: []types.SwapAmountInRoute{
					{
						PoolId:        fooBarPoolId,
						TokenOutDenom: BAR,
					},
				},
				TokenInAmount: twentyFiveBaseUnitsAmount,
			},
		}

		defaultTwoHopRoutes = []types.SwapAmountInRoute{
			{
				PoolId:        fooBarPoolId,
				TokenOutDenom: BAR,
			},
			{
				PoolId:        barBazPoolId,
				TokenOutDenom: BAZ,
			},
		}

		defaultSingleRouteTwoHops = types.SwapAmountInSplitRoute{
			Pools:         defaultTwoHopRoutes,
			TokenInAmount: twentyFiveBaseUnitsAmount,
		}

		fooAbcBazTwoHops = types.SwapAmountInSplitRoute{
			Pools: []types.SwapAmountInRoute{
				{
					PoolId:        fooAbcPoolId,
					TokenOutDenom: abc,
				},
				{
					PoolId:        bazAbcPoolId,
					TokenOutDenom: BAZ,
				},
			},
			TokenInAmount: twentyFiveBaseUnitsAmount,
		}

		defaultSingleRouteThreeHops = types.SwapAmountInSplitRoute{
			Pools: []types.SwapAmountInRoute{
				{
					PoolId:        fooBarPoolId,
					TokenOutDenom: BAR,
				},
				{
					PoolId:        barUosmoPoolId,
					TokenOutDenom: UOSMO,
				},
				{
					PoolId:        bazUosmoPoolId,
					TokenOutDenom: BAZ,
				},
			},
			TokenInAmount: osmomath.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
		}

		priceImpactThreshold = osmomath.NewInt(97866545)
	)

	tests := map[string]struct {
		isInvalidSender   bool
		setupPools        []poolSetup
		routes            []types.SwapAmountInSplitRoute
		tokenInDenom      string
		tokenOutMinAmount osmomath.Int

		// This value was taken from the actual result
		// and not manually calculated. This is acceptable
		// for this test because we are not testing the math
		// but the routing logic.
		// The math should be tested per-module.
		// We keep this assertion to make sure that the
		// actual result is within a reasonable range.
		expectedTokenOutEstimate osmomath.Int
		expectedTakerFees        expectedTakerFees
		checkExactOutput         bool

		expectError error
	}{
		"valid solo route one hop": {
			routes:            defaultSingleRouteOneHop,
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
			expectedTakerFees:        zeroTakerFeeDistr,
		},
		"valid solo route one hop (exact output check)": {
			// Set the pool we swap through to have a 0.3% taker fee
			setupPools: s.withTakerFees(defaultValidPools, []uint64{0}, []osmomath.Dec{pointThreePercent}),

			routes: []types.SwapAmountInSplitRoute{
				withInputSwapIn(defaultSingleRouteOneHop[0], osmomath.NewInt(1000)),
			},
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			// We expect the output to truncate
			expectedTokenOutEstimate: osmomath.NewInt(996),
			checkExactOutput:         true,

			// We expect total taker fees to be 3foo (0.3% of 1000foo)
			// Of this, 67% goes towards community pool and 33% towards stakers.
			// The community pool allocation is truncated in favor of staking reward allocation.
			expectedTakerFees: expectedTakerFees{
				// Since foo is a quote denom, we expect the full community pool allocation
				// to be in the quote asset address.
				// That being said, since 33% of 3 is technically less than 1, we expect the
				// community pool allocation to be truncated to zero, leaving the full amount in
				// staking rewards.
				communityPoolQuoteAssets:    sdk.NewCoins(),
				communityPoolNonQuoteAssets: sdk.NewCoins(),
				stakingRewardAssets:         sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(3))),
			},
		},
		"valid solo route multi hop": {
			routes:            []types.SwapAmountInSplitRoute{defaultSingleRouteTwoHops},
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
			expectedTakerFees:        zeroTakerFeeDistr,
		},
		"valid split route multi hop": {
			routes: []types.SwapAmountInSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			// 1x from single route two hops and 3x from single route three hops
			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount.MulRaw(4),
			expectedTakerFees:        zeroTakerFeeDistr,
		},
		"valid split route multi hop (exact output check)": {
			// Set the pools we swap through to all have a 0.35% taker fee
			setupPools: s.withTakerFees(
				defaultValidPools,
				[]uint64{0, 3, 6, 7},
				[]osmomath.Dec{pointThreeFivePercent, pointThreeFivePercent, pointThreeFivePercent, pointThreeFivePercent},
			),
			routes: []types.SwapAmountInSplitRoute{
				withInputSwapIn(defaultSingleRouteTwoHops, osmomath.NewInt(1000)),
				withInputSwapIn(fooAbcBazTwoHops, osmomath.NewInt(1000)),
			},
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			// We charge taker fee on each hop and expect the output to truncate
			// The output of first hop: (1 - 0.0035) * 1000 = 996.5, truncated to 996 (4foo taker fee)
			// which has an additional truncation after the swap executes, leaving 995.
			// The second hop is similar: (1 - 0.0035) * 995 = 991.5, truncated to 991 (4bar taker fee)
			// which has an additional truncation after the swap executes, leaving 990.
			expectedTokenOutEstimate: osmomath.NewInt(990 + 990),
			checkExactOutput:         true,

			// Expected taker fees from calculation above are:
			// * [4foo, 4bar] for first route
			// * [4foo, 4abc] for second route
			// Due to truncation, 4 units of fees get distributed 1 to community pool, 3 to stakers.
			// Recall that foo and bar are quote assets, while abc is not.
			expectedTakerFees: expectedTakerFees{
				// 1foo & 1bar from first route, 1foo from second route
				// Total: 2foo, 1bar
				communityPoolQuoteAssets: sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(2)), sdk.NewCoin(BAR, osmomath.NewInt(1))),
				// 1abc from second route
				// Total: 1abc
				communityPoolNonQuoteAssets: sdk.NewCoins(sdk.NewCoin(abc, osmomath.NewInt(1))),
				// 3foo & 3bar from first route, 3foo & 3abc from second route
				// Total: 6foo, 3bar, 3abc
				stakingRewardAssets: sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(6)), sdk.NewCoin(BAR, osmomath.NewInt(3)), sdk.NewCoin(abc, osmomath.NewInt(3))),
			},
		},
		"split route multi hop with different taker fees (exact output check)": {
			// Set the pools we swap through to all have varying taker fees
			setupPools: s.withTakerFees(
				defaultValidPools,
				[]uint64{0, 3, 6, 7},
				[]osmomath.Dec{pointThreeFivePercent, osmomath.ZeroDec(), pointOneFivePercent, pointThreePercent},
			),
			routes: []types.SwapAmountInSplitRoute{
				withInputSwapIn(defaultSingleRouteTwoHops, osmomath.NewInt(1000)),
				withInputSwapIn(fooAbcBazTwoHops, osmomath.NewInt(1000)),
			},
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			// Route 1:
			// 	Hop 1: (1 - 0.0035) * 1000 = 996.5, truncated to 996. Post swap truncated to 995.
			//    * 4foo taker fee
			// 	Hop 2: 995 input with no fee = 995 output since spot price is 1. Post swap truncated to 994.
			//    * No taker fee
			//  Route 1 output: 994 (Taker fees: 4foo)
			// Route 2:
			// 	Hop 1: (1 - 0.0015) * 1000 = 998.5, truncated to 998. Post swap truncated to 997.
			//    * 2foo taker fee
			// 	Hop 2: (1 - 0.003) * 997 = 994.009, truncated to 994. Post swap truncated to 993.
			//    * 3abc taker fee
			//  Route 2 output: 993 (Taker fees: 2foo, 3abc)
			expectedTokenOutEstimate: osmomath.NewInt(994 + 993),
			checkExactOutput:         true,

			// Expected taker fees from calculation above are:
			// * [4foo] for first route
			// * [2foo, 3abc] for second route
			// Due to truncation, 4 units of fees get distributed 1 to community pool, 3 to stakers.
			// Anything less than 3 units goes full to stakers.
			// Recall that foo and bar are quote assets, while abc is not.
			expectedTakerFees: expectedTakerFees{
				// 1foo from first route, everything from second route is truncated
				communityPoolQuoteAssets:    sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(1))),
				communityPoolNonQuoteAssets: sdk.NewCoins(),
				// 3foo from first route, 2foo & 3abc from second route
				// Total: 5foo, 3abc
				stakingRewardAssets: sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(5)), sdk.NewCoin(abc, osmomath.NewInt(3))),
			},
		},
		"valid split route multi hop with price impact protection that would fail individual route if given per multihop": {
			routes: []types.SwapAmountInSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenInDenom: FOO,
			// equal to the expected amount
			// every route individually would fail, but the split route should succeed
			tokenOutMinAmount: priceImpactThreshold,

			expectedTokenOutEstimate: priceImpactThreshold,
			expectedTakerFees:        zeroTakerFeeDistr,
		},

		"error: price impact protection triggered": {
			routes: []types.SwapAmountInSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenInDenom: FOO,
			// one greater than expected amount
			tokenOutMinAmount: priceImpactThreshold.Add(osmomath.OneInt()),

			expectError: types.PriceImpactProtectionExactInError{Actual: priceImpactThreshold, MinAmount: priceImpactThreshold.Add(osmomath.OneInt())},
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
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			expectError: types.ErrDuplicateRoutesNotAllowed,
		},

		"error: invalid pool id": {
			routes: []types.SwapAmountInSplitRoute{
				{
					Pools: []types.SwapAmountInRoute{
						{
							PoolId:        uint64(len(defaultValidPools) + 1),
							TokenOutDenom: BAR,
						},
					},
					TokenInAmount: twentyFiveBaseUnitsAmount,
				},
			},
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

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
			ak := s.App.AccountKeeper
			bk := s.App.BankKeeper

			sender := s.TestAccs[1]

			setupPools := defaultValidPools
			if tc.setupPools != nil {
				setupPools = tc.setupPools
			}

			for _, pool := range setupPools {
				s.CreatePoolFromTypeWithCoins(pool.poolType, pool.initialLiquidity)

				// Set taker fee for pool/pair
				k.SetDenomPairTakerFee(s.Ctx, pool.initialLiquidity[0].Denom, pool.initialLiquidity[1].Denom, pool.takerFee)
				k.SetDenomPairTakerFee(s.Ctx, pool.initialLiquidity[1].Denom, pool.initialLiquidity[0].Denom, pool.takerFee)

				// Fund sender with initial liquidity
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

			// Note, we use a 1% error tolerance with rounding down by default
			// because we initialize the reserves 1:1 so by performing
			// the swap we don't expect the price to change significantly.
			// As a result, we roughly expect the amount out to be the same
			// as the amount in given in another token. However, the actual
			// amount must be strictly less than the given due to price impact.
			multiplicativeTolerance := osmomath.OneDec()
			if tc.checkExactOutput {
				// We set to a small value instead of zero since zero is a special case
				// where multiplicative tolerance is skipped/not considered
				multiplicativeTolerance = osmomath.NewDecWithPrec(1, 8)
			}
			errTolerance := osmomath.ErrTolerance{
				RoundingDir:             osmomath.RoundDown,
				MultiplicativeTolerance: multiplicativeTolerance,
			}

			// Ensure output amount is within tolerance
			osmoassert.Equal(s.T(), errTolerance, tc.expectedTokenOutEstimate, tokenOut)

			// -- Ensure taker fee distributions have properly executed --

			// We expect all taker fees collected to be sent to the taker fee module account
			totalTakerFeesExpected := tc.expectedTakerFees.communityPoolQuoteAssets.Add(tc.expectedTakerFees.communityPoolNonQuoteAssets...).Add(tc.expectedTakerFees.stakingRewardAssets...)
			s.Require().Equal(totalTakerFeesExpected, sdk.NewCoins(bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(takerFeeAddrName))...))
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
						TokenInDenom: FOO,
					},
				},
				TokenOutAmount: twentyFiveBaseUnitsAmount,
			},
		}

		defaultTwoHopRoutes = []types.SwapAmountOutRoute{
			{
				PoolId:       fooBarPoolId,
				TokenInDenom: FOO,
			},
			{
				PoolId:       barBazPoolId,
				TokenInDenom: BAR,
			},
		}

		fooUosmoBazTwoHops = types.SwapAmountOutSplitRoute{
			Pools: []types.SwapAmountOutRoute{
				{
					PoolId:       fooUosmoPoolId,
					TokenInDenom: FOO,
				},
				{
					PoolId:       bazUosmoPoolId,
					TokenInDenom: UOSMO,
				},
			},
			TokenOutAmount: twentyFiveBaseUnitsAmount,
		}

		defaultSingleRouteTwoHops = types.SwapAmountOutSplitRoute{
			Pools:          defaultTwoHopRoutes,
			TokenOutAmount: twentyFiveBaseUnitsAmount,
		}

		defaultSingleRouteThreeHops = types.SwapAmountOutSplitRoute{
			Pools: []types.SwapAmountOutRoute{
				{
					PoolId:       fooBarPoolId,
					TokenInDenom: FOO,
				},
				{
					PoolId:       barUosmoPoolId,
					TokenInDenom: BAR,
				},
				{
					PoolId:       bazUosmoPoolId,
					TokenInDenom: UOSMO,
				},
			},
			TokenOutAmount: osmomath.NewInt(twentyFiveBaseUnitsAmount.Int64() * 3),
		}

		priceImpactThreshold = osmomath.NewInt(102239504)
	)

	tests := map[string]struct {
		setupPools       []poolSetup
		isInvalidSender  bool
		routes           []types.SwapAmountOutSplitRoute
		tokenOutDenom    string
		tokenInMaxAmount osmomath.Int

		// This value was taken from the actual result
		// and not manually calculated. This is acceptable
		// for this test because we are not testing the math
		// but the routing logic.
		// The math should be tested per-module.
		// We keep this assertion to make sure that the
		// actual result is within a reasonable range.
		expectedTokenInEstimate osmomath.Int
		checkExactOutput        bool

		expectError error
	}{
		"valid solo route one hop": {
			routes:           defaultSingleRouteOneHop,
			tokenOutDenom:    BAR,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			expectedTokenInEstimate: twentyFiveBaseUnitsAmount,
		},
		"valid solo route one hop (exact output check)": {
			// Set the pool we swap through to have a 0.3% taker fee
			setupPools: s.withTakerFees(defaultValidPools, []uint64{0}, []osmomath.Dec{pointThreePercent}),

			routes: []types.SwapAmountOutSplitRoute{
				withInputSwapOut(defaultSingleRouteOneHop[0], osmomath.NewInt(1000)),
			},
			tokenOutDenom:    BAR,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			// (1000 / (1 - 0.003)) = 1003.009, rounded up = 1004. Post swap rounded up to 1005.
			expectedTokenInEstimate: osmomath.NewInt(1005),
			checkExactOutput:        true,
		},
		"valid solo route multi hop": {
			routes:           []types.SwapAmountOutSplitRoute{defaultSingleRouteTwoHops},
			tokenOutDenom:    BAZ,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			expectedTokenInEstimate: twentyFiveBaseUnitsAmount,
		},
		"valid split route multi hop": {
			routes: []types.SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenOutDenom:    BAZ,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			// 1x from single route two hops and 3x from single route three hops
			expectedTokenInEstimate: twentyFiveBaseUnitsAmount.MulRaw(4),
		},
		"valid split route multi hop (exact output check)": {
			// Set the pools we swap through to all have a 0.35% taker fee
			setupPools: s.withTakerFees(
				defaultValidPools,
				[]uint64{0, 2, 3, 5},
				[]osmomath.Dec{pointThreeFivePercent, pointThreeFivePercent, pointThreeFivePercent, pointThreeFivePercent},
			),
			routes: []types.SwapAmountOutSplitRoute{
				withInputSwapOut(defaultSingleRouteTwoHops, osmomath.NewInt(1000)),
				withInputSwapOut(fooUosmoBazTwoHops, osmomath.NewInt(1000)),
			},
			tokenOutDenom:    BAZ,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			// We charge taker fee on each hop and expect the output to round up at each step
			// The output of first hop: (1000 / (1 - 0.0035)) = 1003.51, rounded up = 1004
			// which has an additional rounding up after the swap executes, leaving 1005.
			// The second hop is similar: (1005 / (1 - 0.0035)) = 1008.52, rounded up = 1009
			// which has an additional rounding up after the swap executes, leaving 1010.
			expectedTokenInEstimate: osmomath.NewInt(1010 + 1010),
			checkExactOutput:        true,
		},
		"split route multi hop with different taker fees (exact output check)": {
			// Set the pools we swap through to have the following taker fees:
			//  Route 1: 0.35% (hop 1) -> 0% (hop 2)
			//  Route 2: 0.15% (hop 1) -> 0.3% (hop 2)
			setupPools: s.withTakerFees(
				defaultValidPools,
				[]uint64{0, 2, 3, 5},
				[]osmomath.Dec{pointThreeFivePercent, osmomath.ZeroDec(), pointOneFivePercent, pointThreePercent},
			),
			routes: []types.SwapAmountOutSplitRoute{
				withInputSwapOut(defaultSingleRouteTwoHops, osmomath.NewInt(1000)),
				withInputSwapOut(fooUosmoBazTwoHops, osmomath.NewInt(1000)),
			},
			tokenOutDenom:    BAZ,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			// Route 1:
			// 	Hop 1: 1000 / (1 - 0.0035) = 1003.5, rounded up to 1004. Post swap rounded up to 1005.
			// 	Hop 2: 1005 output with no fee = 1005 input since spot price is 1. Post swap rounded up to 1006.
			//  Route 1 expected input: 1006
			// Route 2:
			// 	Hop 1: 1000 / (1 - 0.0015) = 1001.5, rounded up to 1002. Post swap rounded up to 1003.
			// 	Hop 2: 1003 / (1 - 0.003) = 1006.01, rounded up to 1007. Post swap rounded up to 1008.
			//  Route 2 expected input: 1008
			expectedTokenInEstimate: osmomath.NewInt(1006 + 1008),
			checkExactOutput:        true,
		},

		"valid split route multi hop with price impact protection that would fail individual route if given per multihop": {
			routes: []types.SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenOutDenom: BAZ,
			// equal to the amount calculated.
			// every route individually would fail, but the split route should succeed
			tokenInMaxAmount: priceImpactThreshold,

			expectedTokenInEstimate: priceImpactThreshold,
		},

		"error: price impact protection triggered": {
			routes: []types.SwapAmountOutSplitRoute{
				defaultSingleRouteTwoHops,
				defaultSingleRouteThreeHops,
			},
			tokenOutDenom: BAZ,
			// one less than expected amount
			// every route individually would fail, but the split route should succeed
			tokenInMaxAmount: priceImpactThreshold.Sub(osmomath.OneInt()),

			expectError: types.PriceImpactProtectionExactOutError{Actual: priceImpactThreshold, MaxAmount: priceImpactThreshold.Sub(osmomath.OneInt())},
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
			tokenOutDenom:    FOO,
			tokenInMaxAmount: poolmanager.IntMaxValue,

			expectError: types.ErrDuplicateRoutesNotAllowed,
		},

		"error: invalid pool id": {
			routes: []types.SwapAmountOutSplitRoute{
				{
					Pools: []types.SwapAmountOutRoute{
						{
							PoolId:       uint64(len(defaultValidPools) + 1),
							TokenInDenom: FOO,
						},
					},
					TokenOutAmount: twentyFiveBaseUnitsAmount,
				},
			},
			tokenOutDenom:    FOO,
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

			setupPools := defaultValidPools
			if tc.setupPools != nil {
				setupPools = tc.setupPools
			}

			for _, pool := range setupPools {
				s.CreatePoolFromTypeWithCoins(pool.poolType, pool.initialLiquidity)

				// Set taker fee for pool/pair
				k.SetDenomPairTakerFee(s.Ctx, pool.initialLiquidity[0].Denom, pool.initialLiquidity[1].Denom, pool.takerFee)
				k.SetDenomPairTakerFee(s.Ctx, pool.initialLiquidity[1].Denom, pool.initialLiquidity[0].Denom, pool.takerFee)

				// Fund sender with initial liquidity
				// If not valid, we don't fund to trigger an error case.
				if !tc.isInvalidSender {
					s.FundAcc(sender, pool.initialLiquidity)
				}
			}

			tokenIn, err := k.SplitRouteExactAmountOut(s.Ctx, sender, tc.routes, tc.tokenOutDenom, tc.tokenInMaxAmount)

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
			// amount must be strictly greater than the given due to price impact.
			multiplicativeTolerance := osmomath.OneDec()
			if tc.checkExactOutput {
				// We set to a small value instead of zero since zero is a special case
				// where multiplicative tolerance is skipped/not considered
				multiplicativeTolerance = osmomath.NewDecWithPrec(1, 8)
			}
			errTolerance := osmomath.ErrTolerance{
				RoundingDir:             osmomath.RoundUp,
				MultiplicativeTolerance: multiplicativeTolerance,
			}

			osmoassert.Equal(s.T(), errTolerance, tc.expectedTokenInEstimate, tokenIn)
		})
	}
}

func (s *KeeperTestSuite) TestGetTotalPoolLiquidity() {
	const (
		cosmWasmPoolId = uint64(3)
	)
	var (
		defaultPoolCoinOne = sdk.NewCoin("usdc", osmomath.OneInt())
		defaultPoolCoinTwo = sdk.NewCoin("eth", osmomath.NewInt(2))
		nonPoolCool        = sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(3))

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
			expectedResult: sdk.Coins{},
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
			name:        "round not found because pool id does not exist",
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
				s.CreatePoolFromTypeWithCoinsAndSpreadFactor(types.CosmWasm, tc.poolLiquidity, osmomath.ZeroDec())
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

func (suite *KeeperTestSuite) TestCreateMultihopExpectedSwapOuts() {
	tests := map[string]struct {
		route                  []types.SwapAmountOutRoute
		tokenOut               sdk.Coin
		balancerPoolCoins      []sdk.Coins
		concentratedPoolDenoms [][]string
		poolCoins              []sdk.Coins

		expectedSwapIns []osmomath.Int
		expectedError   bool
	}{
		"happy path: one route": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: BAR,
				},
			},
			poolCoins: []sdk.Coins{sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(100)), sdk.NewCoin(BAR, osmomath.NewInt(100)))},

			tokenOut: sdk.NewCoin(FOO, osmomath.NewInt(10)),
			// expectedSwapIns = (tokenOut * (poolTokenOutBalance / poolPostSwapOutBalance)).ceil()
			// foo token = 10 * (100 / 90) ~ 12
			expectedSwapIns: []osmomath.Int{osmomath.NewInt(12)},
		},
		"happy path: two route": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: FOO,
				},
				{
					PoolId:       2,
					TokenInDenom: BAR,
				},
			},

			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(100)), sdk.NewCoin(BAR, osmomath.NewInt(100))), // pool 1.
				sdk.NewCoins(sdk.NewCoin(BAR, osmomath.NewInt(100)), sdk.NewCoin(BAZ, osmomath.NewInt(100))), // pool 2.
			},
			tokenOut: sdk.NewCoin(BAZ, osmomath.NewInt(10)),
			// expectedSwapIns = (tokenOut * (poolTokenOutBalance / poolPostSwapOutBalance)).ceil()
			// foo token = 10 * (100 / 90) ~ 12
			// bar token = 12 * (100 / 88) ~ 14
			expectedSwapIns: []osmomath.Int{osmomath.NewInt(14), osmomath.NewInt(12)},
		},
		"error: Invalid Pool": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       100,
					TokenInDenom: FOO,
				},
			},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(100)), sdk.NewCoin(BAR, osmomath.NewInt(100))), // pool 1.
			},
			tokenOut:      sdk.NewCoin(BAZ, osmomath.NewInt(10)),
			expectedError: true,
		},
		"error: calculating in given out": {
			route: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: UOSMO,
				},
			},

			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(FOO, osmomath.NewInt(100)), sdk.NewCoin(BAR, osmomath.NewInt(100))), // pool 1.
			},
			tokenOut:        sdk.NewCoin(BAZ, osmomath.NewInt(10)),
			expectedSwapIns: []osmomath.Int{},

			expectedError: true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()

			suite.createBalancerPoolsFromCoins(tc.poolCoins)

			var actualSwapOuts []osmomath.Int
			var err error

			actualSwapOuts, err = suite.App.PoolManagerKeeper.CreateMultihopExpectedSwapOuts(suite.Ctx, tc.route, tc.tokenOut)
			if tc.expectedError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedSwapIns, actualSwapOuts)
			}
		})
	}
}

// runMultipleTrackVolumes runs TrackVolume on the same pool multiple times
func (s *KeeperTestSuite) runMultipleTrackVolumes(poolId uint64, volume sdk.Coin, times int64) {
	for i := 0; i < int(times); i++ {
		s.App.PoolManagerKeeper.TrackVolume(s.Ctx, poolId, volume)
	}
}

// Testing strategy:
// 1. If applicable, create an OSMO-paired pool
// 2. Set OSMO-paired pool as canonical for that denom pair in state
// 3. Run `trackVolume` on test input amount (cases include both OSMO and non-OSMO volumes)
// 4. Assert correct amount was added to pool volume
func (s *KeeperTestSuite) TestTrackVolume() {
	hundred := osmomath.NewInt(100)
	hundredFoo := sdk.NewCoin(FOO, hundred)
	hundredUosmo := sdk.NewCoin(UOSMO, hundred)
	oneRun := int64(1)
	threeRuns := int64(3)

	tests := map[string]struct {
		generatedVolume     sdk.Coin
		timesRun            int64
		osmoPairedPoolType  types.PoolType
		osmoPairedPoolCoins sdk.Coins

		expectedVolume osmomath.Int
	}{
		"Happy path: volume denominated in OSMO": {
			generatedVolume: hundredUosmo,
			timesRun:        oneRun,

			expectedVolume: hundred.MulRaw(oneRun),
		},

		// --- Pricing against balancer pool ---

		"Non-OSMO volume priced with balancer pool": {
			generatedVolume:    hundredFoo,
			timesRun:           oneRun,
			osmoPairedPoolType: types.Balancer,
			// Spot price = 1
			osmoPairedPoolCoins: fooUosmoCoins,

			expectedVolume: hundred.MulRaw(oneRun),
		},
		"Non-OSMO volume priced with balancer pool, multiple runs": {
			generatedVolume:    hundredFoo,
			timesRun:           threeRuns,
			osmoPairedPoolType: types.Balancer,
			// Spot price = 1
			osmoPairedPoolCoins: fooUosmoCoins,

			expectedVolume: hundred.MulRaw(threeRuns),
		},
		"Non-OSMO volume priced with balancer pool, large spot price": {
			generatedVolume:    hundredFoo,
			timesRun:           oneRun,
			osmoPairedPoolType: types.Balancer,
			// 100 foo corresponds to 1000 osmo (spot price = 10)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(1000)),
			),

			expectedVolume: osmomath.NewInt(10).Mul(hundred).MulRaw(oneRun),
		},
		"Non-OSMO volume priced with balancer pool, small spot price": {
			generatedVolume:    hundredFoo,
			timesRun:           oneRun,
			osmoPairedPoolType: types.Balancer,
			// 100 foo corresponds to 10 osmo (spot price = 0.1)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(10)),
			),

			expectedVolume: hundred.MulRaw(oneRun).Quo(osmomath.NewInt(10)),
		},
		"Non-OSMO volume priced with balancer pool, large spot price, multiple runs": {
			generatedVolume:    hundredFoo,
			timesRun:           threeRuns,
			osmoPairedPoolType: types.Balancer,
			// 100 foo corresponds to 1000 osmo (spot price = 10)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(1000)),
			),

			expectedVolume: osmomath.NewInt(10).Mul(hundred).MulRaw(threeRuns),
		},
		"Non-OSMO volume priced with balancer pool, small spot price, multiple runs": {
			generatedVolume:    hundredFoo,
			timesRun:           threeRuns,
			osmoPairedPoolType: types.Balancer,
			// 100 foo corresponds to 10 osmo (spot price = 0.1)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(10)),
			),

			expectedVolume: hundred.MulRaw(threeRuns).Quo(osmomath.NewInt(10)),
		},

		// --- Pricing against CL pool ---

		"Non-OSMO volume priced with concentrated pool, multiple runs": {
			generatedVolume:    hundredFoo,
			timesRun:           threeRuns,
			osmoPairedPoolType: types.Concentrated,
			// Spot price = 1
			osmoPairedPoolCoins: fooUosmoCoins,

			expectedVolume: hundred.MulRaw(threeRuns),
		},
		"Non-OSMO volume priced with concentrated pool, large spot price": {
			generatedVolume:    hundredFoo,
			timesRun:           oneRun,
			osmoPairedPoolType: types.Concentrated,
			// 100 foo corresponds to 1000 osmo (spot price = 10)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(1000)),
			),

			expectedVolume: osmomath.NewInt(10).Mul(hundred).MulRaw(oneRun),
		},
		"Non-OSMO volume priced with concentrated pool, small spot price": {
			generatedVolume:    hundredFoo,
			timesRun:           oneRun,
			osmoPairedPoolType: types.Concentrated,
			// 100 foo corresponds to 10 osmo (spot price = 0.1)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(10)),
			),

			expectedVolume: hundred.MulRaw(oneRun).Quo(osmomath.NewInt(10)),
		},
		"Non-OSMO volume priced with concentrated pool, large spot price, multiple runs": {
			generatedVolume:    hundredFoo,
			timesRun:           threeRuns,
			osmoPairedPoolType: types.Concentrated,
			// 100 foo corresponds to 1000 osmo (spot price = 10)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(1000)),
			),

			expectedVolume: osmomath.NewInt(10).Mul(hundred).MulRaw(threeRuns),
		},
		"Non-OSMO volume priced with concentrated pool, small spot price, multiple runs": {
			generatedVolume:    hundredFoo,
			timesRun:           threeRuns,
			osmoPairedPoolType: types.Concentrated,
			// 100 foo corresponds to 10 osmo (spot price = 0.1)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(10)),
			),

			expectedVolume: hundred.MulRaw(threeRuns).Quo(osmomath.NewInt(10)),
		},

		// --- Pricing against cosmwasm pool ---

		"Non-OSMO volume priced with CosmWasm pool, multiple runs": {
			generatedVolume:    hundredFoo,
			timesRun:           threeRuns,
			osmoPairedPoolType: types.CosmWasm,
			// Spot price = 1 since our test CW pool is a 1:1 transmuter
			osmoPairedPoolCoins: fooUosmoCoins,

			expectedVolume: hundred.MulRaw(threeRuns),
		},

		// --- Edge cases ---

		"OSMO denominated volume, no volume added": {
			generatedVolume:     sdk.NewCoin(UOSMO, osmomath.NewInt(0)),
			timesRun:            oneRun,
			osmoPairedPoolType:  types.Balancer,
			osmoPairedPoolCoins: fooUosmoCoins,

			expectedVolume: osmomath.NewInt(0),
		},
		"Non-OSMO volume priced with balancer pool, no volume added": {
			generatedVolume:     sdk.NewCoin(FOO, osmomath.NewInt(0)),
			timesRun:            oneRun,
			osmoPairedPoolType:  types.Balancer,
			osmoPairedPoolCoins: fooUosmoCoins,

			expectedVolume: osmomath.NewInt(0),
		},
		"Added volume truncated to zero (no volume added)": {
			generatedVolume:    sdk.NewCoin(FOO, osmomath.NewInt(1)),
			timesRun:           oneRun,
			osmoPairedPoolType: types.Balancer,
			// 100 foo corresponds to 10 osmo (spot price = 0.1)
			osmoPairedPoolCoins: sdk.NewCoins(
				sdk.NewCoin(FOO, osmomath.NewInt(100)),
				sdk.NewCoin(UOSMO, osmomath.NewInt(10)),
			),

			expectedVolume: osmomath.NewInt(0),
		},
		"Non-OSMO denominated volume, no OSMO-paired pool": {
			generatedVolume: hundredFoo,
			timesRun:        oneRun,

			// Note that we expect this to fail quietly and track zero volume,
			// as described in the function spec/comments.
			expectedVolume: osmomath.NewInt(0),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// --- Setup ---

			// Create target pool to track volume for.
			// Note that the actual contents or type of this pool do not matter for this test as we just need the ID.
			targetPoolId := s.PrepareBalancerPool()

			// If applicable, create an OSMO-paired pool and set it in protorev
			if tc.osmoPairedPoolCoins != nil {
				osmoPairedPoolId := s.CreatePoolFromTypeWithCoins(tc.osmoPairedPoolType, tc.osmoPairedPoolCoins)
				s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, FOO, osmoPairedPoolId)
			}

			// --- System under test ---

			// Run TrackVolume the specified number of times. Note that this function fails quietly in all error cases.
			s.runMultipleTrackVolumes(targetPoolId, tc.generatedVolume, tc.timesRun)

			// --- Assertions ---

			// Assert that the correct amount of volume was added to the pool tracker

			// Note that the units should always be in OSMO, even if the input volume was in another token.
			//
			// We wrap with sdk.NewCoins() to sanitize the outputs for comparison in case they are empty.
			totalVolume := s.App.PoolManagerKeeper.GetTotalVolumeForPool(s.Ctx, targetPoolId)
			s.Require().Equal(sdk.NewCoins(sdk.NewCoin(UOSMO, tc.expectedVolume)), sdk.NewCoins(totalVolume...))
		})
	}
}

// TestTakerFee tests starting from the swap that the taker fee is taken from and ends at the after epoch end hook,
// ensuring the resulting values are swapped as intended and sent to the correct destinations.
func (s *KeeperTestSuite) TestTakerFee() {
	var (
		nonZeroTakerFee = osmomath.MustNewDecFromStr("0.0015")
		// Note: all pools have the default amount of 10_000_000_000 in each asset,
		// meaning in their initial state they have a spot price of 1.
		quoteNativeDenomRoute = []types.SwapAmountInSplitRoute{
			{
				Pools: []types.SwapAmountInRoute{
					{
						PoolId:        fooUosmoPoolId,
						TokenOutDenom: FOO,
					},
				},
				TokenInAmount: twentyFiveBaseUnitsAmount,
			},
		}
		quoteQuoteDenomRoute = []types.SwapAmountInSplitRoute{
			{
				Pools: []types.SwapAmountInRoute{
					{
						PoolId:        fooBarPoolId,
						TokenOutDenom: BAR,
					},
				},
				TokenInAmount: twentyFiveBaseUnitsAmount,
			},
		}
		quoteNonquoteDenomRoute = []types.SwapAmountInSplitRoute{
			{
				Pools: []types.SwapAmountInRoute{
					{
						PoolId:        fooAbcPoolId,
						TokenOutDenom: FOO,
					},
				},
				TokenInAmount: twentyFiveBaseUnitsAmount,
			},
		}
		totalExpectedTakerFee = osmomath.NewDecFromInt(twentyFiveBaseUnitsAmount).Mul(nonZeroTakerFee)
		osmoTakerFeeDistr     = s.App.PoolManagerKeeper.GetParams(s.Ctx).TakerFeeParams.OsmoTakerFeeDistribution
		nonOsmoTakerFeeDistr  = s.App.PoolManagerKeeper.GetParams(s.Ctx).TakerFeeParams.NonOsmoTakerFeeDistribution
	)

	tests := map[string]struct {
		routes            []types.SwapAmountInSplitRoute
		tokenInDenom      string
		tokenOutMinAmount osmomath.Int

		expectedTokenOutEstimate                          osmomath.Int
		expectedTakerFees                                 expectedTakerFees
		expectedCommunityPoolBalancesDelta                sdk.Coins // actual community pool
		expectedStakingRewardFeeCollectorMainBalanceDelta sdk.Coins // where fees are staged prior to being distributed to stakers

		expectError error
	}{
		"native denom taker fee": {
			routes:            quoteNativeDenomRoute,
			tokenInDenom:      UOSMO,
			tokenOutMinAmount: osmomath.OneInt(),

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
			expectedTakerFees: expectedTakerFees{
				communityPoolQuoteAssets:    sdk.Coins{},
				communityPoolNonQuoteAssets: sdk.Coins{},
				stakingRewardAssets:         sdk.NewCoins(sdk.NewCoin(UOSMO, totalExpectedTakerFee.Mul(osmoTakerFeeDistr.StakingRewards).TruncateInt())),
			},
			// full native denom set in the main fee collector addr
			expectedStakingRewardFeeCollectorMainBalanceDelta: sdk.NewCoins(sdk.NewCoin(UOSMO, totalExpectedTakerFee.Mul(osmoTakerFeeDistr.StakingRewards).TruncateInt())),
			expectedCommunityPoolBalancesDelta:                sdk.Coins{},
		},
		"quote denom taker fee": {
			routes:            quoteQuoteDenomRoute,
			tokenInDenom:      FOO,
			tokenOutMinAmount: osmomath.OneInt(),

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
			expectedTakerFees: expectedTakerFees{
				communityPoolQuoteAssets:    sdk.NewCoins(sdk.NewCoin(FOO, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.CommunityPool).TruncateInt())),
				communityPoolNonQuoteAssets: sdk.Coins{},
				stakingRewardAssets:         sdk.NewCoins(sdk.NewCoin(FOO, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.StakingRewards).TruncateInt())),
			},
			// since foo is whitelisted token, it is sent directly to community pool
			expectedCommunityPoolBalancesDelta: sdk.NewCoins(sdk.NewCoin(FOO, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.CommunityPool).TruncateInt())),
			// foo swapped for uosmo, uosmo sent to main fee collector, 1 uosmo diff due to slippage from swap
			expectedStakingRewardFeeCollectorMainBalanceDelta: sdk.NewCoins(sdk.NewCoin(UOSMO, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.StakingRewards).Sub(osmomath.OneDec()).TruncateInt())),
		},
		"non quote denom taker fee": {
			routes:            quoteNonquoteDenomRoute,
			tokenInDenom:      abc,
			tokenOutMinAmount: osmomath.OneInt(),

			expectedTokenOutEstimate: twentyFiveBaseUnitsAmount,
			expectedTakerFees: expectedTakerFees{
				communityPoolQuoteAssets:    sdk.Coins{},
				communityPoolNonQuoteAssets: sdk.NewCoins(sdk.NewCoin(abc, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.CommunityPool).TruncateInt())),
				stakingRewardAssets:         sdk.NewCoins(sdk.NewCoin(abc, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.StakingRewards).TruncateInt())),
			},
			// since abc is not whitelisted token, it gets swapped for `CommunityPoolDenomToSwapNonWhitelistedAssetsTo`, which is set to baz, 1 baz diff due to slippage from swap
			expectedCommunityPoolBalancesDelta: sdk.NewCoins(sdk.NewCoin(BAZ, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.CommunityPool).Sub(osmomath.OneDec()).TruncateInt())),
			// abc swapped for uosmo, uosmo sent to main fee collector, 1 uosmo diff due to slippage from swap
			expectedStakingRewardFeeCollectorMainBalanceDelta: sdk.NewCoins(sdk.NewCoin(UOSMO, totalExpectedTakerFee.Mul(nonOsmoTakerFeeDistr.StakingRewards).Sub(osmomath.OneDec()).TruncateInt())),
		},
	}

	s.PrepareBalancerPool()
	s.PrepareConcentratedPool()

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			k := s.App.PoolManagerKeeper
			ak := s.App.AccountKeeper
			bk := s.App.BankKeeper

			sender := s.TestAccs[1]

			for _, pool := range defaultValidPools {
				poolId := s.CreatePoolFromTypeWithCoins(pool.poolType, pool.initialLiquidity)

				// Set taker fee for pool/pair
				k.SetDenomPairTakerFee(s.Ctx, pool.initialLiquidity[0].Denom, pool.initialLiquidity[1].Denom, nonZeroTakerFee)
				k.SetDenomPairTakerFee(s.Ctx, pool.initialLiquidity[1].Denom, pool.initialLiquidity[0].Denom, nonZeroTakerFee)

				// Set the denom pair as a pool route in protorev
				s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, pool.initialLiquidity[0].Denom, pool.initialLiquidity[1].Denom, poolId)

				// Fund sender with initial liqudity
				s.FundAcc(sender, pool.initialLiquidity)
			}

			// Log starting balances to compare against
			communityPoolBalancesPreHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(communityPoolAddrName))
			stakingRewardFeeCollectorMainBalancePreHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(authtypes.FeeCollectorName))
			stakingRewardFeeCollectorTxfeesBalancePreHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(txFeesStakingAddrName))
			takerFeeCollectorBalancePreHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(takerFeeAddrName))

			// Execute swap
			tokenOut, err := k.SplitRouteExactAmountIn(s.Ctx, sender, tc.routes, tc.tokenInDenom, tc.tokenOutMinAmount)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectError.Error())
				return
			}
			s.Require().NoError(err)

			// Note, we use a 1% error tolerance with rounding down by default
			// because we initialize the reserves 1:1 so by performing
			// the swap we don't expect the price to change significantly.
			// As a result, we roughly expect the amount out to be the same
			// as the amount in given in another token. However, the actual
			// amount must be strictly less than the given due to price impact.
			multiplicativeTolerance := osmomath.OneDec()
			errTolerance := osmomath.ErrTolerance{
				RoundingDir:             osmomath.RoundDown,
				MultiplicativeTolerance: multiplicativeTolerance,
			}

			// Ensure output amount is within tolerance
			osmoassert.Equal(s.T(), errTolerance, tc.expectedTokenOutEstimate, tokenOut)

			// -- Ensure taker fee distributions have properly executed --

			// We expect all taker fees collected to be sent directly the taker fee module account at time of swap.
			totalTakerFeesExpected := tc.expectedTakerFees.communityPoolQuoteAssets.Add(tc.expectedTakerFees.communityPoolNonQuoteAssets...).Add(tc.expectedTakerFees.stakingRewardAssets...)
			s.Require().Equal(totalTakerFeesExpected, sdk.NewCoins(bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(takerFeeAddrName))...))

			// Run the afterEpochEnd hook from txfees directly
			s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "day", 1)

			// Store balances after hook
			communityPoolBalancesPostHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(communityPoolAddrName))
			stakingRewardFeeCollectorMainBalancePostHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(authtypes.FeeCollectorName))
			stakingRewardFeeCollectorTxfeesBalancePostHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(txFeesStakingAddrName))
			takerFeeCollectorBalancePostHook := bk.GetAllBalances(s.Ctx, ak.GetModuleAddress(takerFeeAddrName))

			communityPoolBalancesDelta := communityPoolBalancesPostHook.Sub(communityPoolBalancesPreHook...)
			stakingRewardFeeCollectorMainBalanceDelta := stakingRewardFeeCollectorMainBalancePostHook.Sub(stakingRewardFeeCollectorMainBalancePreHook...)
			stakingRewardFeeCollectorTxfeesBalanceDelta := stakingRewardFeeCollectorTxfeesBalancePostHook.Sub(stakingRewardFeeCollectorTxfeesBalancePreHook...)
			takerFeeBalanceDelta := takerFeeCollectorBalancePostHook.Sub(takerFeeCollectorBalancePreHook...)

			// Ensure balances are as expected
			s.Require().Equal(tc.expectedCommunityPoolBalancesDelta, communityPoolBalancesDelta)
			s.Require().Equal(tc.expectedStakingRewardFeeCollectorMainBalanceDelta, stakingRewardFeeCollectorMainBalanceDelta)
			s.Require().Equal(sdk.Coins{}, stakingRewardFeeCollectorTxfeesBalanceDelta) // should always be empty after hook if all routes exist
			s.Require().Equal(sdk.Coins{}, takerFeeBalanceDelta)                        // should always be empty after hook if all routes exist
		})
	}
}

// This test validates that SwapExactAmountIn tracks volume correctly.
// It is a simple check to make sure that trackVolume() is called.
func (s *KeeperTestSuite) TestSwapExactAmountIn_VolumeTracked() {
	const withTakerFee = false

	s.Run("with taker fee", func() {
		s.testSwapExactAmountInVolumeTracked(withTakerFee)
	})

	s.Run("without taker fee", func() {
		s.testSwapExactAmountInVolumeTracked(!withTakerFee)
	})
}

// test for ensuring that volume is tracked by variants of swap exact amount in
func (s *KeeperTestSuite) testSwapExactAmountInVolumeTracked(noTakerFeeVariant bool) {
	s.SetupTest()

	// Set UOSMO as bond denom

	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	stakingParams.BondDenom = UOSMO
	s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)

	// Prepare pool with liquidity
	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], UOSMO, FOO, 1, osmomath.ZeroDec())
	s.CreateFullRangePosition(concentratedPool, sdk.NewCoins(sdk.NewCoin(UOSMO, osmomath.NewInt(1_000_000_000)), sdk.NewCoin(FOO, osmomath.NewInt(5_000_000_000))))

	// Validate that volume is zero
	totalVolume := s.App.PoolManagerKeeper.GetTotalVolumeForPool(s.Ctx, concentratedPool.GetId())
	s.Require().Equal(emptyCoins.String(), totalVolume.String())

	// Fund sender
	tokenIn := sdk.NewCoin(UOSMO, osmomath.NewInt(1000))
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(tokenIn))

	// System under test
	if noTakerFeeVariant {
		_, err := s.App.PoolManagerKeeper.SwapExactAmountInNoTakerFee(s.Ctx, s.TestAccs[0], concentratedPool.GetId(), tokenIn, FOO, osmomath.ZeroInt())
		s.Require().NoError(err)
	} else {
		_, _, err := s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], concentratedPool.GetId(), tokenIn, FOO, osmomath.ZeroInt())
		s.Require().NoError(err)
	}

	// Validate that volume was updated
	totalVolume = s.App.PoolManagerKeeper.GetTotalVolumeForPool(s.Ctx, concentratedPool.GetId())
	s.Require().Equal(tokenIn.String(), totalVolume.String())
}

func (suite *KeeperTestSuite) TestListPoolsByDenom() {
	suite.Setup()

	tests := map[string]struct {
		denom            string
		poolCoins        []sdk.Coins
		expectedNumPools int
		expectedError    bool
		poolType         []types.PoolType
	}{
		"Single pool, pool contain denom": {
			poolType: []types.PoolType{types.Balancer},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1 bar-uosmo
			},
			denom:            BAR,
			expectedNumPools: 1,
		},
		"Single pool, pool does not contain denom": {
			poolType: []types.PoolType{types.Balancer},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1 bar-uosmo
			},
			denom:            FOO,
			expectedNumPools: 0,
		},
		"Two pools, pools contains denom": {
			poolType: []types.PoolType{types.Balancer, types.Balancer},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1 bar-uosmo
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(FOO, defaultInitPoolAmount)),   // pool 2. baz-foo
			},
			denom:            BAR,
			expectedNumPools: 2,
		},
		"Two pools, pools does not contains denom": {
			poolType: []types.PoolType{types.Balancer, types.Concentrated},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1 bar-uosmo
				sdk.NewCoins(sdk.NewCoin(BAZ, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 2. baz-foo
			},
			denom:            FOO,
			expectedNumPools: 0,
		},
		"Many pools": {
			poolType: []types.PoolType{types.Concentrated, types.Balancer, types.Concentrated, types.Balancer},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1 bar-uosmo
				sdk.NewCoins(sdk.NewCoin(BAZ, defaultInitPoolAmount), sdk.NewCoin(FOO, defaultInitPoolAmount)),   // pool 2. baz-foo
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(BAZ, defaultInitPoolAmount)),   // pool 3. bar-baz
				sdk.NewCoins(sdk.NewCoin(BAZ, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 4. baz-uosmo
			},
			denom:            BAR,
			expectedNumPools: 2,
		},
		"A cosmwasm pool": {
			poolType: []types.PoolType{types.CosmWasm},
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(BAR, defaultInitPoolAmount), sdk.NewCoin(UOSMO, defaultInitPoolAmount)), // pool 1 bar-uosmo
			},
			denom:            BAR,
			expectedNumPools: 1,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			poolManagerKeeper := suite.App.PoolManagerKeeper

			for i := range tc.poolType {
				suite.FundAcc(suite.TestAccs[0], tc.poolCoins[i])
				suite.CreatePoolFromTypeWithCoins(tc.poolType[i], tc.poolCoins[i])
			}

			poolsResult, err := poolManagerKeeper.ListPoolsByDenom(ctx, tc.denom)
			if tc.expectedError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedNumPools, len(poolsResult))
			}
		})
	}
}
