package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gogotypes "github.com/cosmos/gogoproto/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// ----------------------- Statistics Stores  ----------------------- //

// GetNumberOfTrades returns the number of trades executed by the ProtoRev module
func (k Keeper) GetNumberOfTrades(ctx sdk.Context) (osmomath.Int, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixNumberOfTrades)

	bz := store.Get(types.KeyPrefixNumberOfTrades)
	if len(bz) == 0 {
		return osmomath.ZeroInt(), errors.New("no trades have been executed by the protorev module")
	}

	trades := osmomath.Int{}
	if err := trades.Unmarshal(bz); err != nil {
		return osmomath.ZeroInt(), err
	}

	return trades, nil
}

// IncrementNumberOfTrades increments the number of trades executed by the ProtoRev module
func (k Keeper) IncrementNumberOfTrades(ctx sdk.Context) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixNumberOfTrades)

	numberOfTrades, _ := k.GetNumberOfTrades(ctx)
	numberOfTrades = numberOfTrades.Add(oneInt)

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
		return sdk.NewCoin(denom, osmomath.ZeroInt()), fmt.Errorf("no profits for denom %s", denom)
	}

	profits := sdk.Coin{}
	if err := profits.Unmarshal(bz); err != nil {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), err
	}

	return profits, nil
}

// GetAllProfits returns all of the profits made by the ProtoRev module.
func (k Keeper) GetAllProfits(ctx sdk.Context) []sdk.Coin {
	profits := make([]sdk.Coin, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixProfitByDenom)

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

func (k Keeper) SetCyclicArbProfitTrackerValue(ctx sdk.Context, cyclicArbProfits sdk.Coins) {
	newCyclicArbProfits := poolmanagertypes.TrackedVolume{
		Amount: cyclicArbProfits,
	}
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyCyclicArbTracker, &newCyclicArbProfits)
}

func (k Keeper) GetCyclicArbProfitTrackerValue(ctx sdk.Context) (currentCyclicArbProfits sdk.Coins) {
	var cyclicArbProfits poolmanagertypes.TrackedVolume
	cyclicArbProfitsFound, err := osmoutils.Get(ctx.KVStore(k.storeKey), types.KeyCyclicArbTracker, &cyclicArbProfits)
	if err != nil {
		// We can only encounter an error if a database or serialization errors occurs, so we panic here.
		// Normally this would be handled by `osmoutils.MustGet`, but since we want to specifically use `osmoutils.Get`,
		// we also have to manually panic here.
		panic(err)
	}

	// If no volume was found, we treat the existing volume as 0.
	// While we can technically require volume to exist, we would need to store empty coins in state for each pool (past and present),
	// which is a high storage cost to pay for a weak guardrail.
	currentCyclicArbProfits = sdk.NewCoins()
	if cyclicArbProfitsFound {
		currentCyclicArbProfits = cyclicArbProfits.Amount
	}

	return currentCyclicArbProfits
}

// GetCyclicArbProfitTrackerStartHeight gets the height from which we started accounting for cyclic arb profits.
func (k Keeper) GetCyclicArbProfitTrackerStartHeight(ctx sdk.Context) int64 {
	startHeight := gogotypes.Int64Value{}
	osmoutils.MustGet(ctx.KVStore(k.storeKey), types.KeyCyclicArbTrackerStartHeight, &startHeight)
	return startHeight.Value
}

// SetCyclicArbProfitTrackerStartHeight sets the height from which we started accounting for cyclic arb profits.
func (k Keeper) SetCyclicArbProfitTrackerStartHeight(ctx sdk.Context, startHeight int64) {
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyCyclicArbTrackerStartHeight, &gogotypes.Int64Value{Value: startHeight})
}

// UpdateProfitsByDenom updates the profits made by the ProtoRev module for the given denom
func (k Keeper) UpdateProfitsByDenom(ctx sdk.Context, denom string, tradeProfit osmomath.Int) error {
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
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixTradesByRoute)

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
func (k Keeper) GetTradesByRoute(ctx sdk.Context, route []uint64) (osmomath.Int, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTradesByRoute)
	key := types.GetKeyPrefixTradesByRoute(route)

	bz := store.Get(key)
	if len(bz) == 0 {
		return osmomath.ZeroInt(), fmt.Errorf("no trades for route %d", route)
	}

	trades := osmomath.Int{}
	if err := trades.Unmarshal(bz); err != nil {
		return osmomath.ZeroInt(), err
	}
	return trades, nil
}

// IncrementTradesByRoute increments the number of trades executed by the ProtoRev module for the given route
func (k Keeper) IncrementTradesByRoute(ctx sdk.Context, route []uint64) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTradesByRoute)
	key := types.GetKeyPrefixTradesByRoute(route)

	trades, _ := k.GetTradesByRoute(ctx, route)
	trades = trades.Add(oneInt)
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
		return sdk.NewCoin(denom, osmomath.ZeroInt()), fmt.Errorf("no profits for route %d", route)
	}

	profits := sdk.Coin{}
	if err := profits.Unmarshal(bz); err != nil {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), err
	}

	return profits, nil
}

// GetAllProfitsByRoute returns all of the profits made by the ProtoRev module for the given route
func (k Keeper) GetAllProfitsByRoute(ctx sdk.Context, route []uint64) []sdk.Coin {
	profits := make([]sdk.Coin, 0)

	store := ctx.KVStore(k.storeKey)
	prefix := append(types.KeyPrefixProfitsByRoute, types.GetKeyPrefixProfitsByRoute(route, "")...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)

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
func (k Keeper) UpdateProfitsByRoute(ctx sdk.Context, route []uint64, denom string, profit osmomath.Int) error {
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
func (k Keeper) UpdateStatistics(ctx sdk.Context, route poolmanagertypes.SwapAmountInRoutes, denom string, profit osmomath.Int) error {
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
