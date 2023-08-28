package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v19/x/protorev/types"
)

// ----------------------- Statistics Stores  ----------------------- //

// GetNumberOfTrades returns the number of trades executed by the ProtoRev module
func (k Keeper) GetNumberOfTrades(ctx sdk.Context) (sdk.Int, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixNumberOfTrades)

	bz := store.Get(types.KeyPrefixNumberOfTrades)
	if len(bz) == 0 {
		return sdk.ZeroInt(), fmt.Errorf("no trades have been executed by the protorev module")
	}

	trades := sdk.Int{}
	if err := trades.Unmarshal(bz); err != nil {
		return sdk.ZeroInt(), err
	}

	return trades, nil
}

// IncrementNumberOfTrades increments the number of trades executed by the ProtoRev module
func (k Keeper) IncrementNumberOfTrades(ctx sdk.Context) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixNumberOfTrades)

	numberOfTrades, _ := k.GetNumberOfTrades(ctx)
	numberOfTrades = numberOfTrades.Add(sdk.OneInt())

	bz, err := numberOfTrades.Marshal()
	if err != nil {
		return err
	}
	store.Set(types.KeyPrefixNumberOfTrades, bz)
	return nil
}

// GetProfitsByDenom returns the profits made by the ProtoRev module for the given denom
func (k Keeper) GetProfitsByDenom(ctx sdk.Context, denom string) (sdk.Coin, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixProfitByDenom)
	key := types.GetKeyPrefixProfitByDenom(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return sdk.NewCoin(denom, sdk.ZeroInt()), fmt.Errorf("no profits for denom %s", denom)
	}

	profits := sdk.Coin{}
	if err := profits.Unmarshal(bz); err != nil {
		return sdk.NewCoin(denom, sdk.ZeroInt()), err
	}

	return profits, nil
}

// GetAllProfits returns all of the profits made by the ProtoRev module.
func (k Keeper) GetAllProfits(ctx sdk.Context) []sdk.Coin {
	profits := make([]sdk.Coin, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixProfitByDenom)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		profit := sdk.Coin{}
		if err := profit.Unmarshal(bz); err == nil {
			profits = append(profits, profit)
		}
	}

	return profits
}

// UpdateProfitsByDenom updates the profits made by the ProtoRev module for the given denom
func (k Keeper) UpdateProfitsByDenom(ctx sdk.Context, denom string, tradeProfit sdk.Int) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixProfitByDenom)
	key := types.GetKeyPrefixProfitByDenom(denom)

	profits, _ := k.GetProfitsByDenom(ctx, denom)
	profits.Amount = profits.Amount.Add(tradeProfit)
	bz, err := profits.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// GetAllRoutes returns all of the routes that the ProtoRev module has traded on
func (k Keeper) GetAllRoutes(ctx sdk.Context) ([][]uint64, error) {
	routes := make([][]uint64, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixTradesByRoute)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		// ignore the portion of the key that is the prefix
		key := iterator.Key()[len(types.KeyPrefixTradesByRoute)+1:]

		// convert the key into a route
		route, err := types.CreateRouteFromKey(key)
		if err != nil {
			return nil, err
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// GetTradesByRoute returns the number of trades executed by the ProtoRev module for the given route
func (k Keeper) GetTradesByRoute(ctx sdk.Context, route []uint64) (sdk.Int, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTradesByRoute)
	key := types.GetKeyPrefixTradesByRoute(route)

	bz := store.Get(key)
	if len(bz) == 0 {
		return sdk.ZeroInt(), fmt.Errorf("no trades for route %d", route)
	}

	trades := sdk.Int{}
	if err := trades.Unmarshal(bz); err != nil {
		return sdk.ZeroInt(), err
	}
	return trades, nil
}

// IncrementTradesByRoute increments the number of trades executed by the ProtoRev module for the given route
func (k Keeper) IncrementTradesByRoute(ctx sdk.Context, route []uint64) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTradesByRoute)
	key := types.GetKeyPrefixTradesByRoute(route)

	trades, _ := k.GetTradesByRoute(ctx, route)
	trades = trades.Add(sdk.OneInt())
	bz, err := trades.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// GetProfitsByRoute returns the profits made by the ProtoRev module for the given route and denom
func (k Keeper) GetProfitsByRoute(ctx sdk.Context, route []uint64, denom string) (sdk.Coin, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixProfitsByRoute)
	key := types.GetKeyPrefixProfitsByRoute(route, denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return sdk.NewCoin(denom, sdk.ZeroInt()), fmt.Errorf("no profits for route %d", route)
	}

	profits := sdk.Coin{}
	if err := profits.Unmarshal(bz); err != nil {
		return sdk.NewCoin(denom, sdk.ZeroInt()), err
	}

	return profits, nil
}

// GetAllProfitsByRoute returns all of the profits made by the ProtoRev module for the given route
func (k Keeper) GetAllProfitsByRoute(ctx sdk.Context, route []uint64) []sdk.Coin {
	profits := make([]sdk.Coin, 0)

	store := ctx.KVStore(k.storeKey)
	prefix := append(types.KeyPrefixProfitsByRoute, types.GetKeyPrefixProfitsByRoute(route, "")...)
	iterator := sdk.KVStorePrefixIterator(store, prefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		profit := sdk.Coin{}
		if err := profit.Unmarshal(bz); err == nil {
			profits = append(profits, profit)
		}
	}

	return profits
}

// UpdateProfitsByRoute updates the profits made by the ProtoRev module for the given route and denom
func (k Keeper) UpdateProfitsByRoute(ctx sdk.Context, route []uint64, denom string, profit sdk.Int) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixProfitsByRoute)
	key := types.GetKeyPrefixProfitsByRoute(route, denom)

	profits, _ := k.GetProfitsByRoute(ctx, route, denom)
	profits.Amount = profits.Amount.Add(profit)
	bz, err := profits.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// UpdateStatistics updates the module statistics after each trade is executed
func (k Keeper) UpdateStatistics(ctx sdk.Context, route poolmanagertypes.SwapAmountInRoutes, denom string, profit sdk.Int) error {
	// Increment the number of trades executed by the ProtoRev module
	if err := k.IncrementNumberOfTrades(ctx); err != nil {
		return err
	}

	// Update the profits made by the ProtoRev module for the denom
	if err := k.UpdateProfitsByDenom(ctx, denom, profit); err != nil {
		return err
	}

	// Increment the number of times the module has executed a trade on a given route
	if err := k.IncrementTradesByRoute(ctx, route.PoolIds()); err != nil {
		return err
	}

	// Update the profits accumulated by the ProtoRev module for the given route and denom
	if err := k.UpdateProfitsByRoute(ctx, route.PoolIds(), denom, profit); err != nil {
		return err
	}

	return nil
}
