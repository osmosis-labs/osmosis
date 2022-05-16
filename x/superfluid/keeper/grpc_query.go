package keeper

import (
	"context"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/types"
)

var _ types.QueryServer = Keeper{}

// Params returns the superfluid module params
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// AssetType Returns superfluid asset type
func (k Keeper) AssetType(goCtx context.Context, req *types.AssetTypeRequest) (*types.AssetTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	asset := k.GetSuperfluidAsset(ctx, req.Denom)
	return &types.AssetTypeResponse{
		AssetType: asset.AssetType,
	}, nil
}

// AllAssets Returns all superfluid assets info
func (k Keeper) AllAssets(goCtx context.Context, req *types.AllAssetsRequest) (*types.AllAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assets := k.GetAllSuperfluidAssets(ctx)
	return &types.AllAssetsResponse{
		Assets: assets,
	}, nil
}

// AssetMultiplier returns superfluid asset multiplier
func (k Keeper) AssetMultiplier(goCtx context.Context, req *types.AssetMultiplierRequest) (*types.AssetMultiplierResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	epochInfo := k.ek.GetEpochInfo(ctx, k.GetEpochIdentifier(ctx))

	return &types.AssetMultiplierResponse{
		OsmoEquivalentMultiplier: &types.OsmoEquivalentMultiplierRecord{
			EpochNumber: epochInfo.CurrentEpoch,
			Denom:       req.Denom,
			Multiplier:  k.GetOsmoEquivalentMultiplier(ctx, req.Denom),
		},
	}, nil
}

// AllIntermediaryAccounts returns all superfluid intermediary accounts
func (k Keeper) AllIntermediaryAccounts(goCtx context.Context, req *types.AllIntermediaryAccountsRequest) (*types.AllIntermediaryAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accounts := k.GetAllIntermediaryAccounts(ctx)
	accInfos := []types.SuperfluidIntermediaryAccountInfo{}
	for _, acc := range accounts {
		accInfos = append(accInfos, types.SuperfluidIntermediaryAccountInfo{
			Denom:   acc.Denom,
			ValAddr: acc.ValAddr,
			GaugeId: acc.GaugeId,
			Address: acc.GetAccAddress().String(),
		})
	}
	return &types.AllIntermediaryAccountsResponse{
		Accounts: accInfos,
	}, nil
}

// ConnectedIntermediaryAccount returns intermediary account connected to a superfluid staked lock by id
func (k Keeper) ConnectedIntermediaryAccount(goCtx context.Context, req *types.ConnectedIntermediaryAccountRequest) (*types.ConnectedIntermediaryAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := k.GetLockIdIntermediaryAccountConnection(ctx, req.LockId)
	acc := k.GetIntermediaryAccount(ctx, address)

	if len(acc.Denom) == 0 && acc.GaugeId == uint64(0) && len(acc.ValAddr) == 0 {
		return &types.ConnectedIntermediaryAccountResponse{
			Account: &types.SuperfluidIntermediaryAccountInfo{
				Denom:   acc.Denom,
				ValAddr: acc.ValAddr,
				GaugeId: acc.GaugeId,
				Address: "",
			},
		}, nil
	}

	return &types.ConnectedIntermediaryAccountResponse{
		Account: &types.SuperfluidIntermediaryAccountInfo{
			Denom:   acc.Denom,
			ValAddr: acc.ValAddr,
			GaugeId: acc.GaugeId,
			Address: acc.GetAccAddress().String(),
		},
	}, nil
}

// SuperfluidDelegationAmount returns the coins superfluid delegated for a
//delegator, validator, denom triplet
func (k Keeper) SuperfluidDelegationAmount(goCtx context.Context, req *types.SuperfluidDelegationAmountRequest) (*types.SuperfluidDelegationAmountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.GetSuperfluidAsset(ctx, req.Denom).Denom == "" {
		return nil, types.ErrNonSuperfluidAsset
	}

	_, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	syntheticDenom := stakingSyntheticDenom(req.Denom, req.ValidatorAddress)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	periodLocks := k.lk.GetAccountLockedLongerDurationDenomNotUnlockingOnly(ctx, delAddr, syntheticDenom, time.Second)

	if len(periodLocks) == 0 {
		return &types.SuperfluidDelegationAmountResponse{sdk.NewCoins()}, nil
	}

	return &types.SuperfluidDelegationAmountResponse{periodLocks[0].GetCoins()}, nil
}

