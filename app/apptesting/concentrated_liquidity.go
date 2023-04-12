package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
)

var (
	ETH                       = "eth"
	USDC                      = "usdc"
	DefaultTickSpacing        = uint64(1)
	DefaultExponentAtPriceOne = sdk.NewInt(-4)
	DefaultLowerTick          = int64(305450)
	DefaultUpperTick          = int64(315000)
)

// PrepareConcentratedPool sets up an eth usdc concentrated liquidity pool with pool ID 1, tick spacing of 1,
// no liquidity and zero swap fee.
func (s *KeeperTestHelper) PrepareConcentratedPool() types.ConcentratedPoolExtension {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, DefaultExponentAtPriceOne, sdk.ZeroDec())
}

func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoins(denom1, denom2 string) types.ConcentratedPoolExtension {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, DefaultExponentAtPriceOne, sdk.ZeroDec())
}

func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoinsAndFullRangePosition(denom1, denom2 string) types.ConcentratedPoolExtension {
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, DefaultExponentAtPriceOne, sdk.ZeroDec())
	fundCoins := sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(1000000000000000000)), sdk.NewCoin(denom2, sdk.NewInt(1000000000000000000)))
	s.FundAcc(s.TestAccs[0], fundCoins)
	s.CreateFullRangePosition(clPool, fundCoins)
	return clPool
}

// PrepareCustomConcentratedPool sets up a concentrated liquidity pool with the custom parameters.
func (s *KeeperTestHelper) PrepareCustomConcentratedPool(owner sdk.AccAddress, denom0, denom1 string, tickSpacing uint64, exponentAtPriceOne sdk.Int, swapFee sdk.Dec) types.ConcentratedPoolExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	// Create a concentrated pool via the poolmanager
	poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(owner, denom0, denom1, tickSpacing, exponentAtPriceOne, swapFee))
	s.Require().NoError(err)

	// Retrieve the poolInterface via the poolID
	poolI, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolID)
	s.Require().NoError(err)

	// Type cast the PoolInterface to a ConcentratedPoolExtension
	pool, ok := poolI.(types.ConcentratedPoolExtension)
	s.Require().True(ok)

	return pool
}

// PrepareMultipleConcentratedPools returns X cl pool's with X being provided by the user.
func (s *KeeperTestHelper) PrepareMultipleConcentratedPools(poolsToCreate uint16) []uint64 {
	var poolIds []uint64
	for i := uint16(0); i < poolsToCreate; i++ {
		pool := s.PrepareConcentratedPool()
		poolIds = append(poolIds, pool.GetId())
	}

	return poolIds
}

// CreateFullRangePosition creates a full range position and returns position id and the liquidity created.
func (s *KeeperTestHelper) CreateFullRangePosition(pool types.ConcentratedPoolExtension, coins sdk.Coins) (uint64, sdk.Dec) {
	s.FundAcc(s.TestAccs[0], coins)
	positionId, _, _, liquidity, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pool, s.TestAccs[0], coins)
	s.Require().NoError(err)
	return positionId, liquidity
}

// WithdrawFullRangePosition withdraws given liquidity from a position specified by id.
func (s *KeeperTestHelper) WithdrawFullRangePosition(pool types.ConcentratedPoolExtension, positionId uint64, liquidityToRemove sdk.Dec) {
	clMsgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

	_, err := clMsgServer.WithdrawPosition(sdk.WrapSDKContext(s.Ctx), &types.MsgWithdrawPosition{
		PositionId:      positionId,
		LiquidityAmount: liquidityToRemove,
		Sender:          s.TestAccs[0].String(),
	})
	s.Require().NoError(err)
}
