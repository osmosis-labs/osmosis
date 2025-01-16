package poolstransformer_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	ingesttypes "github.com/osmosis-labs/osmosis/v29/ingest/types"
	sqscosmwasmpool "github.com/osmosis-labs/osmosis/v29/ingest/types/cosmwasmpool"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v29/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v29/ingest/common/domain"
	poolstransformer "github.com/osmosis-labs/osmosis/v29/ingest/sqs/pools/transformer"
	clqueryproto "github.com/osmosis-labs/osmosis/v29/x/concentrated-liquidity/client/queryproto"
	cltypes "github.com/osmosis-labs/osmosis/v29/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v29/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v29/x/protorev/types"
)

type PoolTransformerTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	USDT = "usdt"

	USDC = poolstransformer.USDC
	// set w for simplicity of reasoning
	// about lexographical order
	USDW = "usdw"

	UOSMO = poolstransformer.UOSMO

	noPoolLiquidityCapErrorStr = poolstransformer.NoPoolLiquidityCapError
)

var (
	defaultAmount       = osmomath.NewInt(1_000_000_000)
	halfDefaultAmount   = defaultAmount.QuoRaw(2)
	doubleDefaultAmount = defaultAmount.MulRaw(2)

	emptyDenomPriceInfoMap      = map[string]osmomath.BigDec{}
	emptyDenomPairToTakerFeeMap = ingesttypes.TakerFeeMap{}

	zeroInt = osmomath.ZeroInt()
)

func TestPoolTransformerTestSuite(t *testing.T) {
	suite.Run(t, new(PoolTransformerTestSuite))
}

// This test validates that converting a pool that has to get pool liquidity capitalization from another
// pool with empty price info map works as expected.
// Additionally, it also validates that the input denomPairToTakerFeeMap is mutated correctly
// with the taker fee retrieved from the pool manager params.
func (s *PoolTransformerTestSuite) TestConvertPool_EmptyPriceInfoMap_TakerFee() {
	s.Setup()

	s.setDefaultPoolManagerTakerFee()

	// Corresponds to the denoms of the pool being converted
	// The taker fee is taken from params.
	expectedDenomPairToTakerFeeMap := ingesttypes.TakerFeeMap{
		{
			Denom0: USDC,
			Denom1: USDT,
		}: defaultPoolManagerTakerFee,
		{
			Denom0: USDT,
			Denom1: USDC,
		}: defaultPoolManagerTakerFee,
	}

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDC, usdcOsmoPoolID)

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	priceInfoMap := map[string]osmomath.BigDec{}
	denomPairToTakerFeeMap := ingesttypes.TakerFeeMap{}

	poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, priceInfoMap, denomPairToTakerFeeMap)
	s.Require().NoError(err)

	// 0.5 defaultAmount OSMO for each token that equals 1 osmo and 2 for each denom
	// 	Multiplied by two in default quote denom conversion
	expectedPoolLiquidityCap := descaleQuoteDenomPrecisionAmount(doubleDefaultAmount)

	expectPoolLiquidityCapErrorStr := noPoolLiquidityCapErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))
	s.validatePoolConversion(pool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)

	// Validate that the input denom pair to taker fee map is updated correctly.
	s.Require().Equal(expectedDenomPairToTakerFeeMap, denomPairToTakerFeeMap)
}

// This test validates that converting a pool that has to pool liquidity capitalization from another
// pool with non-empty denom to routing info map works as expected.
func (s *PoolTransformerTestSuite) TestConvertPool_NonEmptyPriceInfoMap() {
	s.Setup()

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDC, usdcOsmoPoolID)

	denomPriceInfoMap := map[string]osmomath.BigDec{
		// Make the spot price 4 OSMO = 1 USDC
		// This is for testing that the price is picked up from
		// this map, if present, rather than pool.
		USDC: osmomath.NewBigDec(4),
	}

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomPriceInfoMap, emptyDenomPairToTakerFeeMap)

	// 0.5 OSMO per USDT amount + 0.25 OSMO per USDC amount (overwritten by routing info)
	expectedPoolLiquidityCap := defaultAmount.QuoRaw(2).Add(defaultAmount.QuoRaw(4))
	// 	Multiplied by two in default quote denom conversion
	expectedPoolLiquidityCap = descaleQuoteDenomPrecisionAmount(expectedPoolLiquidityCap.MulRaw(2))
	expectPoolLiquidityCapErrorStr := noPoolLiquidityCapErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))
	s.validatePoolConversion(pool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)
}

