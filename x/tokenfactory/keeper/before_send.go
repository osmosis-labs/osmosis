package keeper

import (
	"context"
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
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
	} else {
		// if a contract is being set, call the contract using cache context
		// to test if the contract is an existing, valid contract.
		cacheCtx, _ := ctx.CacheContext()

		cwAddr, err := sdk.AccAddressFromBech32(cosmwasmAddress)
		if err != nil {
			return err
		}

		tempMsg := types.TrackBeforeSendSudoMsg{
			TrackBeforeSend: types.TrackBeforeSendMsg{},
		}
		msgBz, err := json.Marshal(tempMsg)
		if err != nil {
			return err
		}
		_, err = k.contractKeeper.Sudo(cacheCtx, cwAddr, msgBz)

		if err != nil && strings.Contains(err.Error(), "no such contract") {
			return err
		}
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

func (k Keeper) GetAllBeforeSendHooks(ctx sdk.Context) ([]string, []string) {
	denomsList := []string{}
	beforeSendHooksList := []string{}

	iterator := k.GetAllDenomsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Value())

		beforeSendHook := k.GetBeforeSendHook(ctx, denom)
		if beforeSendHook != "" {
			denomsList = append(denomsList, denom)
			beforeSendHooksList = append(beforeSendHooksList, beforeSendHook)
		}
	}
	return denomsList, beforeSendHooksList
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

// TrackBeforeSend calls the before send listener contract suppresses any errors
func (h Hooks) TrackBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) {
	_ = h.k.callBeforeSendListener(ctx, from, to, amount, false)
}

// TrackBeforeSend calls the before send listener contract returns any errors
func (h Hooks) BlockBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	return h.k.callBeforeSendListener(ctx, from, to, amount, true)
}

// callBeforeSendListener iterates over each coin and sends corresponding sudo msg to the contract address stored in state.
// If blockBeforeSend is true, sudoMsg wraps BlockBeforeSendMsg, otherwise sudoMsg wraps TrackBeforeSendMsg.
// Note that we gas meter trackBeforeSend to prevent infinite contract calls.
// CONTRACT: this should not be called in beginBlock or endBlock since out of gas will cause this method to panic.
func (k Keeper) callBeforeSendListener(context context.Context, from, to sdk.AccAddress, amount sdk.Coins, blockBeforeSend bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errorsmod.Wrapf(types.ErrBeforeSendHookOutOfGas, "%v", r)
		}
	}()

	ctx := sdk.UnwrapSDKContext(context)
	for _, coin := range amount {
		cosmwasmAddress := k.GetBeforeSendHook(ctx, coin.Denom)
		if cosmwasmAddress != "" {
			cwAddr, err := sdk.AccAddressFromBech32(cosmwasmAddress)
			if err != nil {
				return err
			}

			var msgBz []byte

			// get msgBz, either BlockBeforeSend or TrackBeforeSend
			// Note that for trackBeforeSend, we need to gas meter computations to prevent infinite loop
			// specifically because module to module sends are not gas metered.
			// We don't need to do this for blockBeforeSend since blockBeforeSend is not called during module to module sends.
			if blockBeforeSend {
				msg := types.BlockBeforeSendSudoMsg{
					BlockBeforeSend: types.BlockBeforeSendMsg{
						From:   from.String(),
						To:     to.String(),
						Amount: osmoutils.CWCoinFromSDKCoin(coin),
					},
				}
				msgBz, err = json.Marshal(msg)
			} else {
				msg := types.TrackBeforeSendSudoMsg{
					TrackBeforeSend: types.TrackBeforeSendMsg{
						From:   from.String(),
						To:     to.String(),
						Amount: osmoutils.CWCoinFromSDKCoin(coin),
					},
				}
				msgBz, err = json.Marshal(msg)
			}
			if err != nil {
				return err
			}
			em := sdk.NewEventManager()

			childCtx := ctx.WithGasMeter(storetypes.NewGasMeter(types.BeforeSendHookGasLimit))
			_, err = k.contractKeeper.Sudo(childCtx.WithEventManager(em), cwAddr, msgBz)
			if err != nil {
				if strings.Contains(err.Error(), "no such contract") {
					return nil
				}
				if k.IsModuleAcc(ctx, from) {
					return nil
				}

				return errorsmod.Wrapf(err, "failed to call before send hook for denom %s", coin.Denom)
			}

			// consume gas used for calling contract to the parent ctx
			ctx.GasMeter().ConsumeGas(childCtx.GasMeter().GasConsumed(), "track before send gas")
		}
	}
	return nil
}
