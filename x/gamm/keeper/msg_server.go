package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

func NewBalancerMsgServerImpl(keeper *Keeper) balancer.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var (
	_ types.MsgServer    = msgServer{}
	_ balancer.MsgServer = msgServer{}
)

// CreateBalancerPool is a create balancer pool message.
func (server msgServer) CreateBalancerPool(goCtx context.Context, msg *balancer.MsgCreateBalancerPool) (*balancer.MsgCreateBalancerPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := server.keeper.GetParams(ctx)

	//validate the pool contains asset which is whitelisted
	found := false
	for _, asset := range msg.PoolAssets {
		if ok, _ := params.PoolCreationFee.Find(asset.Token.Denom); ok {
			found = true
			break
		}
	}
	if !found {
		return nil, types.ErrPoolAssetNotAllowed
	}

	//set global fees
	if params.EnableGlobalPoolFees {
		msg.PoolParams.SwapFee = params.GlobalFees.SwapFee
		msg.PoolParams.ExitFee = params.GlobalFees.ExitFee
	}

	//validate uniqueness of pool assets
	poolAlreadyExists := false
	iter := server.keeper.iterator(ctx, types.KeyPrefixPools)
	defer iter.Close() //nolint:errcheck

	for ; iter.Valid(); iter.Next() {
		pool, err := server.keeper.UnmarshalPool(iter.Value())
		if err != nil {
			return nil, err
		}

		existingPoolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
		sameAssets := true
		for _, asset := range msg.PoolAssets {
			if contains(existingPoolDenoms, asset.Token.Denom) {
				continue
			}
			sameAssets = false
			break
		}
		if sameAssets {
			poolAlreadyExists = true
			break
		}
	}

	if poolAlreadyExists {
		return nil, types.ErrPoolAlreadyExists
	}

	poolId, err := server.CreatePool(goCtx, msg)
	return &balancer.MsgCreateBalancerPoolResponse{PoolID: poolId}, err
}

// Function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// CreatePool attempts to create a pool returning the newly created pool ID or an error upon failure.
// The pool creation fee is used to fund the community pool.
// It will create a dedicated module account for the pool and sends the initial liquidity to the created module account.
func (server msgServer) CreatePool(goCtx context.Context, msg poolmanagertypes.CreatePoolMsg) (poolId uint64, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	poolId, err = server.keeper.poolManager.CreatePool(ctx, msg)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.PoolCreator().String()),
		),
	})

	return poolId, nil
}

