package redis_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v21/app/apptesting"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/common"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/ingester/redis"
	redisingester "github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/ingester/redis"
	clqueryproto "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/client/queryproto"
	cltypes "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v21/x/protorev/types"
)

type IngesterTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	USDT = "usdt"
	USDC = "usdc"
	// set w for simplicity of reasoning
	// about lexographical order
	USDW = "usdw"

	UOSMO = redisingester.UOSMO

	noTotalValueLockedErrorStr = ""
)

var (
	defaultAmount     = osmomath.NewInt(1_000_000_000)
	halfDefaultAmount = defaultAmount.QuoRaw(2)

	// Sets precision tof test tokens to match the precision of UOSMO
	// for simplicity.
	defaultOneToOneUosmoPrecisionMap = map[string]int{
		USDT: redisingester.OneToOnePrecision,
		USDC: redisingester.OneToOnePrecision,
		USDW: redisingester.OneToOnePrecision,
	}
)

func TestIngesterTestSuite(t *testing.T) {
	suite.Run(t, new(IngesterTestSuite))
}

// This test validates that converting a pool that has to get TVL from another
// pool with empty denom to routing info map works as expected.
// Additionally, it also validates that the input denomPairToTakerFeeMap is mutated correctly
// with the taker fee retrieved from the pool manager params.
func (s *IngesterTestSuite) TestConvertPool_EmptyDenomToRoutingInfoMa_TakerFee() {
	s.Setup()

	s.setDefaultPoolManagerTakerFee()

	// Corresponds to the denoms of the pool being converted
	// The taker fee is taken from params.
	expectedDenomPairToTakerFeeMap := domain.TakerFeeMap{
		{
			Denom0: USDC,
			Denom1: USDT,
		}: defaultPoolManagerTakerFee,
	}

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDC, usdcOsmoPoolID)

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)
	s.Require().NoError(err)

	// 0.5 for each token that equals 1 osmo and 2 for each denom
	expectedTVL := defaultAmount
	expectTVLErrorStr := noTotalValueLockedErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)

	// Validate that the input denom pair to taker fee map is updated correctly.
	s.Require().Equal(expectedDenomPairToTakerFeeMap, denomPairToTakerFeeMap)
}

// This test validates that converting a pool that has to get TVL from another
// pool with non-empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_NonEmptyDenomToRoutingInfoMap() {
	s.Setup()

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDC, usdcOsmoPoolID)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{
		USDC: {
			PoolID: usdcOsmoPoolID,
			// Make the spot price 4 OSMO = 1 USDC
			Price: osmomath.NewBigDec(4),
		},
	}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)

	// 0.5 OSMO per USDT amount + 0.25 OSMO per USDC amount (overwritten by routing info)
	expectedTVL := defaultAmount.QuoRaw(2).Add(defaultAmount.QuoRaw(4))
	expectTVLErrorStr := noTotalValueLockedErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)
}

// This test validates that converting an OSMO paired pool that has to get TVL from another
// pool with empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_OSMOPairedPool_WithRoutingInOtherPool() {
	s.Setup()
	// Create OSMO / USDT pool
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolIDSpotPrice := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdcOsmoPoolIDSpotPrice)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)

	// 0.5 OSMO per USDT amount + half amount OSMO itself
	expectedTVL := defaultAmount.QuoRaw(2).Add(halfDefaultAmount)
	expectTVLErrorStr := noTotalValueLockedErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)
}

// This test validates that converting an OSMO paired pool that has to get TVL from itself
// with empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_OSMOPairedPool_WithRoutingAsItself() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolIDConverted)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)

	// 0.5 OSMO per USDT amount + half amount OSMO itself
	expectedTVL := defaultAmount.QuoRaw(2).Add(halfDefaultAmount)
	expectTVLErrorStr := noTotalValueLockedErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)
}

// Tests that if no route is set, the pool is converted correctly and the method does not error.
// However, the error flag is updated.
func (s *IngesterTestSuite) TestConvertPool_NoRouteSet() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	// Purposefully remove the route set in the after pool created hook.
	s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, UOSMO)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedTVL := halfDefaultAmount
	expectTVLErrorStr := protorevtypes.NoPoolForDenomPairError{
		BaseDenom:  UOSMO,
		MatchDenom: USDT,
	}.Error()
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)
}

// Tests that if route is set incorrectly, the error is silently skipped and the error flag is set.
func (s *IngesterTestSuite) TestConvertPool_InvalidPoolSetInRoutes_SilentSpotPriceError() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	// Purposefully set a non-existent pool
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolIDConverted+1)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedTVL := halfDefaultAmount
	// Note: empty string is set for simplicity, omitting the actual error in the assertion
	expectTVLErrorStr := fmt.Sprintf(redis.SpotPriceErrorFmtStr, USDT, "")
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)
}

