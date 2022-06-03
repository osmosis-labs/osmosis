package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appparams "github.com/osmosis-labs/osmosis/v7/app/params"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (q Querier) TotalDelegation(goCtx context.Context, req *types.QueryTotalDelegationRequest) (*types.QueryTotalDelegationResponse, error) {
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

	res := types.QueryTotalDelegationResponse{
		SuperfluidDelegationRecords: []types.SuperfluidDelegationRecord{},
		DelegationResponse:          []stakingtypes.DelegationResponse{},
		TotalDelegatedCoins:         sdk.NewCoins(),
		TotalEquivalentStakedAmount: sdk.NewCoin(appparams.BaseCoinUnit, sdk.ZeroInt()),
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

		// Find how many osmo tokens this delegation is worth at superfluids current risk adjustment
		// and twap of the denom.
		equivalentAmount := q.Keeper.GetSuperfluidOSMOTokens(ctx, baseDenom, lockedCoins.Amount)
		equivalentOsmoCoin := sdk.NewCoin(appparams.BaseCoinUnit, equivalentAmount)

		res.SuperfluidDelegationRecords = append(res.SuperfluidDelegationRecords,
			types.SuperfluidDelegationRecord{
				DelegatorAddress: req.DelegatorAddress,
				ValidatorAddress: valAddr,
				DelegationAmount: lockedCoins,
			},
		)
		res.TotalDelegatedCoins = res.TotalDelegatedCoins.Add(lockedCoins)
		res.TotalEquivalentStakedAmount = res.TotalEquivalentStakedAmount.Add(equivalentOsmoCoin)
	}

	//this is for getting normal staking
	q.sk.IterateDelegations(ctx, delAddr, func(_ int64, del stakingtypes.DelegationI) bool {
		val, found := q.sk.GetValidator(ctx, del.GetValidatorAddr())
		if !found {
			return true
		}

		lockedCoins := sdk.NewCoin(appparams.BaseCoinUnit, val.TokensFromShares(del.GetShares()).TruncateInt().Mul(sdk.NewInt(1000000)))

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
