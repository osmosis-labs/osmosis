package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *app.OsmosisApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.SuperfluidKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) SetupDefaultPool() {
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	poolId := suite.createGammPool([]string{bondDenom, "foo"})
	suite.Require().Equal(poolId, uint64(1))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
