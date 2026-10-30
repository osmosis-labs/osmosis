package v33_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v32/app/params"
	v33 "github.com/osmosis-labs/osmosis/v32/app/upgrades/v33"
	"github.com/osmosis-labs/osmosis/v32/x/gamm/pool-models/balancer"
	txfeestypes "github.com/osmosis-labs/osmosis/v32/x/txfees/types"
)

// poolLiquidity is the amount of each asset an LP deposits in the pool tests.
var poolLiquidity = osmomath.NewInt(1_000_000)

// poolMigration returns a pool migration of Noble USDC into transmuter with the given pool lists.
func poolMigration(transmuter, alloyed string, clConvert, balancerConvert, clDrain []uint64) v33.NobleUSDCPoolMigration {
	return v33.NobleUSDCPoolMigration{
		Transmuter: transmuter, Denom: v33.NobleUSDCDenom, AlloyedDenom: alloyed,
		CLConvertPools: clConvert, BalancerConvertPools: balancerConvert, CLDrainPools: clDrain,
		MaxPositionsPerPool: 10,
	}
}

// nobleCLPool creates a Noble USDC/OSMO concentrated liquidity pool and returns its ID.
func (s *UpgradeTestSuite) nobleCLPool() uint64 {
	return s.nobleCLPoolWith(appparams.BaseCoinUnit)
}

// nobleCLPoolWith creates a Noble USDC/quote concentrated liquidity pool and returns its ID.
func (s *UpgradeTestSuite) nobleCLPoolWith(quote string) uint64 {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], v33.NobleUSDCDenom, quote, 1, osmomath.ZeroDec()).GetId()
}

// fullRangePosition funds owner and creates a full range position for it in a Noble USDC/OSMO pool.
func (s *UpgradeTestSuite) fullRangePosition(poolId uint64, owner sdk.AccAddress) uint64 {
	return s.fullRangePositionWith(poolId, owner, appparams.BaseCoinUnit)
}

// fullRangePositionWith funds owner and creates a full range position for it in a Noble USDC/quote pool.
func (s *UpgradeTestSuite) fullRangePositionWith(poolId uint64, owner sdk.AccAddress, quote string) uint64 {
	coins := sdk.NewCoins(sdk.NewCoin(v33.NobleUSDCDenom, poolLiquidity), sdk.NewCoin(quote, poolLiquidity))
	s.FundAcc(owner, coins)
	position, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, poolId, owner, coins)
	s.Require().NoError(err)
	return position.ID
}

// TestMigrateNobleUSDCPoolsConvertsCLPool checks that a converted concentrated liquidity pool trades
// allUSDC instead of Noble USDC, that its reserve moved into the transmuter 1:1, that its positions
// withdraw allUSDC, and that TWAP records exist for the new pair.
func (s *UpgradeTestSuite) TestMigrateNobleUSDCPoolsConvertsCLPool() {
	transmuter, alloyed := s.prepareAlloyTransmuter(v33.NobleUSDCDenom)
	transmuterAddr := sdk.MustAccAddressFromBech32(transmuter)
	transmuterBefore := s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, v33.NobleUSDCDenom).Amount
	poolId := s.nobleCLPool()
	lp := s.newUser()
	positionId := s.fullRangePosition(poolId, lp)
	pool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolId)
	s.Require().NoError(err)
	reserve := s.App.BankKeeper.GetBalance(s.Ctx, pool.GetAddress(), v33.NobleUSDCDenom).Amount
	s.Require().True(reserve.IsPositive())

	result, err := v33.MigrateNobleUSDCPools(s.Ctx, &s.App.AppKeepers, poolMigration(transmuter, alloyed, []uint64{poolId}, nil, nil))
	s.Require().NoError(err)
	s.Require().Equal(1, result.Converted)
	s.Require().Equal(reserve, result.ConvertedAmount)

	pool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolId)
	s.Require().NoError(err)
	s.Require().Equal(alloyed, pool.GetToken0())
	s.Require().Equal(appparams.BaseCoinUnit, pool.GetToken1())
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, pool.GetAddress(), v33.NobleUSDCDenom).IsZero())
	s.Require().Equal(reserve, s.App.BankKeeper.GetBalance(s.Ctx, pool.GetAddress(), alloyed).Amount)
	s.Require().Equal(transmuterBefore.Add(reserve), s.App.BankKeeper.GetBalance(s.Ctx, transmuterAddr, v33.NobleUSDCDenom).Amount)

	_, err = s.App.TwapKeeper.GetAllMostRecentRecordsForPoolWithDenoms(s.Ctx, poolId, []string{alloyed, appparams.BaseCoinUnit})
	s.Require().NoError(err)

	position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
	s.Require().NoError(err)
	amount0, _, err := s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, lp, positionId, position.Liquidity)
	s.Require().NoError(err)
	s.Require().True(amount0.IsPositive())
	s.Require().Equal(amount0, s.App.BankKeeper.GetBalance(s.Ctx, lp, alloyed).Amount)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, lp, v33.NobleUSDCDenom).IsZero())
}

