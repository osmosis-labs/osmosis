package concentrated_liquidity_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/internal/math"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/v14/app/apptesting"
)

var (
	DefaultExponentAtPriceOne      = sdk.NewInt(-4)
	DefaultMinTick, DefaultMaxTick = math.GetMinAndMaxTicksFromExponentAtPriceOne(DefaultExponentAtPriceOne)
	DefaultLowerPrice              = sdk.NewDec(4545)
	DefaultLowerTick               = int64(305450)
	DefaultUpperPrice              = sdk.NewDec(5500)
	DefaultUpperTick               = int64(315000)
	DefaultCurrPrice               = sdk.NewDec(5000)
	DefaultCurrTick                = sdk.NewInt(310000)
	DefaultCurrSqrtPrice, _        = DefaultCurrPrice.ApproxSqrt() // 70.710678118654752440
	DefaultZeroSwapFee             = sdk.ZeroDec()
	ETH                            = "eth"
	DefaultAmt0                    = sdk.NewInt(1000000)
	DefaultAmt0Expected            = sdk.NewInt(998976)
	USDC                           = "usdc"
	DefaultAmt1                    = sdk.NewInt(5000000000)
	DefaultAmt1Expected            = sdk.NewInt(5000000000)
	DefaultLiquidityAmt            = sdk.MustNewDecFromStr("1517882343.751510418088349649")
	DefaultTickSpacing             = uint64(1)
	PoolCreationFee                = poolmanagertypes.DefaultParams().PoolCreationFee
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

func (s *KeeperTestSuite) SetupPosition(poolId uint64) {
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))
	_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)
}

// validatePositionUpdate validates that position with given parameters has expectedRemainingLiquidity left.
func (s *KeeperTestSuite) validatePositionUpdate(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64, expectedRemainingLiquidity sdk.Dec) {
	position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(ctx, poolId, owner, lowerTick, upperTick)
	s.Require().NoError(err)
	newPositionLiquidity := position.Liquidity
	s.Require().Equal(expectedRemainingLiquidity.String(), newPositionLiquidity.String())
	s.Require().True(newPositionLiquidity.GTE(sdk.ZeroDec()))
}

// validateTickUpdates validates that ticks with the given parameters have expectedRemainingLiquidity left.
func (s *KeeperTestSuite) validateTickUpdates(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64, expectedRemainingLiquidity sdk.Dec) {
	lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, lowerTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityNet.String())

	upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, upperTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), upperTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.Neg().String(), upperTickInfo.LiquidityNet.String())
}
