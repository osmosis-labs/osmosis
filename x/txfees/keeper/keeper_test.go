package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/app"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/osmosis-labs/osmosis/x/txfees/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	// querier sdk.Querier
	app *app.OsmosisApp

	queryClient types.QueryClient
}

var (
	acc1 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc2 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

func (suite *KeeperTestSuite) SetupTest() {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.TxFeesKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx

	suite.queryClient = queryClient

	// Mint some assets to the accounts.
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := suite.app.BankKeeper.AddCoins(
			suite.ctx,
			acc,
			sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
				sdk.NewCoin("uion", sdk.NewInt(10000000)),
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
			),
		)
		if err != nil {
			panic(err)
		}
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) preparePool(assets []gammtypes.PoolAsset) uint64 {
	suite.Require().Len(assets, 2)

	poolId, err := suite.app.GAMMKeeper.CreatePool(suite.ctx, acc1,
		gammtypes.PoolParams{
			SwapFee: sdk.NewDec(0),
			ExitFee: sdk.NewDec(0),
		}, assets, "")
	suite.NoError(err)

	_, err = suite.app.GAMMKeeper.CalculateSpotPrice(suite.ctx, poolId, assets[0].Token.Denom, assets[1].Token.Denom)
	suite.NoError(err)

	_, err = suite.app.GAMMKeeper.CalculateSpotPrice(suite.ctx, poolId, assets[1].Token.Denom, assets[0].Token.Denom)
	suite.NoError(err)

	return poolId
}
