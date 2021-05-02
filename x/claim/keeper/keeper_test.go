package keeper_test

import (
	"testing"
	"time"

	"github.com/c-osmosis/osmosis/app"
	"github.com/c-osmosis/osmosis/x/claim/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	querier sdk.Querier
	app     *app.OsmosisApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	suite.app.ClaimKeeper.SetModuleAccountBalance(suite.ctx, sdk.NewCoins(sdk.NewCoin(types.OsmoBondDenom, sdk.NewInt(10000000))))
	suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
		AirdropStart:       time.Now(),
		DurationUntilDecay: types.DefaultDurationUntilDecay,
		DurationOfDecay:    types.DefaultDurationOfDecay,
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
