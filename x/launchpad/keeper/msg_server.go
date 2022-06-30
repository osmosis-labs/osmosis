package keeper

import (
	"context"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad"
	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

func (k Keeper) CreateSale(goCtx context.Context, msg *types.MsgCreateSale) (*types.MsgCreateSaleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	params := k.GetParams(ctx)
	id, creator, err := k.createSale(msg, ctx.BlockTime(), params, store)
	if err != nil {
		return nil, err
	}
	if params.SaleCreationFeeRecipient != "" && !params.SaleCreationFee.Empty() {
		r, err := sdk.AccAddressFromBech32(params.SaleCreationFeeRecipient)
		if err != nil {
			return nil, err
		}
		k.bank.SendCoins(ctx, creator, r, params.SaleCreationFee)
		ctx.Logger().Info("Sale Creation Fee charged",
			"recipient", params.SaleCreationFeeRecipient, "fee", params.SaleCreationFee)
	} else {
		ctx.Logger().Info("Sale Creation Fee not charged. Params creation fee recipient or fee is not defined")
	}
	err = k.bank.SendCoinsFromAccountToModule(ctx, creator, launchpad.ModuleName, sdk.NewCoins(*msg.TokenOut))
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&types.EventCreateSale{
		Id:       id,
		Creator:  msg.Creator,
		TokenIn:  msg.TokenIn,
		TokenOut: msg.TokenOut,
	})
	return &types.MsgCreateSaleResponse{SaleId: id}, err
}

func (k Keeper) createSale(msg *types.MsgCreateSale, now time.Time, params types.Params, store storetypes.KVStore) (uint64, sdk.AccAddress, error) {
	creator, err := msg.Validate(now, params.MinimumSaleDuration, params.MinimumDurationUntilStartTime)
	if err != nil {
		return 0, nil, err
	}

	id, idBz := k.nextSaleID(store)
	end := msg.StartTime.Add(msg.Duration)
	treasury := msg.Recipient
	if treasury == "" {
		treasury = msg.Creator
	}
	p := newSale(treasury, id, msg.TokenIn, *msg.TokenOut, msg.StartTime, end)
	k.saveSale(store, idBz, &p)
	return id, creator, nil
}

func (k Keeper) Subscribe(goCtx context.Context, msg *types.MsgSubscribe) (*emptypb.Empty, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.subscribe(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&types.EventSubscribe{
		Sender: msg.Sender,
		SaleId: msg.SaleId,
		Amount: msg.Amount.String(),
	})
	return &emptypb.Empty{}, err
}

func (k Keeper) subscribe(ctx sdk.Context, msg *types.MsgSubscribe, store storetypes.KVStore) error {
	if !msg.Amount.IsPositive() {
		return errors.ErrInvalidRequest.Wrap("amount of tokens must be positive")
	}
	sender, p, saleIdBz, u, err := k.getUserAndSale(store, msg.SaleId, msg.Sender, true)
	if err != nil {
		return err
	}

	coin := sdk.NewCoin(p.TokenIn, msg.Amount)
	err = k.bank.SendCoinsFromAccountToModule(ctx, sender, launchpad.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return errors.Wrap(err, "user doesn't have enough tokens to subscribe for a Sale")
	}
	subscribe(p, u, msg.Amount, ctx.BlockTime())

	k.saveSale(store, saleIdBz, p)
	k.saveUserPosition(store, saleIdBz, sender, u)
	// TODO: event
	return nil
}

func (k Keeper) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*emptypb.Empty, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.withdraw(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&types.EventWithdraw{
		Sender: msg.Sender,
		SaleId: msg.SaleId,
		Amount: msg.Amount.String(),
	})
	return &emptypb.Empty{}, err
}

