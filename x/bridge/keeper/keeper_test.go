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

func (s *KeeperTestSuite) AppendNewAssets(assets ...types.Asset) {
	newParams := s.GetParams()
	newParams.Assets = append(newParams.Assets, assets...)

	_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
		Sender:    s.authority,
		NewParams: newParams,
	})
	s.Require().NoError(err)

	// Check that a new denom appeared in tokenfactory
	tfDenoms := s.GetBridgeTFDenoms()
	for _, asset := range assets {
		expectedDenom, err := tokenfactorytypes.GetTokenDenom(s.GetModuleAddress(), asset.Name())
		s.Require().NoError(err)
		s.Require().Contains(tfDenoms, expectedDenom)
	}
}

func (s *KeeperTestSuite) AppendNewSigners(signers ...string) {
	newParams := s.GetParams()
	newParams.Signers = append(newParams.Signers, signers...)

	_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
		Sender:    s.authority,
		NewParams: newParams,
	})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) EnableAssets(assetIDs ...types.AssetID) {
	for _, assetID := range assetIDs {
		_, err := s.msgServer.ChangeAssetStatus(s.Ctx, &types.MsgChangeAssetStatus{
			Sender:    s.authority,
			AssetId:   assetID,
			NewStatus: types.AssetStatus_ASSET_STATUS_OK,
		})
		s.Require().NoError(err)
	}
}

func (s *KeeperTestSuite) SetVotesNeeded(votesNeeded uint64) {
	newParams := s.GetParams()
	newParams.VotesNeeded = votesNeeded

	_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
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

func (s *KeeperTestSuite) GetLastTransferHeight(assetID types.AssetID) uint64 {
	resp, err := s.queryClient.LastTransferHeight(s.Ctx, &types.LastTransferHeightRequest{
		AssetId: assetID,
	})
	s.Require().NoError(err)
	return resp.LastTransferHeight
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

func (s *KeeperTestSuite) GetTFDenom(assetID types.AssetID) string {
	moduleAddr := s.GetModuleAddress()
	denom, err := tokenfactorytypes.GetTokenDenom(moduleAddr, assetID.Name())
	s.Require().NoError(err)
	return denom
}

func (s *KeeperTestSuite) GetAddrFromBech32(addr string) sdk.AccAddress {
	result, err := sdk.AccAddressFromBech32(addr)
	s.Require().NoError(err)
	return result
}

func (s *KeeperTestSuite) HasEvent(eventType string) bool {
	events := s.Ctx.EventManager().Events()
	const eventIdxNotFound = -1
	eventIdx := slices.IndexFunc(events, func(e sdk.Event) bool {
		return e.Type == eventType
	})
	return eventIdx != eventIdxNotFound
}
