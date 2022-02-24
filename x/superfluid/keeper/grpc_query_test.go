package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	suite.SetupTest()
	res, err := suite.app.SuperfluidKeeper.Params(sdk.WrapSDKContext(suite.ctx), &types.ParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().True(res.Params.MinimumRiskFactor.Equal(types.DefaultParams().MinimumRiskFactor))
}

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

func (suite *KeeperTestSuite) TestGRPCQuerySuperfluidDelegations() {
	suite.SetupTest()

	// Generate delegator addresses
	delAddrs := CreateRandomAccounts(2)

	// setup 2 validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	denoms := []string{"gamm/pool/1", "gamm/pool/2"}

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, denoms[0], 1000000},

		{0, 1, denoms[1], 1000000},

		{1, 0, denoms[1], 1000000},

		{1, 1, denoms[0], 1000000},
	}

	// setup superfluid delegations
	suite.SetupSuperfluidDelegations(delAddrs, valAddrs, superfluidDelegations)

	// for each superfluid delegation, query the amount and make sure it is 1000000
	for _, delegation := range superfluidDelegations {
		res, err := suite.queryClient.SuperfluidDelegationAmount(sdk.WrapSDKContext(suite.ctx), &types.SuperfluidDelegationAmountRequest{
			DelegatorAddress: delAddrs[delegation.delIndex].String(),
			ValidatorAddress: valAddrs[delegation.valIndex].String(),
			Denom:            delegation.lpDenom,
		})
		suite.Require().NoError(err)
		suite.Require().Equal(res.Amount.AmountOf(delegation.lpDenom).Int64(), delegation.lpAmount)
	}

	// for each delegator, query all their superfluid delegations and make sure they have 2 delegations
	for _, delegator := range delAddrs {
		res, err := suite.queryClient.SuperfluidDelegationsByDelegator(sdk.WrapSDKContext(suite.ctx), &types.SuperfluidDelegationsByDelegatorRequest{
			DelegatorAddress: delegator.String(),
		})
		suite.Require().NoError(err)
		suite.Require().Len(res.SuperfluidDelegationRecords, 2)
		suite.Require().True(res.TotalDelegatedCoins.IsEqual(sdk.NewCoins(
			sdk.NewInt64Coin("gamm/pool/1", 1000000),
			sdk.NewInt64Coin("gamm/pool/2", 1000000),
		)))
	}

	// for each validator denom pair, make sure they have 1 delegations
	for _, validator := range valAddrs {
		for _, denom := range denoms {
			amountRes, err := suite.queryClient.EstimateSuperfluidDelegatedAmountByValidatorDenom(sdk.WrapSDKContext(suite.ctx), &types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{
				ValidatorAddress: validator.String(),
				Denom:            denom,
			})

			suite.Require().NoError(err)
			suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(denom, 1000000)), amountRes.TotalDelegatedCoins)

			delegationsRes, err := suite.queryClient.SuperfluidDelegationsByValidatorDenom(sdk.WrapSDKContext(suite.ctx), &types.SuperfluidDelegationsByValidatorDenomRequest{
				ValidatorAddress: validator.String(),
				Denom:            denom,
			})
			suite.Require().NoError(err)
			suite.Require().Len(delegationsRes.SuperfluidDelegationRecords, 1)
		}
	}

	totalSuperfluidDelegationsRes, err := suite.queryClient.TotalSuperfluidDelegations(sdk.WrapSDKContext(suite.ctx), &types.TotalSuperfluidDelegationsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(40000000), totalSuperfluidDelegationsRes.TotalDelgations)

}

func (suite *KeeperTestSuite) TestGRPCQuerySuperfluidDelegationsDontIncludeUnbonding() {
	suite.SetupTest()

	// Generate delegator addresses
	delAddrs := CreateRandomAccounts(2)

	// setup 2 validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	denoms := []string{"gamm/pool/1", "gamm/pool/2"}

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, denoms[0], 1000000},

		{0, 1, denoms[1], 1000000},

		{1, 0, denoms[1], 1000000},

		{1, 1, denoms[0], 1000000},
	}

	// setup superfluid delegations
	_, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, superfluidDelegations)

	// start unbonding the superfluid delegations of denom0 from delegator0 to validator0
	err := suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, locks[0].Owner, locks[0].ID)
	suite.Require().NoError(err)

	// query to make sure that the amount delegated for the now unbonding delegation is 0
	res, err := suite.queryClient.SuperfluidDelegationAmount(sdk.WrapSDKContext(suite.ctx), &types.SuperfluidDelegationAmountRequest{
		DelegatorAddress: delAddrs[0].String(),
		ValidatorAddress: valAddrs[0].String(),
		Denom:            denoms[0],
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Amount.AmountOf(denoms[0]).Int64(), int64(0))

	// query to make sure that the unbonding delegation is not included in delegator query
	res2, err := suite.queryClient.SuperfluidDelegationsByDelegator(sdk.WrapSDKContext(suite.ctx), &types.SuperfluidDelegationsByDelegatorRequest{
		DelegatorAddress: delAddrs[0].String(),
	})
	suite.Require().NoError(err)
	suite.Require().Len(res2.SuperfluidDelegationRecords, 1)
	suite.Require().Equal(sdk.NewCoins(
		sdk.NewInt64Coin("gamm/pool/2", 1000000)), res2.TotalDelegatedCoins)

	// query to make sure that the unbonding delegation is not included in the validator denom pair query
	amountRes, err := suite.queryClient.EstimateSuperfluidDelegatedAmountByValidatorDenom(sdk.WrapSDKContext(suite.ctx), &types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{
		ValidatorAddress: valAddrs[1].String(),
		Denom:            denoms[0],
	})

	suite.Require().NoError(err)
	suite.Require().True(amountRes.TotalDelegatedCoins.IsEqual(sdk.NewCoins(
		sdk.NewInt64Coin(denoms[0], 1000000),
	)))

	delegationsRes, err := suite.queryClient.SuperfluidDelegationsByValidatorDenom(sdk.WrapSDKContext(suite.ctx), &types.SuperfluidDelegationsByValidatorDenomRequest{
		ValidatorAddress: valAddrs[1].String(),
		Denom:            denoms[0],
	})
	suite.Require().NoError(err)
	suite.Require().Len(delegationsRes.SuperfluidDelegationRecords, 1)
}
