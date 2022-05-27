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

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v7/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"
)

func CustomMessageDecorator(gammKeeper *gammkeeper.Keeper, bank *bankkeeper.BaseKeeper, tokenFactory *tokenfactorykeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:      old,
			bank:         bank,
			gammKeeper:   gammKeeper,
			tokenFactory: tokenFactory,
		}
	}
}

type CustomMessenger struct {
	wrapped      wasmkeeper.Messenger
	bank         *bankkeeper.BaseKeeper
	gammKeeper   *gammkeeper.Keeper
	tokenFactory *tokenfactorykeeper.Keeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		// only handle the happy path where this is really creating / minting / swapping ...
		// leave everything else for the wrapped version
		var contractMsg wasmbindings.OsmosisMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, sdkerrors.Wrap(err, "osmosis msg")
		}
		if contractMsg.CreateDenom != nil {
			return m.createDenom(ctx, contractAddr, contractMsg.CreateDenom)
		}
		if contractMsg.MintTokens != nil {
			return m.mintTokens(ctx, contractAddr, contractMsg.MintTokens)
		}
		if contractMsg.ChangeAdmin != nil {
			return m.changeAdmin(ctx, contractAddr, contractMsg.ChangeAdmin)
		}
		if contractMsg.BurnTokens != nil {
			return m.burnTokens(ctx, contractAddr, contractMsg.BurnTokens)
		}
		if contractMsg.Swap != nil {
			return m.swapTokens(ctx, contractAddr, contractMsg.Swap)
		}
		if contractMsg.ExitPool != nil {
			return m.exitPool(ctx, contractAddr, contractMsg.ExitPool)
		}
		if contractMsg.JoinPoolNoSwap != nil {
			return m.joinPoolNoSwap(ctx, contractAddr, contractMsg.JoinPoolNoSwap)
		}
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func (m *CustomMessenger) createDenom(ctx sdk.Context, contractAddr sdk.AccAddress, createDenom *wasmbindings.CreateDenom) ([]sdk.Event, [][]byte, error) {
	err := PerformCreateDenom(m.tokenFactory, m.bank, ctx, contractAddr, createDenom)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "perform create denom")
	}
	return nil, nil, nil
}

func PerformCreateDenom(f *tokenfactorykeeper.Keeper, b *bankkeeper.BaseKeeper, ctx sdk.Context, contractAddr sdk.AccAddress, createDenom *wasmbindings.CreateDenom) error {
	if createDenom == nil {
		return wasmvmtypes.InvalidRequest{Err: "create denom null create denom"}
	}

	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)

	msgCreateDenom := tokenfactorytypes.NewMsgCreateDenom(contractAddr.String(), createDenom.Subdenom)

	if err := msgCreateDenom.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "failed validating MsgCreateDenom")
	}

	// Create denom
	_, err := msgServer.CreateDenom(
		sdk.WrapSDKContext(ctx),
		msgCreateDenom,
	)
	if err != nil {
		return sdkerrors.Wrap(err, "creating denom")
	}
	return nil
}

func (m *CustomMessenger) mintTokens(ctx sdk.Context, contractAddr sdk.AccAddress, mint *wasmbindings.MintTokens) ([]sdk.Event, [][]byte, error) {
	err := PerformMint(m.tokenFactory, m.bank, ctx, contractAddr, mint)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "perform mint")
	}
	return nil, nil, nil
}

func PerformMint(f *tokenfactorykeeper.Keeper, b *bankkeeper.BaseKeeper, ctx sdk.Context, contractAddr sdk.AccAddress, mint *wasmbindings.MintTokens) error {
	if mint == nil {
		return wasmvmtypes.InvalidRequest{Err: "mint token null mint"}
	}
	rcpt, err := parseAddress(mint.MintToAddress)
	if err != nil {
		return err
	}

	coin := sdk.Coin{Denom: mint.Denom, Amount: mint.Amount}
	sdkMsg := tokenfactorytypes.NewMsgMint(contractAddr.String(), coin)
	if err = sdkMsg.ValidateBasic(); err != nil {
		return err
	}

	// Mint through token factory / message server
	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)
	_, err = msgServer.Mint(sdk.WrapSDKContext(ctx), sdkMsg)
	if err != nil {
		return sdkerrors.Wrap(err, "minting coins from message")
	}
	err = b.SendCoins(ctx, contractAddr, rcpt, sdk.NewCoins(coin))
	if err != nil {
		return sdkerrors.Wrap(err, "sending newly minted coins from message")
	}
	return nil
}