// This test validates that converting an OSMO paired pool that has to get TVL from another
// pool with empty denom to routing info map works as expected.
func (s *PoolTransformerTestSuite) TestConvertPool_OSMOPairedPool_WithRoutingInOtherPool() {
	s.Setup()
	// Create OSMO / USDT pool
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolIDSpotPrice := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdcOsmoPoolIDSpotPrice)

	defaultQuoteUOSMOPoolID := s.CreateDefaultQuoteDenomUOSMOPool()

	priceInfoMap := map[string]osmomath.BigDec{}
	denomPairToTakerFeeMap := ingesttypes.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester(defaultQuoteUOSMOPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, priceInfoMap, denomPairToTakerFeeMap)

	// 0.5 OSMO per USDT amount + half amount OSMO itself
	expectedPoolLiquidityCap := halfDefaultAmount.Add(halfDefaultAmount)
	// 	Multiplied by two in default quote denom conversion
	expectedPoolLiquidityCap = descaleQuoteDenomPrecisionAmount(expectedPoolLiquidityCap.MulRaw(2))
	expectPoolLiquidityCapErrorStr := noPoolLiquidityCapErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)
}

// This test validates that converting an OSMO paired pool that has to get TVL from itself
// with empty denom to routing info map works as expected.
func (s *PoolTransformerTestSuite) TestConvertPool_OSMOPairedPool_WithRoutingAsItself() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolIDConverted)

	// Create default pool for converting between UOSMO and USDC.
	usdcOsmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()

	denomPriceInfoMap := map[string]osmomath.BigDec{}
	denomPairToTakerFeeMap := ingesttypes.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomPriceInfoMap, denomPairToTakerFeeMap)

	// 0.5 OSMO per USDT amount + half amount OSMO itself
	expectedPoolLiquidityCap := halfDefaultAmount.Add(halfDefaultAmount)
	// 	Multiplied by two in default quote denom conversion
	expectedPoolLiquidityCap = descaleQuoteDenomPrecisionAmount(expectedPoolLiquidityCap.MulRaw(2))
	expectPoolLiquidityCapErrorStr := noPoolLiquidityCapErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)
}

// Tests that if no route is set, the pool is converted correctly and the method does not error.
// However, the error flag is updated.
func (s *PoolTransformerTestSuite) TestConvertPool_NoRouteSet() {
	s.Setup()

	usdcUosmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	// Purposefully remove the route set in the after pool created hook.
	s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, UOSMO)

	denomPriceInfoMap := map[string]osmomath.BigDec{}
	denomPairToTakerFeeMap := ingesttypes.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester(usdcUosmoPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomPriceInfoMap, denomPairToTakerFeeMap)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedPoolLiquidityCap := halfDefaultAmount
	// 	Multiplied by two in default quote denom conversion
	expectedPoolLiquidityCap = descaleQuoteDenomPrecisionAmount(expectedPoolLiquidityCap.MulRaw(2))
	expectPoolLiquidityCapErrorStr := protorevtypes.NoPoolForDenomPairError{
		BaseDenom:  UOSMO,
		MatchDenom: USDT,
	}.Error()
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)
}

// Tests that if route is set incorrectly, the error is silently skipped and the error flag is set.
func (s *PoolTransformerTestSuite) TestConvertPool_InvalidPoolSetInRoutes_SilentSpotPriceError() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	// Purposefully set a non-existent pool
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolIDConverted+1)

	usdcUosmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()

	denomPriceInfoMap := map[string]osmomath.BigDec{}
	denomPairToTakerFeeMap := ingesttypes.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester(usdcUosmoPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomPriceInfoMap, denomPairToTakerFeeMap)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedPoolLiquidityCap := halfDefaultAmount
	// 	Multiplied by two in default quote denom conversion
	expectedPoolLiquidityCap = descaleQuoteDenomPrecisionAmount(expectedPoolLiquidityCap.MulRaw(2))
	// Note: empty string is set for simplicity, omitting the actual error in the assertion
	expectPoolLiquidityCapErrorStr := fmt.Sprintf(poolstransformer.SpotPriceErrorFmtStr, USDT, "")
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)
}

