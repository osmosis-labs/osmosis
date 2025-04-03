package keeper

import (
	"fmt"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount.
func (k Keeper) ConvertToBaseToken(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error) {
	exchangeRatio, err := k.oracleKeeper.GetMelodyExchangeRate(ctx, inputFee.Denom)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("could not get exchange rate for %s: %w", inputFee.Denom, err)
	}

	// Note: spotPrice truncation is done here for maintaining state-compatibility with v19.x
	// It should be changed to support full spot price precision before
	// https://github.com/osmosis-labs/osmosis/issues/6064 is complete
	return sdk.NewCoin(appparams.BaseCoinUnit, exchangeRatio.MulIntMut(inputFee.Amount).RoundInt()), nil
}

func (k Keeper) GetBaseDenom(_ sdk.Context) (denom string, err error) {
	return appparams.BaseCoinUnit, nil
}

func (k Keeper) IsFeeToken(ctx sdk.Context, denom string) bool {
	_, err := k.oracleKeeper.GetMelodyExchangeRate(ctx, denom)
	return err == nil
}