// JoinPool routes `JoinPoolNoSwap` where we do an abstract calculation on needed lp liquidity coins to get the designated
// amount of shares for the pool. (This is done by taking the number of shares we want and then using the total number of shares
// to get the ratio of the pool it accounts for. Using this ratio, we iterate over all pool assets to get the number of tokens we need
// to get the specified number of shares).
// Using the number of tokens needed to actually join the pool, we do a basic sanity check on whether the token does not exceed
// `TokenInMaxs`. Then we hit the actual implementation of `JoinPool` defined by each pool model.
// `JoinPool` takes in the tokensIn calculated above as the parameter rather than using the number of shares provided in the msg.
// This can result in negotiable difference between the number of shares provided within the msg
// and the actual number of share amount resulted from joining pool.
// Internal logic flow for each pool model is as follows:
// Balancer: TokensInMaxs provided as the argument must either contain no tokens or containing all assets in the pool.
// * For the case of a not containing tokens, we simply perform calculation of sharesOut and needed amount of tokens for joining the pool
func (server msgServer) JoinPool(goCtx context.Context, msg *types.MsgJoinPool) (*types.MsgJoinPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	neededLp, sharesOut, err := server.keeper.JoinPoolNoSwap(ctx, sender, msg.PoolId, msg.ShareOutAmount, msg.TokenInMaxs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgJoinPoolResponse{
		ShareOutAmount: sharesOut,
		TokenIn:        neededLp,
	}, nil
}

func (server msgServer) ExitPool(goCtx context.Context, msg *types.MsgExitPool) (*types.MsgExitPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	exitCoins, err := server.keeper.ExitPool(ctx, sender, msg.PoolId, msg.ShareInAmount, msg.TokenOutMins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgExitPoolResponse{
		TokenOut: exitCoins,
	}, nil
}

func (server msgServer) SwapExactAmountIn(goCtx context.Context, msg *types.MsgSwapExactAmountIn) (*types.MsgSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenInAfterSubTakerFee, takerFeesCoins := server.keeper.SubTakerFee(msg.TokenIn, server.keeper.GetParams(ctx).TakerFee)

	tokenOutAmount, err := server.keeper.poolManager.RouteExactAmountIn(ctx, sender, msg.Routes, tokenInAfterSubTakerFee, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	err = server.keeper.chargeTakerFee(ctx, takerFeesCoins, sender)
	if err != nil {
		return nil, err
	}

	// Swap event is handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSwapExactAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}

func (server msgServer) SwapExactAmountOut(goCtx context.Context, msg *types.MsgSwapExactAmountOut) (*types.MsgSwapExactAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	takerFee := server.keeper.GetParams(ctx).TakerFee

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	route := types.SwapAmountOutRoutes(msg.Routes)
	if err := route.Validate(); err != nil {
		return nil, err
	}

	//limit the the TokenInMaxAmount to have enough for taker fee
	maxTokenIn := sdk.NewCoin(msg.Routes[0].TokenInDenom, msg.TokenInMaxAmount)
	tokenInAfterSubTakerFee, _ := server.keeper.SubTakerFee(maxTokenIn, takerFee)

	tokenInAmount, err := server.keeper.poolManager.RouteExactAmountOut(ctx, sender, msg.Routes, tokenInAfterSubTakerFee.Amount, msg.TokenOut)
	if err != nil {
		return nil, err
	}

	tokenInCoin := sdk.NewCoin(msg.Routes[0].TokenInDenom, tokenInAmount)
	tokenInAmountWithTakerFee, takerFeeCoin := server.keeper.AddTakerFee(tokenInCoin, takerFee)

	err = server.keeper.chargeTakerFee(ctx, takerFeeCoin, sender)
	if err != nil {
		return nil, err
	}

	// Swap event is handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSwapExactAmountOutResponse{TokenInAmount: tokenInAmountWithTakerFee.Amount}, nil
}

// JoinSwapExactAmountIn is an LP transaction, that will LP all of the provided tokensIn coins.
// * For the case of a single token, we simply perform single asset join (balancer notation: pAo, pool shares amount out,
// given single asset in).
// For more details on the calculation of the number of shares look at the CalcJoinPoolShares function for the appropriate pool style
func (server msgServer) JoinSwapExternAmountIn(goCtx context.Context, msg *types.MsgJoinSwapExternAmountIn) (*types.MsgJoinSwapExternAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokensIn := sdk.Coins{msg.TokenIn}
	shareOutAmount, err := server.keeper.JoinSwapExactAmountIn(ctx, sender, msg.PoolId, tokensIn, msg.ShareOutMinAmount)
	if err != nil {
		return nil, err
	}

	// Swap and LP events are handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgJoinSwapExternAmountInResponse{ShareOutAmount: shareOutAmount}, nil
}

func (server msgServer) JoinSwapShareAmountOut(goCtx context.Context, msg *types.MsgJoinSwapShareAmountOut) (*types.MsgJoinSwapShareAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenInAmount, err := server.keeper.JoinSwapShareAmountOut(ctx, sender, msg.PoolId, msg.TokenInDenom, msg.ShareOutAmount, msg.TokenInMaxAmount)
	if err != nil {
		return nil, err
	}

	// Swap and LP events are handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgJoinSwapShareAmountOutResponse{TokenInAmount: tokenInAmount}, nil
}

func (server msgServer) ExitSwapExternAmountOut(goCtx context.Context, msg *types.MsgExitSwapExternAmountOut) (*types.MsgExitSwapExternAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	shareInAmount, err := server.keeper.ExitSwapExactAmountOut(ctx, sender, msg.PoolId, msg.TokenOut, msg.ShareInMaxAmount)
	if err != nil {
		return nil, err
	}

	// Swap and LP events are handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgExitSwapExternAmountOutResponse{ShareInAmount: shareInAmount}, nil
}

func (server msgServer) ExitSwapShareAmountIn(goCtx context.Context, msg *types.MsgExitSwapShareAmountIn) (*types.MsgExitSwapShareAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenOutAmount, err := server.keeper.ExitSwapShareAmountIn(ctx, sender, msg.PoolId, msg.TokenOutDenom, msg.ShareInAmount, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	// Swap and LP events are handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgExitSwapShareAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}