// This test validates that converting a concentrated pool works as expected.
// This pool type also has tick data set on it.
func (s *PoolTransformerTestSuite) TestConvertPool_Concentrated() {
	s.Setup()

	// Create default pool for converting between UOSMO and USDC.
	usdcUosmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()

	// Prepare a stablecoin pool that we attempt to convert

	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, UOSMO, 1, osmomath.ZeroDec())

	initialLiquidity := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, defaultAmount))
	s.FundAcc(s.TestAccs[0], initialLiquidity)
	fullRangePositionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, concentratedPool.GetId(), s.TestAccs[0], initialLiquidity)
	s.Require().NoError(err)

	// Refetch the pool from state.
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, concentratedPool.GetId())
	s.Require().NoError(err)

	denomPriceInfoMap := map[string]osmomath.BigDec{}
	denomPairToTakerFeeMap := ingesttypes.TakerFeeMap{}

	poolIngester := s.initializePoolIngester(usdcUosmoPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, concentratedPool, denomPriceInfoMap, denomPairToTakerFeeMap)
	s.Require().NoError(err)

	// 1:1 ration, twice the default amount
	// However, due to CL LP logic UOSMO ands up being LPed truncated.
	// As a result, we take it directly from balances. The precision of CL logic is not
	// the concern of this test so this is acceptable.
	osmoBalance := s.App.BankKeeper.GetBalance(s.Ctx, concentratedPool.GetAddress(), UOSMO)

	// Validate that osmo balance is close to the default amount
	tolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: osmomath.NewDecWithPrec(1, 2), // 1%
		RoundingDir:             osmomath.RoundDown,
	}
	osmoassert.Equal(s.T(), tolerance, defaultAmount, osmoBalance.Amount)

	expectedPoolLiquidityCap := defaultAmount.Add(osmoBalance.Amount)
	// 	Multiplied by two in default quote denom conversion
	expectedPoolLiquidityCap = descaleQuoteDenomPrecisionAmount(expectedPoolLiquidityCap.MulRaw(2))
	expectPoolLiquidityCapErrorStr := noPoolLiquidityCapErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), osmoBalance)
	s.validatePoolConversion(concentratedPool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)

	// Validate that ticks are set for full range position

	actualModel, err := actualPool.GetTickModel()
	s.Require().NoError(err)

	s.Require().Equal(1, len(actualModel.Ticks))
	s.Require().Equal(int64(0), actualModel.CurrentTickIndex)

	expectedTick := clqueryproto.LiquidityDepthWithRange{
		LowerTick:       cltypes.MinInitializedTick,
		UpperTick:       cltypes.MaxTick,
		LiquidityAmount: fullRangePositionData.Liquidity,
	}

	s.Require().Equal(expectedTick, actualModel.Ticks[0])
}

// This test validates that CL pools with no liquidity are converted correctly.
// The relevant no liquidity flag is set where applicable
func (s *PoolTransformerTestSuite) TestConvertPool_Concentrated_NoLiquidity() {
	s.Setup()

	// Create default pool for converting between UOSMO and USDC.
	usdcUosmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()

	// Prepare a stablecoin pool that we attempt to convert

	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, UOSMO, 1, osmomath.ZeroDec())

	denomPriceInfoMap := map[string]osmomath.BigDec{}
	denomPairToTakerFeeMap := ingesttypes.TakerFeeMap{}

	poolIngester := s.initializePoolIngester(usdcUosmoPoolID)

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, concentratedPool, denomPriceInfoMap, denomPairToTakerFeeMap)
	s.Require().NoError(err)

	// 1:1 ration, twice the default amount
	// However, due to CL LP logic UOSMO ands up being LPed truncated.
	// As a result, we take it directly from balances. The precision of CL logic is not
	// the concern of this test so this is acceptable.
	// osmoBalance := s.App.BankKeeper.GetBalance(s.Ctx, concentratedPool.GetAddress(), UOSMO)

	expectedPoolLiquidityCap := osmomath.ZeroInt()
	expectPoolLiquidityCapErrorStr := noPoolLiquidityCapErrorStr
	expectedBalances := sdk.Coins{
		sdk.Coin{
			Denom:  UOSMO,
			Amount: osmomath.ZeroInt(),
		},
		sdk.Coin{
			Denom:  USDT,
			Amount: osmomath.ZeroInt(),
		},
	}
	s.validatePoolConversion(concentratedPool, expectedPoolLiquidityCap, expectPoolLiquidityCapErrorStr, actualPool, expectedBalances)

	// Validate that ticks are set for full range position

	actualModel, err := actualPool.GetTickModel()
	s.Require().NoError(err)

	s.Require().Equal(0, len(actualModel.Ticks))
	s.Require().Equal(int64(-1), actualModel.CurrentTickIndex)
	s.Require().True(actualModel.HasNoLiquidity)
}

