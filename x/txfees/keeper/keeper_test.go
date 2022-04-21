package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	osmosisapp "github.com/osmosis-labs/osmosis/v7/app"
  
	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

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
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)

	encodingConfig := osmosisapp.MakeEncodingConfig()
	suite.clientCtx = client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONCodec(encodingConfig.Marshaler)

	// Mint some assets to the accounts.
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc,
			sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000000)),
				sdk.NewCoin("uosmo", sdk.NewInt(100000000000000000)), // Needed for pool creation fee
				sdk.NewCoin("uion", sdk.NewInt(10000000)),
				sdk.NewCoin("atom", sdk.NewInt(10000000)),
				sdk.NewCoin("ust", sdk.NewInt(10000000)),
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
			))
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