// TestMigrateNobleUSDCPoolsConvertsBalancerPool checks that a converted balancer pool keeps its
// amounts and weights under the allUSDC denom, sorted, and that LPs exit with allUSDC.
func (s *UpgradeTestSuite) TestMigrateNobleUSDCPoolsConvertsBalancerPool() {
	transmuter, alloyed := s.prepareAlloyTransmuter(v33.NobleUSDCDenom)
	poolId := s.PrepareCustomBalancerPool([]balancer.PoolAsset{
		{Token: sdk.NewCoin(v33.NobleUSDCDenom, poolLiquidity), Weight: osmomath.NewInt(1)},
		{Token: sdk.NewCoin("uion", poolLiquidity), Weight: osmomath.NewInt(3)},
	}, balancer.PoolParams{SwapFee: osmomath.ZeroDec(), ExitFee: osmomath.ZeroDec()})

	result, err := v33.MigrateNobleUSDCPools(s.Ctx, &s.App.AppKeepers, poolMigration(transmuter, alloyed, nil, []uint64{poolId}, nil))
	s.Require().NoError(err)
	s.Require().Equal(1, result.Converted)
	s.Require().Equal(poolLiquidity, result.ConvertedAmount)

	pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
	s.Require().NoError(err)
	assets := pool.(*balancer.Pool).PoolAssets
	s.Require().Len(assets, 2)
	// allUSDC is a factory denom, which sorts before uion.
	s.Require().Equal(alloyed, assets[0].Token.Denom)
	s.Require().Equal(poolLiquidity, assets[0].Token.Amount)
	s.Require().Equal(osmomath.NewInt(1), assets[0].Weight.QuoRaw(balancer.GuaranteedWeightPrecision))
	s.Require().Equal("uion", assets[1].Token.Denom)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, pool.GetAddress(), v33.NobleUSDCDenom).IsZero())

	shares := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], "gamm/pool/"+osmomath.NewIntFromUint64(poolId).String())
	exited, err := s.App.GAMMKeeper.ExitPool(s.Ctx, s.TestAccs[0], poolId, shares.Amount.QuoRaw(2), sdk.Coins{})
	s.Require().NoError(err)
	s.Require().True(exited.AmountOf(alloyed).IsPositive())
	s.Require().True(exited.AmountOf(v33.NobleUSDCDenom).IsZero())
}

// TestMigrateNobleUSDCPoolsDrainsCLPool checks that a drained pool's positions are withdrawn to their
// owners, apart from a contract's, and that the holder migration then converts the owners' Noble USDC.
func (s *UpgradeTestSuite) TestMigrateNobleUSDCPoolsDrainsCLPool() {
	transmuter, alloyed := s.prepareAlloyTransmuter(v33.NobleUSDCDenom)
	poolId := s.nobleCLPool()
	lps := []sdk.AccAddress{s.newUser(), s.newUser()}
	for _, lp := range lps {
		s.fullRangePosition(poolId, lp)
	}
	contract := sdk.MustAccAddressFromBech32(s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{"ibc/other", otherAlloyDenom}, []uint16{1, 1}).GetContractAddress())
	contractPosition := s.fullRangePosition(poolId, contract)

	result, err := v33.MigrateNobleUSDCPools(s.Ctx, &s.App.AppKeepers, poolMigration(transmuter, alloyed, nil, nil, []uint64{poolId}))
	s.Require().NoError(err)
	s.Require().Equal(1, result.Drained)
	s.Require().Equal(2, result.PositionsWithdrawn)
	s.Require().Equal(1, result.PositionsKept)

	for _, lp := range lps {
		positions, err := s.App.ConcentratedLiquidityKeeper.GetUserPositions(s.Ctx, lp, poolId)
		s.Require().NoError(err)
		s.Require().Empty(positions)
		s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, lp, v33.NobleUSDCDenom).IsPositive())
	}
	_, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, contractPosition)
	s.Require().NoError(err, "a contract's position is kept")

	// The holder migration runs after the pool migration and converts what the LPs got back.
	holders := s.migrateAlloy(v33.NobleUSDCDenom, transmuter, alloyed, nil, nil)
	s.Require().GreaterOrEqual(holders.Migrated, len(lps))
	for _, lp := range lps {
		s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, lp, v33.NobleUSDCDenom).IsZero())
		s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, lp, alloyed).IsPositive())
	}
}

