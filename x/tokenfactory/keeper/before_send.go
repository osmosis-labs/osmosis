package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/tokenfactory/types"

	errorsmod "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

func (k Keeper) setBeforeSendHook(ctx sdk.Context, denom string, cosmwasmAddress string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(denom)
	if err != nil {
		return err
	}

	store := k.GetDenomPrefixStore(ctx, denom)

	// delete the store for denom prefix store when cosmwasm address is nil
	if cosmwasmAddress == "" {
		store.Delete([]byte(types.BeforeSendHookAddressPrefixKey))
		return nil
	}

	_, err = sdk.AccAddressFromBech32(cosmwasmAddress)
	if err != nil {
		return err
	}

	store.Set([]byte(types.BeforeSendHookAddressPrefixKey), []byte(cosmwasmAddress))

	return nil
}

func (k Keeper) GetBeforeSendHook(ctx sdk.Context, denom string) string {
	store := k.GetDenomPrefixStore(ctx, denom)

	bz := store.Get([]byte(types.BeforeSendHookAddressPrefixKey))
	if bz == nil {
		return ""
	}

	return string(bz)
}

func CWCoinsFromSDKCoins(in sdk.Coins) wasmvmtypes.Coins {
	var cwCoins wasmvmtypes.Coins
	for _, coin := range in {
		cwCoins = append(cwCoins, CWCoinFromSDKCoin(coin))
	}
	return cwCoins
}

func CWCoinFromSDKCoin(in sdk.Coin) wasmvmtypes.Coin {
	return wasmvmtypes.Coin{
		Denom:  in.GetDenom(),
		Amount: in.Amount.String(),
	}
}

// Hooks wrapper struct for bank keeper
type Hooks struct {
	k Keeper
}

var _ types.BankHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// TrackBeforeSend calls the before send listener contract surpresses any errors
func (h Hooks) TrackBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) {
	_ = h.k.callBeforeSendListener(ctx, from, to, amount, false)
}

// TrackBeforeSend calls the before send listener contract returns any errors
func (h Hooks) BlockBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	return h.k.callBeforeSendListener(ctx, from, to, amount, true)
}

// callBeforeSendListener iterates over each coin and sends corresponding sudo msg to the contract address stored in state.
// If blockBeforeSend is true, sudoMsg wraps BlockBeforeSendMsg, otherwise sudoMsg wraps TrackBeforeSendMsg.
func (k Keeper) callBeforeSendListener(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins, blockBeforeSend bool) error {
	ctx.Logger().Error("CALL BEFORE SEND LISTNER IS BEING CALLED")
	ctx.Logger().Error(fmt.Sprint("From, to, amount: ", from.String(), to.String(), amount.String()))
	for _, coin := range amount {
		cosmwasmAddress := k.GetBeforeSendHook(ctx, coin.Denom)
		ctx.Logger().Error(fmt.Sprint("denom: ", coin.Denom))
		ctx.Logger().Error(fmt.Sprint("COSMWASM ADDRESS FOR BEFORE SEND HOOK IS: ", cosmwasmAddress))
		if cosmwasmAddress != "" {
			cwAddr, err := sdk.AccAddressFromBech32(cosmwasmAddress)
			if err != nil {
				return err
			}

			var msgBz []byte

			// get msgBz, either BlockBeforeSend or TrackBeforeSend
			if blockBeforeSend {
				msg := types.BlockBeforeSendSudoMsg{
					BlockBeforeSend: types.BlockBeforeSendMsg{
						From:   from.String(),
						To:     to.String(),
						Amount: CWCoinFromSDKCoin(coin),
					},
				}
				msgBz, err = json.Marshal(msg)
			} else {
				msg := types.TrackBeforeSendSudoMsg{
					TrackBeforeSend: types.TrackBeforeSendMsg{
						From:   from.String(),
						To:     to.String(),
						Amount: CWCoinFromSDKCoin(coin),
					},
				}
				msgBz, err = json.Marshal(msg)
			}
			if err != nil {
				ctx.Logger().Error(fmt.Sprint("FAILED TO CALL SEND SEND HOOK WITH: ", err.Error()))
				return err
			}

			_, err = k.contractKeeper.Sudo(ctx, cwAddr, msgBz)
			if err != nil {
				ctx.Logger().Error(fmt.Sprint("FAILED TO CALL SEND SEND HOOK SUDO WITH: ", err.Error()))
				return errorsmod.Wrapf(err, "failed to call before send hook for denom %s", coin.Denom)
			}
		}
	}
	return nil
}
