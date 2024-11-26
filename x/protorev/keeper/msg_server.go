package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

type MsgServer struct {
	k Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface for the provided Keeper.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return MsgServer{k: keeper}
}

var _ types.MsgServer = MsgServer{}

// SetHotRoutes sets the hot routes for ProtoRev
func (m MsgServer) SetHotRoutes(c context.Context, msg *types.MsgSetHotRoutes) (*types.MsgSetHotRoutesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Ensure the account has the admin role and can make the tx
	if err := m.AdminCheck(ctx, msg.Admin); err != nil {
		return nil, err
	}

	// Delete all previously set hot routes
	m.k.DeleteAllTokenPairArbRoutes(ctx)

	// Set the new hot routes
	for _, tokenPairArbRoutes := range msg.HotRoutes {
		if err := m.k.SetTokenPairArbRoutes(ctx, tokenPairArbRoutes.TokenIn, tokenPairArbRoutes.TokenOut, tokenPairArbRoutes); err != nil {
			return nil, err
		}
	}

	return &types.MsgSetHotRoutesResponse{}, nil
}

// SetDeveloperAccount sets the developer account that will receive fees
func (m MsgServer) SetDeveloperAccount(c context.Context, msg *types.MsgSetDeveloperAccount) (*types.MsgSetDeveloperAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Ensure the account has the admin role and can make the tx
	if err := m.AdminCheck(ctx, msg.Admin); err != nil {
		return nil, err
	}

	// Set the developer account
	developer, err := sdk.AccAddressFromBech32(msg.DeveloperAccount)
	if err != nil {
		return nil, err
	}

	m.k.SetDeveloperAccount(ctx, developer)

	return &types.MsgSetDeveloperAccountResponse{}, nil
}

// SetMaxPoolPointsPerTx sets the maximum number of pool points that can be consumed per tx
func (m MsgServer) SetMaxPoolPointsPerTx(c context.Context, msg *types.MsgSetMaxPoolPointsPerTx) (*types.MsgSetMaxPoolPointsPerTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Ensure the account has the admin role and can make the tx
	if err := m.AdminCheck(ctx, msg.Admin); err != nil {
		return nil, err
	}

	maxPointsPerBlock, err := m.k.GetMaxPointsPerBlock(ctx)
	if err != nil {
		return nil, err
	}

	if msg.MaxPoolPointsPerTx > maxPointsPerBlock {
		return nil, errors.New("max pool points per tx cannot be greater than max pool points per block")
	}

	// Set the max pool points per tx
	if err := m.k.SetMaxPointsPerTx(ctx, msg.MaxPoolPointsPerTx); err != nil {
		return nil, err
	}

	return &types.MsgSetMaxPoolPointsPerTxResponse{}, nil
}

// SetMaxPoolPointsPerBlock sets the maximum number of pool points that can be consumed per block
func (m MsgServer) SetMaxPoolPointsPerBlock(c context.Context, msg *types.MsgSetMaxPoolPointsPerBlock) (*types.MsgSetMaxPoolPointsPerBlockResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Ensure the account has the admin role and can make the tx
	if err := m.AdminCheck(ctx, msg.Admin); err != nil {
		return nil, err
	}

	maxPointsPerTx, err := m.k.GetMaxPointsPerTx(ctx)
	if err != nil {
		return nil, err
	}

	if msg.MaxPoolPointsPerBlock < maxPointsPerTx {
		return nil, errors.New("max pool points per block cannot be less than max pool points per tx")
	}

	// Set the max pool points per block
	if err := m.k.SetMaxPointsPerBlock(ctx, msg.MaxPoolPointsPerBlock); err != nil {
		return nil, err
	}

	return &types.MsgSetMaxPoolPointsPerBlockResponse{}, nil
}

// SetInfoByPoolType sets the execution time/gas consumption parameters corresponding to each pool type.
// This distinction is necessary because the pool types have different execution times / gas consumption.
func (m MsgServer) SetInfoByPoolType(c context.Context, msg *types.MsgSetInfoByPoolType) (*types.MsgSetInfoByPoolTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Ensure the account has the admin role and can make the tx
	if err := m.AdminCheck(ctx, msg.Admin); err != nil {
		return nil, err
	}

	m.k.SetInfoByPoolType(ctx, msg.InfoByPoolType)

	return &types.MsgSetInfoByPoolTypeResponse{}, nil
}

// SetBaseDenoms sets the base denoms that will be used to generate cyclic arbitrage routes
func (m MsgServer) SetBaseDenoms(c context.Context, msg *types.MsgSetBaseDenoms) (*types.MsgSetBaseDenomsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Ensure the account has the admin role and can make the tx
	if err := m.AdminCheck(ctx, msg.Admin); err != nil {
		return nil, err
	}

	// Get the old base denoms
	baseDenoms, err := m.k.GetAllBaseDenoms(ctx)
	if err != nil {
		return nil, err
	}

	// Delete all pools associated with the base denoms
	for _, baseDenom := range baseDenoms {
		m.k.DeleteAllPoolsForBaseDenom(ctx, baseDenom.Denom)
	}

	if err := m.k.SetBaseDenoms(ctx, msg.BaseDenoms); err != nil {
		return nil, err
	}

	// Update all of the pools
	if err := m.k.UpdatePools(ctx); err != nil {
		return nil, err
	}

	return &types.MsgSetBaseDenomsResponse{}, nil
}

// AdminCheck ensures that the sender is the admin account.
func (m MsgServer) AdminCheck(ctx sdk.Context, admin string) error {
	sender, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return err
	}

	adminAccount := m.k.GetAdminAccount(ctx)

	// Ensure the admin and sender are the same
	if !adminAccount.Equals(sender) {
		return fmt.Errorf("sender account %s is not authorized. sender must be %s", sender.String(), adminAccount.String())
	}

	return nil
}
