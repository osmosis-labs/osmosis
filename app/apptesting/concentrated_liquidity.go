package apptesting

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"

	cl "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity"
)

var (
	ETH                = "eth"
	USDC               = "usdc"
	WBTC               = "ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F"
	DefaultTickSpacing = uint64(100)
	DefaultLowerTick   = int64(30545000)
	DefaultUpperTick   = int64(31500000)
	DefaultCoinAmount  = sdk.NewInt(1000000000000000000)
)

// PrepareConcentratedPool sets up an eth usdc concentrated liquidity pool with a tick spacing of 100,
// no liquidity and zero spread factor.
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

// createConcentratedPoolsFromCoinsWithSpreadFactor creates CL pools from given sets of coins and respective swap fees.
// Where element 1 of the input corresponds to the first pool created, element 2 to the second pool created etc.
func (s *KeeperTestHelper) CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(poolDenoms [][]string, spreadFactor []sdk.Dec) {
	for i, curPoolDenoms := range poolDenoms {
		s.Require().Equal(2, len(curPoolDenoms))
		var curSpreadFactor sdk.Dec
		if len(spreadFactor) > i {
			curSpreadFactor = spreadFactor[i]
		} else {
			curSpreadFactor = sdk.ZeroDec()
		}

		clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], curPoolDenoms[0], curPoolDenoms[1], DefaultTickSpacing, curSpreadFactor)
		fundCoins := sdk.NewCoins(sdk.NewCoin(curPoolDenoms[0], DefaultCoinAmount), sdk.NewCoin(curPoolDenoms[1], DefaultCoinAmount))
		s.FundAcc(s.TestAccs[0], fundCoins)
		s.CreateFullRangePosition(clPool, fundCoins)
	}
}

// createConcentratedPoolsFromCoins creates CL pools from given sets of coins (with zero swap fees).
// Where element 1 of the input corresponds to the first pool created, element 2 to the second pool created etc.
func (s *KeeperTestHelper) CreateConcentratedPoolsAndFullRangePosition(poolDenoms [][]string) {
	s.CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(poolDenoms, []sdk.Dec{sdk.ZeroDec()})
}

// PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition sets up a concentrated liquidity pool with custom denoms.
// It also creates a full range position and locks it for 14 days.
func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition(denom1, denom2 string) (types.ConcentratedPoolExtension, uint64, uint64) {
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, sdk.ZeroDec())
	fundCoins := sdk.NewCoins(sdk.NewCoin(denom1, DefaultCoinAmount), sdk.NewCoin(denom2, DefaultCoinAmount))
	s.FundAcc(s.TestAccs[0], fundCoins)
	positionData, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), s.TestAccs[0], fundCoins, time.Hour*24*14)
	s.Require().NoError(err)
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, clPool.GetId())
	s.Require().NoError(err)
	return clPool, concentratedLockId, positionData.ID
}

// PrepareCustomConcentratedPool sets up a concentrated liquidity pool with the custom parameters.
func (s *KeeperTestHelper) PrepareCustomConcentratedPool(owner sdk.AccAddress, denom0, denom1 string, tickSpacing uint64, spreadFactor sdk.Dec) types.ConcentratedPoolExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], DefaultAcctFunds)

	// Create a concentrated pool via the poolmanager
	poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(owner, denom0, denom1, tickSpacing, spreadFactor))
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
	positionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pool.GetId(), s.TestAccs[0], coins)
	s.Require().NoError(err)
	return positionData.ID, positionData.Liquidity
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
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, defaultParams)

	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.AuthorizedQuoteDenoms = append(poolmanagertypes.DefaultParams().AuthorizedQuoteDenoms, ETH, USDC, BAR, BAZ, FOO, UOSMO, STAKE, WBTC)
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
}
