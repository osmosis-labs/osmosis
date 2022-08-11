package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/osmoutils"
	"github.com/osmosis-labs/osmosis/v10/x/twap/types"
)

// GetAllHistoricalTimeIndexedTWAPs returns all historical TWAPs indexed by time.
func (k Keeper) GetAllHistoricalTimeIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStore(ctx.KVStore(k.storeKey), []byte(types.HistoricalTWAPTimeIndexPrefix), []byte(types.HistoricalTWAPTimeIndexPrefix+"a"), types.ParseTwapFromBz)
}

// GetAllHistoricalPoolIndexedTWAPs returns all historical TWAPs indexed by pool.
func (k Keeper) GetAllHistoricalPoolIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStore(ctx.KVStore(k.storeKey), []byte(types.HistoricalTWAPPoolIndexPrefix), []byte(types.HistoricalTWAPPoolIndexPrefix+"a"), types.ParseTwapFromBz)
}
