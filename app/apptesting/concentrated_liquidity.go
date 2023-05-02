package apptesting

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
)

var (
	ETH                = "eth"
	USDC               = "usdc"
	DefaultTickSpacing = uint64(100)
	DefaultLowerTick   = int64(30545000)
	DefaultUpperTick   = int64(31500000)
	DefaultCoinAmount  = sdk.NewInt(1000000000000000000)
)

// PrepareConcentratedPool sets up an eth usdc concentrated liquidity pool with pool ID 1, tick spacing of 100,
// no liquidity and zero swap fee.
func (s *KeeperTestHelper) PrepareConcentratedPool() types.ConcentratedPoolExtension {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, sdk.ZeroDec())
}

// PrepareConcentratedPoolWithCoins sets up a concentrated liquidity pool with custom denoms.
func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoins(denom1, denom2 string) types.ConcentratedPoolExtension {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, sdk.ZeroDec())
}

// PrepareConcentratedPoolWithCoinsAndFullRangePosition sets up a concentrated liquidity pool with custom denoms.
// It also creates a full range position.
func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoinsAndFullRangePosition(denom1, denom2 string) types.ConcentratedPoolExtension {
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, sdk.ZeroDec())
	fundCoins := sdk.NewCoins(sdk.NewCoin(denom1, DefaultCoinAmount), sdk.NewCoin(denom2, DefaultCoinAmount))
	s.FundAcc(s.TestAccs[0], fundCoins)
	s.CreateFullRangePosition(clPool, fundCoins)
	return clPool
}

// PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition sets up a concentrated liquidity pool with custom denoms.
// It also creates a full range position and locks it for 14 days.
func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition(denom1, denom2 string) (types.ConcentratedPoolExtension, uint64, uint64) {
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, sdk.ZeroDec())
	fundCoins := sdk.NewCoins(sdk.NewCoin(denom1, DefaultCoinAmount), sdk.NewCoin(denom2, DefaultCoinAmount))
	s.FundAcc(s.TestAccs[0], fundCoins)
	positionId, _, _, _, _, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), s.TestAccs[0], fundCoins, time.Hour*24*14)
	s.Require().NoError(err)
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolFromPoolIdAndConvertToConcentrated(s.Ctx, clPool.GetId())
	s.Require().NoError(err)
	return clPool, concentratedLockId, positionId
}

// PrepareCustomConcentratedPool sets up a concentrated liquidity pool with the custom parameters.
func (s *KeeperTestHelper) PrepareCustomConcentratedPool(owner sdk.AccAddress, denom0, denom1 string, tickSpacing uint64, swapFee sdk.Dec) types.ConcentratedPoolExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	// Create a concentrated pool via the poolmanager
	poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(owner, denom0, denom1, tickSpacing, swapFee))
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
	positionId, _, _, liquidity, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pool.GetId(), s.TestAccs[0], coins)
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

// SetupConcentratedLiquidityDenomsAndPoolCreation sets up the default authorized quote denoms.
// Additionally, enables permissionless pool creation.
// This is to overwrite the default params set in concentrated liquidity genesis to account for the test cases that
// used various denoms before the authorized quote denoms were introduced.
func (s *KeeperTestHelper) SetupConcentratedLiquidityDenomsAndPoolCreation() {
	// modify authorized quote denoms to include test denoms.
	defaultParams := types.DefaultParams()
	defaultParams.IsPermissionlessPoolCreationEnabled = true
	defaultParams.AuthorizedQuoteDenoms = append(defaultParams.AuthorizedQuoteDenoms, ETH, USDC, BAR, BAZ, FOO, UOSMO, STAKE)
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, defaultParams)
}
