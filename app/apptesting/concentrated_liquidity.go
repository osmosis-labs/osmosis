package apptesting

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clmath "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	clmodel "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"

	cl "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity"
)

var (
	ETH                = "eth"
	USDC               = "usdc"
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
	positionId, _, _, _, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), s.TestAccs[0], fundCoins, time.Hour*24*14)
	s.Require().NoError(err)
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, clPool.GetId())
	s.Require().NoError(err)
	return clPool, concentratedLockId, positionId
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
	positionId, _, _, liquidity, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pool.GetId(), s.TestAccs[0], coins)
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

// PriceToTickRoundDown takes a price and returns the corresponding tick index.
// If tickSpacing is provided, the tick index will be rounded down to the nearest multiple of tickSpacing.
// CONTRACT: tickSpacing must be smaller or equal to the max of 1 << 63 - 1.
// This is not a concern because we have authorized tick spacings that are smaller than this max,
// and we don't expect to ever require it to be this large.
func (s *KeeperTestHelper) PriceToTickRoundDownSpacing(price sdk.Dec, tickSpacing uint64) (int64, error) {
	tickIndex, err := s.PriceToTick(price)
	if err != nil {
		return 0, err
	}

	tickIndex, err = clmath.RoundDownTickToSpacing(tickIndex, int64(tickSpacing))
	if err != nil {
		return 0, err
	}

	return tickIndex, nil
}

// CalculatePriceToTick takes in a price and returns the corresponding tick index.
// This function does not take into consideration tick spacing.
// NOTE: This is really returning a "Bucket index". Bucket index `b` corresponds to
// all prices in range [TickToSqrtPrice(b), TickToSqrtPrice(b+1)).
// We make an erroneous assumption here, that bucket index `b` corresponds to
// all prices in range [TickToPrice(b), TickToPrice(b+1)).
// This currently makes this function unsuitable for the state machine.
func (s *KeeperTestHelper) CalculatePriceToTick(price sdk.Dec) (tickIndex int64) {
	// TODO: Make truncate, since this defines buckets as
	// [TickToPrice(b - .5), TickToPrice(b+.5))
	return clmath.CalculatePriceToTickDec(price).RoundInt64()
}

// PriceToTick takes a price and returns the corresponding tick index assuming
// tick spacing of 1.
func (s *KeeperTestHelper) PriceToTick(price sdk.Dec) (int64, error) {
	if price.Equal(sdk.OneDec()) {
		return 0, nil
	}

	if price.IsNegative() {
		return 0, fmt.Errorf("price must be greater than zero")
	}

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return 0, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	// Determine the tick that corresponds to the price
	// This does not take into account the tickSpacing
	tickIndex := s.CalculatePriceToTick(price)

	return tickIndex, nil
}
