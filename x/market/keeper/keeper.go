package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

// Keeper of the market store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramstypes.Subspace

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	OracleKeeper  types.OracleKeeper
}

// NewKeeper constructs a new keeper for oracle
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramstore paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	oracleKeeper types.OracleKeeper,
) Keeper {
	// ensure market module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramstore,
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		OracleKeeper:  oracleKeeper,
	}
}

// GetOsmosisPoolDelta returns the gap between the OsmosisPool and the OsmosisBasePool
func (k Keeper) GetOsmosisPoolDelta(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OsmosisPoolDeltaKey)
	if bz == nil {
		return sdk.ZeroDec()
	}

	dp := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &dp)
	return dp.Dec
}

// SetOsmosisPoolDelta updates OsmosisPoolDelta which is gap between the OsmosisPool and the BasePool
func (k Keeper) SetOsmosisPoolDelta(ctx sdk.Context, delta sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: delta})
	store.Set(types.OsmosisPoolDeltaKey, bz)
}

// ReplenishPools replenishes each pool(Osmo,Luna) to BasePool
func (k Keeper) ReplenishPools(ctx sdk.Context) {
	poolDelta := k.GetOsmosisPoolDelta(ctx)

	poolRecoveryPeriod := int64(k.PoolRecoveryPeriod(ctx))
	poolRegressionAmt := poolDelta.QuoInt64(poolRecoveryPeriod)

	// Replenish pools towards each base pool
	// regressionAmt cannot make delta zero
	poolDelta = poolDelta.Sub(poolRegressionAmt)

	k.SetOsmosisPoolDelta(ctx, poolDelta)
}
