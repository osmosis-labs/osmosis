package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	suite.SetupTest()
	res, err := suite.querier.Params(sdk.WrapSDKContext(suite.Ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().True(res.Params.MinimumRiskFactor.Equal(types.DefaultParams().MinimumRiskFactor))
}

func (suite *KeeperTestSuite) TestGRPCSuperfluidAsset() {
	suite.SetupTest()

	// initial check
	assets := suite.querier.GetAllSuperfluidAssets(suite.Ctx)
	suite.Require().Len(assets, 0)

	// set asset
	suite.querier.SetSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
		Denom:     "gamm/pool/1",
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// get asset
	res, err := suite.querier.AssetType(sdk.WrapSDKContext(suite.Ctx), &types.AssetTypeRequest{Denom: "gamm/pool/1"})
	suite.Require().NoError(err)
	suite.Require().Equal(res.AssetType, types.SuperfluidAssetTypeLPShare)

	// check assets
	resp, err := suite.querier.AllAssets(sdk.WrapSDKContext(suite.Ctx), &types.AllAssetsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Assets, 1)
}

func (suite *KeeperTestSuite) TestGRPCQuerySuperfluidDelegations() {
	suite.SetupTest()

	// Generate delegator addresses
	delAddrs := CreateRandomAccounts(2)

	// setup 2 validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, 0, 1000000},
		{0, 1, 1, 1000000},
		{1, 0, 1, 1000000},
		{1, 1, 0, 1000000},
	}

	// setup superfluid delegations
	_, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, superfluidDelegations, denoms)

	// setup normal delegations

	// for each superfluid delegation, query the amount and make sure it is 1000000
	for _, delegation := range superfluidDelegations {
		lpDenom := denoms[delegation.lpIndex]
		res, err := suite.queryClient.SuperfluidDelegationAmount(sdk.WrapSDKContext(suite.Ctx), &types.SuperfluidDelegationAmountRequest{
			DelegatorAddress: delAddrs[delegation.delIndex].String(),
			ValidatorAddress: valAddrs[delegation.valIndex].String(),
			Denom:            lpDenom,
		})
		suite.Require().NoError(err)
		suite.Require().Equal(res.Amount.AmountOf(lpDenom).Int64(), delegation.lpAmount)
	}

	// for each delegator, query all their superfluid delegations and make sure they have 2 delegations
	for _, delegator := range delAddrs {
		res, err := suite.queryClient.SuperfluidDelegationsByDelegator(sdk.WrapSDKContext(suite.Ctx), &types.SuperfluidDelegationsByDelegatorRequest{
			DelegatorAddress: delegator.String(),
		})
		suite.Require().NoError(err)
		suite.Require().Len(res.SuperfluidDelegationRecords, 2)
		suite.Require().True(res.TotalDelegatedCoins.IsEqual(sdk.NewCoins(
			sdk.NewInt64Coin(denoms[0], 1000000),
			sdk.NewInt64Coin(denoms[1], 1000000),
		)))
	}

	// for each validator denom pair, make sure they have 1 delegations
	for _, validator := range valAddrs {
		for _, denom := range denoms {
			amountRes, err := suite.queryClient.EstimateSuperfluidDelegatedAmountByValidatorDenom(sdk.WrapSDKContext(suite.Ctx), &types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{
				ValidatorAddress: validator.String(),
				Denom:            denom,
			})

			suite.Require().NoError(err)
			suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(denom, 1000000)), amountRes.TotalDelegatedCoins)

			delegationsRes, err := suite.queryClient.SuperfluidDelegationsByValidatorDenom(sdk.WrapSDKContext(suite.Ctx), &types.SuperfluidDelegationsByValidatorDenomRequest{
				ValidatorAddress: validator.String(),
				Denom:            denom,
			})
			suite.Require().NoError(err)
			suite.Require().Len(delegationsRes.SuperfluidDelegationRecords, 1)
		}
	}

	totalSuperfluidDelegationsRes, err := suite.queryClient.TotalSuperfluidDelegations(sdk.WrapSDKContext(suite.Ctx), &types.TotalSuperfluidDelegationsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(40000000), totalSuperfluidDelegationsRes.TotalDelegations)

	for _, lockID := range locks {
		connectedIntermediaryAccountRes, err := suite.queryClient.ConnectedIntermediaryAccount(sdk.WrapSDKContext(suite.Ctx), &types.ConnectedIntermediaryAccountRequest{LockId: lockID.ID})
		suite.Require().NoError(err)
		suite.Require().NotEqual("", connectedIntermediaryAccountRes.Account.Denom)
		suite.Require().NotEqual("", connectedIntermediaryAccountRes.Account.Address)
		suite.Require().NotEqual(uint64(0), connectedIntermediaryAccountRes.Account.GaugeId)

	}
	connectedIntermediaryAccountRes, err := suite.queryClient.ConnectedIntermediaryAccount(sdk.WrapSDKContext(suite.Ctx), &types.ConnectedIntermediaryAccountRequest{LockId: 123})
	suite.Require().NoError(err)
	suite.Require().Equal("", connectedIntermediaryAccountRes.Account.Denom)
	suite.Require().Equal("", connectedIntermediaryAccountRes.Account.ValAddr)
	suite.Require().Equal(uint64(0), connectedIntermediaryAccountRes.Account.GaugeId)
}