// This test validates that converting a concentrated pool works as expected.
// This pool type also has tick data set on it.
func (s *IngesterTestSuite) TestConvertPool_Concentrated() {
	s.Setup()

	// Prepare a stablecoin pool that we attempt to convert

	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, UOSMO, 1, osmomath.ZeroDec())

	initialLiquidity := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, defaultAmount))
	s.FundAcc(s.TestAccs[0], initialLiquidity)
	fullRangePositionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, concentratedPool.GetId(), s.TestAccs[0], initialLiquidity)
	s.Require().NoError(err)

	// Refetch the pool from state.
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, concentratedPool.GetId())
	s.Require().NoError(err)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, concentratedPool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)
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

	expectedTVL := defaultAmount.Add(osmoBalance.Amount)
	expectTVLErrorStr := noTotalValueLockedErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), osmoBalance)
	s.validatePoolConversion(concentratedPool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)

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
func (s *IngesterTestSuite) TestConvertPool_Concentrated_NoLiquidity() {
	s.Setup()

	// Prepare a stablecoin pool that we attempt to convert

	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, UOSMO, 1, osmomath.ZeroDec())

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, concentratedPool, denomToRoutingInfoMap, denomPairToTakerFeeMap, defaultOneToOneUosmoPrecisionMap)
	s.Require().NoError(err)

	// 1:1 ration, twice the default amount
	// However, due to CL LP logic UOSMO ands up being LPed truncated.
	// As a result, we take it directly from balances. The precision of CL logic is not
	// the concern of this test so this is acceptable.
	// osmoBalance := s.App.BankKeeper.GetBalance(s.Ctx, concentratedPool.GetAddress(), UOSMO)

	expectedTVL := osmomath.ZeroInt()
	expectTVLErrorStr := noTotalValueLockedErrorStr
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
	s.validatePoolConversion(concentratedPool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)

	// Validate that ticks are set for full range position

	actualModel, err := actualPool.GetTickModel()
	s.Require().NoError(err)

	s.Require().Equal(0, len(actualModel.Ticks))
	s.Require().Equal(int64(-1), actualModel.CurrentTickIndex)
	s.Require().True(actualModel.HasNoLiquidity)
}

// This test validates TVL calculation for a pool that has a token with non-osmo (6) precision
// that is set by the precision map given as a parameter.c
func (s *IngesterTestSuite) TestConvertPool_NonOsmoPrecision() {
	s.Setup()

	s.setDefaultPoolManagerTakerFee()

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolID)

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}
	customPrecisionMap := map[string]int{
		USDT: 18,
	}

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, customPrecisionMap)
	s.Require().NoError(err)

	// 1 for the OSMO-denominated TVL. 3 for the USDT-denominated TVL.
	// 1:1 spot price based on the reserves. However, precision multiplier is 3 (18 (osmo) / 6 (usdt) = 3)
	expectedTVL := defaultAmount.MulRaw(4)
	expectTVLErrorStr := noTotalValueLockedErrorStr
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, defaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)
}

func (s *IngesterTestSuite) TestConvertPool_NoPrecisionInMap() {
	s.Setup()

	s.setDefaultPoolManagerTakerFee()

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, UOSMO, USDT, usdtOsmoPoolID)

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}
	denomPairToTakerFeeMap := domain.TakerFeeMap{}
	emptyPrecisionMap := map[string]int{}

	poolIngester := s.initializePoolIngester()

	// System under test
	actualPool, err := poolIngester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, emptyPrecisionMap)
	s.Require().NoError(err)

	// 1 for the OSMO-denominated TVL. No USDT-denominated TVL because precision is not given in the map
	expectedTVL := defaultAmount.MulRaw(1)
	// TVL error due to no USDT precision in the map.
	expectTVLErrorStr := fmt.Sprintf(redis.NoTokenPrecisionErrorFmtStr, USDT)
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(UOSMO, defaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLErrorStr, actualPool, expectedBalances)
}

