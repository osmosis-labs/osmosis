package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	s.SetupTest()
	res, err := s.querier.Params(s.Ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().True(res.Params.MinimumRiskFactor.Equal(types.DefaultParams().MinimumRiskFactor))
}

func (s *KeeperTestSuite) TestAllIntermediaryAccounts() {
	s.SetupTest()
	// set account 1
	valAddr1 := sdk.ValAddress([]byte("test1-AllIntermediaryAccounts"))
	acc1 := types.NewSuperfluidIntermediaryAccount("test1", valAddr1.String(), 0)
	s.App.SuperfluidKeeper.SetIntermediaryAccount(s.Ctx, acc1)

	// set account 2
	valAddr2 := sdk.ValAddress([]byte("test2-AllIntermediaryAccounts"))
	acc2 := types.NewSuperfluidIntermediaryAccount("test2", valAddr2.String(), 0)
	s.App.SuperfluidKeeper.SetIntermediaryAccount(s.Ctx, acc2)

	// set account 3
	valAddr3 := sdk.ValAddress([]byte("test3-AllIntermediaryAccounts"))
	acc3 := types.NewSuperfluidIntermediaryAccount("test3", valAddr3.String(), 0)
	s.App.SuperfluidKeeper.SetIntermediaryAccount(s.Ctx, acc3)

	res, err := s.querier.AllIntermediaryAccounts(s.Ctx, &types.AllIntermediaryAccountsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(3, len(res.Accounts))
	s.Require().Equal(uint64(3), res.Pagination.Total)
}

func (s *KeeperTestSuite) TestTotalDelegationByValidatorForAsset() {
	s.SetupTest()
	ctx := s.Ctx
	querier := s.querier
	delegation_amount := int64(1000000)

	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})
	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

	superfluidDelegations := []superfluidDelegation{
		{0, 0, 0, delegation_amount},
		{0, 1, 1, delegation_amount},
		{1, 0, 1, delegation_amount},
		{1, 1, 0, delegation_amount},
	}

	s.setupSuperfluidDelegations(valAddrs, superfluidDelegations, denoms)

	for _, denom := range denoms {
		res, err := querier.TotalDelegationByValidatorForDenom(ctx, &types.QueryTotalDelegationByValidatorForDenomRequest{Denom: denom})

		s.Require().NoError(err)
		s.Require().Equal(len(valAddrs), len(res.Assets))

		for _, result := range res.Assets {
			// check osmo equivalent is correct
			actual_response_osmo := result.OsmoEquivalent
			needed_response_osmo, err := s.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(ctx, denom, osmomath.NewInt(delegation_amount))
			s.Require().NoError(err)

			s.Require().Equal(actual_response_osmo, needed_response_osmo)

			// check sfs'd asset amount correct
			actual_response_asset := result.AmountSfsd
			needed_response_asset := osmomath.NewInt(delegation_amount)
			s.Require().Equal(actual_response_asset, needed_response_asset)

			// check validator addresses correct
			actual_val := result.ValAddr
			checks := 0
			for _, val := range valAddrs {
				if val.String() == actual_val {
					checks++
					break
				}
			}
			s.Require().True(checks == 1)
		}
	}
}

func (s *KeeperTestSuite) TestGRPCSuperfluidAsset() {
	s.SetupTest()

	// initial check
	assets := s.querier.GetAllSuperfluidAssets(s.Ctx)
	s.Require().Len(assets, 0)

	// set asset
	s.querier.SetSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     DefaultGammAsset,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})

	// get asset
	res, err := s.querier.AssetType(s.Ctx, &types.AssetTypeRequest{Denom: DefaultGammAsset})
	s.Require().NoError(err)
	s.Require().Equal(res.AssetType, types.SuperfluidAssetTypeLPShare)

	// check assets
	resp, err := s.querier.AllAssets(s.Ctx, &types.AllAssetsRequest{})
	s.Require().NoError(err)
	s.Require().Len(resp.Assets, 1)
}

