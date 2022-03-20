package wasm

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	wasmbindings "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func CustomMessageDecorator(gammKeeper *gammkeeper.Keeper, bank *bankkeeper.BaseKeeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &MintTokenMessenger{
			wrapped:    old,
			bank:       bank,
			gammKeeper: gammKeeper,
		}
	}
}

type MintTokenMessenger struct {
	wrapped    wasmkeeper.Messenger
	bank       *bankkeeper.BaseKeeper
	gammKeeper *gammkeeper.Keeper
}

var _ wasmkeeper.Messenger = (*MintTokenMessenger)(nil)

func (m *MintTokenMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		// only handle the happy path where this is really minting / swapping ...
		// leave everything else for the wrapped version
		var contractMsg wasmbindings.OsmosisMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, sdkerrors.Wrap(err, "osmosis msg")
		}
		if contractMsg.MintTokens != nil {
			return m.mintTokens(ctx, contractAddr, contractMsg.MintTokens)
		}
		if contractMsg.Swap != nil {
			return m.swapTokens(ctx, contractAddr, contractMsg.Swap)
		}
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func (m *MintTokenMessenger) mintTokens(ctx sdk.Context, contractAddr sdk.AccAddress, mint *wasmbindings.MintTokens) ([]sdk.Event, [][]byte, error) {
	rcpt, err := parseAddress(mint.Recipient)
	if err != nil {
		return nil, nil, err
	}

	denom, err := GetFullDenom(contractAddr.String(), mint.SubDenom)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "mint token denom")
	}
	coins := []sdk.Coin{sdk.NewCoin(denom, mint.Amount)}

	err = m.bank.MintCoins(ctx, gammtypes.ModuleName, coins)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "minting coins from message")
	}
	err = m.bank.SendCoinsFromModuleToAccount(ctx, gammtypes.ModuleName, rcpt, coins)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "sending newly minted coins from message")
	}

	return nil, nil, nil
}

// TODO: this is very close to QueryPlugin.EstimatePrice, maybe we can pull out common code into one function
// that these both use? at least the routes / token In/Out calculation
func (m *MintTokenMessenger) swapTokens(ctx sdk.Context, contractAddr sdk.AccAddress, swap *wasmbindings.SwapMsg) ([]sdk.Event, [][]byte, error) {
	if len(swap.Route) != 0 {
		return nil, nil, wasmvmtypes.UnsupportedRequest{Kind: "TODO: multi-hop swaps"}
	}
	if swap.Amount.ExactIn != nil {
		routes := []gammtypes.SwapAmountInRoute{{
			PoolId:        swap.First.PoolId,
			TokenOutDenom: swap.First.DenomOut,
		}}
		tokenIn := sdk.Coin{
			Denom:  swap.First.DenomIn,
			Amount: swap.Amount.ExactIn.Input,
		}
		tokenOutMinAmount := swap.Amount.ExactIn.MinOutput
		_, err := m.gammKeeper.MultihopSwapExactAmountIn(ctx, contractAddr, routes, tokenIn, tokenOutMinAmount)
		if err != nil {
			return nil, nil, sdkerrors.Wrap(err, "gamm estimate price exact amount in")
		}
		return nil, nil, nil
	} else if swap.Amount.ExactOut != nil {
		routes := []gammtypes.SwapAmountOutRoute{{
			PoolId:       swap.First.PoolId,
			TokenInDenom: swap.First.DenomIn,
		}}
		tokenInMaxAmount := swap.Amount.ExactOut.MaxInput
		tokenOut := sdk.Coin{
			Denom:  swap.First.DenomOut,
			Amount: swap.Amount.ExactOut.Output,
		}
		_, err := m.gammKeeper.MultihopSwapExactAmountOut(ctx, contractAddr, routes, tokenInMaxAmount, tokenOut)
		if err != nil {
			return nil, nil, sdkerrors.Wrap(err, "gamm estimate price exact amount out")
		}
		return nil, nil, nil
	} else {
		return nil, nil, wasmvmtypes.UnsupportedRequest{Kind: "must support either Swap.ExactIn or Swap.ExactOut"}
	}
}

// this is a function, not method, so the message_plugin can use it
func GetFullDenom(contract string, subDenom string) (string, error) {
	// Address validation
	if _, err := parseAddress(contract); err != nil {
		return "", err
	}
	err := ValidateSubDenom(subDenom)
	if err != nil {
		return "", sdkerrors.Wrap(err, "validate sub-denom")
	}
	// TODO: Confirm "cw" prefix
	fullDenom := fmt.Sprintf("cw/%s/%s", contract, subDenom)

	return fullDenom, nil
}

func parseAddress(addr string) (sdk.AccAddress, error) {
	parsed, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "address from bech32")
	}
	err = sdk.VerifyAddressFormat(parsed)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "verify address format")
	}
	return parsed, nil
}

func ValidateSubDenom(subDenom string) error {
	if len(subDenom) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty sub-denom")
	}
	// TODO: Extra validations
	return nil
}