func (suite *KeeperTestSuite) TestGRPCQuerySuperfluidDelegationsDontIncludeUnbonding() {
	suite.SetupTest()

	// Generate delegator addresses
	delAddrs := CreateRandomAccounts(2)

	// setup 2 validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, 0, 1000000},
		{0, 1, 1, 1000000},
		{1, 0, 1, 1000000},
		{1, 1, 0, 1000000},
	}

	// setup superfluid delegations
	_, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, superfluidDelegations, denoms)

	// start unbonding the superfluid delegations of denom0 from delegator0 to validator0
	err := suite.querier.SuperfluidUndelegate(suite.Ctx, locks[0].Owner, locks[0].ID)
	suite.Require().NoError(err)

	// query to make sure that the amount delegated for the now unbonding delegation is 0
	res, err := suite.queryClient.SuperfluidDelegationAmount(sdk.WrapSDKContext(suite.Ctx), &types.SuperfluidDelegationAmountRequest{
		DelegatorAddress: delAddrs[0].String(),
		ValidatorAddress: valAddrs[0].String(),
		Denom:            denoms[0],
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Amount.AmountOf(denoms[0]).Int64(), int64(0))

	// query to make sure that the unbonding delegation is not included in delegator query
	res2, err := suite.queryClient.SuperfluidDelegationsByDelegator(sdk.WrapSDKContext(suite.Ctx), &types.SuperfluidDelegationsByDelegatorRequest{
		DelegatorAddress: delAddrs[0].String(),
	})
	suite.Require().NoError(err)
	suite.Require().Len(res2.SuperfluidDelegationRecords, 1)
	suite.Require().Equal(sdk.NewCoins(
		sdk.NewInt64Coin(denoms[1], 1000000)), res2.TotalDelegatedCoins)

	// query to make sure that the unbonding delegation is not included in the validator denom pair query
	amountRes, err := suite.queryClient.EstimateSuperfluidDelegatedAmountByValidatorDenom(sdk.WrapSDKContext(suite.Ctx), &types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{
		ValidatorAddress: valAddrs[1].String(),
		Denom:            denoms[0],
	})

	suite.Require().NoError(err)
	suite.Require().True(amountRes.TotalDelegatedCoins.IsEqual(sdk.NewCoins(
		sdk.NewInt64Coin(denoms[0], 1000000),
	)))

	delegationsRes, err := suite.queryClient.SuperfluidDelegationsByValidatorDenom(sdk.WrapSDKContext(suite.Ctx), &types.SuperfluidDelegationsByValidatorDenomRequest{
		ValidatorAddress: valAddrs[1].String(),
		Denom:            denoms[0],
	})
	suite.Require().NoError(err)
	suite.Require().Len(delegationsRes.SuperfluidDelegationRecords, 1)

	totalSuperfluidDelegationsRes, err := suite.queryClient.TotalSuperfluidDelegations(sdk.WrapSDKContext(suite.Ctx), &types.TotalSuperfluidDelegationsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(totalSuperfluidDelegationsRes.TotalDelegations, sdk.NewInt(30000000))
}

func (suite *KeeperTestSuite) TestGRPCQueryTotalDelegationByDelegator() {
	suite.SetupTest()

	// Generate delegator addresses
	delAddrs := CreateRandomAccounts(2)

	// setup 2 validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, 0, 1000000},
		{0, 1, 1, 1000000},
		{1, 0, 1, 1000000},
		{1, 1, 0, 1000000},
	}

	// setup superfluid delegations
	suite.SetupSuperfluidDelegations(delAddrs, valAddrs, superfluidDelegations, denoms)

	// setup normal delegations
	bond0to0 := stakingtypes.NewDelegation(delAddrs[0], valAddrs[0], sdk.NewDec(9))
	bond0to1 := stakingtypes.NewDelegation(delAddrs[0], valAddrs[1], sdk.NewDec(9))
	bond1to0 := stakingtypes.NewDelegation(delAddrs[1], valAddrs[0], sdk.NewDec(9))
	bond1to1 := stakingtypes.NewDelegation(delAddrs[1], valAddrs[1], sdk.NewDec(9))

	suite.App.StakingKeeper.SetDelegation(suite.Ctx, bond0to0)
	suite.App.StakingKeeper.SetDelegation(suite.Ctx, bond0to1)
	suite.App.StakingKeeper.SetDelegation(suite.Ctx, bond1to0)
	suite.App.StakingKeeper.SetDelegation(suite.Ctx, bond1to1)

	multiplier0 := suite.querier.Keeper.GetOsmoEquivalentMultiplier(suite.Ctx, denoms[0])
	multiplier1 := suite.querier.Keeper.GetOsmoEquivalentMultiplier(suite.Ctx, denoms[1])
	minRiskFactor := suite.querier.Keeper.GetParams(suite.Ctx).MinimumRiskFactor

	expectAmount0 := multiplier0.Mul(sdk.NewDec(1000000)).Sub(multiplier0.Mul(sdk.NewDec(1000000)).Mul(minRiskFactor))
	expectAmount1 := multiplier1.Mul(sdk.NewDec(1000000)).Sub(multiplier1.Mul(sdk.NewDec(1000000)).Mul(minRiskFactor))

	// for each delegator, query all their superfluid delegations and normal delegations. Making sure they have 4 delegations
	// Making sure TotalEquivalentStakedAmount is equal to converted amount + normal delegations
	for _, delegator := range delAddrs {
		res, err := suite.queryClient.TotalDelegationByDelegator(sdk.WrapSDKContext(suite.Ctx), &types.QueryTotalDelegationByDelegatorRequest{
			DelegatorAddress: delegator.String(),
		})

		suite.Require().NoError(err)
		suite.Require().Len(res.SuperfluidDelegationRecords, 2)
		suite.Require().Len(res.DelegationResponse, 2)
		suite.Require().True(res.TotalDelegatedCoins.IsEqual(sdk.NewCoins(
			sdk.NewInt64Coin(denoms[0], 1000000),
			sdk.NewInt64Coin(denoms[1], 1000000),
			sdk.NewInt64Coin("uosmo", 18000000),
		)))

		total_osmo_equivalent := sdk.NewCoin("uosmo", expectAmount0.RoundInt().Add(expectAmount1.RoundInt()).Add(sdk.NewInt(18000000)))

		suite.Require().True(res.TotalEquivalentStakedAmount.IsEqual(total_osmo_equivalent))
	}
}
