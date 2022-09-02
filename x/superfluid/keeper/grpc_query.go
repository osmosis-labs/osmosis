package keeper

import (
	"context"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appparams "github.com/osmosis-labs/osmosis/v11/app/params"

	lockuptypes "github.com/osmosis-labs/osmosis/v11/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v11/x/superfluid/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/superfluid keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params returns the superfluid module params.
func (q Querier) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// AssetType Returns superfluid asset type.
func (q Querier) AssetType(goCtx context.Context, req *types.AssetTypeRequest) (*types.AssetTypeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Denom) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	asset := q.Keeper.GetSuperfluidAsset(ctx, req.Denom)
	return &types.AssetTypeResponse{
		AssetType: asset.AssetType,
	}, nil
}

// AllAssets Returns all superfluid assets info.
func (q Querier) AllAssets(goCtx context.Context, _ *types.AllAssetsRequest) (*types.AllAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assets := q.Keeper.GetAllSuperfluidAssets(ctx)
	return &types.AllAssetsResponse{
		Assets: assets,
	}, nil
}

// AssetMultiplier returns superfluid asset multiplier.
func (q Querier) AssetMultiplier(goCtx context.Context, req *types.AssetMultiplierRequest) (*types.AssetMultiplierResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Denom) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	epochInfo := q.Keeper.ek.GetEpochInfo(ctx, q.Keeper.GetEpochIdentifier(ctx))

	return &types.AssetMultiplierResponse{
		OsmoEquivalentMultiplier: &types.OsmoEquivalentMultiplierRecord{
			EpochNumber: epochInfo.CurrentEpoch,
			Denom:       req.Denom,
			Multiplier:  q.Keeper.GetOsmoEquivalentMultiplier(ctx, req.Denom),
		},
	}, nil
}

// AllIntermediaryAccounts returns all superfluid intermediary accounts.
func (q Querier) AllIntermediaryAccounts(goCtx context.Context, _ *types.AllIntermediaryAccountsRequest) (*types.AllIntermediaryAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accounts := q.Keeper.GetAllIntermediaryAccounts(ctx)
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

// ConnectedIntermediaryAccount returns intermediary account connected to a superfluid staked lock by id.
func (q Querier) ConnectedIntermediaryAccount(goCtx context.Context, req *types.ConnectedIntermediaryAccountRequest) (*types.ConnectedIntermediaryAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	address := q.Keeper.GetLockIdIntermediaryAccountConnection(ctx, req.LockId)
	acc := q.Keeper.GetIntermediaryAccount(ctx, address)

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

func (q Querier) SuperfluidDelegationAmount(goCtx context.Context, req *types.SuperfluidDelegationAmountRequest) (*types.SuperfluidDelegationAmountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Denom) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}
	if len(req.ValidatorAddress) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty validator address")
	}
	if len(req.DelegatorAddress) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if q.Keeper.GetSuperfluidAsset(ctx, req.Denom).Denom == "" {
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

	periodLocks := q.Keeper.lk.GetAccountLockedLongerDurationDenomNotUnlockingOnly(ctx, delAddr, syntheticDenom, time.Second)

	if len(periodLocks) == 0 {
		return &types.SuperfluidDelegationAmountResponse{Amount: sdk.NewCoins()}, nil
	}

	return &types.SuperfluidDelegationAmountResponse{Amount: periodLocks[0].GetCoins()}, nil
}

// SuperfluidDelegationsByDelegator returns all the superfluid poistions for a specific delegator.
func (q Querier) SuperfluidDelegationsByDelegator(goCtx context.Context, req *types.SuperfluidDelegationsByDelegatorRequest) (*types.SuperfluidDelegationsByDelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.DelegatorAddress) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	res := types.SuperfluidDelegationsByDelegatorResponse{
		SuperfluidDelegationRecords: []types.SuperfluidDelegationRecord{},
		TotalDelegatedCoins:         sdk.NewCoins(),
		TotalEquivalentStakedAmount: sdk.NewCoin(appparams.BaseCoinUnit, sdk.ZeroInt()),
	}

	syntheticLocks := q.Keeper.lk.GetAllSyntheticLockupsByAddr(ctx, delAddr)

	for _, syntheticLock := range syntheticLocks {
		// don't include unbonding delegations
		if strings.Contains(syntheticLock.SynthDenom, "superunbonding") {
			continue
		}

		periodLock, err := q.Keeper.lk.GetLockByID(ctx, syntheticLock.UnderlyingLockId)
		if err != nil {
			return nil, err
		}

		baseDenom := periodLock.Coins.GetDenomByIndex(0)
		lockedCoins := sdk.NewCoin(baseDenom, periodLock.GetCoins().AmountOf(baseDenom))
		valAddr, err := ValidatorAddressFromSyntheticDenom(syntheticLock.SynthDenom)

		// Find how many osmo tokens this delegation is worth at superfluids current risk adjustment
		// and twap of the denom.
		equivalentAmount := q.Keeper.GetSuperfluidOSMOTokens(ctx, baseDenom, lockedCoins.Amount)
		coin := sdk.NewCoin(appparams.BaseCoinUnit, equivalentAmount)

		if err != nil {
			return nil, err
		}
		res.SuperfluidDelegationRecords = append(res.SuperfluidDelegationRecords,
			types.SuperfluidDelegationRecord{
				DelegatorAddress:       req.DelegatorAddress,
				ValidatorAddress:       valAddr,
				DelegationAmount:       lockedCoins,
				EquivalentStakedAmount: &coin,
			},
		)
		res.TotalDelegatedCoins = res.TotalDelegatedCoins.Add(lockedCoins)
		res.TotalEquivalentStakedAmount = res.TotalEquivalentStakedAmount.Add(coin)
	}

	return &res, nil
}

