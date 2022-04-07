package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmosisapp "github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	clientCtx client.Context

	queryClient types.QueryClient
}

var (
	acc1 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc2 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest(isCheckTx bool) {
	app := osmosisapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(*app.TxFeesKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = app
	suite.Ctx = ctx

	suite.queryClient = queryClient

	encodingConfig := osmosisapp.MakeEncodingConfig()
	suite.clientCtx = client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONCodec(encodingConfig.Marshaler)

	// Mint some assets to the accounts.
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, acc,
			sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000000)),
				sdk.NewCoin("uosmo", sdk.NewInt(100000000000000000)), // Needed for pool creation fee
				sdk.NewCoin("uion", sdk.NewInt(10000000)),
				sdk.NewCoin("atom", sdk.NewInt(10000000)),
				sdk.NewCoin("ust", sdk.NewInt(10000000)),
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
			))
		if err != nil {
			panic(err)
		}
	}
}

func (suite *KeeperTestSuite) ExecuteUpgradeFeeTokenProposal(feeToken string, poolId uint64) error {
	upgradeProp := types.NewUpdateFeeTokenProposal(
		"Test Proposal",
		"test",
		types.FeeToken{
			Denom:  feeToken,
			PoolID: poolId,
		},
	)
	return suite.App.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.Ctx, &upgradeProp)
}
