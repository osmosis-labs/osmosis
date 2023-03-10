package concentrated_liquidity_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
)

var (
	DefaultExponentAtPriceOne                      = sdk.NewInt(-4)
	DefaultMinTick, DefaultMaxTick                 = cl.GetMinAndMaxTicksFromExponentAtPriceOne(DefaultExponentAtPriceOne)
	DefaultLowerPrice                              = sdk.NewDec(4545)
	DefaultLowerTick                               = int64(305450)
	DefaultUpperPrice                              = sdk.NewDec(5500)
	DefaultUpperTick                               = int64(315000)
	DefaultCurrPrice                               = sdk.NewDec(5000)
	DefaultCurrTick                                = sdk.NewInt(310000)
	DefaultCurrSqrtPrice, _                        = DefaultCurrPrice.ApproxSqrt() // 70.710678118654752440
	DefaultZeroSwapFee                             = sdk.ZeroDec()
	DefaultFeeAccumCoins                           = sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(50)))
	DefaultFreezeDuration                          = time.Duration(time.Hour * 24)
	ETH                                            = "eth"
	DefaultAmt0                                    = sdk.NewInt(1000000)
	DefaultAmt0Expected                            = sdk.NewInt(998976)
	DefaultCoin0                                   = sdk.NewCoin(ETH, DefaultAmt0)
	USDC                                           = "usdc"
	DefaultAmt1                                    = sdk.NewInt(5000000000)
	DefaultAmt1Expected                            = sdk.NewInt(5000000000)
	DefaultCoin1                                   = sdk.NewCoin(USDC, DefaultAmt1)
	DefaultLiquidityAmt                            = sdk.MustNewDecFromStr("1517882343.751510418088349649")
	DefaultTickSpacing                             = uint64(1)
	PoolCreationFee                                = poolmanagertypes.DefaultParams().PoolCreationFee
	DefaultExponentConsecutivePositionLowerTick, _ = math.PriceToTick(sdk.NewDec(5500), DefaultExponentAtPriceOne)
	DefaultExponentConsecutivePositionUpperTick, _ = math.PriceToTick(sdk.NewDec(6250), DefaultExponentAtPriceOne)
	DefaultExponentOverlappingPositionLowerTick, _ = math.PriceToTick(sdk.NewDec(4000), DefaultExponentAtPriceOne)
	DefaultExponentOverlappingPositionUpperTick, _ = math.PriceToTick(sdk.NewDec(4999), DefaultExponentAtPriceOne)
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

func (s *KeeperTestSuite) SetupDefaultPosition(poolId uint64) {
	s.SetupPosition(poolId, s.TestAccs[0], DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, s.Ctx.BlockTime(), DefaultFreezeDuration)
}

func (s *KeeperTestSuite) SetupPosition(poolId uint64, owner sdk.AccAddress, coin0, coin1 sdk.Coin, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration) model.Position {
	s.FundAcc(owner, sdk.NewCoins(coin0, coin1))
	_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, owner, coin0.Amount, coin1.Amount, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick, freezeDuration)
	s.Require().NoError(err)
	position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
	s.Require().NoError(err)
	return *position
}

// SetupDefaultPositions sets up four different positions to the given pool with different accounts for each position./
// Sets up the following positions:
// 1. Default position
// 2. Full range position
// 3. Postion with consecutive price range from the default position
// 4. Position with overlapping price range from the default position
func (s *KeeperTestSuite) SetupDefaultPositions(poolId uint64) {
	// ----------- set up positions ----------
	// 1. Default position
	s.SetupDefaultPosition(poolId)

	// 2. Full range position
	s.SetupFullRangePositionAcc(poolId, s.TestAccs[1])

	// 3. Position with consecutive price range from the default position
	s.SetupOverlappingRangePositionAcc(poolId, s.TestAccs[2])

	// 4. Position with overlapping price range from the default position
	s.SetupOverlappingRangePositionAcc(poolId, s.TestAccs[3])
}

func (s *KeeperTestSuite) SetupDefaultPositionAcc(poolId uint64, owner sdk.AccAddress) {
	s.SetupPosition(poolId, owner, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, s.Ctx.BlockTime(), DefaultFreezeDuration)
}

