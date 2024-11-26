package poolmanager

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

// TODO: spec and tests, including events
func (server msgServer) SwapExactAmountIn(goCtx context.Context, msg *types.MsgSwapExactAmountIn) (*types.MsgSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenOutAmount, err := server.keeper.RouteExactAmountIn(ctx, sender, msg.Routes, msg.TokenIn, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	// Swap event is handled elsewhere

	return &types.MsgSwapExactAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}

// TODO: spec and tests, including events
func (server msgServer) SwapExactAmountOut(goCtx context.Context, msg *types.MsgSwapExactAmountOut) (*types.MsgSwapExactAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenInAmount, err := server.keeper.RouteExactAmountOut(ctx, sender, msg.Routes, msg.TokenInMaxAmount, msg.TokenOut)
	if err != nil {
		return nil, err
	}

	// Swap event is handled elsewhere

	return &types.MsgSwapExactAmountOutResponse{TokenInAmount: tokenInAmount}, nil
}

func (server msgServer) SplitRouteSwapExactAmountIn(goCtx context.Context, msg *types.MsgSplitRouteSwapExactAmountIn) (*types.MsgSplitRouteSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenOutAmount, err := server.keeper.SplitRouteExactAmountIn(ctx, sender, msg.Routes, msg.TokenInDenom, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	// Swap event is handled in each pool module's SwapExactAmountIn

	return &types.MsgSplitRouteSwapExactAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}

func (server msgServer) SplitRouteSwapExactAmountOut(goCtx context.Context, msg *types.MsgSplitRouteSwapExactAmountOut) (*types.MsgSplitRouteSwapExactAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenInAmount, err := server.keeper.SplitRouteExactAmountOut(ctx, sender, msg.Routes, msg.TokenOutDenom, msg.TokenInMaxAmount)
	if err != nil {
		return nil, err
	}

	// Swap event is handled in each pool module's SwapExactAmountOut

	return &types.MsgSplitRouteSwapExactAmountOutResponse{TokenInAmount: tokenInAmount}, nil
}

func (server msgServer) SetDenomPairTakerFee(goCtx context.Context, msg *types.MsgSetDenomPairTakerFee) (*types.MsgSetDenomPairTakerFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, denomPair := range msg.DenomPairTakerFee {
		err := server.keeper.SenderValidationSetDenomPairTakerFee(ctx, msg.Sender, denomPair.TokenInDenom, denomPair.TokenOutDenom, denomPair.TakerFee)
		if err != nil {
			return nil, err
		}
	}

	// Set denom pair taker fee event is handled in each iteration of the loop above

	return &types.MsgSetDenomPairTakerFeeResponse{Success: true}, nil
}

func (server msgServer) SetTakerFeeShareAgreementForDenom(goCtx context.Context, msg *types.MsgSetTakerFeeShareAgreementForDenom) (*types.MsgSetTakerFeeShareAgreementForDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.accountKeeper.GetModuleAccount(ctx, govtypes.ModuleName)
	if msg.Sender != govAddr.GetAddress().String() {
		return nil, types.ErrUnauthorizedGov
	}

	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return nil, err
	}

	takerFeeShareAgreement := types.TakerFeeShareAgreement{
		Denom:       msg.Denom,
		SkimPercent: msg.SkimPercent,
		SkimAddress: msg.SkimAddress,
	}

	err := server.keeper.SetTakerFeeShareAgreementForDenom(ctx, takerFeeShareAgreement)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetTakerFeeShareAgreementForDenomPair,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyTakerFeeShareDenom, takerFeeShareAgreement.Denom),
			sdk.NewAttribute(types.AttributeKeyTakerFeeShareSkimPercent, takerFeeShareAgreement.SkimPercent.String()),
			sdk.NewAttribute(types.AttributeKeyTakerFeeShareSkimAddress, takerFeeShareAgreement.SkimAddress),
		),
	})

	return &types.MsgSetTakerFeeShareAgreementForDenomResponse{}, nil
}

func (server msgServer) SetRegisteredAlloyedPool(goCtx context.Context, msg *types.MsgSetRegisteredAlloyedPool) (*types.MsgSetRegisteredAlloyedPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.accountKeeper.GetModuleAccount(ctx, govtypes.ModuleName)
	if msg.Sender != govAddr.GetAddress().String() {
		return nil, types.ErrUnauthorizedGov
	}

	err := server.keeper.setRegisteredAlloyedPool(ctx, msg.PoolId)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetRegisteredAlloyedPool,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(msg.PoolId, 10)),
		),
	})

	return &types.MsgSetRegisteredAlloyedPoolResponse{}, nil
}
