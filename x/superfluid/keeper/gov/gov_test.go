package gov_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	minttypes "github.com/osmosis-labs/osmosis/v15/x/mint/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper/gov"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

func (suite *KeeperTestSuite) createGammPool(denoms []string) uint64 {
	coins := suite.App.GAMMKeeper.GetParams(suite.Ctx).PoolCreationFee
	poolAssets := []balancer.PoolAsset{}
	for _, denom := range denoms {
		coins = coins.Add(sdk.NewInt64Coin(denom, 1000000000000000000))
		poolAssets = append(poolAssets, balancer.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom, sdk.NewInt(1000000000000000000)),
		})
	}

	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	err := suite.App.BankKeeper.MintCoins(suite.Ctx, minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.App.BankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, minttypes.ModuleName, acc1, coins)
	suite.Require().NoError(err)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, poolAssets, "")
	poolId, err := suite.App.PoolManagerKeeper.CreatePool(suite.Ctx, msg)
	suite.Require().NoError(err)

	return poolId
}

func (suite *KeeperTestSuite) TestHandleSetSuperfluidAssetsProposal() {
	nativeAsset := types.SuperfluidAsset{
		Denom:     "stake",
		AssetType: types.SuperfluidAssetTypeNative,
	}
	gammAsset := types.SuperfluidAsset{
		Denom:     "gamm/pool/1",
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	concentratedAsset := types.SuperfluidAsset{
		Denom:     cltypes.GetConcentratedLockupDenomFromPoolId(2),
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	}
	concentratedAssetWrongAssetType := types.SuperfluidAsset{
		Denom:     cltypes.GetConcentratedLockupDenomFromPoolId(2),
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	nonExistentToken := types.SuperfluidAsset{
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
		name          string
		actions       []Action
		expectedEvent []string
	}{
		{
			"happy path flow (GAMM shares)",
			[]Action{
				{
					true, []types.SuperfluidAsset{gammAsset}, []types.SuperfluidAsset{gammAsset}, false,
				},
				{
					false, []types.SuperfluidAsset{gammAsset}, []types.SuperfluidAsset{}, false,
				},
			},
			[]string{types.TypeEvtSetSuperfluidAsset, types.TypeEvtRemoveSuperfluidAsset},
		},
		{
			"happy path flow (concentrated shares)",
			[]Action{
				{
					true, []types.SuperfluidAsset{concentratedAsset}, []types.SuperfluidAsset{concentratedAsset}, false,
				},
				{
					false, []types.SuperfluidAsset{concentratedAsset}, []types.SuperfluidAsset{}, false,
				},
			},
			[]string{types.TypeEvtSetSuperfluidAsset, types.TypeEvtRemoveSuperfluidAsset},
		},
		{
			"token does not exist",
			[]Action{
				{
					true, []types.SuperfluidAsset{gammAsset}, []types.SuperfluidAsset{gammAsset}, false,
				},
				{
					false, []types.SuperfluidAsset{nonExistentToken}, []types.SuperfluidAsset{gammAsset}, true,
				},
			},
			[]string{types.TypeEvtSetSuperfluidAsset, types.TypeEvtRemoveSuperfluidAsset},
		},
		{
			"concentrated share must be of type ConcentratedShare",
			[]Action{
				{
					true, []types.SuperfluidAsset{concentratedAssetWrongAssetType}, []types.SuperfluidAsset{}, true,
				},
			},
			[]string{types.TypeEvtSetSuperfluidAsset},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			// initial check
			resp, err := suite.querier.AllAssets(sdk.WrapSDKContext(suite.Ctx), &types.AllAssetsRequest{})
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
					suite.PrepareConcentratedPoolWithCoinsAndFullRangePosition(apptesting.STAKE, apptesting.USDC)
					// set superfluid assets via proposal
					err = gov.HandleSetSuperfluidAssetsProposal(suite.Ctx, *suite.App.SuperfluidKeeper, *suite.App.EpochsKeeper, &types.SetSuperfluidAssetsProposal{
						Title:       "title",
						Description: "description",
						Assets:      action.assets,
					})
				} else {
					// remove existing superfluid asset via proposal
					err = gov.HandleRemoveSuperfluidAssetsProposal(suite.Ctx, *suite.App.SuperfluidKeeper, &types.RemoveSuperfluidAssetsProposal{
						Title:                 "title",
						Description:           "description",
						SuperfluidAssetDenoms: govDenoms,
					})
				}
				if action.expectErr {
					suite.Require().Error(err)
				} else {
					suite.Require().NoError(err)
					suite.AssertEventEmitted(suite.Ctx, tc.expectedEvent[i], 1)
				}

				// check assets individually
				for _, asset := range action.expectedAssets {
					res, err := suite.querier.AssetType(sdk.WrapSDKContext(suite.Ctx), &types.AssetTypeRequest{Denom: asset.Denom})
					suite.Require().NoError(err)
					suite.Require().Equal(res.AssetType, asset.AssetType, "tcname %s, action num %d", tc.name, i)
				}

				// check assets
				resp, err = suite.querier.AllAssets(sdk.WrapSDKContext(suite.Ctx), &types.AllAssetsRequest{})
				suite.Require().NoError(err)
				suite.Require().Equal(resp.Assets, action.expectedAssets)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestHandleUnpoolWhiteListChange() {
	const (
		testTitle       = "test title"
		testDescription = "test description"
	)

	basePoolIds := []uint64{1, 2, 3}

	tests := map[string]struct {
		preCreatedPoolCount int
		preSetWhiteList     []uint64

		p               types.UpdateUnpoolWhiteListProposal
		expectError     bool
		expectedPoolIds []uint64
	}{
		"success; 3 pre-created poold ids and no pre-set whitelist, no overwrite": {
			preCreatedPoolCount: 3,

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         basePoolIds,
			},

			expectedPoolIds: basePoolIds,
		},
		"success; 3 pre-created poold ids and no pre-set whitelist, overwrite": {
			preCreatedPoolCount: 3,

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         basePoolIds,
				IsOverwrite: true,
			},

			expectedPoolIds: []uint64{1, 2, 3},
		},
		"success; 3 pre-created poold ids and pre-set whitelist, no overwrite": {
			preCreatedPoolCount: 3,

			preSetWhiteList: []uint64{1},

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{2, 3},
			},

			expectedPoolIds: basePoolIds,
		},
		"success; 3 pre-created poold ids and pre-set whitelist, overwrite": {
			preCreatedPoolCount: 3,

			preSetWhiteList: []uint64{1},

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{2, 3},
				IsOverwrite: true,
			},

			expectedPoolIds: []uint64{2, 3},
		},
		"success; duplicate id set, no overwrite": {
			preCreatedPoolCount: 1,

			preSetWhiteList: []uint64{1},

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{1},
			},

			expectedPoolIds: []uint64{1},
		},
		"success; duplicate id set, overwrite": {
			preCreatedPoolCount: 1,

			preSetWhiteList: []uint64{1},

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{1},
				IsOverwrite: true,
			},

			expectedPoolIds: []uint64{1},
		},
		"success; many duplicates with old values but not all, no overwrite": {
			preCreatedPoolCount: 10,

			preSetWhiteList: []uint64{1, 2, 3, 6, 9, 10},

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{3, 5, 6, 7, 10},
			},

			expectedPoolIds: []uint64{1, 2, 3, 5, 6, 7, 9, 10},
		},
		"success; many duplicates with old values but not all, overwrite": {
			preCreatedPoolCount: 10,

			preSetWhiteList: []uint64{1, 2, 3, 6, 9, 10},

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{3, 5, 6, 7, 10},
				IsOverwrite: true,
			},

			expectedPoolIds: []uint64{3, 5, 6, 7, 10},
		},
		"error; non-existent poold id provided": {
			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{1},
			},

			expectError: true,
		},
		"error; pool ids of 0": {
			preCreatedPoolCount: 1,

			p: types.UpdateUnpoolWhiteListProposal{
				Title:       testTitle,
				Description: testDescription,
				Ids:         []uint64{0},
				IsOverwrite: true,
			},

			expectError: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			ctx := suite.Ctx
			superfluidKeeper := suite.App.SuperfluidKeeper
			gammKeeper := suite.App.GAMMKeeper

			// Setup.
			for i := 0; i < tc.preCreatedPoolCount; i++ {
				suite.PrepareBalancerPool()
			}

			superfluidKeeper.SetUnpoolAllowedPools(ctx, tc.preSetWhiteList)

			// System under test.
			err := gov.HandleUnpoolWhiteListChange(ctx, *superfluidKeeper, gammKeeper, &tc.p)

			if tc.expectError {
				suite.Require().Error(err)
				return
			}

			suite.Require().NoError(err)

			// Validate that whitelist is set correctly.
			actualAllowedPools := superfluidKeeper.GetUnpoolAllowedPools(ctx)
			suite.Require().Equal(tc.expectedPoolIds, actualAllowedPools)
		})
	}
}
