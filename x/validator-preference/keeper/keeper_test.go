package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/validator-preference/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	cleanup func()
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *KeeperTestSuite) Cleanup() {
	suite.cleanup()
}

func (suite *KeeperTestSuite) SetupValidators(bondStatuses []stakingtypes.BondStatus) []sdk.ValAddress {
	valAddrs := []sdk.ValAddress{}
	for _, status := range bondStatuses {
		valAddr := suite.SetupValidator(status)
		valAddrs = append(valAddrs, valAddr)
	}
	return valAddrs
}

// SetupMultipleValidators setups "numValidator" validators and returns their address in string
func (suite *KeeperTestSuite) SetupMultipleValidators(numValidator int) []string {
	valAddrs := []string{}
	for i := 0; i < numValidator; i++ {
		valAddr := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})
		valAddrs = append(valAddrs, valAddr[0].String())
	}
	return valAddrs
}

func (suite *KeeperTestSuite) PrepareStakeToValidatorSet() []types.ValidatorPreference {
	valAddrs := suite.SetupMultipleValidators(3)
	valPreferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         sdk.NewDecWithPrec(5, 1),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         sdk.NewDecWithPrec(3, 1),
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         sdk.NewDecWithPrec(2, 1),
		},
	}
	return valPreferences
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
