package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest(isCheckTx bool) {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)

	// Mint some assets to the accounts.
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc,
			sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000000)),
				sdk.NewCoin("uosmo", sdk.NewInt(100000000000000000)), // Needed for pool creation fee
				sdk.NewCoin("uion", sdk.NewInt(10000000)),
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
			))
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
	return suite.App.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.Ctx, &upgradeProp)
}
