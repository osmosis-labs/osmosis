package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v8/app/apptesting"
	"github.com/osmosis-labs/osmosis/v8/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v8/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v8/x/gamm/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func (suite *KeeperTestSuite) prepareCustomBalancerPool(
	balances sdk.Coins,
	poolAssets []balancertypes.PoolAsset,
	poolParams balancer.PoolParams,
) uint64 {
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, balances)
	}

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}
