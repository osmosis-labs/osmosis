package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
)

type HooksTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

func (s *HooksTestSuite) SetupTest() {
	s.Setup()
	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroUSDDenom, sdrPriceInMelody)

	// Mint initial tokens
	totalUsddSupply := sdk.NewCoins(sdk.NewCoin(assets.MicroUSDDenom, InitTokens.MulRaw(int64(len(Addr)*10))))
	err := s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalUsddSupply)
	s.Require().NoError(err)

	// Fund test accounts
	staker, err := sdk.AccAddressFromBech32("symphony137jfmwnjgzuy4fvd60mmg50uyfye877q56uca6")
	require.NoError(s.T(), err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker, InitUSDDCoins)
	s.Require().NoError(err)

	s.App.StableStakingKeeper.SetParams(s.Ctx, types.DefaultParams())
}

func (s *HooksTestSuite) TestAfterEpochEnd() {
	staker, err := sdk.AccAddressFromBech32("symphony137jfmwnjgzuy4fvd60mmg50uyfye877q56uca6")
	require.NoError(s.T(), err)

	s.Run("no action on wrong epoch identifier", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)

		// Call AfterEpochEnd with the wrong identifier
		err = s.App.StableStakingKeeper.AfterEpochEnd(s.Ctx, "wrong_epoch", 1)
		require.NoError(s.T(), err)

		// Verify no snapshot was taken
		snapshot, err := s.App.StableStakingKeeper.GetEpochSnapshot(s.Ctx, assets.MicroUSDDenom)
		require.Error(s.T(), err, "epoch snapshot not found")
		require.Equal(s.T(), snapshot, types.EpochSnapshot{})
	})

	s.Run("take snapshot and distribute rewards", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)

		// Call AfterEpochEnd with correct identifier
		params := s.App.StableStakingKeeper.GetParams(s.Ctx)
		err = s.App.StableStakingKeeper.AfterEpochEnd(s.Ctx, params.RewardEpochIdentifier, 1)
		require.NoError(s.T(), err)

		// Verify snapshot was taken
		snapshot, err := s.App.StableStakingKeeper.GetEpochSnapshot(s.Ctx, assets.MicroUSDDenom)
		require.False(s.T(), snapshot.TotalShares.IsZero())
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(200)), snapshot.TotalShares)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(200)), snapshot.TotalStaked)
		require.Len(s.T(), snapshot.Stakers, 1)
		require.Equal(s.T(), staker.String(), snapshot.Stakers[0].Address)
		require.Equal(s.T(), math.LegacyNewDecFromInt(math.NewInt(100)), snapshot.Stakers[0].Shares)

		// Move to the next epoch
		err = s.App.StableStakingKeeper.AfterEpochEnd(s.Ctx, params.RewardEpochIdentifier, 2)
		require.NoError(s.T(), err)

		// Verify rewards were distributed
		balance := s.App.BankKeeper.GetBalance(s.Ctx, staker, assets.MicroUSDDenom)
		require.True(s.T(), balance.Amount.GT(math.NewInt(100))) // Should have received rewards
	})
}

func (s *HooksTestSuite) TestGetEpochReward() {
	staker, err := sdk.AccAddressFromBech32("symphony137jfmwnjgzuy4fvd60mmg50uyfye877q56uca6")
	require.NoError(s.T(), err)

	s.Run("calculate rewards correctly", func() {
		// First stake some tokens
		token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(1000))
		_, err := s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
		require.NoError(s.T(), err)

		// Get epoch reward
		reward := s.App.StableStakingKeeper.GetEpochReward(s.Ctx)
		require.True(s.T(), reward.IsPositive())

		// Verify reward calculation
		params := s.App.StableStakingKeeper.GetParams(s.Ctx)
		rewardRate, err := osmomath.NewDecFromStr(params.RewardRate)
		require.NoError(s.T(), err)
		expectedReward := math.LegacyNewDecFromInt(math.NewInt(1000)).Mul(rewardRate).TruncateInt()
		require.Equal(s.T(), expectedReward, reward)
	})

	s.Run("zero reward for no stakes", func() {
		reward := s.App.StableStakingKeeper.GetEpochReward(s.Ctx)
		require.True(s.T(), reward.IsZero())
	})
}

func (s *HooksTestSuite) TestDistributeRewardsToLastEpochStakers() {
	staker1, err := sdk.AccAddressFromBech32("symphony137jfmwnjgzuy4fvd60mmg50uyfye877q56uca6")
	require.NoError(s.T(), err)
	staker2, err := sdk.AccAddressFromBech32("symphony1cvtrs9jhacf0p7xlmeq0ejhq83udmcqx40nyg9")
	require.NoError(s.T(), err)

	s.Run("distribute rewards proportionally", func() {
		// Fund second staker
		err := s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, staker2, InitUSDDCoins)
		require.NoError(s.T(), err)

		// First staker stakes 1000 tokens
		token1 := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(1000))
		_, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker1, token1)
		require.NoError(s.T(), err)

		// Second staker stakes 2000 tokens
		token2 := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(2000))
		_, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker2, token2)
		require.NoError(s.T(), err)

		// Take snapshot
		params := s.App.StableStakingKeeper.GetParams(s.Ctx)
		err = s.App.StableStakingKeeper.AfterEpochEnd(s.Ctx, params.RewardEpochIdentifier, 1)
		require.NoError(s.T(), err)

		// Calculate total reward
		totalReward := s.App.StableStakingKeeper.GetEpochReward(s.Ctx)

		// Distribute rewards
		s.App.StableStakingKeeper.DistributeRewardsToLastEpochStakers(s.Ctx)

		// Verify rewards were distributed proportionally
		balance1 := s.App.BankKeeper.GetBalance(s.Ctx, staker1, assets.MicroUSDDenom)
		balance2 := s.App.BankKeeper.GetBalance(s.Ctx, staker2, assets.MicroUSDDenom)

		// The First staker should get 1/3 of rewards
		expectedReward1 := totalReward.QuoRaw(3)
		// Second staker should get 2/3 of rewards
		expectedReward2 := totalReward.MulRaw(2).QuoRaw(3)

		require.Equal(s.T(), expectedReward1, balance1.Amount.Sub(token1.Amount))
		require.Equal(s.T(), expectedReward2, balance2.Amount.Sub(token2.Amount))
	})
}
