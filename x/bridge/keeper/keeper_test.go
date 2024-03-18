package keeper_test

import (
	"slices"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
	msgServer   types.MsgServer

	tfQueryClient tokenfactorytypes.QueryClient

	authority string
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.App.BridgeKeeper)

	s.tfQueryClient = tokenfactorytypes.NewQueryClient(s.QueryHelper)

	s.authority = s.App.GovKeeper.GetAuthority()
}

func (s *KeeperTestSuite) AppendNewAsset(asset types.Asset) {
	resp, err := s.queryClient.Params(s.Ctx, new(types.QueryParamsRequest))
	s.Require().NoError(err)

	newParams := resp.GetParams()
	newParams.Assets = append(newParams.Assets, asset)

	_, err = s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
		Sender:    s.authority,
		NewParams: newParams,
	})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) GetParams() types.Params {
	resp, err := s.queryClient.Params(s.Ctx, new(types.QueryParamsRequest))
	s.Require().NoError(err)
	return resp.GetParams()
}

func (s *KeeperTestSuite) GetModuleAddress() string {
	return s.App.AccountKeeper.GetModuleAddress(types.ModuleName).String()
}

// GetBridgeTFDenoms returns list of denoms created by the bridge module.
// Demons are requested from the tokenfactory.
func (s *KeeperTestSuite) GetBridgeTFDenoms() []string {
	req := &tokenfactorytypes.QueryDenomsFromCreatorRequest{
		Creator: s.GetModuleAddress(),
	}
	resp, err := s.tfQueryClient.DenomsFromCreator(s.Ctx, req)
	s.Require().NoError(err)
	return resp.Denoms
}

// GetBridgeDenoms generates list of denoms based on the assets module param.
func (s *KeeperTestSuite) GetBridgeDenoms() []string {
	params := s.GetParams()
	bridgeDenoms := make([]string, 0, len(params.Assets))
	moduleAddr := s.GetModuleAddress()
	for _, asset := range params.Assets {
		denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr, asset.Name())
		s.Require().NoError(err)
		bridgeDenoms = append(bridgeDenoms, denom)
	}
	return bridgeDenoms
}

func (s *KeeperTestSuite) HasEvent(eventType string) bool {
	events := s.Ctx.EventManager().Events()
	const eventIdxNotFound = -1
	eventIdx := slices.IndexFunc(events, func(e sdk.Event) bool {
		return e.Type == eventType
	})
	return eventIdx != eventIdxNotFound
}
