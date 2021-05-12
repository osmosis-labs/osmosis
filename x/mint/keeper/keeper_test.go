package keeper_test

import (
	"testing"
	"time"

	"github.com/c-osmosis/osmosis/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.OsmosisApp
	ctx sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, Time: time.Now().UTC()})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestGetPoolAllocatableAsset() {
	mintKeeper := suite.app.MintKeeper

	// Params would be set as the stake coin with 20% ratio from the default genesis state.

	// At this time, the fee collector doesn't have any assets.
	// So, it should be return the empty coins.
	coin := mintKeeper.GetPoolAllocatableAsset(suite.ctx)
	suite.Equal("0stake", coin.String())

	// Mint the stake coin to the fee collector.
	err := suite.app.BankKeeper.AddCoins(
		suite.ctx,
		suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName),
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000))),
	)
	suite.NoError(err)

	// In this time, should return the 20% of 100000stake
	coin = mintKeeper.GetPoolAllocatableAsset(suite.ctx)
	suite.Equal("20000stake", coin.String())

	// Mint some random coins to the fee collector.
	err = suite.app.BankKeeper.AddCoins(
		suite.ctx,
		suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName),
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1481290)), sdk.NewCoin("test", sdk.NewInt(12389190))),
	)
	suite.NoError(err)

	coin = mintKeeper.GetPoolAllocatableAsset(suite.ctx)
	suite.Equal("316258stake", coin.String())
}