func (s *KeeperTestSuite) SetupFullRangePositionAcc(poolId uint64, owner sdk.AccAddress) {
	s.SetupPosition(poolId, owner, DefaultCoin0, DefaultCoin1, DefaultMinTick, DefaultMaxTick, s.Ctx.BlockTime(), DefaultFreezeDuration)
}

func (s *KeeperTestSuite) SetupConsecutiveRangePositionAcc(poolId uint64, owner sdk.AccAddress) {
	s.SetupPosition(poolId, owner, DefaultCoin0, DefaultCoin1, DefaultExponentConsecutivePositionLowerTick.Int64(), DefaultExponentConsecutivePositionUpperTick.Int64(), s.Ctx.BlockTime(), DefaultFreezeDuration)
}

func (s *KeeperTestSuite) SetupOverlappingRangePositionAcc(poolId uint64, owner sdk.AccAddress) {
	s.SetupPosition(poolId, owner, DefaultCoin0, DefaultCoin1, DefaultExponentOverlappingPositionLowerTick.Int64(), DefaultExponentOverlappingPositionUpperTick.Int64(), s.Ctx.BlockTime(), DefaultFreezeDuration)
}

// validatePositionUpdate validates that position with given parameters has expectedRemainingLiquidity left.
func (s *KeeperTestSuite) validatePositionUpdate(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64, joinTime time.Time, freezeDuration time.Duration, expectedRemainingLiquidity sdk.Dec) {
	position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
	s.Require().NoError(err)
	newPositionLiquidity := position.Liquidity
	s.Require().Equal(expectedRemainingLiquidity.String(), newPositionLiquidity.String())
	s.Require().True(newPositionLiquidity.GTE(sdk.ZeroDec()))
}

// validateTickUpdates validates that ticks with the given parameters have expectedRemainingLiquidity left.
func (s *KeeperTestSuite) validateTickUpdates(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64, expectedRemainingLiquidity sdk.Dec, expectedLowerFeeGrowthOutside, expectedUpperFeeGrowthOutside sdk.DecCoins) {
	lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, lowerTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityNet.String())
	s.Require().Equal(lowerTickInfo.FeeGrowthOutside.String(), expectedLowerFeeGrowthOutside.String())

	upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, upperTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), upperTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.Neg().String(), upperTickInfo.LiquidityNet.String())
	s.Require().Equal(upperTickInfo.FeeGrowthOutside.String(), expectedUpperFeeGrowthOutside.String())
}

func (s *KeeperTestSuite) initializeTick(ctx sdk.Context, currentTick int64, tickIndex int64, initialLiquidity sdk.Dec, feeGrowthOutside sdk.DecCoins, uptimeTrackers []model.UptimeTracker, isLower bool) {
	err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(ctx, validPoolId, currentTick, tickIndex, initialLiquidity, isLower)
	s.Require().NoError(err)

	tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, validPoolId, tickIndex)
	s.Require().NoError(err)

	tickInfo.FeeGrowthOutside = feeGrowthOutside
	tickInfo.UptimeTrackers = uptimeTrackers

	s.App.ConcentratedLiquidityKeeper.SetTickInfo(ctx, validPoolId, tickIndex, tickInfo)
}

// initializeFeeAccumulatorPositionWithLiquidity initializes fee accumulator position with given parameters and updates it with given liquidity.
func (s *KeeperTestSuite) initializeFeeAccumulatorPositionWithLiquidity(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidity sdk.Dec) {
	err := s.App.ConcentratedLiquidityKeeper.InitializeFeeAccumulatorPosition(ctx, poolId, owner, lowerTick, upperTick)
	s.Require().NoError(err)

	err = s.App.ConcentratedLiquidityKeeper.UpdateFeeAccumulatorPosition(ctx, poolId, owner, liquidity, lowerTick, upperTick)
	s.Require().NoError(err)
}

// validatePositionFeeAccUpdate validates that the position's accumulator with given parameters
// has been updated with liquidity.
func (s *KeeperTestSuite) validatePositionFeeAccUpdate(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64, liquidity sdk.Dec) {
	accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, poolId)
	s.Require().NoError(err)

	accumulatorPosition, err := accum.GetPositionSize(cl.FormatPositionAccumulatorKey(poolId, owner, lowerTick, upperTick))
	s.Require().NoError(err)

	s.Require().Equal(liquidity.String(), accumulatorPosition.String())
}
