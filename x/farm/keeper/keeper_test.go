package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/app"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.OsmosisApp
	ctx sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

var (
	allocatorAcc = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

	acc1 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc2 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

func (suite *KeeperTestSuite) prepareAccounts() {
	err := suite.app.BankKeeper.AddCoins(
		suite.ctx,
		allocatorAcc,
		sdk.NewCoins(
			sdk.NewCoin("foo", sdk.NewInt(10000000000)),
			sdk.NewCoin("bar", sdk.NewInt(10000000000)),
			sdk.NewCoin("baz", sdk.NewInt(10000000000)),
		),
	)
	if err != nil {
		panic(err)
	}
}
