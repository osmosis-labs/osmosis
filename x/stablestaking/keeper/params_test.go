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

type ParamsTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (s *ParamsTestSuite) SetupTest() {
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

func (s *ParamsTestSuite) TestGetSetParams() {
	// Get initial params
	params := s.App.StableStakingKeeper.GetParams(s.Ctx)
	require.NotNil(s.T(), params)

	// Modify params
	newParams := types.Params{
		RewardEpochIdentifier:    "day",
		UnbondingEpochIdentifier: "day",
		RewardRate:               "0.4",
		SupportedTokens:          []string{"uusd"},
		UnbondingDuration:        time.Hour * 24 * 7,
		MaxStakingAmount:         math.LegacyNewDecFromInt(math.NewInt(1000)),
	}

	// Set new params
	s.App.StableStakingKeeper.SetParams(s.Ctx, newParams)

	updatedParams := s.App.StableStakingKeeper.GetParams(s.Ctx)
	require.Equal(s.T(), "day", updatedParams.RewardEpochIdentifier)
	require.Equal(s.T(), newParams.RewardRate, updatedParams.RewardRate)
	require.Equal(s.T(), newParams.UnbondingDuration, updatedParams.UnbondingDuration)
	require.Equal(s.T(), math.LegacyDec{}, updatedParams.MaxStakingAmount)
}

func (s *ParamsTestSuite) TestParamValidationInUnbonding() {
	// Set unbonding period
	params := s.App.StableStakingKeeper.GetParams(s.Ctx)
	params.UnbondingDuration = time.Hour * 24 * 7
	s.App.StableStakingKeeper.SetParams(s.Ctx, params)

	staker, err := sdk.AccAddressFromBech32("symphony1cvtrs9jhacf0p7xlmeq0ejhq83udmcqx40nyg9")
	require.NoError(s.T(), err)

	// Stake some tokens
	token := sdk.NewCoin(assets.MicroUSDDenom, math.NewInt(100))
	_, err = s.App.StableStakingKeeper.StakeTokens(s.Ctx, staker, token)
	require.NoError(s.T(), err)

	// Unbond tokens
	unbondAmount := math.NewInt(50)
	unbonding, err := s.App.StableStakingKeeper.UnStakeTokens(s.Ctx, staker, sdk.NewCoin(assets.MicroUSDDenom, unbondAmount))
	require.NoError(s.T(), err)

	// Verify unbonding info
	require.Equal(s.T(), staker.String(), unbonding.Staker)
	require.Equal(s.T(), unbondAmount.String(), unbonding.Amount.Amount.TruncateInt().String())
	require.Equal(s.T(), "50", unbonding.TotalStaked.TruncateInt().String())
	require.Equal(s.T(), "50", unbonding.TotalShares.TruncateInt().String())
}
