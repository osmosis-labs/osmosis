package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
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
	}
	testCases := []struct {
		name    string
		actions []Action
	}{
		{
			"happy path flow",
			[]Action{
				{
					true, []types.SuperfluidAsset{asset1, asset2}, []types.SuperfluidAsset{asset1, asset2},
				},
				{
					false, []types.SuperfluidAsset{asset2}, []types.SuperfluidAsset{asset1},
				},
				{
					false, []types.SuperfluidAsset{asset3}, []types.SuperfluidAsset{asset1},
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

			for _, action := range tc.actions {
				if action.isAdd {
					// set superfluid assets via proposal
					err = suite.app.SuperfluidKeeper.HandleSetSuperfluidAssetsProposal(suite.ctx, &types.SetSuperfluidAssetsProposal{
						Title:       "title",
						Description: "description",
						Assets:      action.assets,
					})
					suite.Require().NoError(err)
				} else {
					assetDenoms := []string{}
					for _, asset := range action.assets {
						assetDenoms = append(assetDenoms, asset.Denom)
					}
					// remove existing superfluid asset via proposal
					err = suite.app.SuperfluidKeeper.HandleRemoveSuperfluidAssetsProposal(suite.ctx, &types.RemoveSuperfluidAssetsProposal{
						Title:                 "title",
						Description:           "description",
						SuperfluidAssetDenoms: assetDenoms,
					})
					suite.Require().NoError(err)
				}

				// check assets individually
				for _, asset := range action.expectedAssets {
					res, err := suite.app.SuperfluidKeeper.AssetType(sdk.WrapSDKContext(suite.ctx), &types.AssetTypeRequest{Denom: asset.Denom})
					suite.Require().NoError(err)
					suite.Require().Equal(res.AssetType, asset.AssetType)
				}

				// check assets
				resp, err = suite.app.SuperfluidKeeper.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
				suite.Require().NoError(err)
				suite.Require().Equal(resp.Assets, action.expectedAssets)
			}
		})
	}
}
