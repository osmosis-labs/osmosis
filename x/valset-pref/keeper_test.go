package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/valset-pref/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

// SetupMultipleValidators setups "numValidator" validators and returns their address in string
func (suite *KeeperTestSuite) SetupMultipleValidators(numValidator int) []string {
	valAddrs := []string{}
	for i := 0; i < numValidator; i++ {
		valAddr := suite.SetupValidator(stakingtypes.Bonded)
		valAddrs = append(valAddrs, valAddr.String())
	}
	return valAddrs
}

func (suite *KeeperTestSuite) PrepareDelegateToValidatorSet() []types.ValidatorPreference {
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