// SuperfluidUndelegationsByDelegator returns total amount undelegating by delegator.
func (q Querier) SuperfluidUndelegationsByDelegator(goCtx context.Context, req *types.SuperfluidUndelegationsByDelegatorRequest) (*types.SuperfluidUndelegationsByDelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.DelegatorAddress) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
	}

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

	syntheticLocks := q.Keeper.lk.GetAllSyntheticLockupsByAddr(ctx, delAddr)

	for _, syntheticLock := range syntheticLocks {
		if strings.Contains(syntheticLock.SynthDenom, "superbonding") {
			continue
		}

		periodLock, err := q.Keeper.lk.GetLockByID(ctx, syntheticLock.UnderlyingLockId)
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
// of a specific denom delegated to one validator.
func (q Querier) SuperfluidDelegationsByValidatorDenom(goCtx context.Context, req *types.SuperfluidDelegationsByValidatorDenomRequest) (*types.SuperfluidDelegationsByValidatorDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Denom) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}
	if len(req.ValidatorAddress) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty validator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if q.Keeper.GetSuperfluidAsset(ctx, req.Denom).Denom == "" {
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

	periodLocks := q.Keeper.lk.GetLocksLongerThanDurationDenom(ctx, syntheticDenom, time.Second)

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
// lead rounding errors from the true delegated amount.
func (q Querier) EstimateSuperfluidDelegatedAmountByValidatorDenom(goCtx context.Context, req *types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest) (*types.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Denom) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}
	if len(req.ValidatorAddress) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty validator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if q.Keeper.GetSuperfluidAsset(ctx, req.Denom).Denom == "" {
		return nil, types.ErrNonSuperfluidAsset
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	intermediaryAccAddress := types.GetSuperfluidIntermediaryAccountAddr(req.Denom, req.ValidatorAddress)
	intermediaryAcc := q.Keeper.GetIntermediaryAccount(ctx, intermediaryAccAddress)
	if intermediaryAcc.Empty() {
		return nil, err
	}

	val, found := q.Keeper.sk.GetValidator(ctx, valAddr)
	if !found {
		return nil, stakingtypes.ErrNoValidatorFound
	}

	delegation, found := q.Keeper.sk.GetDelegation(ctx, intermediaryAcc.GetAccAddress(), valAddr)
	if !found {
		return nil, stakingtypes.ErrNoDelegation
	}

	syntheticOsmoAmt := delegation.Shares.Quo(val.DelegatorShares).MulInt(val.Tokens)
	baseAmount := q.Keeper.UnriskAdjustOsmoValue(ctx, syntheticOsmoAmt).Quo(q.Keeper.GetOsmoEquivalentMultiplier(ctx, req.Denom)).RoundInt()

	return &types.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse{
		TotalDelegatedCoins: sdk.NewCoins(sdk.NewCoin(req.Denom, baseAmount)),
	}, nil
}

// TotalSuperfluidDelegations returns total amount of osmo delegated via superfluid staking.
func (q Querier) TotalSuperfluidDelegations(goCtx context.Context, _ *types.TotalSuperfluidDelegationsRequest) (*types.TotalSuperfluidDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalSuperfluidDelegated := sdk.NewInt(0)

	intermediaryAccounts := q.Keeper.GetAllIntermediaryAccounts(ctx)
	for _, intermediaryAccount := range intermediaryAccounts {
		valAddr, err := sdk.ValAddressFromBech32(intermediaryAccount.ValAddr)
		if err != nil {
			return nil, err
		}

		val, found := q.Keeper.sk.GetValidator(ctx, valAddr)
		if !found {
			return nil, stakingtypes.ErrNoValidatorFound
		}

		delegation, found := q.Keeper.sk.GetDelegation(ctx, intermediaryAccount.GetAccAddress(), valAddr)
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

func (q Querier) TotalDelegationByDelegator(goCtx context.Context, req *types.QueryTotalDelegationByDelegatorRequest) (*types.QueryTotalDelegationByDelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.DelegatorAddress) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	superfluidDelegationResp, err := q.SuperfluidDelegationsByDelegator(goCtx, &types.SuperfluidDelegationsByDelegatorRequest{
		DelegatorAddress: req.DelegatorAddress,
	})
	if err != nil {
		return nil, err
	}

	res := types.QueryTotalDelegationByDelegatorResponse{
		SuperfluidDelegationRecords: superfluidDelegationResp.SuperfluidDelegationRecords,
		DelegationResponse:          []stakingtypes.DelegationResponse{},
		TotalDelegatedCoins:         superfluidDelegationResp.TotalDelegatedCoins,
		TotalEquivalentStakedAmount: superfluidDelegationResp.TotalEquivalentStakedAmount,
	}

	// this is for getting normal staking
	q.sk.IterateDelegations(ctx, delAddr, func(_ int64, del stakingtypes.DelegationI) bool {
		val, found := q.sk.GetValidator(ctx, del.GetValidatorAddr())
		if !found {
			return true
		}

		lockedCoins := sdk.NewCoin(appparams.BaseCoinUnit, val.TokensFromShares(del.GetShares()).TruncateInt())

		res.DelegationResponse = append(res.DelegationResponse,
			stakingtypes.DelegationResponse{
				Delegation: stakingtypes.Delegation{
					DelegatorAddress: del.GetDelegatorAddr().String(),
					ValidatorAddress: del.GetValidatorAddr().String(),
					Shares:           del.GetShares(),
				},
				Balance: lockedCoins,
			},
		)

		res.TotalDelegatedCoins = res.TotalDelegatedCoins.Add(lockedCoins)
		res.TotalEquivalentStakedAmount = res.TotalEquivalentStakedAmount.Add(lockedCoins)

		return false
	})

	return &res, nil
}
