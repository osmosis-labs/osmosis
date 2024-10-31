package keeper

import (
	"context"

	"net/url"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) DenomAuthorityMetadata(ctx context.Context, req *types.QueryDenomAuthorityMetadataRequest) (*types.QueryDenomAuthorityMetadataResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	decodedDenom, err := url.QueryUnescape(req.Denom)
	if err == nil {
		req.Denom = decodedDenom
	}
	authorityMetadata, err := k.GetAuthorityMetadata(sdkCtx, req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomAuthorityMetadataResponse{AuthorityMetadata: authorityMetadata}, nil
}

func (k Keeper) DenomsFromCreator(ctx context.Context, req *types.QueryDenomsFromCreatorRequest) (*types.QueryDenomsFromCreatorResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	denoms := k.getDenomsFromCreator(sdkCtx, req.GetCreator())
	return &types.QueryDenomsFromCreatorResponse{Denoms: denoms}, nil
}

func (k Keeper) BeforeSendHookAddress(ctx context.Context, req *types.QueryBeforeSendHookAddressRequest) (*types.QueryBeforeSendHookAddressResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	decodedDenom, err := url.QueryUnescape(req.Denom)
	if err == nil {
		req.Denom = decodedDenom
	}

	cosmwasmAddress := k.GetBeforeSendHook(sdkCtx, req.GetDenom())

	return &types.QueryBeforeSendHookAddressResponse{CosmwasmAddress: cosmwasmAddress}, nil
}

func (k Keeper) AllBeforeSendHooksAddresses(ctx context.Context, req *types.QueryAllBeforeSendHooksAddressesRequest) (*types.QueryAllBeforeSendHooksAddressesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	denoms, beforesendHookAddresses := k.GetAllBeforeSendHooks(sdkCtx)

	return &types.QueryAllBeforeSendHooksAddressesResponse{Denoms: denoms, BeforeSendHookAddresses: beforesendHookAddresses}, nil
}
