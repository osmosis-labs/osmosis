package keeper

import (
	"github.com/c-osmosis/osmosis/x/lockup/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier returns an instance of querier
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		case types.QueryModuleBalance:
			return queryModuleBalance(ctx, req, k, legacyQuerierCdc)

		case types.QueryModuleLockedAmount:
			return queryModuleLockedAmount(ctx, req, k, legacyQuerierCdc)

		case types.QueryAccountUnlockableCoins:
			return queryAccountUnlockableCoins(ctx, req, k, legacyQuerierCdc)

		case types.QueryAccountLockedCoins:
			return queryAccountLockedCoins(ctx, req, k, legacyQuerierCdc)

		case types.QueryAccountLockedPastTime:
			return queryAccountLockedPastTime(ctx, req, k, legacyQuerierCdc)

		case types.QueryAccountUnlockedBeforeTime:
			return queryAccountUnlockedBeforeTime(ctx, req, k, legacyQuerierCdc)

		case types.QueryAccountLockedPastTimeDenom:
			return queryAccountLockedPastTimeDenom(ctx, req, k, legacyQuerierCdc)

		case types.QueryLockedByID:
			return queryLockedByID(ctx, req, k, legacyQuerierCdc)

		case types.QueryAccountLockedLongerDuration:
			return queryAccountLockedLongerDuration(ctx, req, k, legacyQuerierCdc)

		case types.QueryAccountLockedLongerDurationDenom:
			return queryAccountLockedLongerDurationDenom(ctx, req, k, legacyQuerierCdc)

		default:
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}

		return res, err
	}
}

func queryModuleBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	coins := k.GetModuleBalance(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryModuleLockedAmount(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	coins := k.GetModuleLockedCoins(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountUnlockableCoins(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountUnlockableCoinsRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	coins := k.GetAccountUnlockableCoins(ctx, params.Owner)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountUnlockingCoins(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountUnlockableCoinsRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	coins := k.GetAccountUnlockingCoins(ctx, params.Owner)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountLockedCoins(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountLockedCoinsRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	coins := k.GetAccountLockedCoins(ctx, params.Owner)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountLockedPastTime(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountLockedPastTimeRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	locks := k.GetAccountLockedPastTime(ctx, params.Owner, params.Timestamp)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, locks)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountUnlockedBeforeTime(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountUnlockedBeforeTimeRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	unlocks := k.GetAccountUnlockedBeforeTime(ctx, params.Owner, params.Timestamp)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, unlocks)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountLockedPastTimeDenom(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountLockedPastTimeDenomRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	locks := k.GetAccountLockedPastTimeDenom(ctx, params.Owner, params.Denom, params.Timestamp)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, locks)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryLockedByID(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.LockedRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	lock, err := k.GetLockByID(ctx, params.LockId)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, lock)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountLockedLongerDuration(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountLockedLongerDurationRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	locks := k.GetAccountLockedLongerDuration(ctx, params.Owner, params.Duration)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, locks)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAccountLockedLongerDurationDenom(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.AccountLockedLongerDurationDenomRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	locks := k.GetAccountLockedLongerDuration(ctx, params.Owner, params.Duration)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, locks)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
