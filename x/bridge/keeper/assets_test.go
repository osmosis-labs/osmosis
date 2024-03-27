package keeper_test

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func (s *KeeperTestSuite) TestChangeAssetStatus() {
	testCases := []struct {
		name          string
		assetID       types.AssetID
		newStatus     types.AssetStatus
		expectedError error
	}{
		{
			name:          "success",
			assetID:       assetID1,
			newStatus:     types.AssetStatus_ASSET_STATUS_OK,
			expectedError: nil,
		},
		{
			name: "invalid asset",
			assetID: types.AssetID{
				SourceChain: "",
				Denom:       "",
			},
			newStatus:     types.AssetStatus_ASSET_STATUS_OK,
			expectedError: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "asset not found",
			assetID: types.AssetID{
				SourceChain: "unknown_chain",
				Denom:       "unknown_denom",
			},
			newStatus:     types.AssetStatus_ASSET_STATUS_OK,
			expectedError: types.ErrInvalidAssetID,
		},
		{
			name:          "unknown status",
			assetID:       assetID1,
			newStatus:     100,
			expectedError: sdkerrors.ErrInvalidRequest,
		},
		{
			name:          "invalid status",
			assetID:       assetID1,
			newStatus:     types.AssetStatus_ASSET_STATUS_UNSPECIFIED,
			expectedError: sdkerrors.ErrInvalidRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// Prepare the app
			s.AppendNewAssets(asset1)
			s.EnableAssets(assetID1)

			// Send a message
			_, err := s.msgServer.ChangeAssetStatus(s.Ctx, &types.MsgChangeAssetStatus{
				Sender:    s.authority,
				AssetId:   tc.assetID,
				NewStatus: tc.newStatus,
			})
			s.Require().ErrorIs(err, tc.expectedError)

			// Prepare results
			asset, found := s.App.BridgeKeeper.GetParams(s.Ctx).GetAsset(assetID1)
			s.Require().True(found)

			// Check results
			switch {
			case tc.expectedError != nil:
				// Status hasn't changed
				s.Require().Equal(asset.Status, types.AssetStatus_ASSET_STATUS_OK)
			case tc.expectedError == nil:
				// Status has changed
				s.Require().Equal(asset.Status, tc.newStatus)
			}
		})
	}
}
