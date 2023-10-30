package redis_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	redisingester "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/ingester/redis"
	clqueryproto "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/client/queryproto"
	cltypes "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type IngesterTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	USDT = "uatom"
	USDC = "usdc"
)

var (
	defaultAmount     = osmomath.NewInt(1_000_000_000)
	halfDefaultAmount = defaultAmount.QuoRaw(2)
)

func TestIngesterTestSuite(t *testing.T) {
	suite.Run(t, new(IngesterTestSuite))
}

// This test validates that converting a pool that has to get TVL from another
// pool with empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_EmptyDenomToRoutingInfoMap() {
	s.Setup()

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, redisingester.UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, redisingester.UOSMO, USDC, usdcOsmoPoolID)

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)
	s.Require().NoError(err)

	// 2 for the spot price (each denom is worth 2 OSMO) and 2 for each denom
	expectedTVL := defaultAmount.MulRaw(2 * 2)
	expectTVLError := false
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// This test validates that converting a pool that has to get TVL from another
// pool with non-empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_NonEmptyDenomToRoutingInfoMap() {
	s.Setup()

	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, redisingester.UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, redisingester.UOSMO, USDC, usdcOsmoPoolID)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{
		USDC: {
			PoolID: usdcOsmoPoolID,
			// Make the spot price 4 OSMO = 1 USDC
			Price: osmomath.NewBigDec(4),
		},
	}

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)

	// 2 OSMO per USDT amount + 4 OSMO per USDC amount (overwritten by routing info)
	expectedTVL := defaultAmount.MulRaw(2).Add(defaultAmount.MulRaw(4))
	expectTVLError := false
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// This test validates that converting an OSMO paired pool that has to get TVL from another
// pool with empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_OSMOPairedPool_WithRoutingInOtherPool() {
	s.Setup()
	// Create OSMO / USDT pool
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolIDSpotPrice := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, redisingester.UOSMO, USDT, usdcOsmoPoolIDSpotPrice)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)

	// 2 OSMO per USDT amount + half amount OSMO itself
	expectedTVL := defaultAmount.MulRaw(2).Add(halfDefaultAmount)
	expectTVLError := false
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// This test validates that converting an OSMO paired pool that has to get TVL from itself
// with empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_OSMOPairedPool_WithRoutingAsItself() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, redisingester.UOSMO, USDT, usdtOsmoPoolIDConverted)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)

	// 2 OSMO per USDT amount + half amount OSMO itself
	expectedTVL := defaultAmount.MulRaw(2).Add(halfDefaultAmount)
	expectTVLError := false
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// Tests that if no route is set, the pool is converted correctly and the method does not error.
// However, the error flag is updated.
func (s *IngesterTestSuite) TestConvertPool_NoRouteSet() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	// Purposefully remove the route set in the after pool created hook.
	s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, redisingester.UOSMO)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedTVL := halfDefaultAmount
	expectTVLError := true
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// Tests that if route is set incorrectly, the error is silently skipped and the error flag is set.
func (s *IngesterTestSuite) TestConvertPool_InvalidPoolSetInRoutes_SilentSpotPriceError() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	// Purposefully set a non-existent pool
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, redisingester.UOSMO, USDT, usdtOsmoPoolIDConverted+1)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedTVL := halfDefaultAmount
	expectTVLError := true
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// This test validates that converting a concentrated pool works as expected.
// This pool type also has tick data set on it.
func (s *IngesterTestSuite) TestConvertPool_Concentrated() {
	s.Setup()

	// Prepare a stablecoin pool that we attempt to convert

	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, redisingester.UOSMO, 1, osmomath.ZeroDec())

	initialLiquidity := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(redisingester.UOSMO, defaultAmount))
	s.FundAcc(s.TestAccs[0], initialLiquidity)
	fullRangePositionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, concentratedPool.GetId(), s.TestAccs[0], initialLiquidity)
	s.Require().NoError(err)

	// Refetch the pool from state.
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, concentratedPool.GetId())
	s.Require().NoError(err)

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, concentratedPool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)
	s.Require().NoError(err)

	// 1:1 ration, twice the default amount
	// However, due to CL LP logic UOSMO ands up being LPed truncated.
	// As a result, we take it directly from balances. The precision of CL logic is not
	// the concern of this test so this is acceptable.
	osmoBalance := s.App.BankKeeper.GetBalance(s.Ctx, concentratedPool.GetAddress(), redisingester.UOSMO)

	// Validate that osmo balance is close to the default amount
	tolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: osmomath.NewDecWithPrec(1, 2), // 1%
		RoundingDir:             osmomath.RoundDown,
	}
	osmoassert.Equal(s.T(), tolerance, defaultAmount, osmoBalance.Amount)

	expectedTVL := defaultAmount.Add(osmoBalance.Amount)
	expectTVLError := false
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), osmoBalance)
	s.validatePoolConversion(concentratedPool, expectedTVL, expectTVLError, actualPool, expectedBalances)

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

	concentratedPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], USDT, redisingester.UOSMO, 1, osmomath.ZeroDec())

	denomToRoutingInfoMap := map[string]redisingester.DenomRoutingInfo{}

	// System under test
	actualPool, err := redisingester.ConvertPool(s.Ctx, concentratedPool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper, s.App.ConcentratedLiquidityKeeper)
	s.Require().NoError(err)

	// 1:1 ration, twice the default amount
	// However, due to CL LP logic UOSMO ands up being LPed truncated.
	// As a result, we take it directly from balances. The precision of CL logic is not
	// the concern of this test so this is acceptable.
	// osmoBalance := s.App.BankKeeper.GetBalance(s.Ctx, concentratedPool.GetAddress(), redisingester.UOSMO)

	expectedTVL := osmomath.ZeroInt()
	expectTVLError := false
	expectedBalances := sdk.Coins{
		sdk.Coin{
			Denom:  USDT,
			Amount: osmomath.ZeroInt(),
		},
		sdk.Coin{
			Denom:  redisingester.UOSMO,
			Amount: osmomath.ZeroInt(),
		},
	}
	s.validatePoolConversion(concentratedPool, expectedTVL, expectTVLError, actualPool, expectedBalances)

	// Validate that ticks are set for full range position

	actualModel, err := actualPool.GetTickModel()
	s.Require().NoError(err)

	s.Require().Equal(0, len(actualModel.Ticks))
	s.Require().Equal(int64(-1), actualModel.CurrentTickIndex)
	s.Require().True(actualModel.HasNoLiquidity)
}

// validatePoolConversion validates that the pool conversion is correct.
// It asserts that
// - the pool ID of the actual pool is equal to the expected pool ID.
// - the pool type of the actual pool is equal to the expected pool type.
// - the TVL of the actual pool is equal to the expected TVL.
// - the balances of the actual pool is equal to the expected balances.
func (s *IngesterTestSuite) validatePoolConversion(expectedPool poolmanagertypes.PoolI, expectedTVL osmomath.Int, expectTVLError bool, actualPool domain.PoolI, expectedBalances sdk.Coins) {
	// Correct ID
	s.Require().Equal(expectedPool.GetId(), actualPool.GetId())

	// Correct type
	s.Require().Equal(expectedPool.GetType(), actualPool.GetType())

	// Validate TVL
	s.Require().Equal(expectedTVL.String(), actualPool.GetTotalValueLockedUOSMO().String())
	sqsPoolModel := actualPool.GetSQSPoolModel()
	s.Require().Equal(expectTVLError, sqsPoolModel.IsErrorInTotalValueLocked)

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