// SuperfluidDelegationsByDelegator returns all the superfluid poistions for a specific delegator
func (k Keeper) SuperfluidDelegationsByDelegator(goCtx context.Context, req *types.SuperfluidDelegationsByDelegatorRequest) (*types.SuperfluidDelegationsByDelegatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	res := types.SuperfluidDelegationsByDelegatorResponse{
		SuperfluidDelegationRecords: []types.SuperfluidDelegationRecord{},
		TotalDelegatedCoins:         sdk.NewCoins(),
	}

	syntheticLocks := k.lk.GetAllSyntheticLockupsByAddr(ctx, delAddr)

	for _, syntheticLock := range syntheticLocks {
		// don't include unbonding delegations
		if strings.Contains(syntheticLock.SynthDenom, "superunbonding") {
			continue
		}

		periodLock, err := k.lk.GetLockByID(ctx, syntheticLock.UnderlyingLockId)
		if err != nil {
			return nil, err
		}

		baseDenom := periodLock.Coins.GetDenomByIndex(0)
		lockedCoins := sdk.NewCoin(baseDenom, periodLock.GetCoins().AmountOf(baseDenom))
		valAddr, err := ValidatorAddressFromSyntheticDenom(syntheticLock.SynthDenom)
		if err != nil {
			return nil, err
		}
		res.SuperfluidDelegationRecords = append(res.SuperfluidDelegationRecords,
			types.SuperfluidDelegationRecord{
				DelegatorAddress: req.DelegatorAddress,
				ValidatorAddress: valAddr,
				DelegationAmount: lockedCoins,
			},
		)
		res.TotalDelegatedCoins = res.TotalDelegatedCoins.Add(lockedCoins)
	}
	return &res, nil

}

// SuperfluidUndelegationsByDelegator returns total amount undelegating by delegator
func (k Keeper) SuperfluidUndelegationsByDelegator(goCtx context.Context, req *types.SuperfluidUndelegationsByDelegatorRequest) (*types.SuperfluidUndelegationsByDelegatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	res := types.SuperfluidUndelegationsByDelegatorResponse{
		SuperfluidDelegationRecords: []types.SuperfluidDelegationRecord{},
		TotalUndelegatedCoins:       sdk.NewCoins(),
		SyntheticLocks:              []lockuptypes.SyntheticLock{},
	}

	syntheticLocks := k.lk.GetAllSyntheticLockupsByAddr(ctx, delAddr)

	for _, syntheticLock := range syntheticLocks {
		if strings.Contains(syntheticLock.SynthDenom, "superbonding") {
			continue
		}

		periodLock, err := k.lk.GetLockByID(ctx, syntheticLock.UnderlyingLockId)
		if err != nil {
			return nil, err
		}

		baseDenom := periodLock.Coins.GetDenomByIndex(0)
		lockedCoin := sdk.NewCoin(baseDenom, periodLock.GetCoins().AmountOf(baseDenom))
		valAddr, err := ValidatorAddressFromSyntheticDenom(syntheticLock.SynthDenom)
		if err != nil {
			return nil, err
		}
		res.SuperfluidDelegationRecords = append(res.SuperfluidDelegationRecords,
			types.SuperfluidDelegationRecord{
				DelegatorAddress: req.DelegatorAddress,
				ValidatorAddress: valAddr,
				DelegationAmount: lockedCoin,
			},
		)
		res.SyntheticLocks = append(res.SyntheticLocks, syntheticLock)
		res.TotalUndelegatedCoins = res.TotalUndelegatedCoins.Add(lockedCoin)
	}
	return &res, nil
}