// This test validates that the block is processes correctly.
// That is, it checks that:
// - the appropriate pools are ingested
// - taker fee is set correctly (either from params or custom)
//
// TODO: add tests for TVL setting.
// https://app.clickup.com/t/86a1b3t6p
func (s *PoolTransformerTestSuite) TestProcessBlock() {
	s.Setup()

	// Set the default taker fee
	s.setDefaultPoolManagerTakerFee()

	// Create one pool of each type.
	poolsData := s.PrepareAllSupportedPools()

	// Create a custom denom pair taker fee and set its taker fee to non-default
	customTakerFeeConcentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, USDC, 1, osmomath.ZeroDec())
	s.App.PoolManagerKeeper.SetDenomPairTakerFee(s.Ctx, customTakerFeeConcentratedPool.GetToken0(), customTakerFeeConcentratedPool.GetToken1(), defaultCustomTakerFee)
	s.App.PoolManagerKeeper.SetDenomPairTakerFee(s.Ctx, customTakerFeeConcentratedPool.GetToken1(), customTakerFeeConcentratedPool.GetToken0(), defaultCustomTakerFee)

	sqsKeepers := commondomain.PoolExtractorKeepers{
		GammKeeper:         s.App.GAMMKeeper,
		ConcentratedKeeper: s.App.ConcentratedLiquidityKeeper,
		WasmKeeper:         s.App.WasmKeeper,
		BankKeeper:         s.App.BankKeeper,
		ProtorevKeeper:     s.App.ProtoRevKeeper,
		PoolManagerKeeper:  s.App.PoolManagerKeeper,
		CosmWasmPoolKeeper: s.App.CosmwasmPoolKeeper,
	}

	// Get concentrated pool
	concentratedPool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolsData.ConcentratedPoolID)
	s.Require().NoError(err)

	// Get balancer pool
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolsData.BalancerPoolID)
	s.Require().NoError(err)

	// Get stable swap pool
	stableSwapPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolsData.StableSwapPoolID)
	s.Require().NoError(err)

	// Get cosm wasm pool
	cosmWasmPool, err := s.App.CosmwasmPoolKeeper.GetPool(s.Ctx, poolsData.CosmWasmPoolID)
	s.Require().NoError(err)

	// Create default pool for converting between UOSMO and USDC.
	usdcUosmoPoolID := s.CreateDefaultQuoteDenomUOSMOPool()
	poolTransformer := poolstransformer.NewPoolTransformer(sqsKeepers, usdcUosmoPoolID)

	blockPools := commondomain.BlockPools{
		ConcentratedPools: []poolmanagertypes.PoolI{
			concentratedPool,
			customTakerFeeConcentratedPool,
		},
		CFMMPools: []poolmanagertypes.PoolI{
			balancerPool,
			stableSwapPool,
		},
		CosmWasmPools: []poolmanagertypes.PoolI{
			cosmWasmPool,
		},
	}

	allPools, takerFeesMap, err := poolTransformer.Transform(s.Ctx, blockPools)
	s.Require().NoError(err)

	s.Require().Len(allPools, 2+2+1)

	// Order of pools is by order of writes:
	// 1. CFMM
	// 2. Concentrated
	// 3. Cosmwasm

	s.Require().Equal(poolsData.BalancerPoolID, allPools[0].GetId())
	s.Require().Equal(poolsData.StableSwapPoolID, allPools[1].GetId())

	s.Require().Equal(poolsData.ConcentratedPoolID, allPools[2].GetId())
	s.Require().Equal(customTakerFeeConcentratedPool.GetId(), allPools[3].GetId())

	s.Require().Equal(poolsData.CosmWasmPoolID, allPools[4].GetId())

	// Validate taker fee for the custom pool
	actualTakerFee := takerFeesMap.GetTakerFee(customTakerFeeConcentratedPool.GetToken0(), customTakerFeeConcentratedPool.GetToken1())
	// Custom taker fee
	s.Require().Equal(defaultCustomTakerFee, actualTakerFee)

	// Validate taker fee for one of the default taker fee pools
	defaultConcentratedPool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolsData.ConcentratedPoolID)
	s.Require().NoError(err)
	actualTakerFee = takerFeesMap.GetTakerFee(defaultConcentratedPool.GetToken0(), defaultConcentratedPool.GetToken1())
	// Poolmanager params taker fee
	s.Require().Equal(defaultPoolManagerTakerFee, actualTakerFee)
}

