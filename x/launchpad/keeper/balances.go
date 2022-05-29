package keeper

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Updates balances in a given substore.
// Expectations:
// + coins are validated
// TODO: probably we can remove it
func (k Keeper) addBalances(store storetypes.KVStore, coins sdk.Coins) error {
	for _, c := range coins {
		if c.IsZero() {
			continue
		}
		denom := []byte(c.Denom)
		bz := store.Get(denom)
		if bz == nil {
			bz, err := c.Amount.Marshal()
			if err != nil {
				return err
			}
			store.Set(denom, bz)
			continue
		}
		a := sdk.ZeroInt()
		if err := a.Unmarshal(bz); err != nil {
			return err
		}
		bz, err := a.Add(c.Amount).Marshal()
		if err != nil {
			return err
		}
		store.Set(denom, bz)
	}
	return nil
}