// TestMigrateNobleUSDCPoolsSkipsUnexpectedPools checks that listed pools that do not hold Noble USDC,
// have the wrong type, have too many positions or do not exist are left unchanged.
func (s *UpgradeTestSuite) TestMigrateNobleUSDCPoolsSkipsUnexpectedPools() {
	transmuter, alloyed := s.prepareAlloyTransmuter(v33.NobleUSDCDenom)
	otherCLPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], "uion", appparams.BaseCoinUnit, 1, osmomath.ZeroDec()).GetId()
	nobleCLPool := s.nobleCLPool()
	crowded := s.nobleCLPool()
	for i := 0; i < 11; i++ {
		s.fullRangePosition(crowded, s.newUser())
	}
	const missingPool = 999_999

	m := poolMigration(transmuter, alloyed,
		[]uint64{otherCLPool, missingPool},
		[]uint64{nobleCLPool},
		[]uint64{otherCLPool, crowded, missingPool})
	result, err := v33.MigrateNobleUSDCPools(s.Ctx, &s.App.AppKeepers, m)
	s.Require().NoError(err)
	s.Require().Equal(0, result.Converted)
	s.Require().Equal(0, result.Drained)
	s.Require().Equal(6, result.SkippedPools)

	pool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, nobleCLPool)
	s.Require().NoError(err)
	s.Require().Equal(v33.NobleUSDCDenom, pool.GetToken0())
	positions, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdsForPool(s.Ctx, crowded, 100)
	s.Require().NoError(err)
	s.Require().Len(positions, 11)
}

// TestMigrateNobleUSDCPoolsPreflight checks that nothing changes when the transmuter fails the
// pre-flight check.
func (s *UpgradeTestSuite) TestMigrateNobleUSDCPoolsPreflight() {
	_, alloyed := s.prepareAlloyTransmuter(v33.NobleUSDCDenom)
	poolId := s.nobleCLPool()
	s.fullRangePosition(poolId, s.newUser())

	notAContract := s.TestAccs[1].String()
	_, err := v33.MigrateNobleUSDCPools(s.Ctx, &s.App.AppKeepers, poolMigration(notAContract, alloyed, []uint64{poolId}, nil, []uint64{poolId}))
	s.Require().ErrorContains(err, "is not a contract")

	pool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolId)
	s.Require().NoError(err)
	s.Require().Equal(v33.NobleUSDCDenom, pool.GetToken0())
	positions, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdsForPool(s.Ctx, poolId, 100)
	s.Require().NoError(err)
	s.Require().Len(positions, 1)
}

// TestMigrateNobleUSDCPoolsRemovesFeeToken checks that Noble USDC stops being a fee token when its fee
// pool is drained, and stays one when its fee pool is not migrated.
func (s *UpgradeTestSuite) TestMigrateNobleUSDCPoolsRemovesFeeToken() {
	for _, drainFeePool := range []bool{true, false} {
		s.Run(map[bool]string{true: "fee pool drained", false: "fee pool not migrated"}[drainFeePool], func() {
			s.SetupTest()
			transmuter, alloyed := s.prepareAlloyTransmuter(v33.NobleUSDCDenom)
			baseDenom, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
			s.Require().NoError(err)
			feePool := s.nobleCLPoolWith(baseDenom)
			s.fullRangePositionWith(feePool, s.newUser(), baseDenom)
			s.Require().NoError(s.App.TxFeesKeeper.SetFeeTokens(s.Ctx, []txfeestypes.FeeToken{{Denom: v33.NobleUSDCDenom, PoolID: feePool}}))

			drain := []uint64{}
			if drainFeePool {
				drain = append(drain, feePool)
			}
			result, err := v33.MigrateNobleUSDCPools(s.Ctx, &s.App.AppKeepers, poolMigration(transmuter, alloyed, nil, nil, drain))
			s.Require().NoError(err)
			s.Require().Equal(drainFeePool, result.FeeTokenRemoved)
			_, err = s.App.TxFeesKeeper.GetFeeToken(s.Ctx, v33.NobleUSDCDenom)
			s.Require().Equal(drainFeePool, err != nil)
		})
	}
}

// TestNobleUSDCPoolListsAreDisjoint checks that no mainnet pool is both converted and drained.
func (s *UpgradeTestSuite) TestNobleUSDCPoolListsAreDisjoint() {
	m := v33.MainnetNobleUSDCPoolMigration()
	seen := map[uint64]bool{}
	for _, list := range [][]uint64{m.CLConvertPools, m.BalancerConvertPools, m.CLDrainPools} {
		for _, id := range list {
			s.Require().False(seen[id], "pool %d is listed twice", id)
			seen[id] = true
		}
	}
	s.Require().Len(seen, 606)
}
