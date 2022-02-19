package gov_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper/gov"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestHandleSetSuperfluidAssetsProposal() {
	asset1 := types.SuperfluidAsset{
		Denom:     "gamm/pool/1",
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	asset2 := types.SuperfluidAsset{
		Denom:     "nativetoken",
		AssetType: types.SuperfluidAssetTypeNative,
	}
	asset3 := types.SuperfluidAsset{
		Denom:     "nonexistanttoken",
		AssetType: types.SuperfluidAssetTypeNative,
	}

	type Action struct {
		isAdd          bool
		assets         []types.SuperfluidAsset
		expectedAssets []types.SuperfluidAsset
		expectErr      bool
	}
	testCases := []struct {
		name    string
		actions []Action
	}{
		{
			"happy path flow",
			[]Action{
				{
					true, []types.SuperfluidAsset{asset1, asset2}, []types.SuperfluidAsset{asset1, asset2}, false,
				},
				{
					false, []types.SuperfluidAsset{asset2}, []types.SuperfluidAsset{asset1}, false,
				},
			},
		},
		{
			"token does not exist",
			[]Action{
				{
					true, []types.SuperfluidAsset{asset1, asset2}, []types.SuperfluidAsset{asset1, asset2}, false,
				},
				{
					false, []types.SuperfluidAsset{asset3}, []types.SuperfluidAsset{asset1, asset2}, true,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			// initial check
			resp, err := suite.app.SuperfluidKeeper.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
			suite.Require().NoError(err)
			suite.Require().Len(resp.Assets, 0)

			for i, action := range tc.actions {
				if action.isAdd {
					// set superfluid assets via proposal
					err = gov.HandleSetSuperfluidAssetsProposal(suite.ctx, *suite.app.SuperfluidKeeper, &types.SetSuperfluidAssetsProposal{
						Title:       "title",
						Description: "description",
						Assets:      action.assets,
					})
				} else {
					assetDenoms := []string{}
					for _, asset := range action.assets {
						assetDenoms = append(assetDenoms, asset.Denom)
					}
					// remove existing superfluid asset via proposal
					err = gov.HandleRemoveSuperfluidAssetsProposal(suite.ctx, *suite.app.SuperfluidKeeper, &types.RemoveSuperfluidAssetsProposal{
						Title:                 "title",
						Description:           "description",
						SuperfluidAssetDenoms: assetDenoms,
					})
				}
				if action.expectErr {
					suite.Require().Error(err)
				} else {
					suite.Require().NoError(err)
				}

				// check assets individually
				for _, asset := range action.expectedAssets {
					res, err := suite.app.SuperfluidKeeper.AssetType(sdk.WrapSDKContext(suite.ctx), &types.AssetTypeRequest{Denom: asset.Denom})
					suite.Require().NoError(err)
					suite.Require().Equal(res.AssetType, asset.AssetType, "tcname %s, action num %d", tc.name, i)
				}

				// check assets
				resp, err = suite.app.SuperfluidKeeper.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
				suite.Require().NoError(err)
				suite.Require().Equal(resp.Assets, action.expectedAssets)
			}
		})
	}
}
