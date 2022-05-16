package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSuperfluidAssetSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	assets := suite.App.SuperfluidKeeper.GetAllSuperfluidAssets(suite.Ctx)
	suite.Require().Len(assets, 0)

	// set asset
	suite.App.SuperfluidKeeper.SetSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
		Denom:     "gamm/pool/1",
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// get asset
	asset := suite.App.SuperfluidKeeper.GetSuperfluidAsset(suite.Ctx, "gamm/pool/1")
	suite.Require().Equal(asset.Denom, "gamm/pool/1")
	suite.Require().Equal(asset.AssetType, types.SuperfluidAssetTypeLPShare)

	// check assets
	assets = suite.App.SuperfluidKeeper.GetAllSuperfluidAssets(suite.Ctx)
	suite.Require().Equal(assets, []types.SuperfluidAsset{asset})

	// delete asset
	suite.App.SuperfluidKeeper.DeleteSuperfluidAsset(suite.Ctx, "gamm/pool/1")

	// get asset
	asset = suite.App.SuperfluidKeeper.GetSuperfluidAsset(suite.Ctx, "gamm/pool/1")
	suite.Require().Equal(asset.Denom, "")
	suite.Require().Equal(asset.AssetType, types.SuperfluidAssetTypeNative)

	// check assets
	assets = suite.App.SuperfluidKeeper.GetAllSuperfluidAssets(suite.Ctx)
	suite.Require().Len(assets, 0)
}

func (suite *KeeperTestSuite) TestGetRiskAdjustedOsmoValue() {
	suite.SetupTest()

	adjustedValue := suite.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(
		suite.Ctx,
		types.SuperfluidAsset{Denom: "gamm/pool/1", AssetType: types.SuperfluidAssetTypeLPShare},
		sdk.NewInt(100),
	)
	suite.Require().Equal(sdk.NewInt(50), adjustedValue)
}
