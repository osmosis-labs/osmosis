package swaprouter_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

const testExpectedPoolId = 3

var testPoolCreationFee = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)}

func TestKeeperTestSuite(t *testing.T) {

	// TODO: re-enable this once swaprouter is fully merged.
	t.SkipNow()

	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

// createPoolFromType creates a basic pool of the given type for testing.
func (suite *KeeperTestSuite) createPoolFromType(poolType types.PoolType) {
	switch poolType {
	case types.Balancer:
		suite.PrepareBalancerPool()
		return
	case types.StableSwap:
		// TODO
		return
	case types.Concentrated:
		// TODO
		return
	}
}

// createBalancerPoolsFromCoins creates balancer pools from given sets of coins.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (suite *KeeperTestSuite) createBalancerPoolsFromCoins(poolCoins []sdk.Coins) {
	for _, curPoolCoins := range poolCoins {
		suite.FundAcc(suite.TestAccs[0], curPoolCoins)
		suite.PrepareBalancerPoolWithCoins(curPoolCoins...)
	}
}
