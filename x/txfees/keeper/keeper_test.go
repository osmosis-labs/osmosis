package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	osmosisapp "github.com/osmosis-labs/osmosis/v19/app"

	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	protorevtypes "github.com/osmosis-labs/osmosis/v19/x/protorev/types"
	"github.com/osmosis-labs/osmosis/v19/x/txfees/types"
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

	// We set the base denom here in order for highest liquidity routes to get generated.
	// This is used in the tx fees epoch hook to swap the non OSMO to other tokens.
	baseDenom, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	s.Require().NoError(err)

	// Configure protorev base denoms
	baseDenomPriorities := []protorevtypes.BaseDenom{
		{
			Denom:    baseDenom,
			StepSize: sdk.NewInt(1_000_000),
		},
	}
	err = s.App.ProtoRevKeeper.SetBaseDenoms(s.Ctx, baseDenomPriorities)
	s.Require().NoError(err)

	// Mint some assets to the accounts.
	for _, acc := range s.TestAccs {
		s.FundAcc(acc,
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
