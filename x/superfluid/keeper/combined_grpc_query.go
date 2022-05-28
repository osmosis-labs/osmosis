package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (q Querier) Delegation(goCtx context.Context, req *types.QueryDelegationRequest) (*types.QueryDelegationResponse, error) {
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

	res := types.QueryDelegationResponse{
		SuperfluidDelegationRecords: []types.SuperfluidDelegationRecord{},
		DelegationResponse:          []stakingtypes.DelegationResponse{},
		TotalDelegatedCoins:         sdk.NewCoins(),
	}

	syntheticLocks := q.Keeper.lk.GetAllSyntheticLockupsByAddr(ctx, delAddr)

	// this if for getting superfluid staking
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
		res.SuperfluidDelegationRecords = append(res.SuperfluidDelegationRecords,
			types.SuperfluidDelegationRecord{
				DelegatorAddress: req.DelegatorAddress,
				ValidatorAddress: valAddr,
				DelegationAmount: lockedCoins,
			},
		)
		res.TotalDelegatedCoins = res.TotalDelegatedCoins.Add(lockedCoins)
	}

	//this is for getting normal staking
	q.sk.IterateDelegations(ctx, delAddr, func(_ int64, del stakingtypes.DelegationI) bool {
		val, found := q.sk.GetValidator(ctx, del.GetValidatorAddr())
		if !found {
			return true
		}

		lockedCoins := sdk.NewCoin(q.sk.BondDenom(ctx), val.TokensFromShares(del.GetShares()).TruncateInt())

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

		return false
	})

	return &res, nil
}
