package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

// Keeper of the treasury store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramstypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	marketKeeper  types.MarketKeeper
	stakingKeeper types.StakingKeeper
	distrKeeper   types.DistributionKeeper
	oracleKeeper  types.OracleKeeper
	wasmKeeper    *wasmkeeper.Keeper

	distributionModuleName string
}

// NewKeeper creates a new treasury Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	marketKeeper types.MarketKeeper,
	oracleKeeper types.OracleKeeper,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistributionKeeper,
	wasmKeeper *wasmkeeper.Keeper,
	distributionModuleName string,
) Keeper {
	// ensure treasury module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                    cdc,
		storeKey:               storeKey,
		paramSpace:             paramSpace,
		accountKeeper:          accountKeeper,
		bankKeeper:             bankKeeper,
		marketKeeper:           marketKeeper,
		oracleKeeper:           oracleKeeper,
		stakingKeeper:          stakingKeeper,
		distrKeeper:            distrKeeper,
		wasmKeeper:             wasmKeeper,
		distributionModuleName: distributionModuleName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetTaxRate loads the tax rate
func (k Keeper) GetTaxRate(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.TaxRateKey)
	if b == nil {
		return types.DefaultTaxRate
	}

	dp := sdk.DecProto{}
	k.cdc.MustUnmarshal(b, &dp)
	return dp.Dec
}

// SetTaxRate sets the tax rate
func (k Keeper) SetTaxRate(ctx sdk.Context, taxRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sdk.DecProto{Dec: taxRate})
	store.Set(types.TaxRateKey, b)
}
