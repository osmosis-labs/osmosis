package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
)

type UnbondingTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUnbondingTestSuite(t *testing.T) {
	suite.Run(t, new(UnbondingTestSuite))
}

func (s *UnbondingTestSuite) SetupTest() {
	s.Setup()
	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroUSDDenom, sdrPriceInMelody)

	// Mint initial tokens
	totalUsddSupply := sdk.NewCoins(sdk.NewCoin(assets.MicroUSDDenom, InitTokens.MulRaw(int64(len(Addr)*10))))
	err := s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalUsddSupply)
	s.Require().NoError(err)

	// Fund test accounts
	staker, err := sdk.AccAddressFromBech32("symphony1cvtrs9jhacf0p7xlmeq0ejhq83udmcqx40nyg9")
	require.NoError(s.T(), err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker, InitUSDDCoins)
	s.Require().NoError(err)
}

func (s *UnbondingTestSuite) TestUnbondTokens() {
	staker, err := sdk.AccAddressFromBech32("symphony1cvtrs9jhacf0p7xlmeq0ejhq83udmcqx40nyg9")
	require.NoError(s.T(), err)

	s.Run("fail on non-existent stake", func() {
		_, err := s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100)))
		require.Error(s.T(), err)
		require.Contains(s.T(), err.Error(), "not found pool for denom uusd")
	})

	s.Run("fail on insufficient stake", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)

		// Try to unbond more than staked
		_, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(200)))
		require.Error(s.T(), err)
		require.Contains(s.T(), err.Error(), "unstake amount exceeds user's share: 100.000000000000000000")
	})

	s.Run("successful unbonding", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)

		// Unbond half of the stake
		unbondAmount := math.NewInt(50)
		unbondingInfo, err := s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount))
		require.NoError(s.T(), err)
		require.NotNil(s.T(), unbondingInfo)

		// Verify unbonding info
		require.Equal(s.T(), staker.String(), unbondingInfo.Staker)
		require.Equal(s.T(), assets.MicroUSDDenom, unbondingInfo.Amount.Denom)
		require.Equal(s.T(), unbondAmount.String(), unbondingInfo.Amount.Amount.TruncateInt().String())

		// Verify user stake is reduced
		userStake, found := s.App.StableStakingKeeper.GetUserStake(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(150)), userStake.Shares)

		// Verify pool is updated
		pool, found := s.App.StableStakingKeeper.GetPool(s.Ctx, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(150)), pool.TotalShares)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(150)), pool.TotalStaked)
	})
}

func (s *UnbondingTestSuite) TestCompleteUnbonding() {
	staker, err := sdk.AccAddressFromBech32("symphony1cvtrs9jhacf0p7xlmeq0ejhq83udmcqx40nyg9")
	require.NoError(s.T(), err)
	s.Run("complete unbonding after period", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)
		startTime := time.Now()
		s.Ctx = s.Ctx.WithBlockTime(startTime)

		// Unbond tokens
		unbondAmount := math.NewInt(50)
		_, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount))
		require.NoError(s.T(), err)

		// Get initial balances
		initialBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		moduleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), assets.MicroUSDDenom)

		// Complete unbonding
		unbondingEpoch := s.App.StableStakingKeeper.GetParams(s.Ctx).UnbondingDuration.Milliseconds() / 1000 / 60 / 60 / 24
		err = s.App.StableStakingKeeper.AfterEpochEnd(s.Ctx, "day", unbondingEpoch)
		s.Require().NoError(err)

		// Verify unbonding info is removed
		unbondInfo, found := s.App.StableStakingKeeper.GetUnbondingInfo(s.Ctx, staker, assets.MicroUSDDenom)
		require.False(s.T(), found)
		require.Equal(s.T(), unbondInfo, types.UnbondingInfo{})

		// Verify tokens are returned to user
		finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		require.Equal(s.T(), initialBalance.Amount.Add(unbondAmount), finalBalance.Amount)

		// Verify module balance is reduced
		finalModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), assets.MicroUSDDenom)
		require.Equal(s.T(), moduleBalance.Amount.Sub(unbondAmount), finalModuleBalance.Amount)
	})

	s.Run("multiple unbonding requests", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(200))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)

		// Create multiple unbonding requests
		unbondAmount1 := math.NewInt(50)
		unbondAmount2 := math.NewInt(75)
		_, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount1))
		require.NoError(s.T(), err)
		_, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount2))
		require.NoError(s.T(), err)

		// Get initial balances
		initialBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		moduleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), assets.MicroUSDDenom)

		// Move time forward past an unbonding period
		unbondingDuration := s.App.StableStakingKeeper.GetParams(s.Ctx).UnbondingDuration
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(unbondingDuration + time.Hour))

		// Complete unbonding
		unbondingEpoch := s.App.StableStakingKeeper.GetParams(s.Ctx).UnbondingDuration.Milliseconds() / 1000 / 60 / 60 / 24
		err = s.App.StableStakingKeeper.AfterEpochEnd(s.Ctx, "day", unbondingEpoch)
		s.Require().NoError(err)

		// Verify unbonding info is removed
		_, found := s.App.StableStakingKeeper.GetUnbondingInfo(s.Ctx, staker, assets.MicroUSDDenom)
		require.False(s.T(), found)

		// Verify total tokens are returned to user
		finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		totalUnbonded := unbondAmount1.Add(unbondAmount2)
		require.Equal(s.T(), initialBalance.Amount.Add(totalUnbonded), finalBalance.Amount)

		// Verify module balance is reduced
		finalModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), assets.MicroUSDDenom)
		require.Equal(s.T(), moduleBalance.Amount.Sub(totalUnbonded), finalModuleBalance.Amount)
	})

	s.Run("unbonding not completed before period", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)

		// Unbond tokens
		unbondAmount := math.NewInt(50)
		_, err = s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount))
		require.NoError(s.T(), err)

		// Get initial balances
		initialBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		moduleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), assets.MicroUSDDenom)

		// Move time forward but not past unbonding period
		unbondingDuration := s.App.StableStakingKeeper.GetParams(s.Ctx).UnbondingDuration
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(unbondingDuration - time.Hour))

		// Complete unbonding
		unbondingEpoch := unbondingDuration.Milliseconds() / 1000 / 60 / 60 / 24
		err = s.App.StableStakingKeeper.AfterEpochEnd(s.Ctx, "day", unbondingEpoch-1)
		s.Require().NoError(err)

		// Verify unbonding info still exists
		unbondInfo, found := s.App.StableStakingKeeper.GetUnbondingInfo(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), found)
		require.Equal(s.T(), unbondAmount, unbondInfo.Amount.TruncateInt())

		// Verify balances haven't changed
		finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		require.Equal(s.T(), initialBalance.Amount, finalBalance.Amount)

		finalModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), assets.MicroUSDDenom)
		require.Equal(s.T(), moduleBalance.Amount, finalModuleBalance.Amount)
	})
}
