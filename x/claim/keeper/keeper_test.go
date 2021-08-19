package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/claim/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	// querier sdk.Querier
	app *app.OsmosisApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	airdropStartTime := time.Now()
	suite.app.ClaimKeeper.CreateModuleAccount(suite.ctx, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)))

	suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: types.DefaultDurationUntilDecay,
		DurationOfDecay:    types.DefaultDurationOfDecay,
		ClaimDenom:         sdk.DefaultBondDenom,
	})

	suite.ctx = suite.ctx.WithBlockTime(airdropStartTime)
}

func (suite *KeeperTestSuite) SetBalances(addr sdk.AccAddress, coins sdk.Coins) {
	suite.app.MintKeeper.MintCoins(suite.ctx, coins)
	suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, coins)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
