package keeper_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func (s *KeeperTestSuite) TestSuperfluidAssetSetGetDeleteFlow() {
	s.SetupTest()

	// initial check
	assets := s.App.SuperfluidKeeper.GetAllSuperfluidAssets(s.Ctx)
	s.Require().Len(assets, 0)

	// set asset
	s.App.SuperfluidKeeper.SetSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     DefaultGammAsset,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// get asset
	asset, err := s.App.SuperfluidKeeper.GetSuperfluidAsset(s.Ctx, DefaultGammAsset)
	s.Require().NoError(err)
	s.Require().Equal(asset.Denom, DefaultGammAsset)
	s.Require().Equal(asset.AssetType, types.SuperfluidAssetTypeLPShare)

	// check assets
	assets = s.App.SuperfluidKeeper.GetAllSuperfluidAssets(s.Ctx)
	s.Require().Equal(assets, []types.SuperfluidAsset{asset})

	// delete asset
	s.App.SuperfluidKeeper.DeleteSuperfluidAsset(s.Ctx, DefaultGammAsset)

	// get asset
	asset, err = s.App.SuperfluidKeeper.GetSuperfluidAsset(s.Ctx, DefaultGammAsset)
	s.Require().Error(err)
	s.Require().Equal(asset.Denom, "")
	s.Require().Equal(asset.AssetType, types.SuperfluidAssetTypeNative)

	// check assets
	assets = s.App.SuperfluidKeeper.GetAllSuperfluidAssets(s.Ctx)
	s.Require().Len(assets, 0)
}

func (s *KeeperTestSuite) TestGetRiskAdjustedOsmoValue() {
	s.SetupTest()

	adjustedValue := s.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(
		s.Ctx,
		osmomath.NewInt(100),
	)
	s.Require().Equal(osmomath.NewInt(50), adjustedValue)
}