// This tests validates that uosmo pool liquidity cap is computed correctly
// by validating the happy path cases. Validates that if no computation method is found
// the error string and zero is returned without error or panic
func (s *PoolTransformerTestSuite) TestComputeUOSMOPoolLiquidityCap() {
	var (
		uosmoCoins = sdk.NewCoins(sdk.NewCoin(UOSMO, defaultAmount))
		usdcCoins  = sdk.NewCoins(sdk.NewCoin(USDC, defaultAmount))
	)

	tests := []struct {
		name string

		balances               sdk.Coins
		priceInfoMap           map[string]osmomath.BigDec
		shouldSetProtorevRoute bool

		expectedPoolLiquidityCap         osmomath.Int
		expectedPoolLiquidityErrorSubstr string
	}{
		{
			name:         "UOSMO balance -> returns the same amount",
			balances:     uosmoCoins,
			priceInfoMap: map[string]osmomath.BigDec{},

			expectedPoolLiquidityCap:         defaultAmount,
			expectedPoolLiquidityErrorSubstr: noPoolLiquidityCapErrorStr,
		},
		{
			name:                   "USDC Balance with no routing info but protorev route -> returns half by using protorev route",
			balances:               usdcCoins,
			priceInfoMap:           map[string]osmomath.BigDec{},
			shouldSetProtorevRoute: true,

			expectedPoolLiquidityCap:         halfDefaultAmount,
			expectedPoolLiquidityErrorSubstr: noPoolLiquidityCapErrorStr,
		},
		{
			name:     "USDC Balance with price info & protorev route present -> returns the amount using the price info price",
			balances: usdcCoins,
			priceInfoMap: map[string]osmomath.BigDec{
				USDC: osmomath.NewBigDec(4),
			},
			shouldSetProtorevRoute: true,

			expectedPoolLiquidityCap:         defaultAmount.QuoRaw(4),
			expectedPoolLiquidityErrorSubstr: noPoolLiquidityCapErrorStr,
		},
		{
			name:                   "USDC balance with no routing info and no protorev route -> use stables overwrite",
			balances:               usdcCoins,
			priceInfoMap:           map[string]osmomath.BigDec{},
			shouldSetProtorevRoute: false,

			// defaultAmount from usdcCoins * spot price of 0.5
			expectedPoolLiquidityCap:         halfDefaultAmount,
			expectedPoolLiquidityErrorSubstr: noPoolLiquidityCapErrorStr,
		},

		{
			name:                   "USDT balance with no routing info and no protorev route -> return zero and not found error string",
			balances:               sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount)),
			priceInfoMap:           map[string]osmomath.BigDec{},
			shouldSetProtorevRoute: false,

			expectedPoolLiquidityCap:         zeroInt,
			expectedPoolLiquidityErrorSubstr: "not found",
		},

		{
			name:                   "UOSMO & USDC from skip route",
			balances:               uosmoCoins.Add(usdcCoins...),
			priceInfoMap:           map[string]osmomath.BigDec{},
			shouldSetProtorevRoute: true,

			// default for UOSMO and half for USDC
			expectedPoolLiquidityCap:         defaultAmount.Add(halfDefaultAmount),
			expectedPoolLiquidityErrorSubstr: noPoolLiquidityCapErrorStr,
		},
	}

	for _, tc := range tests {
		tc := tc

		s.Run(tc.name, func() {
			s.Setup()
			// Create OSMO / USDC pool and set the protorev route
			// Note that spot price is 1 OSMO = 2 USDC
			usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

			// Delete all protorev pools for UOSMO by default (set in the after pool created hook)
			s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, UOSMO)

			// Set the protorev route if the test case requires it
			if tc.shouldSetProtorevRoute {
				s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDC, usdcOsmoPoolID)
			}

			// Initialize the pool ingester
			poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

			// System under test
			actualPoolLiquidityCap, actualPoolLiquidityCapError := poolIngester.ComputeUOSMOPoolLiquidityCap(s.Ctx, tc.balances, tc.priceInfoMap)

			// Validate the results
			s.Require().Equal(tc.expectedPoolLiquidityCap.String(), actualPoolLiquidityCap.String())
			s.Require().Contains(actualPoolLiquidityCapError, tc.expectedPoolLiquidityErrorSubstr)
		})
	}
}

