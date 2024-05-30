package keeper_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"
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
	type riskFactorTest struct {
		name            string
		riskFactorDenom string
		riskFactor      string
		queryDenom      string
		initialAmt      int64
		expected        int64
	}

	tests := []riskFactorTest{
		{"test risk factor not set ", "random", "", "btc", 100, 50},
		{"test risk factor set high", "btc", "0.8", "btc", 100, 20},
		{"test risk factor set low", "btc", "0.1", "btc", 100, 90},
		{"test risk set diff query", "gamm/pool/1", "0.1", "btc", 100, 50},         // default
		{"test risk set same query", "gamm/pool/1", "0.1", "gamm/pool/1", 100, 90}, // set

		// wildcards. TODO: In the future it would be nice to add wildcards
		//{"test wildcard 1", "gamm/pool/*", "0.1", "gamm/pool/1", 100, 90},
		//{"test wildcard 2", "gamm/pool/*", "0.1", "gamm/pool/2", 100, 90},
		//{"test wildcard not matching", "gamm/pool/*", "0.1", "gamm", 100, 50},
		//{"test wildcard not matching 2", "gamm/pool/*", "0.1", "btc", 100, 50},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			if tt.riskFactor != "" {
				s.App.SuperfluidKeeper.SetDenomRiskFactor(s.Ctx, tt.riskFactorDenom, osmomath.MustNewDecFromStr(tt.riskFactor))
			}
			adjustedValue := s.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(s.Ctx, osmomath.NewInt(tt.initialAmt), tt.queryDenom)
			s.Require().Equal(osmomath.NewInt(tt.expected), adjustedValue)
		})
	}
}
