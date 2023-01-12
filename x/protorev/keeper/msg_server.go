package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

type MsgServer struct {
	k Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface for the provided Keeper.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return &MsgServer{k: keeper}
}

var _ types.MsgServer = MsgServer{}

// SetHotRoutes sets the hot routes for the protocol
func (m MsgServer) SetHotRoutes(c context.Context, msg *types.MsgSetHotRoutes) (*types.MsgSetHotRoutesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Error checked in msg validation
	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	// If the admin account has not been set, ignore
	admin, err := m.k.GetAdminAccount(ctx)
	if err != nil {
		return nil, err
	}

	// If the admin account has been set, and the sender is not the admin, ignore
	if !admin.Equals(sender) {
		return nil, fmt.Errorf("sender account %s is not authorized to set hot routes. sender must be %s", sender.String(), admin.String())
	}

	// Set the hot routes
	m.k.DeleteAllTokenPairArbRoutes(ctx)
	for _, tokenPairArbRoutes := range msg.HotRoutes {
		m.k.SetTokenPairArbRoutes(ctx, tokenPairArbRoutes.TokenIn, tokenPairArbRoutes.TokenOut, tokenPairArbRoutes)
	}

	return &types.MsgSetHotRoutesResponse{}, nil
}

// SetDeveloperAccount sets the developer account that will receive fees
func (m MsgServer) SetDeveloperAccount(c context.Context, msg *types.MsgSetDeveloperAccount) (*types.MsgSetDeveloperAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	// If the admin account has not been set, ignore
	admin, err := m.k.GetAdminAccount(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure the admin and sender are the same
	if !admin.Equals(sender) {
		return nil, fmt.Errorf("sender account %s is not authorized to set developer account. sender must be %s", sender.String(), admin.String())
	}

	// Set the developer account
	developer, err := sdk.AccAddressFromBech32(msg.DeveloperAccount)
	if err != nil {
		return nil, err
	}

	m.k.SetDeveloperAccount(ctx, developer)

	return &types.MsgSetDeveloperAccountResponse{}, nil
}

// SetMaxRoutesPerTx sets the max routes per tx
func (m MsgServer) SetMaxRoutesPerTx(c context.Context, msg *types.MsgSetMaxRoutesPerTx) (*types.MsgSetMaxRoutesPerTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	// If the admin account has not been set, ignore
	admin, err := m.k.GetAdminAccount(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure the admin and sender are the same
	if !admin.Equals(sender) {
		return nil, fmt.Errorf("sender account %s is not authorized to set max routes per tx. sender must be %s", sender.String(), admin.String())
	}

	// Set the max routes per tx
	if err := m.k.SetMaxRoutesPerTx(ctx, msg.MaxRoutesPerTx); err != nil {
		return nil, err
	}

	return &types.MsgSetMaxRoutesPerTxResponse{}, nil
}

// SetMaxRoutesPerBlock sets the max routes per block
func (m MsgServer) SetMaxRoutesPerBlock(c context.Context, msg *types.MsgSetMaxRoutesPerBlock) (*types.MsgSetMaxRoutesPerBlockResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	// If the admin account has not been set, ignore
	admin, err := m.k.GetAdminAccount(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure the admin and sender are the same
	if !admin.Equals(sender) {
		return nil, fmt.Errorf("sender account %s is not authorized to set max routes per block. sender must be %s", sender.String(), admin.String())
	}

	// Set the max routes per block
	if err := m.k.SetMaxRoutesPerBlock(ctx, msg.MaxRoutesPerBlock); err != nil {
		return nil, err
	}

	return &types.MsgSetMaxRoutesPerBlockResponse{}, nil
}

// SetPoolWeights sets the weights corresponding to each pool type. This distinction is necessary because the
// pool types have different execution times. Each weight roughly corresponds to the amount of time it takes
// to simulate and execute a trade.
func (m MsgServer) SetPoolWeights(c context.Context, msg *types.MsgSetPoolWeights) (*types.MsgSetPoolWeightsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	// If the admin account has not been set, ignore
	admin, err := m.k.GetAdminAccount(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure the admin and sender are the same
	if !admin.Equals(sender) {
		return nil, fmt.Errorf("sender account %s is not authorized to set pool weights. sender must be %s", sender.String(), admin.String())
	}

	// Set the pool weights
	if err := m.k.SetPoolWeights(ctx, *msg.PoolWeights); err != nil {
		return nil, err
	}

	return &types.MsgSetPoolWeightsResponse{}, nil
}
