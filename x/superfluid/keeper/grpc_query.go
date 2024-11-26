package keeper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
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
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	asset, err := q.Keeper.GetSuperfluidAsset(ctx, req.Denom)
	if err != nil {
		return nil, err
	}
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
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
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
func (q Querier) AllIntermediaryAccounts(goCtx context.Context, req *types.AllIntermediaryAccountsRequest) (*types.AllIntermediaryAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	store := sdkCtx.KVStore(q.Keeper.storeKey)
	accStore := prefix.NewStore(store, types.KeyPrefixIntermediaryAccount)
	iterator := storetypes.KVStorePrefixIterator(accStore, nil)
	defer iterator.Close()

	accInfos := []types.SuperfluidIntermediaryAccountInfo{}

	pageRes, err := query.FilteredPaginate(accStore, req.Pagination,
		func(key, value []byte, accumulate bool) (bool, error) {
			account := types.SuperfluidIntermediaryAccount{}
			err := proto.Unmarshal(iterator.Value(), &account)
			if err != nil {
				return false, err
			}
			iterator.Next()

			accountInfo := types.SuperfluidIntermediaryAccountInfo{
				Denom:   account.Denom,
				ValAddr: account.ValAddr,
				GaugeId: account.GaugeId,
				Address: account.GetAccAddress().String(),
			}
			if accumulate {
				accInfos = append(accInfos, accountInfo)
			}
			return true, nil
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.AllIntermediaryAccountsResponse{
		Accounts:   accInfos,
		Pagination: pageRes,
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
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}
	if len(req.ValidatorAddress) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty validator address")
	}
	if len(req.DelegatorAddress) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := q.Keeper.GetSuperfluidAsset(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	_, err = sdk.ValAddressFromBech32(req.ValidatorAddress)
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

// SuperfluidDelegationsByDelegator returns all the superfluid positions for a specific delegator.
func (q Querier) SuperfluidDelegationsByDelegator(goCtx context.Context, req *types.SuperfluidDelegationsByDelegatorRequest) (*types.SuperfluidDelegationsByDelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.DelegatorAddress) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	res := types.SuperfluidDelegationsByDelegatorResponse{
		SuperfluidDelegationRecords: []types.SuperfluidDelegationRecord{},
		TotalDelegatedCoins:         sdk.NewCoins(),
		TotalEquivalentStakedAmount: sdk.NewCoin(appparams.BaseCoinUnit, osmomath.ZeroInt()),
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
		if err != nil {
			return nil, err
		}

		// Find how many osmo tokens this delegation is worth at superfluids current risk adjustment
		// and twap of the denom.
		equivalentAmount, err := q.Keeper.GetSuperfluidOSMOTokens(ctx, baseDenom, lockedCoins.Amount)
		if err != nil {
			return nil, err
		}

		coin := sdk.NewCoin(appparams.BaseCoinUnit, equivalentAmount)

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

// UserConcentratedSuperfluidPositionsDelegated returns all the cl superfluid positions for the specified delegator across all concentrated pools that are bonded.
func (q Querier) UserConcentratedSuperfluidPositionsDelegated(goCtx context.Context, req *types.UserConcentratedSuperfluidPositionsDelegatedRequest) (*types.UserConcentratedSuperfluidPositionsDelegatedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// Get the position IDs across all pools for the given user address.
	positions, err := q.Keeper.clk.GetUserPositions(ctx, delAddr, 0)
	if err != nil {
		return nil, err
	}

	// Query each position ID and determine if it has a lock ID associated with it.
	// Construct a response with the position ID, lock ID, the amount of cl shares staked, and what those shares are worth in staked osmo tokens.
	clPoolUserPositionRecords, err := q.filterConcentratedPositionLocks(ctx, positions, false)
	if err != nil {
		return nil, err
	}

	return &types.UserConcentratedSuperfluidPositionsDelegatedResponse{
		ClPoolUserPositionRecords: clPoolUserPositionRecords,
	}, nil
}

// UserConcentratedSuperfluidPositionsUndelegating returns all the cl superfluid positions for the specified delegator across all concentrated pools that are unbonding.
func (q Querier) UserConcentratedSuperfluidPositionsUndelegating(goCtx context.Context, req *types.UserConcentratedSuperfluidPositionsUndelegatingRequest) (*types.UserConcentratedSuperfluidPositionsUndelegatingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// Get the position IDs across all pools for the given user address.
	positions, err := q.Keeper.clk.GetUserPositions(ctx, delAddr, 0)
	if err != nil {
		return nil, err
	}

	// Query each position ID and determine if it has a lock ID associated with it.
	// Construct a response with the position ID, lock ID, the amount of cl shares staked, and what those shares are worth in staked osmo tokens.
	clPoolUserPositionRecords, err := q.filterConcentratedPositionLocks(ctx, positions, true)
	if err != nil {
		return nil, err
	}

	return &types.UserConcentratedSuperfluidPositionsUndelegatingResponse{
		ClPoolUserPositionRecords: clPoolUserPositionRecords,
	}, nil
}

// SuperfluidUndelegationsByDelegator returns total amount undelegating by delegator.
func (q Querier) SuperfluidUndelegationsByDelegator(goCtx context.Context, req *types.SuperfluidUndelegationsByDelegatorRequest) (*types.SuperfluidUndelegationsByDelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.DelegatorAddress) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
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
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}
	if len(req.ValidatorAddress) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty validator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := q.Keeper.GetSuperfluidAsset(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	_, err = sdk.ValAddressFromBech32(req.ValidatorAddress)
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
		baseDenom := lock.Coins.GetDenomByIndex(0)

		equivalentAmount, err := q.Keeper.GetSuperfluidOSMOTokens(ctx, baseDenom, lockedCoins.Amount)
		if err != nil {
			return nil, err
		}

		coin := sdk.NewCoin(appparams.BaseCoinUnit, equivalentAmount)

		res.SuperfluidDelegationRecords = append(res.SuperfluidDelegationRecords,
			types.SuperfluidDelegationRecord{
				DelegatorAddress:       lock.GetOwner(),
				ValidatorAddress:       req.ValidatorAddress,
				DelegationAmount:       lockedCoins,
				EquivalentStakedAmount: &coin,
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
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}
	if len(req.ValidatorAddress) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty validator address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := q.Keeper.GetSuperfluidAsset(ctx, req.Denom)
	if err != nil {
		return nil, err
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

	val, err := q.Keeper.sk.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, stakingtypes.ErrNoValidatorFound
	}

	delegation, err := q.Keeper.sk.GetDelegation(ctx, intermediaryAcc.GetAccAddress(), valAddr)
	if err != nil {
		return nil, err
	}

	syntheticOsmoAmt := delegation.Shares.Quo(val.DelegatorShares).MulInt(val.Tokens)
	baseAmount := q.Keeper.UnriskAdjustOsmoValue(ctx, syntheticOsmoAmt).Quo(q.Keeper.GetOsmoEquivalentMultiplier(ctx, req.Denom)).RoundInt()

	return &types.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse{
		TotalDelegatedCoins: sdk.NewCoins(sdk.NewCoin(req.Denom, baseAmount)),
	}, nil
}

func (q Querier) TotalDelegationByValidatorForDenom(goCtx context.Context, req *types.QueryTotalDelegationByValidatorForDenomRequest) (*types.QueryTotalDelegationByValidatorForDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var intermediaryAccount types.SuperfluidIntermediaryAccount

	delegationsByValidator := []types.Delegations{}
	intermediaryAccounts := q.Keeper.GetAllIntermediaryAccounts(ctx)
	for _, intermediaryAccount = range intermediaryAccounts {
		if intermediaryAccount.Denom != req.Denom {
			continue
		}

		valAddr, err := sdk.ValAddressFromBech32(intermediaryAccount.ValAddr)
		if err != nil {
			return nil, err
		}

		delegation, _ := q.SuperfluidDelegationsByValidatorDenom(goCtx, &types.SuperfluidDelegationsByValidatorDenomRequest{ValidatorAddress: valAddr.String(), Denom: req.Denom})

		amount := osmomath.ZeroInt()
		for _, record := range delegation.SuperfluidDelegationRecords {
			amount = amount.Add(record.DelegationAmount.Amount)
		}

		equivalentAmountOSMO, err := q.Keeper.GetSuperfluidOSMOTokens(ctx, req.Denom, amount)
		if err != nil {
			return nil, err
		}

		result := types.Delegations{
			ValAddr:        valAddr.String(),
			AmountSfsd:     amount,
			OsmoEquivalent: equivalentAmountOSMO,
		}

		delegationsByValidator = append(delegationsByValidator, result)
	}

	return &types.QueryTotalDelegationByValidatorForDenomResponse{
		Assets: delegationsByValidator,
	}, nil
}

// TotalSuperfluidDelegations returns total amount of osmo delegated via superfluid staking.
func (q Querier) TotalSuperfluidDelegations(goCtx context.Context, _ *types.TotalSuperfluidDelegationsRequest) (*types.TotalSuperfluidDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalSuperfluidDelegated := osmomath.NewInt(0)

	intermediaryAccounts := q.Keeper.GetAllIntermediaryAccounts(ctx)
	for _, intermediaryAccount := range intermediaryAccounts {
		valAddr, err := sdk.ValAddressFromBech32(intermediaryAccount.ValAddr)
		if err != nil {
			return nil, err
		}

		val, err := q.Keeper.sk.GetValidator(ctx, valAddr)
		if err != nil {
			return nil, stakingtypes.ErrNoValidatorFound
		}

		delegation, err := q.Keeper.sk.GetDelegation(ctx, intermediaryAccount.GetAccAddress(), valAddr)
		if err != nil {
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
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty delegator address")
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
	err = q.sk.IterateDelegations(ctx, delAddr, func(_ int64, del stakingtypes.DelegationI) bool {
		valAddr, err := sdk.ValAddressFromBech32(del.GetValidatorAddr())
		if err != nil {
			return true
		}
		val, err := q.sk.GetValidator(ctx, valAddr)
		if err != nil {
			return true
		}

		lockedCoins := sdk.NewCoin(appparams.BaseCoinUnit, val.TokensFromShares(del.GetShares()).TruncateInt())

		res.DelegationResponse = append(res.DelegationResponse,
			stakingtypes.DelegationResponse{
				Delegation: stakingtypes.Delegation{
					DelegatorAddress: del.GetDelegatorAddr(),
					ValidatorAddress: del.GetValidatorAddr(),
					Shares:           del.GetShares(),
				},
				Balance: lockedCoins,
			},
		)

		res.TotalDelegatedCoins = res.TotalDelegatedCoins.Add(lockedCoins)
		res.TotalEquivalentStakedAmount = res.TotalEquivalentStakedAmount.Add(lockedCoins)

		return false
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (q Querier) UnpoolWhitelist(goCtx context.Context, req *types.QueryUnpoolWhitelistRequest) (*types.QueryUnpoolWhitelistResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	allowedPools := q.GetUnpoolAllowedPools(sdk.UnwrapSDKContext(goCtx))

	return &types.QueryUnpoolWhitelistResponse{
		PoolIds: allowedPools,
	}, nil
}

func (q Querier) filterConcentratedPositionLocks(ctx sdk.Context, positions []model.Position, isUnbonding bool) ([]types.ConcentratedPoolUserPositionRecord, error) {
	// Query each position ID and determine if it has a lock ID associated with it.
	// Construct a response with the position ID, lock ID, the amount of cl shares staked, and what those shares are worth in staked osmo tokens.
	var clPoolUserPositionRecords []types.ConcentratedPoolUserPositionRecord
	for _, pos := range positions {
		lockId, err := q.Keeper.clk.GetLockIdFromPositionId(ctx, pos.PositionId)
		if errors.Is(err, cltypes.PositionIdToLockNotFoundError{PositionId: pos.PositionId}) {
			continue
		} else if err != nil {
			return nil, err
		}

		// If we have hit this logic branch, it means that, at one point, the lockId provided existed. If we fetch it again
		// and it doesn't exist, that means that the lock has matured.
		lock, err := q.Keeper.lk.GetLockByID(ctx, lockId)
		if errors.Is(err, errorsmod.Wrap(lockuptypes.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", lock.GetID()))) {
			continue
		} else if err != nil {
			return nil, err
		}

		syntheticLock, _, err := q.Keeper.lk.GetSyntheticLockupByUnderlyingLockId(ctx, lockId)
		if err != nil {
			return nil, err
		}

		// Its possible for a non superfluid lock to be attached to a position. This can happen for users migrating non superfluid positions that
		// they intend to let mature so they can eventually set non full range positions.
		if syntheticLock.UnderlyingLockId == 0 {
			continue
		}

		if isUnbonding {
			// We only want to return unbonding positions.
			if !strings.Contains(syntheticLock.SynthDenom, "/superunbonding") {
				continue
			}
		} else {
			// We only want to return bonding positions.
			if !strings.Contains(syntheticLock.SynthDenom, "/superbonding") {
				continue
			}
		}

		valAddr, err := ValidatorAddressFromSyntheticDenom(syntheticLock.SynthDenom)
		if err != nil {
			return nil, err
		}

		baseDenom := lock.Coins.GetDenomByIndex(0)
		lockedCoins := sdk.NewCoin(baseDenom, lock.GetCoins().AmountOf(baseDenom))
		equivalentAmount, err := q.Keeper.GetSuperfluidOSMOTokens(ctx, baseDenom, lockedCoins.Amount)
		if err != nil {
			return nil, err
		}
		coin := sdk.NewCoin(appparams.BaseCoinUnit, equivalentAmount)

		clPoolUserPositionRecords = append(clPoolUserPositionRecords, types.ConcentratedPoolUserPositionRecord{
			ValidatorAddress:       valAddr,
			PositionId:             pos.PositionId,
			LockId:                 lockId,
			SyntheticLock:          syntheticLock,
			DelegationAmount:       lockedCoins,
			EquivalentStakedAmount: &coin,
		})
	}
	return clPoolUserPositionRecords, nil
}

// TEMPORARY CODE
func (q Querier) RestSupply(goCtx context.Context, req *types.QueryRestSupplyRequest) (*types.QueryRestSupplyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	supply := q.bk.GetSupply(sdk.UnwrapSDKContext(goCtx), req.Denom)
	return &types.QueryRestSupplyResponse{Amount: supply}, nil
}
