package concentrated_liquidity_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
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
	lowerTick := int64(84222)
	upperTick := int64(86129)
	amount0Desired := sdk.NewInt(1)
	amount1Desired := sdk.NewInt(5000)

	asset0, asset1, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], amount0Desired, amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
	s.Require().NoError(err)
	s.Require().Equal(amount0Desired.String(), asset0.String())
	s.Require().Equal(amount1Desired.String(), asset1.String())
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
