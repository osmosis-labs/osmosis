package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the market MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) Swap(goCtx context.Context, msg *types.MsgSwap) (*types.MsgSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, err
	}

	return k.handleSwapRequest(ctx, addr, addr, msg.OfferCoin, msg.AskDenom)
}

func (k msgServer) SwapSend(goCtx context.Context, msg *types.MsgSwapSend) (*types.MsgSwapSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	fromAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	res, err := k.handleSwapRequest(ctx, fromAddr, toAddr, msg.OfferCoin, msg.AskDenom)
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapSendResponse{
		SwapCoin: res.SwapCoin,
		SwapFee:  res.SwapFee,
	}, nil
}

// handleMsgSwap handles the logic of a MsgSwap
// This function does not repeat checks that have already been performed in msg.ValidateBasic()
// Ex) assert(offerCoin.Denom != askDenom)
func (k msgServer) handleSwapRequest(ctx sdk.Context,
	trader sdk.AccAddress, receiver sdk.AccAddress,
	offerCoin sdk.Coin, askDenom string,
) (*types.MsgSwapResponse, error) {
	// Compute exchange rates between the ask and offer
	swapDecCoin, spread, err := k.ComputeSwap(ctx, offerCoin, askDenom)
	if err != nil {
		return nil, err
	}

	// Charge a spread if applicable; the spread is burned
	var feeDecCoin sdk.DecCoin
	if spread.IsPositive() {
		feeDecCoin = sdk.NewDecCoinFromDec(swapDecCoin.Denom, spread.Mul(swapDecCoin.Amount))
	} else {
		feeDecCoin = sdk.NewDecCoin(swapDecCoin.Denom, sdk.ZeroInt())
	}

	// Subtract fee from the swap coin
	swapDecCoin.Amount = swapDecCoin.Amount.Sub(feeDecCoin.Amount)

	// Send offer coins to module account
	offerCoins := sdk.NewCoins(offerCoin)
	err = k.BankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.ModuleName, offerCoins)
	if err != nil {
		return nil, err
	}

	if offerCoin.Denom != appparams.BaseCoinUnit {
		// Burn offered coins and subtract from the trader's account
		err = k.BankKeeper.BurnCoins(ctx, types.ModuleName, offerCoins)
		if err != nil {
			return nil, err
		}
	}

	// Mint asked coins and credit Trader's account
	swapCoin, decimalCoin := swapDecCoin.TruncateDecimal()

	// Ensure to fail the swap tx when zero swap coin
	if !swapCoin.IsPositive() {
		return nil, types.ErrZeroSwapCoin
	}

	feeDecCoin = feeDecCoin.Add(decimalCoin) // add truncated decimalCoin to swapFee
	feeCoin, _ := feeDecCoin.TruncateDecimal()

	mintCoins := sdk.NewCoins(swapCoin.Add(feeCoin))

	// mint only stable coin
	if askDenom != appparams.BaseCoinUnit {
		err = k.BankKeeper.MintCoins(ctx, types.ModuleName, mintCoins)
		if err != nil {
			return nil, err
		}

		// Send swap coin to the trader
		swapCoins := sdk.NewCoins(swapCoin)
		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, swapCoins)
		if err != nil {
			return nil, err
		}
	} else {
		// native coin transfer using exchange vault
		calculatedAskCoin := swapCoin.Add(feeCoin)

		// check if vault have enough balance to make swap
		marketAcc := k.GetMarketAccount(ctx)
		if marketAcc == nil {
			panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
		}

		marketVaultBalance := k.GetExchangePoolBalance(ctx)
		var neededBalanceFromReserve = osmomath.ZeroInt()
		if marketVaultBalance.Amount.LT(calculatedAskCoin.Amount) {
			neededBalanceFromReserve = calculatedAskCoin.Amount.Sub(marketVaultBalance.Amount)
			reserveVaultBalance := k.GetExchangePoolBalance(ctx)
			if reserveVaultBalance.Amount.LT(neededBalanceFromReserve) {
				return nil, errorsmod.Wrapf(types.ErrNotEnoughBalanceOnMarketVaults, "Market vaults do not have enough coins to swap. Available amount: (main: %v), (reserve: %v), needed amount: %v",
					marketVaultBalance.Amount, reserveVaultBalance.Amount, calculatedAskCoin.Amount)
			}
		}

		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, sdk.NewCoins(swapCoin.SubAmount(neededBalanceFromReserve)))
		if err != nil {
			return nil, fmt.Errorf("could not send from main vault to recipient: %w", err)
		}
		if neededBalanceFromReserve.IsPositive() {
			err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ReserveModuleName, receiver, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, neededBalanceFromReserve)))
			if err != nil {
				return nil, fmt.Errorf("could not send from reserve vault to recipient: %w", err)
			}
		}
	}

	// Send swap fee to reserve
	if feeCoin.IsPositive() {
		feeCoins := sdk.NewCoins(feeCoin)
		if err := k.BankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.ReserveModuleName, feeCoins); err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventSwap,
			sdk.NewAttribute(types.AttributeKeyOffer, offerCoin.String()),
			sdk.NewAttribute(types.AttributeKeyTrader, trader.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, receiver.String()),
			sdk.NewAttribute(types.AttributeKeySwapCoin, swapCoin.String()),
			sdk.NewAttribute(types.AttributeKeySwapFee, feeCoin.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgSwapResponse{
		SwapCoin: swapCoin,
		SwapFee:  feeCoin,
	}, nil
}
