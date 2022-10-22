package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
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
	suite.fundAllAccountsWith(balances)

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

func (suite *KeeperTestSuite) prepareCustomStableswapPool(
	balances sdk.Coins,
	poolParams stableswap.PoolParams,
	initialLiquidity sdk.Coins,
	scalingFactors []uint64,
) uint64 {
	suite.fundAllAccountsWith(balances)

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		stableswap.NewMsgCreateStableswapPool(suite.TestAccs[0], poolParams, initialLiquidity, scalingFactors, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

func (suite *KeeperTestSuite) fundAllAccountsWith(balances sdk.Coins) {
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, balances)
	}
}
