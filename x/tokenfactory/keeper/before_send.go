package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/tokenfactory/types"

	errorsmod "cosmossdk.io/errors"
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
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
	k          Keeper
	wasmkeeper wasmKeeper.Keeper
}

var _ types.BankHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks(wasmkeeper wasmKeeper.Keeper) Hooks {
	return Hooks{k, wasmkeeper}
}

// TrackBeforeSend calls the before send listener contract surpresses any errors
func (h Hooks) TrackBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) {
	_ = h.k.callBeforeSendListener(ctx, h.wasmkeeper, from, to, amount, false)
}

// TrackBeforeSend calls the before send listener contract returns any errors
func (h Hooks) BlockBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	return h.k.callBeforeSendListener(ctx, h.wasmkeeper, from, to, amount, true)
}

// callBeforeSendListener iterates over each coin and sends corresponding sudo msg to the contract address stored in state.
// If blockBeforeSend is true, sudoMsg wraps BlockBeforeSendMsg, otherwise sudoMsg wraps TrackBeforeSendMsg.
func (k Keeper) callBeforeSendListener(ctx sdk.Context, wasmKeeper wasmKeeper.Keeper, from, to sdk.AccAddress, amount sdk.Coins, blockBeforeSend bool) error {
	for _, coin := range amount {
		cosmwasmAddress := k.GetBeforeSendHook(ctx, coin.Denom)
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
				return err
			}

			em := sdk.NewEventManager()

			_, err = wasmKeeper.Sudo(ctx.WithEventManager(em), cwAddr, msgBz)
			if err != nil {
				return errorsmod.Wrapf(err, "failed to call before send hook for denom %s", coin.Denom)
			}
		}
	}
	return nil
}
