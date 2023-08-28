package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	osmosisapp "github.com/osmosis-labs/osmosis/v17/app"

	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	"github.com/osmosis-labs/osmosis/v17/x/txfees/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	clientCtx   client.Context
	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest(isCheckTx bool) {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	encodingConfig := osmosisapp.MakeEncodingConfig()
	s.clientCtx = client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithCodec(encodingConfig.Marshaler)

	// Mint some assets to the accounts.
	for _, acc := range s.TestAccs {
		s.FundAcc(acc,
			sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10000000000)),
				sdk.NewCoin("uosmo", osmomath.NewInt(100000000000000000)), // Needed for pool creation fee
				sdk.NewCoin("uion", osmomath.NewInt(10000000)),
				sdk.NewCoin("atom", osmomath.NewInt(10000000)),
				sdk.NewCoin("ust", osmomath.NewInt(10000000)),
				sdk.NewCoin("foo", osmomath.NewInt(10000000)),
				sdk.NewCoin("bar", osmomath.NewInt(10000000)),
			))
	}
}

func (s *KeeperTestSuite) ExecuteUpgradeFeeTokenProposal(feeToken string, poolId uint64) error {
	upgradeProp := types.NewUpdateFeeTokenProposal(
		"Test Proposal",
		"test",
		[]types.FeeToken{
			{
				Denom:  feeToken,
				PoolID: poolId,
			},
		},
	)
	return s.App.TxFeesKeeper.HandleUpdateFeeTokenProposal(s.Ctx, &upgradeProp)
}
