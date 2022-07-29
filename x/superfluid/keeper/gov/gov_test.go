package gov_test

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/pool-models/balancer"
	minttypes "github.com/osmosis-labs/osmosis/v10/x/mint/types"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/keeper/gov"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/types"
)

func (suite *KeeperTestSuite) createGammPool(denoms []string) uint64 {
	coins := suite.app.GAMMKeeper.GetParams(suite.ctx).PoolCreationFee
	poolAssets := []balancer.PoolAsset{}
	for _, denom := range denoms {
		coins = coins.Add(sdk.NewInt64Coin(denom, 1000000000000000000))
		poolAssets = append(poolAssets, balancer.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom, sdk.NewInt(1000000000000000000)),
		})
	}

	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, acc1, coins)
	suite.Require().NoError(err)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, poolAssets, "")
	poolId, err := suite.app.GAMMKeeper.CreatePool(suite.ctx, msg)
	suite.Require().NoError(err)

	return poolId
}

func (suite *KeeperTestSuite) TestHandleSetSuperfluidAssetsProposal() {
	nativeAsset := types.SuperfluidAsset{
		Denom:     "stake",
		AssetType: types.SuperfluidAssetTypeNative,
	}
	asset1 := types.SuperfluidAsset{
		Denom:     "gamm/pool/1",
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	asset2 := types.SuperfluidAsset{
		Denom:     "nonexistanttoken",
		AssetType: types.SuperfluidAssetTypeNative,
	}

	type Action struct {
		isAdd          bool
		assets         []types.SuperfluidAsset
		expectedAssets []types.SuperfluidAsset
		expectErr      bool
		expectedEvent  string
	}
	testCases := []struct {
		name    string
		actions []Action
	}{
		{
			"happy path flow",
			[]Action{
				{
					true, []types.SuperfluidAsset{asset1}, []types.SuperfluidAsset{asset1}, false, types.TypeEvtSetSuperfluidAsset,
				},
				{
					false, []types.SuperfluidAsset{asset1}, []types.SuperfluidAsset{}, false, types.TypeEvtRemoveSuperfluidAsset,
				},
			},
		},
		{
			"token does not exist",
			[]Action{
				{
					true, []types.SuperfluidAsset{asset1}, []types.SuperfluidAsset{asset1}, false, types.TypeEvtSetSuperfluidAsset,
				},
				{
					false, []types.SuperfluidAsset{asset2}, []types.SuperfluidAsset{asset1}, true, types.TypeEvtRemoveSuperfluidAsset,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			// initial check
			resp, err := suite.querier.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
			suite.Require().NoError(err)
			suite.Require().Len(resp.Assets, 0)

			for i, action := range tc.actions {

				// here we set two different string arrays of denoms.
				// The reason we do this is because native denom should be an asset within the pool,
				// while we do not want native asset to be in gov proposals.
				govDenoms := []string{}
				poolDenoms := []string{nativeAsset.Denom}

				for _, asset := range action.assets {
					poolDenoms = append(poolDenoms, asset.Denom)
					govDenoms = append(govDenoms, asset.Denom)
				}

				if action.isAdd {
					suite.createGammPool(poolDenoms)
					// set superfluid assets via proposal
					err = gov.HandleSetSuperfluidAssetsProposal(suite.ctx, *suite.app.SuperfluidKeeper, *suite.app.EpochsKeeper, &types.SetSuperfluidAssetsProposal{
						Title:       "title",
						Description: "description",
						Assets:      action.assets,
					})
				} else {
					// remove existing superfluid asset via proposal
					err = gov.HandleRemoveSuperfluidAssetsProposal(suite.ctx, *suite.app.SuperfluidKeeper, &types.RemoveSuperfluidAssetsProposal{
						Title:                 "title",
						Description:           "description",
						SuperfluidAssetDenoms: govDenoms,
					})
				}
				if action.expectErr {
					suite.Require().Error(err)
				} else {
					suite.Require().NoError(err)
					assertEventEmitted(suite, suite.ctx, action.expectedEvent, 1)
				}

				// check assets individually
				for _, asset := range action.expectedAssets {
					res, err := suite.querier.AssetType(sdk.WrapSDKContext(suite.ctx), &types.AssetTypeRequest{Denom: asset.Denom})
					suite.Require().NoError(err)
					suite.Require().Equal(res.AssetType, asset.AssetType, "tcname %s, action num %d", tc.name, i)
				}

				// check assets
				resp, err = suite.querier.AllAssets(sdk.WrapSDKContext(suite.ctx), &types.AllAssetsRequest{})
				fmt.Println(resp)
				suite.Require().NoError(err)
				suite.Require().Equal(resp.Assets, action.expectedAssets)
			}
		})
	}
}