func (s *KeeperTestSuite) TestGRPCQuerySuperfluidDelegations() {
	s.SetupTest()

	// setup 2 validators
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, 0, 1000000},
		{0, 1, 1, 1000000},
		{1, 0, 1, 1000000},
		{1, 1, 0, 1000000},
	}

	// setup superfluid delegations
	delegatorAddresses, _, locks := s.setupSuperfluidDelegations(valAddrs, superfluidDelegations, denoms)

	// for each superfluid delegation, query the amount and make sure it is 1000000
	for _, delegation := range superfluidDelegations {
		lpDenom := denoms[delegation.lpIndex]
		res, err := s.queryClient.SuperfluidDelegationAmount(s.Ctx, &types.SuperfluidDelegationAmountRequest{
			DelegatorAddress: delegatorAddresses[delegation.delIndex].String(),
			ValidatorAddress: valAddrs[delegation.valIndex].String(),
			Denom:            lpDenom,
		})
		s.Require().NoError(err)
		s.Require().Equal(res.Amount.AmountOf(lpDenom).Int64(), delegation.lpAmount)
	}

	// for each delegator, query all their superfluid delegations and make sure they have 2 delegations
	for _, delegator := range delegatorAddresses {
		res, err := s.queryClient.SuperfluidDelegationsByDelegator(s.Ctx, &types.SuperfluidDelegationsByDelegatorRequest{
			DelegatorAddress: delegator.String(),
		})

		multiplier0 := s.querier.Keeper.GetOsmoEquivalentMultiplier(s.Ctx, denoms[0])
		multiplier1 := s.querier.Keeper.GetOsmoEquivalentMultiplier(s.Ctx, denoms[1])
		minRiskFactor := s.querier.Keeper.GetParams(s.Ctx).MinimumRiskFactor

		expectAmount0 := multiplier0.Mul(osmomath.NewDec(1000000)).Sub(multiplier0.Mul(osmomath.NewDec(1000000)).Mul(minRiskFactor))
		expectAmount1 := multiplier1.Mul(osmomath.NewDec(1000000)).Sub(multiplier1.Mul(osmomath.NewDec(1000000)).Mul(minRiskFactor))

		s.Require().NoError(err)
		s.Require().Len(res.SuperfluidDelegationRecords, 2)
		s.Require().True(res.TotalDelegatedCoins.Equal(sdk.NewCoins(
			sdk.NewInt64Coin(denoms[0], 1000000),
			sdk.NewInt64Coin(denoms[1], 1000000),
		)))
		s.Require().True(res.SuperfluidDelegationRecords[0].EquivalentStakedAmount.Equal(sdk.NewCoin(appparams.BaseCoinUnit, expectAmount0.RoundInt())))
		s.Require().True(res.SuperfluidDelegationRecords[1].EquivalentStakedAmount.Equal(sdk.NewCoin(appparams.BaseCoinUnit, expectAmount1.RoundInt())))
	}

	// for each validator denom pair, make sure they have 1 delegations
	for _, validator := range valAddrs {
		for _, denom := range denoms {
			amountRes, err := s.queryClient.EstimateSuperfluidDelegatedAmountByValidatorDenom(s.Ctx, &types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{
				ValidatorAddress: validator.String(),
				Denom:            denom,
			})

			s.Require().NoError(err)
			s.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(denom, 1000000)), amountRes.TotalDelegatedCoins)

			delegationsRes, err := s.queryClient.SuperfluidDelegationsByValidatorDenom(s.Ctx, &types.SuperfluidDelegationsByValidatorDenomRequest{
				ValidatorAddress: validator.String(),
				Denom:            denom,
			})
			s.Require().NoError(err)
			s.Require().Len(delegationsRes.SuperfluidDelegationRecords, 1)
		}
	}

	totalSuperfluidDelegationsRes, err := s.queryClient.TotalSuperfluidDelegations(s.Ctx, &types.TotalSuperfluidDelegationsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(40000000), totalSuperfluidDelegationsRes.TotalDelegations)

	for _, lockID := range locks {
		connectedIntermediaryAccountRes, err := s.queryClient.ConnectedIntermediaryAccount(s.Ctx, &types.ConnectedIntermediaryAccountRequest{LockId: lockID.ID})
		s.Require().NoError(err)
		s.Require().NotEqual("", connectedIntermediaryAccountRes.Account.Denom)
		s.Require().NotEqual("", connectedIntermediaryAccountRes.Account.Address)
		s.Require().NotEqual(uint64(0), connectedIntermediaryAccountRes.Account.GaugeId)
	}
	connectedIntermediaryAccountRes, err := s.queryClient.ConnectedIntermediaryAccount(s.Ctx, &types.ConnectedIntermediaryAccountRequest{LockId: 123})
	s.Require().NoError(err)
	s.Require().Equal("", connectedIntermediaryAccountRes.Account.Denom)
	s.Require().Equal("", connectedIntermediaryAccountRes.Account.ValAddr)
	s.Require().Equal(uint64(0), connectedIntermediaryAccountRes.Account.GaugeId)
}

