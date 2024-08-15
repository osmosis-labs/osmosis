package keeper

import (
	"fmt"
	"math"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
	markettypes "github.com/osmosis-labs/osmosis/v23/x/market/types"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"
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

// RefillExchangePool sends coins from the treasury module account to the market module account whenever there is a need to
// refill it. It returns the number of coins sent to the market module account.
func (k Keeper) RefillExchangePool(ctx sdk.Context) sdk.Dec {
	exchangeAmount := k.marketKeeper.GetExchangePoolBalance(ctx).Amount.ToLegacyDec()
	reserveAmount := k.GetReservePoolBalance(ctx).Amount.ToLegacyDec()
	exchangeRequirement := k.marketKeeper.GetExchangeRequirement(ctx)

	if exchangeAmount.LT(exchangeRequirement) {
		params := k.GetParams(ctx)
		percentMissing := 100 - (exchangeAmount.Quo(exchangeRequirement).Mul(sdk.NewDec(100))).TruncateInt64()
		if sdk.NewDec(percentMissing).GT(params.ReserveAllowableOffset) {
			refillAmount := sdk.MinDec(reserveAmount, exchangeRequirement.Sub(exchangeAmount))
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
	return sdk.ZeroDec()
}

// UpdateReserveFee updates the ReserveFeeMultiplier based on the current reserve balance and requirement.
// If reserve is insufficient, the fee multiplier is increased based on the percentage difference.
func (k Keeper) UpdateReserveFee(ctx sdk.Context) sdk.Dec {
	currentTaxRate := k.GetTaxRate(ctx)
	newTaxRate := sdk.ZeroDec()
	reserveAmount := k.GetReservePoolBalance(ctx).Amount.ToLegacyDec()
	exchangeRequirement := k.marketKeeper.GetExchangeRequirement(ctx)
	if reserveAmount.LT(exchangeRequirement) {
		params := k.GetParams(ctx)
		percentMissing := 100 - (reserveAmount.Quo(exchangeRequirement).Mul(sdk.NewDec(100))).TruncateInt64()
		if sdk.NewDec(percentMissing).GT(params.ReserveAllowableOffset) {
			// Determine the power of 2 that the percentMissing falls beneath
			powerOf2 := uint64(math.Log2(float64(percentMissing)))
			newTaxRate = sdk.MinDec(params.MaxFeeMultiplier, currentTaxRate.Mul(sdk.NewDec(2).Power(powerOf2+1)))
		} else {
			// Double the base fee to fill the remaining difference
			newTaxRate = currentTaxRate.Mul(sdk.NewDec(2))
		}
	}
	k.SetTaxRate(ctx, newTaxRate)
	return newTaxRate
}
