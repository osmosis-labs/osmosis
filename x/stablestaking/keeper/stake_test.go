package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

var (
	pubKey = secp256k1.GenPrivKey().PubKey()
	Addr   = sdk.AccAddress(pubKey.Address())

	InitTokens    = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitUSDDCoins = sdk.NewCoins(sdk.NewCoin(assets.MicroUSDDenom, InitTokens))
	InitUSDRCoins = sdk.NewCoins(sdk.NewCoin(assets.MicroHKDDenom, InitTokens))

	FaucetAccountName = tokenfactorytypes.ModuleName
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestStakeTokens() {
	s.Setup()
	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroUSDDenom, sdrPriceInMelody)

	totalUsddSupply := sdk.NewCoins(sdk.NewCoin(assets.MicroUSDDenom, InitTokens.MulRaw(int64(len(Addr)*10))))
	err := s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalUsddSupply)
	s.Require().NoError(err)

	totalSdrSupply := sdk.NewCoins(sdk.NewCoin(assets.MicroHKDDenom, InitTokens.MulRaw(int64(len(Addr)*10))))
	err = s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalSdrSupply)
	s.Require().NoError(err)

	staker, err := sdk.AccAddressFromBech32("symphony1cvtrs9jhacf0p7xlmeq0ejhq83udmcqx40nyg9")
	require.NoError(s.T(), err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker, InitUSDRCoins)
	s.Require().NoError(err)

	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker, InitUSDDCoins)
	s.Require().NoError(err)

	staker2, err := sdk.AccAddressFromBech32("symphony137jfmwnjgzuy4fvd60mmg50uyfye877q56uca6")
	require.NoError(s.T(), err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker2, InitUSDRCoins)
	s.Require().NoError(err)

	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker2, InitUSDDCoins)
	s.Require().NoError(err)

	stakerBaseBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)

	s.Run("fail on unsupported token", func() {
		unsupportedToken := sdk.NewCoin("stable1", math.NewInt(100000))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, unsupportedToken)
		require.NotNil(s.T(), err)
		require.Equal(s.T(), "unsupported token: stable1", err.Error())
	})

	s.Run("stake token successfully", func() {
		// Stake for the first time
		token1 := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		resp, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token1)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp)

		// Verify staking pool
		pool, found := s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(token1.Amount), pool.TotalShares)
		require.Equal(s.T(), math.LegacyNewDecFromInt(token1.Amount), pool.TotalStaked)

		// Verify user stake
		userStake, found := s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDec(token1.Amount.Int64()), userStake.Shares)
		require.Equal(s.T(), s.App.EpochsKeeper.GetEpochInfo(s.Ctx, "week").CurrentEpoch, userStake.Epoch)

		// check the balance of the MicroUSDDenom in module
		moduleAddr := s.App.AccountKeeper.GetModuleAddress(types.ModuleName)
		moduleBaseDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroUSDDenom)

		// check that the total osmo amount has been transferred to a module account
		s.Equal(moduleBaseDenomBalance.Amount.String(), token1.Amount.String())

		// stake additional tokens
		amount1 := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(600))
		resp, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, amount1)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp)

		amount2 := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(300))
		resp, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker2, amount2)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp)

		// Verify updated staking pool
		pool, found = s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(1000)), pool.TotalShares) // 100 + 600 + 300
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(1000)), pool.TotalStaked) // 100 + 600 + 300

		// Verify updated user stake
		staker1Stake, found := s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(700)), staker1Stake.Shares) // 100 + 600

		staker2Stake, found := s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker2, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(300)), staker2Stake.Shares) // 100 + 600

		// get total stake of staker1
		totalStake := s.App.StableStakingKeeper.GetUserTotalStake(s.Ctx, staker)
		require.Equal(s.T(), sdk.DecCoins{
			{Denom: assets.MicroUSDDenom, Amount: math.LegacyNewDec(700)},
		}, totalStake) // 100 + 600

		// check the balance of the MicroUSDDenom in module
		moduleBaseDenomBalance = s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroUSDDenom)
		s.Equal(moduleBaseDenomBalance.Amount.String(), "1000")

		//  unstake tokens
		resp1, err := s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, token1)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp1)

		// Verify staking pool
		pool, found = s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDec(900), pool.TotalShares)
		require.Equal(s.T(), math.LegacyNewDec(900), pool.TotalStaked)

		// Verify user stake
		userStake, found = s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDec(600), userStake.Shares)
		require.Equal(s.T(), s.App.EpochsKeeper.GetEpochInfo(s.Ctx, "week").CurrentEpoch, userStake.Epoch)

		// check unbonding
		userUnbondInfo, found := s.App.StableStakingKeeper.GetUnbondingInfo(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), types.UnbondingInfo{
			Address:     staker.String(),
			Amount:      math.LegacyNewDec(100),
			Denom:       assets.MicroUSDDenom,
			UnbondEpoch: math.NewInt(14).Int64(),
		}, userUnbondInfo)

		// check user balance
		stakerBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		s.Equal(stakerBalance.Amount.String(), stakerBaseBalance.Amount.Sub(math.NewInt(700)).String())

		// check module balance
		moduleBaseDenomBalance = s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroUSDDenom)
		s.Equal(moduleBaseDenomBalance.Amount.String(), "1000")

		//  full unstake tokens
		resp1, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, amount1)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp1)

		// Verify staking pool
		pool, found = s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDec(300), pool.TotalShares)
		require.Equal(s.T(), math.LegacyNewDec(300), pool.TotalStaked)

		// Verify user stake
		userStake, found = s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDec(0), userStake.Shares)

		// check unbonding
		userUnbondInfo, found = s.App.StableStakingKeeper.GetUnbondingInfo(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), types.UnbondingInfo{
			Address:     staker.String(),
			Amount:      math.LegacyNewDec(700),
			Denom:       assets.MicroUSDDenom,
			UnbondEpoch: math.NewInt(14).Int64(),
		}, userUnbondInfo)

		// get total unbondings
		userUnbondTotalInfo := s.App.StableStakingKeeper.GetUnbondingTotalInfo(s.Ctx, staker)
		require.Equal(s.T(), []types.UnbondingInfo{
			{
				Address:     staker.String(),
				Amount:      math.LegacyNewDec(700),
				Denom:       assets.MicroUSDDenom,
				UnbondEpoch: math.NewInt(14).Int64(),
			},
		}, userUnbondTotalInfo)

		// check user balance
		stakerBalance = s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		s.Equal(stakerBalance.Amount.String(), stakerBaseBalance.Amount.Sub(math.NewInt(700)).String())

		// check module balance
		moduleBaseDenomBalance = s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroUSDDenom)
		s.Equal(moduleBaseDenomBalance.Amount.String(), "1000")
	})

	s.Run("stake multiple token successfully", func() {
		// Stake for the first time
		amountUsd := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		resp, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, amountUsd)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp)

		// Verify staking pool
		poolUsd, found := s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(400)), poolUsd.TotalShares)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(400)), poolUsd.TotalStaked)

		// Verify user stake
		stakerUsdStake, found := s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(100)), stakerUsdStake.Shares)
		require.Equal(s.T(), s.App.EpochsKeeper.GetEpochInfo(s.Ctx, "week").CurrentEpoch, stakerUsdStake.Epoch)

		// check the balance of the MicroUSDDenom in module
		moduleAddr := s.App.AccountKeeper.GetModuleAddress(types.ModuleName)
		moduleUsdBalance := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroUSDDenom)

		// check that the total osmo amount has been transferred to a module account
		s.Equal("1100", moduleUsdBalance.Amount.String())

		amountSdr := sdk.NewCoin(assets.MicroHKDDenom, math.NewInt(100))
		resp, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, amountSdr)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp)

		// Verify staking pool
		poolSdr, found := s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroHKDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(amountSdr.Amount), poolSdr.TotalShares)
		require.Equal(s.T(), math.LegacyNewDecFromInt(amountSdr.Amount), poolSdr.TotalStaked)

		// Verify user stake
		userSdrStake, found := s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroHKDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDec(amountSdr.Amount.Int64()), userSdrStake.Shares)
		require.Equal(s.T(), s.App.EpochsKeeper.GetEpochInfo(s.Ctx, "week").CurrentEpoch, userSdrStake.Epoch)

		// check the balance of the MicroUSDDenom in module
		moduleSdrBalance := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroHKDDenom)

		// check that the total osmo amount has been transferred to a module account
		s.Equal(moduleSdrBalance.Amount.String(), amountSdr.Amount.String())

		// stake additional tokens
		amount1 := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(600))
		resp, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, amount1)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp)

		amount2 := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(300))
		resp, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker2, amount2)
		require.Nil(s.T(), err)
		require.NotNil(s.T(), resp)

		// Verify updated staking pool
		poolUsd, found = s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(1300)), poolUsd.TotalShares)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(1300)), poolUsd.TotalStaked)

		// Verify updated user stake
		stakerUsdStake, found = s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(700)), stakerUsdStake.Shares)

		staker2UsdStake, found := s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker2, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(600)), staker2UsdStake.Shares)

		// check the balance of the MicroUSDDenom in module
		moduleUsdBalance = s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroUSDDenom)
		s.Equal(moduleUsdBalance.Amount.String(), "2000")

		moduleSdrBalance = s.App.BankKeeper.GetBalance(s.Ctx, moduleAddr, assets.MicroHKDDenom)
		s.Equal(moduleSdrBalance.Amount.String(), "100")

		// check the total stake
		stakerStakes := s.App.StableStakingKeeper.GetUserTotalStake(s.Ctx, staker)
		require.Equal(s.T(),
			sdk.DecCoins{
				sdk.NewDecCoin("ukhd", math.NewInt(100)),
				sdk.NewDecCoin("uusd", math.NewInt(700)),
			},
			stakerStakes,
		)

		// check staker2 total stake
		staker2Stakes := s.App.StableStakingKeeper.GetUserTotalStake(s.Ctx, staker2)
		require.Equal(s.T(),
			sdk.DecCoins{
				sdk.NewDecCoin("uusd", math.NewInt(600)),
			},
			staker2Stakes,
		)

		// check staked pools
		pools := s.App.StableStakingKeeper.GetPools(s.Ctx)
		require.Equal(
			s.T(),
			[]types.StakingPool{
				{
					Denom:       "ukhd",
					TotalStaked: math.LegacyNewDecFromInt(math.NewInt(100)),
					TotalShares: math.LegacyNewDecFromInt(math.NewInt(100)),
				},
				{
					Denom:       "uusd",
					TotalStaked: math.LegacyNewDecFromInt(math.NewInt(1300)),
					TotalShares: math.LegacyNewDecFromInt(math.NewInt(1300)),
				}},
			pools,
		)

	})
}