func (s *KeeperTestSuite) TestGRPCQuerySuperfluidDelegationsDontIncludeUnbonding() {
	s.SetupTest()

	// setup 2 validators
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})
	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, 0, 1000000},
		{0, 1, 1, 1000000},
		{1, 0, 1, 1000000},
		{1, 1, 0, 1000000},
	}

	// setup superfluid delegations
	delegatorAddresses, _, locks := s.setupSuperfluidDelegations(valAddrs, superfluidDelegations, denoms)

	// start unbonding the superfluid delegations of denom0 from delegator0 to validator0
	err := s.querier.SuperfluidUndelegate(s.Ctx, locks[0].Owner, locks[0].ID)
	s.Require().NoError(err)

	// query to make sure that the amount delegated for the now unbonding delegation is 0
	res, err := s.queryClient.SuperfluidDelegationAmount(s.Ctx, &types.SuperfluidDelegationAmountRequest{
		DelegatorAddress: delegatorAddresses[0].String(),
		ValidatorAddress: valAddrs[0].String(),
		Denom:            denoms[0],
	})
	s.Require().NoError(err)
	s.Require().Equal(res.Amount.AmountOf(denoms[0]).Int64(), int64(0))

	// query to make sure that the unbonding delegation is not included in delegator query
	res2, err := s.queryClient.SuperfluidDelegationsByDelegator(s.Ctx, &types.SuperfluidDelegationsByDelegatorRequest{
		DelegatorAddress: delegatorAddresses[0].String(),
	})
	s.Require().NoError(err)
	s.Require().Len(res2.SuperfluidDelegationRecords, 1)
	s.Require().Equal(sdk.NewCoins(
		sdk.NewInt64Coin(denoms[1], 1000000)), res2.TotalDelegatedCoins)

	// query to make sure that the unbonding delegation is not included in the validator denom pair query
	amountRes, err := s.queryClient.EstimateSuperfluidDelegatedAmountByValidatorDenom(s.Ctx, &types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{
		ValidatorAddress: valAddrs[1].String(),
		Denom:            denoms[0],
	})

	s.Require().NoError(err)
	s.Require().True(amountRes.TotalDelegatedCoins.Equal(sdk.NewCoins(
		sdk.NewInt64Coin(denoms[0], 1000000),
	)))

	delegationsRes, err := s.queryClient.SuperfluidDelegationsByValidatorDenom(s.Ctx, &types.SuperfluidDelegationsByValidatorDenomRequest{
		ValidatorAddress: valAddrs[1].String(),
		Denom:            denoms[0],
	})
	s.Require().NoError(err)
	s.Require().Len(delegationsRes.SuperfluidDelegationRecords, 1)

	totalSuperfluidDelegationsRes, err := s.queryClient.TotalSuperfluidDelegations(s.Ctx, &types.TotalSuperfluidDelegationsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(totalSuperfluidDelegationsRes.TotalDelegations, osmomath.NewInt(30000000))
}

