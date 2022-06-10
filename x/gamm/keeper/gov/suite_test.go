package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx      sdk.Context
	querier  keeper.Querier
	app      *app.OsmosisApp
	TestAccs []sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
	suite.querier = keeper.NewQuerier(*suite.app.GAMMKeeper)
	suite.TestAccs = CreateRandomAccounts(3)
}

func (s *KeeperTestSuite) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	err := simapp.FundAccount(s.app.BankKeeper, s.ctx, acc, amounts)
	s.Require().NoError(err)
}

func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
