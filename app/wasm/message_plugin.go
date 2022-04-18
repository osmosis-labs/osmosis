package wasm

import (
	"encoding/json"

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
		return &CustomMessenger{
			wrapped:    old,
			bank:       bank,
			gammKeeper: gammKeeper,
		}
	}
}

type CustomMessenger struct {
	wrapped    wasmkeeper.Messenger
	bank       *bankkeeper.BaseKeeper
	gammKeeper *gammkeeper.Keeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		// only handle the happy path where this is really minting / swapping ...
		// leave everything else for the wrapped version
		var contractMsg wasmbindings.OsmosisMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, sdkerrors.Wrap(err, "osmosis msg")
		}
		// if contractMsg.MintTokens != nil {
		// 	return m.mintTokens(ctx, contractAddr, contractMsg.MintTokens)
		// }
		if contractMsg.Swap != nil {
			return m.swapTokens(ctx, contractAddr, contractMsg.Swap)
		}
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

// func (m *CustomMessenger) mintTokens(ctx sdk.Context, contractAddr sdk.AccAddress, mint *wasmbindings.MintTokens) ([]sdk.Event, [][]byte, error) {
// 	err := PerformMint(m.bank, ctx, contractAddr, mint)
// 	if err != nil {
// 		return nil, nil, sdkerrors.Wrap(err, "perform mint")
// 	}
// 	return nil, nil, nil
// }

// func PerformMint(b *bankkeeper.BaseKeeper, ctx sdk.Context, contractAddr sdk.AccAddress, mint *wasmbindings.MintTokens) error {
// 	if mint == nil {
// 		return wasmvmtypes.InvalidRequest{Err: "mint token null mint"}
// 	}
// 	rcpt, err := parseAddress(mint.Recipient)
// 	if err != nil {
// 		return err
// 	}

// 	denom, err := GetFullDenom(contractAddr.String(), mint.SubDenom)
// 	if err != nil {
// 		return sdkerrors.Wrap(err, "mint token denom")
// 	}
// 	if mint.Amount.IsNegative() {
// 		return wasmvmtypes.InvalidRequest{Err: "mint token negative amount"}
// 	}
// 	coins := []sdk.Coin{sdk.NewCoin(denom, mint.Amount)}

// 	err = b.MintCoins(ctx, gammtypes.ModuleName, coins)
// 	if err != nil {
// 		return sdkerrors.Wrap(err, "minting coins from message")
// 	}
// 	err = b.SendCoinsFromModuleToAccount(ctx, gammtypes.ModuleName, rcpt, coins)
// 	if err != nil {
// 		return sdkerrors.Wrap(err, "sending newly minted coins from message")
// 	}
// 	return nil
// }

func (m *CustomMessenger) swapTokens(ctx sdk.Context, contractAddr sdk.AccAddress, swap *wasmbindings.SwapMsg) ([]sdk.Event, [][]byte, error) {
	_, err := PerformSwap(m.gammKeeper, ctx, contractAddr, swap)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "perform swap")
	}
	return nil, nil, nil
}

// PerformSwap can be used both for the real swap, and the EstimateSwap query
func PerformSwap(keeper *gammkeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, swap *wasmbindings.SwapMsg) (*wasmbindings.SwapAmount, error) {
	if swap == nil {
		return nil, wasmvmtypes.InvalidRequest{Err: "gamm perform swap null swap"}
	}
	if swap.Amount.ExactIn != nil {
		routes := []gammtypes.SwapAmountInRoute{{
			PoolId:        swap.First.PoolId,
			TokenOutDenom: swap.First.DenomOut,
		}}
		for _, step := range swap.Route {
			routes = append(routes, gammtypes.SwapAmountInRoute{
				PoolId:        step.PoolId,
				TokenOutDenom: step.DenomOut,
			})
		}
		if swap.Amount.ExactIn.Input.IsNegative() {
			return nil, wasmvmtypes.InvalidRequest{Err: "gamm perform swap negative amount in"}
		}
		tokenIn := sdk.Coin{
			Denom:  swap.First.DenomIn,
			Amount: swap.Amount.ExactIn.Input,
		}
		tokenOutMinAmount := swap.Amount.ExactIn.MinOutput
		tokenOutAmount, err := keeper.MultihopSwapExactAmountIn(ctx, contractAddr, routes, tokenIn, tokenOutMinAmount)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "gamm perform swap exact amount in")
		}
		return &wasmbindings.SwapAmount{Out: &tokenOutAmount}, nil
	} else if swap.Amount.ExactOut != nil {
		routes := []gammtypes.SwapAmountOutRoute{{
			PoolId:       swap.First.PoolId,
			TokenInDenom: swap.First.DenomIn,
		}}
		output := swap.First.DenomOut
		for _, step := range swap.Route {
			routes = append(routes, gammtypes.SwapAmountOutRoute{
				PoolId:       step.PoolId,
				TokenInDenom: output,
			})
			output = step.DenomOut
		}
		tokenInMaxAmount := swap.Amount.ExactOut.MaxInput
		if swap.Amount.ExactOut.Output.IsNegative() {
			return nil, wasmvmtypes.InvalidRequest{Err: "gamm perform swap negative amount out"}
		}
		tokenOut := sdk.Coin{
			Denom:  output,
			Amount: swap.Amount.ExactOut.Output,
		}
		tokenInAmount, err := keeper.MultihopSwapExactAmountOut(ctx, contractAddr, routes, tokenInMaxAmount, tokenOut)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "gamm perform swap exact amount out")
		}
		return &wasmbindings.SwapAmount{In: &tokenInAmount}, nil
	} else {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "must support either Swap.ExactIn or Swap.ExactOut"}
	}
}

// // GetFullDenom is a function, not method, so the message_plugin can use it
// func GetFullDenom(contract string, subDenom string) (string, error) {
// 	// Address validation
// 	if _, err := parseAddress(contract); err != nil {
// 		return "", err
// 	}
// 	err := ValidateSubDenom(subDenom)
// 	if err != nil {
// 		return "", sdkerrors.Wrap(err, "validate sub-denom")
// 	}
// 	fullDenom := fmt.Sprintf("cw/%s/%s", contract, subDenom)

// 	return fullDenom, nil
// }

// func parseAddress(addr string) (sdk.AccAddress, error) {
// 	parsed, err := sdk.AccAddressFromBech32(addr)
// 	if err != nil {
// 		return nil, sdkerrors.Wrap(err, "address from bech32")
// 	}
// 	err = sdk.VerifyAddressFormat(parsed)
// 	if err != nil {
// 		return nil, sdkerrors.Wrap(err, "verify address format")
// 	}
// 	return parsed, nil
// }

// const reSubdenomStr = `^[a-zA-Z][a-zA-Z0-9]{2,31}$`

// var reSubdenom *regexp.Regexp

// func init() {
// 	reSubdenom = regexp.MustCompile(reSubdenomStr)
// }

// func ValidateSubDenom(subDenom string) error {
// 	if !reSubdenom.MatchString(subDenom) {
// 		return fmt.Errorf("invalid subdenom: %s", subDenom)
// 	}
// 	return nil
// }
