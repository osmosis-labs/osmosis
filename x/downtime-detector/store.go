package downtimedetector

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
)

func (k *Keeper) GetLastBlockTime(ctx sdk.Context) (time.Time, error) {
	store := ctx.KVStore(k.storeKey)
	timeBz := store.Get(types.GetLastBlockTimestampKey())
	if len(timeBz) == 0 {
		return time.Time{}, errors.New("no last block time stored in state. Should not happen, did initialization happen correctly?")
	}
	timeV, err := osmoutils.ParseTimeString(string(timeBz))
	if err != nil {
		return time.Time{}, err
	}
	return timeV, nil
}

func (k *Keeper) StoreLastBlockTime(ctx sdk.Context, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	timeBz := osmoutils.FormatTimeString(t)
	store.Set(types.GetLastBlockTimestampKey(), []byte(timeBz))
}

func (k *Keeper) GetLastDowntimeOfLength(ctx sdk.Context, dur types.Downtime) (time.Time, error) {
	store := ctx.KVStore(k.storeKey)
	timeBz := store.Get(types.GetLastDowntimeOfLengthKey(dur))
	if len(timeBz) == 0 {
		return time.Time{}, errors.New("no last time stored in state. Should not happen, did initialization happen correctly?")
	}
	timeV, err := osmoutils.ParseTimeString(string(timeBz))
	if err != nil {
		return time.Time{}, err
	}
	return timeV, nil
}

func (k *Keeper) StoreLastDowntimeOfLength(ctx sdk.Context, dur types.Downtime, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	timeBz := osmoutils.FormatTimeString(t)
	store.Set(types.GetLastDowntimeOfLengthKey(dur), []byte(timeBz))
}
