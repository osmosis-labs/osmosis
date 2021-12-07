package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestHandleSetSuperfluidAssetsProposal() {
	suite.SetupTest()

	// initial check
	resp, err := suite.app.SuperfluidKeeper.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Assets, 0)

	// set superfluid assets via proposal
	err = suite.app.SuperfluidKeeper.HandleSetSuperfluidAssetsProposal(suite.ctx, &types.SetSuperfluidAssetsProposal{
		Title:       "title",
		Description: "description",
		Assets: []types.SuperfluidAsset{
			{
				Denom:     "gamm/pool/1",
				AssetType: types.SuperfluidAssetTypeLPShare,
			},
			{
				Denom:     "nativetoken",
				AssetType: types.SuperfluidAssetTypeNative,
			},
		},
	})
	suite.Require().NoError(err)

	// get asset
	res, err := suite.app.SuperfluidKeeper.AssetType(sdk.WrapSDKContext(suite.ctx), &types.AssetTypeRequest{Denom: "gamm/pool/1"})
	suite.Require().NoError(err)
	suite.Require().Equal(res.AssetType, types.SuperfluidAssetTypeLPShare)

	res, err = suite.app.SuperfluidKeeper.AssetType(sdk.WrapSDKContext(suite.ctx), &types.AssetTypeRequest{Denom: "nativetoken"})
	suite.Require().NoError(err)
	suite.Require().Equal(res.AssetType, types.SuperfluidAssetTypeNative)

	// check assets
	resp, err = suite.app.SuperfluidKeeper.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Assets, 2)

	// remove superfluid assets via proposal
	err = suite.app.SuperfluidKeeper.HandleRemoveSuperfluidAssetsProposal(suite.ctx, &types.RemoveSuperfluidAssetsProposal{
		Title:                 "title",
		Description:           "description",
		SuperfluidAssetDenoms: []string{"nativetoken"},
	})
	suite.Require().NoError(err)

	// check assets
	resp, err = suite.app.SuperfluidKeeper.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Assets, 1)
}
