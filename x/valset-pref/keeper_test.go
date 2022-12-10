package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *KeeperTestSuite) PrepareDelegateToValidatorSet() []types.ValidatorPreference {
	valAddrs := suite.SetupMultipleValidators(4)
	valPreferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         sdk.NewDecWithPrec(2, 1), // 0.2
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         sdk.NewDecWithPrec(332, 3), // 0.332
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         sdk.NewDecWithPrec(12, 2), // 0.12
		},
		{
			ValOperAddress: valAddrs[3],
			Weight:         sdk.NewDecWithPrec(348, 3), // 0.348
		},
	}

	return valPreferences
}

func (suite *KeeperTestSuite) GetDelegationRewards(ctx sdk.Context, val types.ValidatorPreference, delegator sdk.AccAddress) (sdk.DecCoins, stakingtypes.Validator) {
	valAddr, err := sdk.ValAddressFromBech32(val.ValOperAddress)
	suite.Require().NoError(err)

	validator, found := suite.App.StakingKeeper.GetValidator(ctx, valAddr)
	suite.Require().True(found)

	endingPeriod := suite.App.DistrKeeper.IncrementValidatorPeriod(ctx, validator)

	delegation, found := suite.App.StakingKeeper.GetDelegation(ctx, delegator, valAddr)
	suite.Require().True(found)

	rewards := suite.App.DistrKeeper.CalculateDelegationRewards(ctx, validator, delegation, endingPeriod)

	return rewards, validator
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
