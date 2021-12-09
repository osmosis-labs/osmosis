package keeper

import (
	"context"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/osmolbp"
	"github.com/osmosis-labs/osmosis/x/osmolbp/proto"
)

var _ proto.MsgServer = Keeper{}

func (k Keeper) CreateLBP(goCtx context.Context, msg *proto.MsgCreateLBP) (*proto.MsgCreateLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	id, err := k.createLBP(msg, ctx.BlockTime(), store)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&proto.EventCreateLBP{
		Id:       id,
		Creator:  msg.Creator,
		TokenIn:  msg.TokenIn,
		TokenOut: msg.TokenOut,
	})
	return &proto.MsgCreateLBPResponse{PoolId: id}, err
}

func (k Keeper) createLBP(msg *proto.MsgCreateLBP, now time.Time, store storetypes.KVStore) (uint64, error) {
	if err := msg.Validate(now); err != nil {
		return 0, err
	}
	id, idBz := k.nextPoolID(store)
	p := proto.LBP{
		TokenOut:       msg.TokenOut,
		TokenIn:        msg.TokenIn,
		Start:          msg.Start,
		End:            msg.Start.Add(msg.Duration),
		Rate:           msg.TotalSale.Quo(sdk.NewInt(int64(msg.Duration / proto.ROUND))),
		AccumulatorOut: sdk.ZeroInt(),
		Round:          0,
		Staked:         sdk.ZeroInt(),
	}
	k.savePool(store, idBz, &p)
	return id, nil

}

func (k Keeper) Stake(goCtx context.Context, msg *proto.MsgStake) (*proto.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.stake(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&proto.EventStake{
		Sender: msg.Sender,
		PoolId: msg.PoolId,
		Amount: msg.Amount.String(),
	})
	return &proto.EmptyResponse{}, err
}

func (k Keeper) stake(ctx sdk.Context, msg *proto.MsgStake, store storetypes.KVStore) error {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	p, poolIdBz, err := k.getPool(store, msg.PoolId)
	if err != nil {
		return err
	}
	coins := []sdk.Coin{{Denom: p.TokenIn, Amount: msg.Amount}}
	err = k.bank.SendCoinsFromAccountToModule(ctx, sender, osmolbp.ModuleName, coins)
	if err != nil {
		return errors.Wrap(err, "user doesn't have enough tokens to stake")
	}

	v, found, err := k.getUserVault(store, poolIdBz, sender)
	if err != nil {
		return err
	}
	if !found {
		v.Accumulator = p.AccumulatorOut
		v.Staked = sdk.ZeroInt()
	}

	stakeInPool(&p, &v, msg.Amount, ctx.BlockTime())

	k.savePool(store, poolIdBz, &p)
	k.saveUserVault(store, poolIdBz, sender, &v)
	return nil
}

func (k Keeper) ExitLBP(goCtx context.Context, msg *proto.MsgExitLBP) (*proto.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.exitLBP(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&proto.EventExit{
		Sender: msg.Sender,
		PoolId: msg.PoolId,
		// TODO: Purchased: ,
	})
	return &proto.EmptyResponse{}, err
}

func (k Keeper) exitLBP(ctx sdk.Context, msg *proto.MsgExitLBP, store storetypes.KVStore) error {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	p, poolIdBz, err := k.getPool(store, msg.PoolId)
	if err != nil {
		return err
	}
	v, found, err := k.getUserVault(store, poolIdBz, sender)
	if err != nil {
		return err
	}
	if !found {
		return errors.Wrap(errors.ErrKeyNotFound, "user doesn't have a stake")
	}

	// TODO: check if v.Staked makes sense, maybe we should first ping and evaulate
	if err = unstakeFromPool(&p, &v, v.Staked, ctx.BlockTime()); err != nil {
		return err
	}

	k.savePool(store, poolIdBz, &p)

	return nil
}