// it will update msg.Amount to the withdrawn amount (this changes only when msg.Amount == nil)
func (k Keeper) withdraw(ctx sdk.Context, msg *types.MsgWithdraw, store storetypes.KVStore) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	sender, p, saleIdBz, u, err := k.getUserAndSale(store, msg.SaleId, msg.Sender, false)
	if err != nil {
		return err
	}
	// withdraw updates msg.Amount
	amount, err := withdraw(p, u, msg.Amount, ctx.BlockTime())
	if err != nil {
		return err
	}
	msg.Amount = &amount
	coin := sdk.NewCoin(p.TokenIn, *msg.Amount)
	err = k.bank.SendCoinsFromModuleToAccount(ctx, launchpad.ModuleName, sender, sdk.NewCoins(coin))
	if err != nil {
		return err
	}

	k.saveSale(store, saleIdBz, p)
	k.saveUserPosition(store, saleIdBz, sender, u)
	return nil
}

func (k Keeper) ExitSale(goCtx context.Context, msg *types.MsgExitSale) (*types.MsgExitSaleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	purchased, err := k.exitSale(ctx, msg, store)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&types.EventExit{
		Sender:    msg.Sender,
		SaleId:    msg.SaleId,
		Purchased: purchased.String(),
	})
	return &types.MsgExitSaleResponse{Purchased: purchased}, err
}

// returns amount of tokens purchased
func (k Keeper) exitSale(ctx sdk.Context, msg *types.MsgExitSale, store storetypes.KVStore) (sdk.Int, error) {
	sender, p, saleIdBz, u, err := k.getUserAndSale(store, msg.SaleId, msg.Sender, false)
	if err != nil {
		return sdk.Int{}, err
	}
	if err := msg.Validate(ctx.BlockTime(), p.EndTime); err != nil {
		return sdk.Int{}, err
	}

	pingSale(p, ctx.BlockTime())
	triggerUserPurchase(p, u)
	// we don't need to update u.Spent, because we delete user record

	coin := sdk.NewCoin(p.TokenOut, u.Purchased)
	err = k.bank.SendCoinsFromModuleToAccount(ctx, launchpad.ModuleName, sender, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Int{}, err
	}
	// TODO: make double check with p.OutSold?

	if u.Shares.IsPositive() || u.Staked.IsPositive() {
		ctx.Logger().Error("user has outstanding token_in balance", "user", msg.Sender, "balance", u.Staked)
		coin = sdk.NewCoin(p.TokenIn, u.Staked)
		err = k.bank.SendCoinsFromModuleToAccount(ctx, launchpad.ModuleName, sender, sdk.NewCoins(coin))
		if err != nil {
			return sdk.Int{}, err
		}
	}

	k.delUserPosition(store, saleIdBz, sender)
	return u.Purchased, nil
}

func (k Keeper) FinalizeSale(goCtx context.Context, msg *types.MsgFinalizeSale) (*types.MsgFinalizeSaleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	income, err := k.finalizeSale(ctx, msg, store)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&types.EventFinalizeSale{
		SaleId: msg.SaleId,
		Income: income.String(),
	})
	return &types.MsgFinalizeSaleResponse{Income: income}, err
}

// returns Sale income
func (k Keeper) finalizeSale(ctx sdk.Context, msg *types.MsgFinalizeSale, store storetypes.KVStore) (sdk.Int, error) {
	p, saleIdBz, err := k.getSale(store, msg.SaleId)
	if err != nil {
		return sdk.Int{}, err
	}
	if err := msg.Validate(ctx.BlockTime(), p.EndTime); err != nil {
		return sdk.Int{}, err
	}
	if p.Income.IsZero() {
		return sdk.Int{}, errors.ErrInvalidRequest.Wrap("Sale already finalized")
	}
	treasury, err := sdk.AccAddressFromBech32(p.Treasury)
	if err != nil {
		return sdk.Int{}, err
	}

	pingSale(&p, ctx.BlockTime())
	coin := sdk.NewCoin(p.TokenOut, p.Income)
	err = k.bank.SendCoinsFromModuleToAccount(ctx, launchpad.ModuleName, treasury, sdk.NewCoins(coin))
	if err != nil {
		return sdk.Int{}, err
	}
	income := p.Income
	p.Income = sdk.ZeroInt()
	k.saveSale(store, saleIdBz, &p)
	return income, nil
}
