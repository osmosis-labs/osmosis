package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/tokenfactory/types"

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

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k          Keeper
	wasmkeeper wasmKeeper.Keeper
}

var _ types.BankHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks(wasmkeeper wasmKeeper.Keeper) Hooks {
	return Hooks{k, wasmkeeper}
}

// Implements sdk.ValidatorHooks
func (h Hooks) BeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	cwCoins := CWCoinsFromSDKCoins(amount)

	for _, coin := range amount {
		cosmwasmAddress := h.k.GetBeforeSendHook(ctx, coin.Denom)
		if cosmwasmAddress != "" {
			cwAddr, err := sdk.AccAddressFromBech32(cosmwasmAddress)
			if err != nil {
				return err
			}

			msg := types.SudoMsg{
				BeforeSend: types.BeforeSendMsg{
					From:   from.String(),
					To:     to.String(),
					Amount: cwCoins,
				},
			}

			msgBz, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			em := sdk.NewEventManager()

			_, err = h.wasmkeeper.Sudo(ctx.WithEventManager(em), cwAddr, msgBz)
			fmt.Println(err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
