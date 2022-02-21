package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

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

// func (suite *KeeperTestSuite) TestGRPCQuerySuperfluidDelegations() {
// 	suite.SetupTest()

// 	poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 	suite.Require().Equal(poolId, uint64(1))

// 	// setup 2 validators
// 	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

// 	// // setup superfluid delegations
// 	// intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, []superfluidDelegation{
// 	// 	{0, "gamm/pool/1"}, {0, "gamm/pool/2"}, {1, "gamm/pool/1"}, {1, "gamm/pool/2"}})
// 	// suite.checkIntermediaryAccountDelegations(intermediaryAccs)
// }