func (s *KeeperTestSuite) TestUserConcentratedSuperfluidPositionsBondedAndUnbonding() {
	s.SetupTest()

	// Setup 2 validators.
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	// Set staking parameters (needed since stake is not a valid quote denom).
	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	stakingParams.BondDenom = appparams.BaseCoinUnit
	s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)

	coins := sdk.NewCoins(sdk.NewCoin("token0", osmomath.NewInt(1000000000000)), sdk.NewCoin(stakingParams.BondDenom, osmomath.NewInt(1000000000000)))

	// Prepare 2 concentrated pools.
	clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(coins[0].Denom, coins[1].Denom)
	clPoolId := clPool.GetId()
	denom := cltypes.GetConcentratedLockupDenomFromPoolId(1)

	clPool2 := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(coins[0].Denom, coins[1].Denom)
	clPoolId2 := clPool2.GetId()
	denom2 := cltypes.GetConcentratedLockupDenomFromPoolId(2)

	// Add both pools as superfluid assets.
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     denom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})
	s.Require().NoError(err)

	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     denom2,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})
	s.Require().NoError(err)

	duration := stakingParams.UnbondingTime

	// Create 4 positions in pool 1 that are superfluid delegated.
	expectedBondedPositionIds := []uint64{}
	expectedBondedLockIds := []uint64{}
	expectedBondedTotalSharesLocked := sdk.Coins{}
	for i := 0; i < 4; i++ {
		positionData, lockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPoolId, s.TestAccs[0], coins, duration)
		s.Require().NoError(err)

		lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
		s.Require().NoError(err)

		err = s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, lock.Owner, lock.ID, valAddrs[0].String())
		s.Require().NoError(err)

		expectedBondedPositionIds = append(expectedBondedPositionIds, positionData.ID)
		expectedBondedLockIds = append(expectedBondedLockIds, lockId)
		expectedBondedTotalSharesLocked = expectedBondedTotalSharesLocked.Add(lock.Coins[0])
	}

	// Create 1 position in pool 1 that is not superfluid delegated.
	_, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPoolId, s.TestAccs[0], coins)
	s.Require().NoError(err)

	// Create 4 positions in pool 2 that are superfluid undelegating.
	expectedUnbondingPositionIds := []uint64{}
	expectedUnbondingLockIds := []uint64{}
	expectedUnbondingTotalSharesLocked := sdk.Coins{}
	for i := 0; i < 4; i++ {
		positionData, lockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPoolId2, s.TestAccs[0], coins, duration)
		s.Require().NoError(err)

		lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
		s.Require().NoError(err)

		err = s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, lock.Owner, lock.ID, valAddrs[0].String())
		s.Require().NoError(err)

		_, err = s.App.SuperfluidKeeper.SuperfluidUndelegateAndUnbondLock(s.Ctx, lockId, lock.Owner, lock.Coins[0].Amount)
		s.Require().NoError(err)

		expectedUnbondingPositionIds = append(expectedUnbondingPositionIds, positionData.ID)
		expectedUnbondingLockIds = append(expectedUnbondingLockIds, lockId)
		expectedUnbondingTotalSharesLocked = expectedUnbondingTotalSharesLocked.Add(lock.Coins[0])
	}

	// Create 1 position in pool 2 that is not superfluid delegated.
	_, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPoolId2, s.TestAccs[0], coins)
	s.Require().NoError(err)

	// Query the bonded positions.
	bondedRes, err := s.queryClient.UserConcentratedSuperfluidPositionsDelegated(s.Ctx, &types.UserConcentratedSuperfluidPositionsDelegatedRequest{
		DelegatorAddress: s.TestAccs[0].String(),
	})
	s.Require().NoError(err)

	// The result should only have the four bonded superfluid positions
	s.Require().Equal(4, len(bondedRes.ClPoolUserPositionRecords))
	s.Require().Equal(4, len(expectedBondedPositionIds))
	s.Require().Equal(4, len(expectedBondedLockIds))

	actualBondedPositionIds := []uint64{}
	actualBondedLockIds := []uint64{}
	actualBondedTotalSharesLocked := sdk.Coins{}
	for _, record := range bondedRes.ClPoolUserPositionRecords {
		s.Require().Equal(record.ValidatorAddress, valAddrs[0].String()) // User 0 only used this validator
		actualBondedPositionIds = append(actualBondedPositionIds, record.PositionId)
		actualBondedLockIds = append(actualBondedLockIds, record.LockId)
		actualBondedTotalSharesLocked = actualBondedTotalSharesLocked.Add(record.DelegationAmount)
	}

	s.Require().True(osmoutils.ContainsDuplicateDeepEqual([]interface{}{expectedBondedPositionIds, actualBondedPositionIds}))
	s.Require().True(osmoutils.ContainsDuplicateDeepEqual([]interface{}{expectedBondedLockIds, actualBondedLockIds}))
	s.Require().Equal(expectedBondedTotalSharesLocked, actualBondedTotalSharesLocked)

	// Query the unbonding positions.
	unbondingRes, err := s.queryClient.UserConcentratedSuperfluidPositionsUndelegating(s.Ctx, &types.UserConcentratedSuperfluidPositionsUndelegatingRequest{
		DelegatorAddress: s.TestAccs[0].String(),
	})
	s.Require().NoError(err)

	// The result should only have the four unbonding superfluid positions
	s.Require().Equal(4, len(unbondingRes.ClPoolUserPositionRecords))
	s.Require().Equal(4, len(expectedUnbondingPositionIds))
	s.Require().Equal(4, len(expectedUnbondingLockIds))

	actualUnbondingPositionIds := []uint64{}
	actualUnbondingLockIds := []uint64{}
	actualUnbondingTotalSharesLocked := sdk.Coins{}
	for _, record := range unbondingRes.ClPoolUserPositionRecords {
		s.Require().Equal(record.ValidatorAddress, valAddrs[0].String()) // User 0 only used this validator
		actualUnbondingPositionIds = append(actualUnbondingPositionIds, record.PositionId)
		actualUnbondingLockIds = append(actualUnbondingLockIds, record.LockId)
		actualUnbondingTotalSharesLocked = actualUnbondingTotalSharesLocked.Add(record.DelegationAmount)
	}

	s.Require().True(osmoutils.ContainsDuplicateDeepEqual([]interface{}{expectedUnbondingPositionIds, actualUnbondingPositionIds}))
	s.Require().True(osmoutils.ContainsDuplicateDeepEqual([]interface{}{expectedUnbondingLockIds, actualUnbondingLockIds}))
	s.Require().Equal(expectedUnbondingTotalSharesLocked, actualUnbondingTotalSharesLocked)
}

