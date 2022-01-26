package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestSuperfluidAssetSetGetDeleteFlow() {
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
	asset := suite.app.SuperfluidKeeper.GetSuperfluidAsset(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(asset.Denom, "gamm/pool/1")
	suite.Require().Equal(asset.AssetType, types.SuperfluidAssetTypeLPShare)

	// check assets
	assets = suite.app.SuperfluidKeeper.GetAllSuperfluidAssets(suite.ctx)
	suite.Require().Equal(assets, []types.SuperfluidAsset{asset})

	// delete asset
	suite.app.SuperfluidKeeper.DeleteSuperfluidAsset(suite.ctx, "gamm/pool/1")

	// get asset
	asset = suite.app.SuperfluidKeeper.GetSuperfluidAsset(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(asset.Denom, "")
	suite.Require().Equal(asset.AssetType, types.SuperfluidAssetTypeNative)

	// check assets
	assets = suite.app.SuperfluidKeeper.GetAllSuperfluidAssets(suite.ctx)
	suite.Require().Len(assets, 0)
}

func (suite *KeeperTestSuite) TestGetRiskAdjustedOsmoValue() {
	suite.SetupTest()

	adjustedValue := suite.app.SuperfluidKeeper.GetRiskAdjustedOsmoValue(
		suite.ctx,
		types.SuperfluidAsset{Denom: "gamm/pool/1", AssetType: types.SuperfluidAssetTypeLPShare},
		sdk.NewInt(100),
	)
	suite.Require().Equal(adjustedValue, sdk.NewInt(95))
}
