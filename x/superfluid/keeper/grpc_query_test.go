package keeper_test

import (
	"github.com/osmosis-labs/osmosis/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGRPCSuperfluidAsset() {
	suite.SetupTest()

	// initial check
	assets := suite.app.SuperfluidKeeper.GetAllSuperfluidAssets(suite.ctx)
	suite.Require().Len(assets, 0)

	// set asset
	suite.app.SuperfluidKeeper.SetSuperfluidAsset(suite.ctx, types.SuperfluidAsset{
		Denom:     "gamm/pool/1",
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// get asset
	res, err := suite.app.SuperfluidKeeper.AssetType(sdk.WrapSDKContext(suite.ctx), &types.AssetTypeRequest{Denom: "gamm/pool/1"})
	suite.Require().NoError(err)
	suite.Require().Equal(res.AssetType, types.SuperfluidAssetTypeLPShare)

	// check assets
	resp, err := suite.app.SuperfluidKeeper.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Assets, 1)
}
