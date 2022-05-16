package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v8/app"
	balancertypes "github.com/osmosis-labs/osmosis/v8/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v8/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v8/x/txfees/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.OsmosisApp

	clientCtx client.Context

	queryClient types.QueryClient
}

var (
	acc1 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc2 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

func (suite *KeeperTestSuite) SetupTest(isCheckTx bool) {
	app := app.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.TxFeesKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx

	suite.queryClient = queryClient

	// Mint some assets to the accounts.
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc,
			sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000000)),
				sdk.NewCoin("uosmo", sdk.NewInt(100000000000000000)), // Needed for pool creation fee
				sdk.NewCoin("uion", sdk.NewInt(10000000)),
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
			))
		if err != nil {
			panic(err)
		}
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
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
	return suite.app.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.ctx, &upgradeProp)
}

func (suite *KeeperTestSuite) PreparePoolWithAssets(asset1, asset2 sdk.Coin) uint64 {
	return suite.preparePool(
		[]gammtypes.PoolAsset{
			{
				Weight: sdk.NewInt(1),
				Token:  asset1,
			},
			{
				Weight: sdk.NewInt(1),
				Token:  asset2,
			},
		},
	)
}

func (suite *KeeperTestSuite) preparePool(assets []gammtypes.PoolAsset) uint64 {
	suite.Require().Len(assets, 2)

	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1,
		balancertypes.PoolParams{
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
