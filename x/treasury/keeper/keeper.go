package keeper

import (
	"fmt"
	"math"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/osmomath"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"
	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

// Keeper of the treasury store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramstypes.Subspace

	accountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	marketKeeper  types.MarketKeeper
	oracleKeeper  types.OracleKeeper
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
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		accountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		marketKeeper:  marketKeeper,
		oracleKeeper:  oracleKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetTaxRate loads the tax rate. Returned value is in percents.
func (k Keeper) GetTaxRate(ctx sdk.Context) osmomath.Dec {
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
func (k Keeper) SetTaxRate(ctx sdk.Context, taxRate osmomath.Dec) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sdk.DecProto{Dec: taxRate})
	store.Set(types.TaxRateKey, b)
}

// RefillExchangePool sends coins from the treasury module account to the market module account whenever there is a need to
// refill it. It returns the number of coins sent to the market module account.
func (k Keeper) RefillExchangePool(ctx sdk.Context) osmomath.Dec {
	exchangeAmount := k.marketKeeper.GetExchangePoolBalance(ctx).Amount.ToLegacyDec()
	reserveAmount := k.GetReservePoolBalance(ctx).Amount.ToLegacyDec()
	exchangeRequirement := k.marketKeeper.GetExchangeRequirement(ctx)

	if exchangeAmount.LT(exchangeRequirement) {
		params := k.GetParams(ctx)
		percentMissing := 100 - (exchangeAmount.Quo(exchangeRequirement).Mul(osmomath.NewDec(100))).TruncateInt64()
		if osmomath.NewDec(percentMissing).GT(params.ReserveAllowableOffset) {
			refillAmount := osmomath.MinDec(reserveAmount, exchangeRequirement.Sub(exchangeAmount))
			if refillAmount.IsPositive() {
				err := k.BankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, markettypes.ModuleName,
					sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, refillAmount.TruncateInt())))
				if err != nil {
					panic(err)
				}
			}
			return refillAmount
		}
	}
	return osmomath.ZeroDec()
}

// UpdateReserveFee updates the ReserveFeeMultiplier based on the current reserve balance and requirement.
// If reserve is insufficient, the fee multiplier is increased based on the percentage difference.
func (k Keeper) UpdateReserveFee(ctx sdk.Context) osmomath.Dec {
	baseFeeRate := types.DefaultTaxRate
	newTaxRate := osmomath.ZeroDec()
	reserveAmount := k.GetReservePoolBalance(ctx).Amount.ToLegacyDec()
	exchangeRequirement := k.marketKeeper.GetExchangeRequirement(ctx)
	if reserveAmount.LT(exchangeRequirement) {
		params := k.GetParams(ctx)
		percentMissing := 100 - (reserveAmount.Quo(exchangeRequirement).Mul(osmomath.NewDec(100))).TruncateInt64()
		if osmomath.NewDec(percentMissing).GT(params.ReserveAllowableOffset) {
			// Determine the power of 2 that the percentMissing falls beneath
			powerOf2 := uint64(math.Log2(float64(percentMissing)))
			newTaxRate = osmomath.MinDec(params.MaxFeeMultiplier, baseFeeRate.Mul(osmomath.NewDec(2).Power(powerOf2+1)))
		} else {
			// Double the base fee to fill the remaining difference
			newTaxRate = baseFeeRate.Mul(osmomath.NewDec(2))
		}
	} else {
		newTaxRate = osmomath.ZeroDec()
	}
	k.SetTaxRate(ctx, newTaxRate)
	return newTaxRate
}