// This tests validates that usdc pool liquidity cap is computed correctly from uosmo.
// by validating the happy path cases. Validates that if no computation method is found
// the error string and zero is returned without error or panic
func (s *PoolTransformerTestSuite) TestComputeUSDCPoolLiquidityCapFromUOSMO() {
	defaultAmountDescaled := osmomath.BigDecFromSDKInt(defaultAmount).QuoMut(poolstransformer.UsdcPrecisionScalingFactor).Dec().TruncateInt()

	tests := []struct {
		name string

		uosmoPoolLiquidityCap osmomath.Int

		isInvalidUOSMOUSDCPool bool

		expectedPoolLiquidityCap         osmomath.Int
		expectedPoolLiquidityErrorSubstr string
	}{
		{
			name:                  "UOSMO balance -> returns the same amount",
			uosmoPoolLiquidityCap: osmomath.ZeroInt(),

			expectedPoolLiquidityCap:         zeroInt,
			expectedPoolLiquidityErrorSubstr: noPoolLiquidityCapErrorStr,
		},

		{
			name:                  "UOSMO balance with USDC pool set -> computes the amount correctly",
			uosmoPoolLiquidityCap: halfDefaultAmount,

			// halfDefaultAmount * price of two
			expectedPoolLiquidityCap:         defaultAmountDescaled,
			expectedPoolLiquidityErrorSubstr: noPoolLiquidityCapErrorStr,
		},

		{
			name:                   "Invalid UOSMO-USDC pool - returns zero and error string.",
			uosmoPoolLiquidityCap:  halfDefaultAmount,
			isInvalidUOSMOUSDCPool: true,

			expectedPoolLiquidityCap:         osmomath.ZeroInt(),
			expectedPoolLiquidityErrorSubstr: fmt.Sprintf(poolstransformer.SpotPriceErrorFmtStr, USDC, ""),
		},
	}

	for _, tc := range tests {
		tc := tc

		s.Run(tc.name, func() {
			s.Setup()

			// Create OSMO / USDC pool and set the protorev route
			// Note that spot price is 1 OSMO = 2 USDC
			usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

			if tc.isInvalidUOSMOUSDCPool {
				usdcOsmoPoolID += 1
			}

			// Initialize the pool ingester
			poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

			// System under test
			actualPoolLiquidityCap, actualPoolLiquidityCapError := poolIngester.ComputeUSDCPoolLiquidityCapFromUOSMO(s.Ctx, tc.uosmoPoolLiquidityCap)

			// Validate the results
			s.Require().Equal(tc.expectedPoolLiquidityCap.String(), actualPoolLiquidityCap.String())
			s.Require().Contains(actualPoolLiquidityCapError, tc.expectedPoolLiquidityErrorSubstr)
		})
	}
}