func (m *CustomMessenger) changeAdmin(ctx sdk.Context, contractAddr sdk.AccAddress, changeAdmin *wasmbindings.ChangeAdmin) ([]sdk.Event, [][]byte, error) {
	err := ChangeAdmin(m.tokenFactory, ctx, contractAddr, changeAdmin)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to change admin")
	}
	return nil, nil, nil
}

func ChangeAdmin(f *tokenfactorykeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, changeAdmin *wasmbindings.ChangeAdmin) error {
	if changeAdmin == nil {
		return wasmvmtypes.InvalidRequest{Err: "changeAdmin is nil"}
	}
	newAdminAddr, err := parseAddress(changeAdmin.NewAdminAddress)
	if err != nil {
		return err
	}

	changeAdminMsg := tokenfactorytypes.NewMsgChangeAdmin(contractAddr.String(), changeAdmin.Denom, newAdminAddr.String())
	if err := changeAdminMsg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)
	_, err = msgServer.ChangeAdmin(sdk.WrapSDKContext(ctx), changeAdminMsg)
	if err != nil {
		return sdkerrors.Wrap(err, "failed changing admin from message")
	}
	return nil
}

func (m *CustomMessenger) burnTokens(ctx sdk.Context, contractAddr sdk.AccAddress, burn *wasmbindings.BurnTokens) ([]sdk.Event, [][]byte, error) {
	err := PerformBurn(m.tokenFactory, ctx, contractAddr, burn)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "perform mint")
	}
	return nil, nil, nil
}

func PerformBurn(f *tokenfactorykeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, burn *wasmbindings.BurnTokens) error {
	if burn == nil {
		return wasmvmtypes.InvalidRequest{Err: "burn token null mint"}
	}
	if burn.BurnFromAddress != "" && burn.BurnFromAddress != contractAddr.String() {
		return wasmvmtypes.InvalidRequest{Err: "BurnFromAddress must be \"\""}
	}

	coin := sdk.Coin{Denom: burn.Denom, Amount: burn.Amount}
	sdkMsg := tokenfactorytypes.NewMsgBurn(contractAddr.String(), coin)
	if err := sdkMsg.ValidateBasic(); err != nil {
		return err
	}

	// Burn through token factory / message server
	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), sdkMsg)
	if err != nil {
		return sdkerrors.Wrap(err, "burning coins from message")
	}
	return nil
}

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

func (m *CustomMessenger) exitPool(ctx sdk.Context, contractAddr sdk.AccAddress, exitPool *wasmbindings.ExitPool) ([]sdk.Event, [][]byte, error) {
	err := PerformExit(m.gammKeeper, ctx, contractAddr, exitPool)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "exit pool")
	}
	return nil, nil, nil
}

func (m *CustomMessenger) joinPoolNoSwap(ctx sdk.Context, contractAddr sdk.AccAddress, joinPool *wasmbindings.JoinPoolNoSwap) ([]sdk.Event, [][]byte, error) {
	err := PerformJoinPoolNoSwap(m.gammKeeper, ctx, contractAddr, joinPool)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "join pool no swap")
	}
	return nil, nil, nil
}

func PerformExit(g *gammkeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, exitPool *wasmbindings.ExitPool) error {
	if exitPool == nil {
		return wasmvmtypes.InvalidRequest{Err: "join pool null"}
	}

	_, err := g.ExitPool(ctx, contractAddr, exitPool.PoolId, exitPool.ShareInAmount, exitPool.TokenOutMins)

	if err != nil {
		return err
	}

	return nil
}

func PerformJoinPoolNoSwap(g *gammkeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, joinPool *wasmbindings.JoinPoolNoSwap) error {
	if joinPool == nil {
		return wasmvmtypes.InvalidRequest{Err: "join pool no swap null"}
	}

	err := g.JoinPoolNoSwap(ctx, contractAddr, joinPool.PoolId, joinPool.ShareOutAmount, joinPool.TokenInMaxs)

	if err != nil {
		return err
	}

	return nil
}

// GetFullDenom is a function, not method, so the message_plugin can use it
func GetFullDenom(contract string, subDenom string) (string, error) {
	// Address validation
	if _, err := parseAddress(contract); err != nil {
		return "", err
	}
	fullDenom, err := tokenfactorytypes.GetTokenDenom(contract, subDenom)
	if err != nil {
		return "", sdkerrors.Wrap(err, "validate sub-denom")
	}

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
