package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	"github.com/osmosis-labs/osmosis/v27/x/market/types"
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

func (k Keeper) GetExchangePoolBalance(ctx sdk.Context) sdk.Coin {
	account := k.GetMarketAccount(ctx)
	if account == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return k.BankKeeper.GetBalance(ctx, account.GetAddress(), appparams.BaseCoinUnit)
}

// GetExchangeRequirement calculates the total amount of Melody asset required to back the assets in the oracle module.
func (k Keeper) GetExchangeRequirement(ctx sdk.Context) osmomath.Dec {
	total := osmomath.ZeroDec()
	for _, req := range k.getExchangeRates(ctx) {
		total = total.Add(req.BaseCurrency.Amount.ToLegacyDec().Mul(req.ExchangeRate))
	}
	return total
}

func (k Keeper) getExchangeRates(ctx sdk.Context) []types.ExchangeRequirement {
	var result []types.ExchangeRequirement
	k.OracleKeeper.IterateNoteExchangeRates(ctx, func(denom string, exchangeRate osmomath.Dec) (stop bool) {
		supply := k.BankKeeper.GetSupply(ctx, denom)
		result = append(result, types.ExchangeRequirement{
			BaseCurrency: supply,
			ExchangeRate: exchangeRate,
		})
		return false
	})
	return result
}
