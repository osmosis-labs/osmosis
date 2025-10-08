package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/x/mint/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/mint keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params returns params of the mint module.
func (q Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// EpochProvisions returns minter.EpochProvisions of the mint module.
func (q Querier) EpochProvisions(c context.Context, _ *types.QueryEpochProvisionsRequest) (*types.QueryEpochProvisionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := q.Keeper.GetMinter(ctx)

	return &types.QueryEpochProvisionsResponse{EpochProvisions: minter.EpochProvisions}, nil
}

// Inflation returns the current minting inflation value.
func (q Querier) Inflation(c context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	inflation, err := q.Keeper.GetInflation(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryInflationResponse{Inflation: inflation}, nil
}

// BurnedSupply returns the total amount of burned tokens.
func (q Querier) BurnedSupply(c context.Context, _ *types.QueryBurnedRequest) (*types.QueryBurnedResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	// The burn address is osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030
	burnAddr, err := sdk.AccAddressFromBech32("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030")
	if err != nil {
		return nil, err
	}

	burnedBalance := q.Keeper.bankKeeper.GetBalance(ctx, burnAddr, params.MintDenom)

	return &types.QueryBurnedResponse{Burned: burnedBalance.Amount}, nil
}

// TotalSupply returns the total supply (minted - burned).
func (q Querier) TotalSupply(c context.Context, _ *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	// Get the minted supply (from bank module)
	mintedSupply := q.Keeper.bankKeeper.GetSupply(ctx, params.MintDenom)

	// Get the burned supply
	burnAddr, err := sdk.AccAddressFromBech32("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030")
	if err != nil {
		return nil, err
	}
	burnedBalance := q.Keeper.bankKeeper.GetBalance(ctx, burnAddr, params.MintDenom)

	// Total supply = minted - burned
	totalSupply := mintedSupply.Amount.Sub(burnedBalance.Amount)

	return &types.QueryTotalSupplyResponse{TotalSupply: totalSupply}, nil
}

// RestrictedSupply returns the supply held in restricted addresses.
// This includes:
// - Developer vesting account balance
// - Community pool balance
// - Developer vested addresses balance and staked amounts from the mint module parameters
// - Restricted addresses balance and staked amounts, typically known entity holdings
func (q Querier) RestrictedSupply(c context.Context, _ *types.QueryRestrictedSupplyRequest) (*types.QueryRestrictedSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	restrictedSupply := q.calculateRestrictedSupply(ctx, params)

	return &types.QueryRestrictedSupplyResponse{RestrictedSupply: restrictedSupply}, nil
}

// CirculatingSupply returns the circulating supply (minted - burned - restricted).
func (q Querier) CirculatingSupply(c context.Context, _ *types.QueryCirculatingSupplyRequest) (*types.QueryCirculatingSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	// Get the minted supply (from bank module)
	mintedSupply := q.Keeper.bankKeeper.GetSupply(ctx, params.MintDenom)

	// Get the burned supply
	burnAddr, err := sdk.AccAddressFromBech32("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030")
	if err != nil {
		return nil, err
	}
	burnedBalance := q.Keeper.bankKeeper.GetBalance(ctx, burnAddr, params.MintDenom)

	// Calculate restricted supply by calling the same logic as RestrictedSupply query
	restrictedSupply := q.calculateRestrictedSupply(ctx, params)

	// Circulating supply = minted - burned - restricted
	circulatingSupply := mintedSupply.Amount.Sub(burnedBalance.Amount).Sub(restrictedSupply)

	return &types.QueryCirculatingSupplyResponse{CirculatingSupply: circulatingSupply}, nil
}

// calculateRestrictedSupply is a helper that calculates the restricted supply.
// This is extracted to be reusable by both RestrictedSupply and CirculatingSupply queries.
func (q Querier) calculateRestrictedSupply(ctx sdk.Context, params types.Params) osmomath.Int {
	restrictedSupply := osmomath.ZeroInt()

	// 1. Developer vesting account balance
	devVestingAddr := q.Keeper.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	if devVestingAddr != nil {
		devVestingBalance := q.Keeper.bankKeeper.GetBalance(ctx, devVestingAddr, params.MintDenom)
		restrictedSupply = restrictedSupply.Add(devVestingBalance.Amount)
	}

	// 2. Community pool balance
	communityPoolAddr := q.Keeper.accountKeeper.GetModuleAddress(distributiontypes.ModuleName)
	if communityPoolAddr != nil {
		communityPoolBalance := q.Keeper.bankKeeper.GetBalance(ctx, communityPoolAddr, params.MintDenom)
		restrictedSupply = restrictedSupply.Add(communityPoolBalance.Amount)
	}

	// 3. Developer vested addresses (from weighted_developer_rewards_receivers)
	for _, devAddr := range params.WeightedDeveloperRewardsReceivers {
		if devAddr.Address == "" {
			continue // Skip empty addresses (community pool allocations)
		}
		addr, err := sdk.AccAddressFromBech32(devAddr.Address)
		if err != nil {
			continue // Skip invalid addresses
		}
		// Add balance
		balance := q.Keeper.bankKeeper.GetBalance(ctx, addr, params.MintDenom)
		restrictedSupply = restrictedSupply.Add(balance.Amount)

		// Add staked amount
		stakedAmount := q.getStakedAmount(ctx, addr, params.MintDenom)
		restrictedSupply = restrictedSupply.Add(stakedAmount)
	}

	// 4. Restricted addresses (from restricted_asset_addresses)
	for _, addrStr := range params.RestrictedAssetAddresses {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			continue // Skip invalid addresses
		}
		// Add balance
		balance := q.Keeper.bankKeeper.GetBalance(ctx, addr, params.MintDenom)
		restrictedSupply = restrictedSupply.Add(balance.Amount)

		// Add staked amount
		stakedAmount := q.getStakedAmount(ctx, addr, params.MintDenom)
		restrictedSupply = restrictedSupply.Add(stakedAmount)
	}

	return restrictedSupply
}

// getStakedAmount returns the total amount staked by a delegator.
// It properly converts delegation shares to tokens using the validator's exchange rate.
func (q Querier) getStakedAmount(ctx sdk.Context, delegator sdk.AccAddress, denom string) osmomath.Int {
	totalStaked := osmomath.ZeroInt()

	// Iterate through all delegations for this delegator
	err := q.Keeper.stakingKeeper.IterateDelegations(ctx, delegator, func(_ int64, delegation stakingtypes.DelegationI) bool {
		shares := delegation.GetShares()

		// Get the validator to convert shares to tokens
		valAddr, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
		if err != nil {
			return false // Continue iteration
		}

		validator, err := q.Keeper.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			// If validator not found (perhaps unbonded/removed), use shares as approximation
			totalStaked = totalStaked.Add(shares.TruncateInt())
			return false // Continue iteration
		}

		// Convert shares to tokens using the validator's exchange rate
		// This accounts for slashing and other events that affect the share-to-token ratio
		tokens := validator.TokensFromShares(shares)
		totalStaked = totalStaked.Add(tokens.TruncateInt())

		return false // Continue iteration
	})

	if err != nil {
		return osmomath.ZeroInt()
	}

	return totalStaked
}