// This test validates that the block is processes correctly.
// That is, it checks that:
// - the appropriate pools are ingested
// - taker fee is set correctly (either from params or custom)
//
// TODO: add tests for TVL setting.
// https://app.clickup.com/t/86a1b3t6p
func (s *IngesterTestSuite) TestProcessBlock() {
	s.Setup()
	var (
		redisRepoMock   = &mocks.RedisPoolsRepositoryMock{}
		redisRouterMock = &mocks.RedisRouterRepositoryMock{
			TakerFees: domain.TakerFeeMap{},
		}
		tokensUseCaseMock = &mocks.TokensUseCaseMock{}

		// Note: this is a dummy tx that is not initialized correctly.
		// We do note expect it to be called or used by the system under test
		// due to using the mock repository.
		redisTx = mvc.NewRedisTx(nil)
	)

	// Set the default taker fee
	s.setDefaultPoolManagerTakerFee()

	// Create one pool of each type.
	poolsData := s.PrepareAllSupportedPools()

	// Create a custom denom pair taker fee and set its taker fee to non-default
	customTakerFeeConcentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, USDC, 1, osmomath.ZeroDec())
	s.App.PoolManagerKeeper.SetDenomPairTakerFee(s.Ctx, customTakerFeeConcentratedPool.GetToken0(), customTakerFeeConcentratedPool.GetToken1(), defaultCustomTakerFee)

	sqsKeepers := common.SQSIngestKeepers{
		GammKeeper:         s.App.GAMMKeeper,
		ConcentratedKeeper: s.App.ConcentratedLiquidityKeeper,
		BankKeeper:         s.App.BankKeeper,
		ProtorevKeeper:     s.App.ProtoRevKeeper,
		PoolManagerKeeper:  s.App.PoolManagerKeeper,
		CosmWasmPoolKeeper: s.App.CosmwasmPoolKeeper,
	}

	poolIngester := redisingester.NewPoolIngester(redisRepoMock, redisRouterMock, tokensUseCaseMock, nil, domain.RouterConfig{}, sqsKeepers)
	poolIngester.SetLogger(&log.NoOpLogger{})

	err := poolIngester.ProcessBlock(s.Ctx, redisTx)
	s.Require().NoError(err)

	allPools, err := redisRepoMock.GetAllPools(sdk.WrapSDKContext(s.Ctx))
	s.Require().NoError(err)

	s.Require().Len(allPools, 2+2+1)

	// Order of pooks is by order of writes:
	// 1. CFMM
	// 2. Concentrated
	// 3. Cosmwasm

	s.Require().Equal(poolsData.BalancerPoolID, allPools[0].GetId())
	s.Require().Equal(poolsData.StableSwapPoolID, allPools[1].GetId())

	s.Require().Equal(poolsData.ConcentratedPoolID, allPools[2].GetId())
	s.Require().Equal(customTakerFeeConcentratedPool.GetId(), allPools[3].GetId())

	s.Require().Equal(poolsData.CosmWasmPoolID, allPools[4].GetId())

	// Validate taker fee for the custom pool
	actualTakerFee, err := redisRouterMock.GetTakerFee(sdk.WrapSDKContext(s.Ctx), customTakerFeeConcentratedPool.GetToken0(), customTakerFeeConcentratedPool.GetToken1())
	s.Require().NoError(err)
	// Custom taker fee
	s.Require().Equal(defaultCustomTakerFee, actualTakerFee)

	// Validate taker fee for one of the default taker fee pools
	defaultConcentratedPool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolsData.ConcentratedPoolID)
	s.Require().NoError(err)
	actualTakerFee, err = redisRouterMock.GetTakerFee(sdk.WrapSDKContext(s.Ctx), defaultConcentratedPool.GetToken0(), defaultConcentratedPool.GetToken1())
	s.Require().NoError(err)
	// Poolmanager params taker fee
	s.Require().Equal(defaultPoolManagerTakerFee, actualTakerFee)
}

// validatePoolConversion validates that the pool conversion is correct.
// It asserts that
// - the pool ID of the actual pool is equal to the expected pool ID.
// - the pool type of the actual pool is equal to the expected pool type.
// - the TVL of the actual pool is equal to the expected TVL.
// - the balances of the actual pool is equal to the expected balances.
func (s *IngesterTestSuite) validatePoolConversion(expectedPool poolmanagertypes.PoolI, expectedTVL osmomath.Int, expectTVLErrorStr string, actualPool domain.PoolI, expectedBalances sdk.Coins) {
	// Correct ID
	s.Require().Equal(expectedPool.GetId(), actualPool.GetId())

	// Correct type
	s.Require().Equal(expectedPool.GetType(), actualPool.GetType())

	// Validate TVL
	s.Require().Equal(expectedTVL.String(), actualPool.GetTotalValueLockedUOSMO().String())
	sqsPoolModel := actualPool.GetSQSPoolModel()
	s.Require().Contains(sqsPoolModel.TotalValueLockedError, expectTVLErrorStr)

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

func (s *IngesterTestSuite) initializePoolIngester() *redisingester.PoolIngester {

	sqsKeepers := common.SQSIngestKeepers{
		GammKeeper:         s.App.GAMMKeeper,
		ConcentratedKeeper: s.App.ConcentratedLiquidityKeeper,
		BankKeeper:         s.App.BankKeeper,
		ProtorevKeeper:     s.App.ProtoRevKeeper,
		PoolManagerKeeper:  s.App.PoolManagerKeeper,
		CosmWasmPoolKeeper: s.App.CosmwasmPoolKeeper,
	}

	atomicIngester := redisingester.NewPoolIngester(nil, nil, nil, nil, domain.RouterConfig{}, sqsKeepers)
	poolIngester, ok := atomicIngester.(*redisingester.PoolIngester)
	poolIngester.SetLogger(&log.NoOpLogger{})
	s.Require().True(ok)
	return poolIngester
}