func (s *KeeperTestSuite) TestGRPCQueryTotalDelegationByDelegator() {
	s.SetupTest()

	// setup 2 validators
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded})

	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

	// create a delegation of 1000000 for every combination of 2 delegations, 2 validators, and 2 superfluid denoms
	superfluidDelegations := []superfluidDelegation{
		{0, 0, 0, 1000000},
		{0, 1, 1, 1000000},
		{1, 0, 1, 1000000},
		{1, 1, 0, 1000000},
	}

	// setup superfluid delegations
	delegatorAddresses, _, _ := s.setupSuperfluidDelegations(valAddrs, superfluidDelegations, denoms)

	// setup normal delegations
	bond0to0 := stakingtypes.NewDelegation(delegatorAddresses[0].String(), valAddrs[0].String(), osmomath.NewDec(9000000))
	bond0to1 := stakingtypes.NewDelegation(delegatorAddresses[0].String(), valAddrs[1].String(), osmomath.NewDec(9000000))
	bond1to0 := stakingtypes.NewDelegation(delegatorAddresses[1].String(), valAddrs[0].String(), osmomath.NewDec(9000000))
	bond1to1 := stakingtypes.NewDelegation(delegatorAddresses[1].String(), valAddrs[1].String(), osmomath.NewDec(9000000))

	s.App.StakingKeeper.SetDelegation(s.Ctx, bond0to0)
	s.App.StakingKeeper.SetDelegation(s.Ctx, bond0to1)
	s.App.StakingKeeper.SetDelegation(s.Ctx, bond1to0)
	s.App.StakingKeeper.SetDelegation(s.Ctx, bond1to1)

	multiplier0 := s.querier.Keeper.GetOsmoEquivalentMultiplier(s.Ctx, denoms[0])
	multiplier1 := s.querier.Keeper.GetOsmoEquivalentMultiplier(s.Ctx, denoms[1])
	minRiskFactor := s.querier.Keeper.GetParams(s.Ctx).MinimumRiskFactor

	expectAmount0 := multiplier0.Mul(osmomath.NewDec(1000000)).Sub(multiplier0.Mul(osmomath.NewDec(1000000)).Mul(minRiskFactor))
	expectAmount1 := multiplier1.Mul(osmomath.NewDec(1000000)).Sub(multiplier1.Mul(osmomath.NewDec(1000000)).Mul(minRiskFactor))

	// for each delegator, query all their superfluid delegations and normal delegations. Making sure they have 4 delegations
	// Making sure TotalEquivalentStakedAmount is equal to converted amount + normal delegations
	for _, delegator := range delegatorAddresses {
		res, err := s.queryClient.TotalDelegationByDelegator(s.Ctx, &types.QueryTotalDelegationByDelegatorRequest{
			DelegatorAddress: delegator.String(),
		})

		fmt.Printf("res = %v \n", res)

		s.Require().NoError(err)
		s.Require().Len(res.SuperfluidDelegationRecords, 2)
		s.Require().Len(res.DelegationResponse, 2)
		s.Require().True(res.TotalDelegatedCoins.Equal(sdk.NewCoins(
			sdk.NewInt64Coin(denoms[0], 1000000),
			sdk.NewInt64Coin(denoms[1], 1000000),
			sdk.NewInt64Coin(appparams.BaseCoinUnit, 18000000),
		)))

		total_osmo_equivalent := sdk.NewCoin(appparams.BaseCoinUnit, expectAmount0.RoundInt().Add(expectAmount1.RoundInt()).Add(osmomath.NewInt(18000000)))

		s.Require().True(res.TotalEquivalentStakedAmount.Equal(total_osmo_equivalent))
	}
}
