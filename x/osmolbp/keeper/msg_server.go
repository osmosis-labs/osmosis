package keeper

import (
	"context"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

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
	err = ctx.EventManager().EmitTypedEvent(&api.EventCreateLBP{
		Id:       id,
		Creator:  msg.Creator,
		TokenIn:  msg.TokenIn,
		TokenOut: msg.TokenOut,
	})
	return &api.MsgCreateLBPResponse{PoolId: id}, err
}

func (k Keeper) createLBP(msg *api.MsgCreateLBP, now time.Time, store storetypes.KVStore) (uint64, error) {
	if err := msg.Validate(now); err != nil { // handle.ValidateMsgCreateLBP(msg)
		return 0, err
	}
	id, idBz := k.nextPoolID(store)
	end := msg.StartTime.Add(msg.Duration)
	p := newLBP(msg.Treasury, id, msg.TokenIn, msg.TokenOut, msg.StartTime, end, msg.InitialDeposit.Amount)
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
	if !msg.Amount.IsPositive() {
		return errors.ErrInvalidRequest.Wrap("amount of tokens must be positive")
	}
	sender, p, poolIdBz, u, err := k.getUserAndLBP(store, msg.PoolId, msg.Sender, false)
	if err != nil {
		return err
	}

	coin := sdk.NewCoin(p.TokenIn, msg.Amount)
	err = k.bank.SendCoinsFromAccountToModule(ctx, sender, api.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return errors.Wrap(err, "user doesn't have enough tokens to subscribe for a LBP")
	}
	subscribe(p, u, msg.Amount, ctx.BlockTime())

	k.saveLBP(store, poolIdBz, p)
	k.saveUserPosition(store, poolIdBz, sender, u)
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
	sender, p, poolIdBz, u, err := k.getUserAndLBP(store, msg.PoolId, msg.Sender, false)
	if err != nil {
		return err
	}
	// withdraw updates msg.Amount
	err = withdraw(p, u, msg.Amount, ctx.BlockTime())
	if err != nil {
		return err
	}
	coin := sdk.NewCoin(p.TokenIn, *msg.Amount)
	err = k.bank.SendCoinsFromModuleToAccount(ctx, api.ModuleName, sender, sdk.NewCoins(coin))
	if err != nil {
		return err
	}

	k.saveLBP(store, poolIdBz, p)
	k.saveUserPosition(store, poolIdBz, sender, u)
	return nil
}

func (k Keeper) ExitLBP(goCtx context.Context, msg *api.MsgExitLBP) (*api.MsgExitLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	purchased, err := k.exitLBP(ctx, msg, store)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&api.EventExit{
		Sender:    msg.Sender,
		PoolId:    msg.PoolId,
		Purchased: purchased.String(),
	})
	return &api.MsgExitLBPResponse{Purchased: purchased}, err
}

// returns amount of tokens purchased
func (k Keeper) exitLBP(ctx sdk.Context, msg *api.MsgExitLBP, store storetypes.KVStore) (sdk.Int, error) {
	sender, p, poolIdBz, u, err := k.getUserAndLBP(store, msg.PoolId, msg.Sender, false)
	if err != nil {
		return sdk.Int{}, err
	}
	if err := msg.Validate(ctx.BlockTime(), p.EndTime); err != nil {
		return sdk.Int{}, err
	}

	pingLBP(p, ctx.BlockTime())
	triggerUserPurchase(p, u)
	// we don't need to update u.Spent, because we delete user record

	coin := sdk.NewCoin(p.TokenOut, u.Purchased)
	err = k.bank.SendCoinsFromModuleToAccount(ctx, api.ModuleName, sender, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Int{}, err
	}
	// TODO: make double check with p.OutSold?

	if u.Shares.IsPositive() || u.Staked.IsPositive() {
		ctx.Logger().Error("user has outstanding token_in balance", "user", msg.Sender, "balance", u.Staked)
		coin = sdk.NewCoin(p.TokenIn, u.Staked)
		err = k.bank.SendCoinsFromModuleToAccount(ctx, api.ModuleName, sender, sdk.NewCoins(coin))
		if err != nil {
			return sdk.Int{}, err
		}
	}

	k.delUserPosition(store, poolIdBz, sender)
	return u.Purchased, nil
}

func (k Keeper) FinalizeLBP(goCtx context.Context, msg *api.MsgFinalizeLBP) (*api.MsgFinalizeLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	income, err := k.finalizeLBP(ctx, msg, store)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&api.EventFinalizeLBP{
		PoolId: msg.PoolId,
		Income: income.String(),
	})
	return &api.MsgFinalizeLBPResponse{Income: income}, err
}

// returns LBP income
func (k Keeper) finalizeLBP(ctx sdk.Context, msg *api.MsgFinalizeLBP, store storetypes.KVStore) (sdk.Int, error) {
	p, poolIdBz, err := k.getLBP(store, msg.PoolId)
	if err != nil {
		return sdk.Int{}, err
	}
	if err := msg.Validate(ctx.BlockTime(), p.EndTime); err != nil {
		return sdk.Int{}, err
	}
	if p.Income.IsZero() {
		return sdk.Int{}, errors.ErrInvalidRequest.Wrap("LBP already finalized")
	}

	treasury, err := sdk.AccAddressFromBech32(p.Treasury)
	if err != nil {
		return sdk.Int{}, err
	}

	pingLBP(&p, ctx.BlockTime())
	coin := sdk.NewCoin(p.TokenOut, p.Income)
	err = k.bank.SendCoinsFromModuleToAccount(ctx, api.ModuleName, treasury, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Int{}, err
	}
	income := p.Income
	p.Income = sdk.ZeroInt()
	k.saveLBP(store, poolIdBz, &p)
	return income, nil
}