// SuperfluidDelegationsByValidatorDenom returns all the superfluid positions
// of a specific denom delegated to one validator
func (k Keeper) SuperfluidDelegationsByValidatorDenom(goCtx context.Context, req *types.SuperfluidDelegationsByValidatorDenomRequest) (*types.SuperfluidDelegationsByValidatorDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.GetSuperfluidAsset(ctx, req.Denom).Denom == "" {
		return nil, types.ErrNonSuperfluidAsset
	}

	_, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	syntheticDenom := stakingSyntheticDenom(req.Denom, req.ValidatorAddress)

	res := types.SuperfluidDelegationsByValidatorDenomResponse{
		SuperfluidDelegationRecords: []types.SuperfluidDelegationRecord{},
	}

	periodLocks := k.lk.GetLocksLongerThanDurationDenom(ctx, syntheticDenom, time.Second)

	for _, lock := range periodLocks {
		lockedCoins := sdk.NewCoin(req.Denom, lock.GetCoins().AmountOf(req.Denom))
		res.SuperfluidDelegationRecords = append(res.SuperfluidDelegationRecords,
			types.SuperfluidDelegationRecord{
				DelegatorAddress: lock.GetOwner(),
				ValidatorAddress: req.ValidatorAddress,
				DelegationAmount: lockedCoins,
			},
		)
	}
	return &res, nil
}

// EstimateSuperfluidDelegatedAmountByValidatorDenom returns the amount of a
// specific denom delegated to a specific validator
// This is labeled an estimate, because the way it calculates the amount can
// lead rounding errors from the true delegated amount
func (k Keeper) EstimateSuperfluidDelegatedAmountByValidatorDenom(goCtx context.Context, req *types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest) (*types.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.GetSuperfluidAsset(ctx, req.Denom).Denom == "" {
		return nil, types.ErrNonSuperfluidAsset
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	intermediaryAccAddress := types.GetSuperfluidIntermediaryAccountAddr(req.Denom, req.ValidatorAddress)
	intermediaryAcc := k.GetIntermediaryAccount(ctx, intermediaryAccAddress)
	if intermediaryAcc.Empty() {
		return nil, err
	}

	val, found := k.sk.GetValidator(ctx, valAddr)
	if !found {
		return nil, stakingtypes.ErrNoValidatorFound
	}

	delegation, found := k.sk.GetDelegation(ctx, intermediaryAcc.GetAccAddress(), valAddr)
	if !found {
		return nil, stakingtypes.ErrNoDelegation
	}

	syntheticOsmoAmt := delegation.Shares.Quo(val.DelegatorShares).MulInt(val.Tokens)

	baseAmount := k.UnriskAdjustOsmoValue(ctx, syntheticOsmoAmt).Quo(k.GetOsmoEquivalentMultiplier(ctx, req.Denom)).RoundInt()
	return &types.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse{
		TotalDelegatedCoins: sdk.NewCoins(sdk.NewCoin(req.Denom, baseAmount)),
	}, nil
}

// TotalSuperfluidDelegations returns total amount of osmo delegated via superfluid staking
func (k Keeper) TotalSuperfluidDelegations(goCtx context.Context, req *types.TotalSuperfluidDelegationsRequest) (*types.TotalSuperfluidDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalSuperfluidDelegated := sdk.NewInt(0)

	intermediaryAccounts := k.GetAllIntermediaryAccounts(ctx)
	for _, intermediaryAccount := range intermediaryAccounts {
		valAddr, err := sdk.ValAddressFromBech32(intermediaryAccount.ValAddr)
		if err != nil {
			return nil, err
		}

		val, found := k.sk.GetValidator(ctx, valAddr)
		if !found {
			return nil, stakingtypes.ErrNoValidatorFound
		}

		delegation, found := k.sk.GetDelegation(ctx, intermediaryAccount.GetAccAddress(), valAddr)
		if !found {
			continue
		}

		syntheticOsmoAmt := delegation.Shares.Quo(val.DelegatorShares).MulInt(val.Tokens).RoundInt()
		totalSuperfluidDelegated = totalSuperfluidDelegated.Add(syntheticOsmoAmt)
	}
	return &types.TotalSuperfluidDelegationsResponse{
		TotalDelegations: totalSuperfluidDelegated,
	}, nil
}