func (s *PoolTransformerTestSuite) TestFilterBalances() {
	var (
		oneHundredInt    = osmomath.NewInt(100)
		twoHundreadInt   = osmomath.NewInt(200)
		threeHundreadInt = osmomath.NewInt(300)

		// Starting balances and pool denoms
		originalBalances = sdk.NewCoins(
			sdk.NewCoin(UOSMO, oneHundredInt),
			sdk.NewCoin(USDC, twoHundreadInt),
			sdk.NewCoin(USDT, threeHundreadInt),
		)
		poolDenomsMap = map[string]struct{}{
			UOSMO: {},
			USDC:  {},
		}
	)

	tests := []struct {
		name     string
		balances sdk.Coins
		poolMap  map[string]struct{}
		expected sdk.Coins
	}{
		{
			name:     "Filters out non-pool tokens",
			balances: originalBalances,
			poolMap:  poolDenomsMap,
			expected: sdk.NewCoins(
				sdk.NewCoin(UOSMO, oneHundredInt),
				sdk.NewCoin(USDC, twoHundreadInt),
			),
		},
		{
			name:     "Handles empty balances",
			balances: sdk.Coins{},
			poolMap:  poolDenomsMap,
			expected: sdk.Coins{},
		},
		{
			name:     "No pool denoms in balances",
			balances: sdk.NewCoins(sdk.NewCoin(USDT, threeHundreadInt)),
			poolMap:  poolDenomsMap,
			expected: sdk.Coins{},
		},
		{
			name: "All balances are pool denoms",
			balances: sdk.NewCoins(
				sdk.NewCoin(UOSMO, oneHundredInt),
				sdk.NewCoin(USDC, twoHundreadInt),
			),
			poolMap: poolDenomsMap,
			expected: sdk.NewCoins(
				sdk.NewCoin(UOSMO, oneHundredInt),
				sdk.NewCoin(USDC, twoHundreadInt),
			),
		},
		{
			name:     "Mixed valid and invalid denoms",
			balances: originalBalances,
			poolMap: map[string]struct{}{
				UOSMO: {},
			},
			expected: sdk.NewCoins(sdk.NewCoin(UOSMO, oneHundredInt)),
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			result := poolstransformer.FilterBalances(tc.balances, tc.poolMap)

			s.Require().Equal(tc.expected, result)
		})
	}
}

func (s *PoolTransformerTestSuite) TestInitCosmWasmPoolModel() {
	s.Setup()
	// Create OSMO / USDC pool and
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

	// Initialize the pool ingester
	poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(
		sdk.NewCoin(apptesting.DefaultTransmuterDenomA, osmomath.NewInt(100000000)),
		sdk.NewCoin(apptesting.DefaultTransmuterDenomB, osmomath.NewInt(100000000)),
	))

	pool := s.PrepareCosmWasmPool()
	cwpm := poolIngester.InitCosmWasmPoolModel(s.Ctx, pool)
	s.Equal(sqscosmwasmpool.CosmWasmPoolModel{
		ContractInfo: sqscosmwasmpool.ContractInfo{
			Contract: "crates.io:transmuter",
			Version:  "0.1.0",
		},
	}, cwpm)

	pool = s.PrepareAlloyTransmuterPool(s.TestAccs[0], apptesting.AlloyTransmuterInstantiateMsg{
		PoolAssetConfigs:                []apptesting.AssetConfig{{Denom: apptesting.DefaultTransmuterDenomA, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomANormFactor)}, {Denom: apptesting.DefaultTransmuterDenomB, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomBNormFactor)}},
		AlloyedAssetSubdenom:            apptesting.DefaultAlloyedSubDenom,
		AlloyedAssetNormalizationFactor: osmomath.NewInt(apptesting.DefaultAlloyedDenomNormFactor),
		Admin:                           s.TestAccs[0].String(),
		Moderator:                       s.TestAccs[1].String(),
	})

	cwpm = poolIngester.InitCosmWasmPoolModel(s.Ctx, pool)
	s.Equal(sqscosmwasmpool.CosmWasmPoolModel{
		ContractInfo: sqscosmwasmpool.ContractInfo{
			Contract: "crates.io:transmuter",
			Version:  "3.0.0",
		},
	}, cwpm)
}

