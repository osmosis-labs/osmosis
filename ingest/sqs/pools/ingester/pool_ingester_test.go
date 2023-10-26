package ingester_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/ingester"
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
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, ingester.UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, ingester.UOSMO, USDC, usdcOsmoPoolID)

	// Prepare a stablecoin pool that we attempt to convert
	stableCoinPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(USDC, defaultAmount))

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, stableCoinPoolID)
	s.Require().NoError(err)

	denomToRoutingInfoMap := map[string]ingester.DenomRoutingInfo{}

	// System under test
	actualPool, err := ingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper)
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
	usdtOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, ingester.UOSMO, USDT, usdtOsmoPoolID)

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, ingester.UOSMO, USDC, usdcOsmoPoolID)

	denomToRoutingInfoMap := map[string]ingester.DenomRoutingInfo{
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
	actualPool, err := ingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper)

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
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))

	// Create OSMO / USDC pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolIDSpotPrice := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, ingester.UOSMO, USDT, usdcOsmoPoolIDSpotPrice)

	denomToRoutingInfoMap := map[string]ingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := ingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper)

	// 2 OSMO per USDT amount + half amount OSMO itself
	expectedTVL := defaultAmount.MulRaw(2).Add(halfDefaultAmount)
	expectTVLError := false
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// This test validates that converting an OSMO paired pool that has to get TVL from itself
// with empty denom to routing info map works as expected.
func (s *IngesterTestSuite) TestConvertPool_OSMOPairedPool_WithRoutingAsItself() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, ingester.UOSMO, USDT, usdtOsmoPoolIDConverted)

	denomToRoutingInfoMap := map[string]ingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := ingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper)

	// 2 OSMO per USDT amount + half amount OSMO itself
	expectedTVL := defaultAmount.MulRaw(2).Add(halfDefaultAmount)
	expectTVLError := false
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// Tests that if no route is set, the pool is converted correctly and the method does not error.
// However, the error flag is updated.
func (s *IngesterTestSuite) TestConvertPool_NoRouteSet() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	// Purposefully remove the route set in the after pool created hook.
	s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, ingester.UOSMO)

	denomToRoutingInfoMap := map[string]ingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := ingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedTVL := halfDefaultAmount
	expectTVLError := true
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
}

// Tests that if route is set incorrectly, the error is silently skipped and the error flag is set.
func (s *IngesterTestSuite) TestConvertPool_InvalidPoolSetInRoutes_SilentSpotPriceError() {
	s.Setup()
	// Create OSMO / USDT pool and set the protorev route
	// Note that spot price is 1 OSMO = 2 USDT
	usdtOsmoPoolIDConverted := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	// Purposefully set a non-existent pool
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, ingester.UOSMO, USDT, usdtOsmoPoolIDConverted+1)

	denomToRoutingInfoMap := map[string]ingester.DenomRoutingInfo{}

	// Fetch the pool from state.
	pool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, usdtOsmoPoolIDConverted)
	s.Require().NoError(err)

	// System under test
	actualPool, err := ingester.ConvertPool(s.Ctx, pool, denomToRoutingInfoMap, s.App.BankKeeper, s.App.ProtoRevKeeper, s.App.PoolManagerKeeper)

	// Only counts half amount of OSMO because USDT has no route set.
	expectedTVL := halfDefaultAmount
	expectTVLError := true
	expectedBalances := sdk.NewCoins(sdk.NewCoin(USDT, defaultAmount), sdk.NewCoin(ingester.UOSMO, halfDefaultAmount))
	s.validatePoolConversion(pool, expectedTVL, expectTVLError, actualPool, expectedBalances)
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

	// Validate balances
	s.Require().Equal(expectedBalances.String(), sqsPoolModel.Balances.String())

	// Validate pool denoms
	poolDenoms := actualPool.GetPoolDenoms()
	s.Require().Equal(2, len(poolDenoms))
	s.Require().Equal(expectedBalances[0].Denom, poolDenoms[0])
	s.Require().Equal(expectedBalances[1].Denom, poolDenoms[1])
}
