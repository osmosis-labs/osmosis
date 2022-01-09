package keeper

import (
	"context"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"github.com/osmosis-labs/osmosis/x/osmolbp"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

var _ api.MsgServer = Keeper{}

func (k Keeper) CreateLBP(goCtx context.Context, msg *api.MsgCreateLBP) (*api.MsgCreateLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	id, err := k.createLBP(msg, ctx.BlockTime(), store)
	if err != nil {
		return nil, err
	}
	// TODO: Add a fee?

	err = ctx.EventManager().EmitTypedEvent(&api.EventCreateLBP{
		Id:       id,
		Creator:  msg.Creator,
		TokenIn:  msg.TokenIn,
		TokenOut: msg.TokenOut,
	})
	return &api.MsgCreateLBPResponse{PoolId: id}, err
}

func (k Keeper) createLBP(msg *api.MsgCreateLBP, now time.Time, store storetypes.KVStore) (uint64, error) {
	if err := msg.Validate(now); err != nil {
		return 0, err
	}
	id, idBz := k.nextPoolID(store)
	end := msg.StartTime.Add(msg.Duration)
	p := api.LBP{
		Treasury:  msg.Treasury,
		Id:        id,
		TokenOut:  msg.TokenOut,
		TokenIn:   msg.TokenIn,
		StartTime: msg.StartTime,
		EndTime:   end,

		Rate:           msg.InitialDeposit.Amount.Quo(sdk.NewInt(int64(msg.Duration / api.ROUND))),
		AccumulatorOut: sdk.ZeroInt(),

		OutRemaining: msg.InitialDeposit.Amount,
		OutSold:      sdk.ZeroInt(),
		OutPerShare:  sdk.ZeroInt(),

		Staked: sdk.ZeroInt(),
		Income: sdk.ZeroInt(),

		Shares:   sdk.ZeroInt(),
		Round:    0,
		EndRound: currentRound(msg.StartTime, end, end),
	}
	k.saveLBP(store, idBz, &p)
	// TODO:
	// + send initial deposit from sender to the pool
	// + use ADR-28 addresses?
	return id, nil
}

func (k Keeper) Subscribe(goCtx context.Context, msg *api.MsgSubscribe) (*emptypb.Empty, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.subscribe(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&api.EventSubscribe{
		Sender: msg.Sender,
		PoolId: msg.PoolId,
		Amount: msg.Amount.String(),
	})
	return &emptypb.Empty{}, err
}

func (k Keeper) subscribe(ctx sdk.Context, msg *api.MsgSubscribe, store storetypes.KVStore) error {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	if !msg.Amount.IsPositive() {
		return errors.ErrInvalidRequest.Wrap("amount of tokens must be positive")
	}
	p, poolIdBz, err := k.getLBP(store, msg.PoolId)
	if err != nil {
		return err
	}
	coin := sdk.NewCoin(p.TokenIn, msg.Amount)
	err = k.bank.SendCoinsFromAccountToModule(ctx, sender, osmolbp.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return errors.Wrap(err, "user doesn't have enough tokens to subscribe for a LBP")
	}

	u, err := k.getUserPosition(store, poolIdBz, sender, true)
	if err != nil {
		return err
	}

	subscribe(&p, &u, msg.Amount, ctx.BlockTime())

	k.saveLBP(store, poolIdBz, &p)
	k.saveUserPosition(store, poolIdBz, sender, &u)
	// TODO: event
	return nil
}

func (k Keeper) Withdraw(goCtx context.Context, msg *api.MsgWithdraw) (*emptypb.Empty, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.withdraw(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&api.EventWithdraw{
		Sender: msg.Sender,
		PoolId: msg.PoolId,
		Amount: msg.Amount.String(),
	})
	return &emptypb.Empty{}, err
}

func (k Keeper) withdraw(ctx sdk.Context, msg *api.MsgWithdraw, store storetypes.KVStore) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	p, poolIdBz, err := k.getLBP(store, msg.PoolId)
	if err != nil {
		return err
	}
	u, err := k.getUserPosition(store, poolIdBz, sender, false)
	if err != nil {
		return err
	}
	// withdraw updates msg.Amount
	err = withdraw(&p, &u, msg.Amount, ctx.BlockTime())
	if err != nil {
		return err
	}
	coin := sdk.NewCoin(p.TokenIn, *msg.Amount)
	err = k.bank.SendCoinsFromModuleToAccount(ctx, osmolbp.ModuleName, sender, sdk.NewCoins(coin))
	if err != nil {
		return err
	}

	k.saveLBP(store, poolIdBz, &p)
	k.saveUserPosition(store, poolIdBz, sender, &u)
	return nil
}

func (k Keeper) ExitLBP(goCtx context.Context, msg *api.MsgExitLBP) (*api.MsgExitLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: finish
	// store := ctx.KVStore(k.storeKey)
	// if err := k.withdraw(ctx, msg, store); err != nil {
	// 	return nil, err
	// }
	err := ctx.EventManager().EmitTypedEvent(&api.EventWithdraw{
		Sender: msg.Sender,
		PoolId: msg.PoolId,
		// TODO Amount: msg.Amount.String(),
	})
	// TODO: fill response
	return &api.MsgExitLBPResponse{}, err
}