// validatePoolConversion validates that the pool conversion is correct.
// It asserts that
// - the pool ID of the actual pool is equal to the expected pool ID.
// - the pool type of the actual pool is equal to the expected pool type.
// - the TVL of the actual pool is equal to the expected TVL.
// - the balances of the actual pool is equal to the expected balances.
func (s *PoolTransformerTestSuite) validatePoolConversion(expectedPool poolmanagertypes.PoolI, expectedPoolLiquidityCap osmomath.Int, expectPoolLiquidityCapError string, actualPool ingesttypes.PoolI, expectedBalances sdk.Coins) {
	// Correct ID
	s.Require().Equal(expectedPool.GetId(), actualPool.GetId())

	// Correct type
	s.Require().Equal(expectedPool.GetType(), actualPool.GetType())

	// Validate TVL
	s.Require().Equal(expectedPoolLiquidityCap.String(), actualPool.GetPoolLiquidityCap().String())
	sqsPoolModel := actualPool.GetSQSPoolModel()
	s.Require().Contains(sqsPoolModel.PoolLiquidityCapError, expectPoolLiquidityCapError)

	// Validate pool denoms
	poolDenoms := actualPool.GetPoolDenoms()
	s.Require().Equal(2, len(poolDenoms))
	s.Require().Equal(expectedBalances[0].Denom, poolDenoms[0])
	s.Require().Equal(expectedBalances[1].Denom, poolDenoms[1])

	// Validate balances
	// Filter out zero balances with coins constructor.
	// The reason we do this is because for no liqudity cases we supply
	// zero coins for getting the expected denoms of the pool.
	expectedBalances = sdk.NewCoins(expectedBalances...)
	s.Require().Equal(expectedBalances.String(), sqsPoolModel.Balances.String())
}

func (s *PoolTransformerTestSuite) initializePoolIngester(defaultUSDCUOSMOPoolID uint64) *poolstransformer.PoolTransformer {

	sqsKeepers := commondomain.PoolExtractorKeepers{
		GammKeeper:         s.App.GAMMKeeper,
		ConcentratedKeeper: s.App.ConcentratedLiquidityKeeper,
		BankKeeper:         s.App.BankKeeper,
		ProtorevKeeper:     s.App.ProtoRevKeeper,
		PoolManagerKeeper:  s.App.PoolManagerKeeper,
		CosmWasmPoolKeeper: s.App.CosmwasmPoolKeeper,
		WasmKeeper:         s.App.WasmKeeper,
	}

	atomicIngester := poolstransformer.NewPoolTransformer(sqsKeepers, defaultUSDCUOSMOPoolID)
	poolIngester, ok := atomicIngester.(*poolstransformer.PoolTransformer)
	s.Require().True(ok)
	return poolIngester
}

func (s *PoolTransformerTestSuite) TestGetPoolDenomsMap() {
	tests := []struct {
		name     string
		input    []string
		expected map[string]struct{}
	}{
		{
			name:     "Handles empty slice",
			input:    []string{},
			expected: map[string]struct{}{},
		},
		{
			name:  "Handles single element",
			input: []string{UOSMO},
			expected: map[string]struct{}{
				UOSMO: {},
			},
		},
		{
			name:  "Converts multiple denoms to map",
			input: []string{UOSMO, USDC, USDT},
			expected: map[string]struct{}{
				UOSMO: {},
				USDC:  {},
				USDT:  {},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			result := poolstransformer.GetPoolDenomsMap(tt.input)

			s.Require().Equal(tt.expected, result)
		})
	}
}

// CreateDefaultQuoteDenomUOSMOPool Create OSMO / USDC pool with price of 1 UOSMO = 2 USDC
// and return the pool ID
func (s *PoolTransformerTestSuite) CreateDefaultQuoteDenomUOSMOPool() uint64 {
	return s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
}

// descaleQuoteDenomPrecisionAmount descales the amount with the quote denom precision scaling factor.
func descaleQuoteDenomPrecisionAmount(amount osmomath.Int) osmomath.Int {
	return osmomath.BigDecFromSDKInt(amount).QuoMut(poolstransformer.UsdcPrecisionScalingFactor).Dec().Ceil().TruncateInt()
}
