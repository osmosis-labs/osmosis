package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"
	// this line is used by starport scaffolding # ibc/keeper/import
)

// GetFeeToken returns the fee token record for a specific denom
func (k Keeper) GetNativeHRP(ctx sdk.Context) (hrp string, err error) {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.NativeHrpKey) {
		return "", types.ErrNoNativeHrp
	}

	bz := store.Get(types.NativeHrpKey)

	return string(bz), nil
}

// setNativeHrp sets the native prefix for the chain. Should only be used once.
func (k Keeper) setNativeHrp(ctx sdk.Context, hrp string) error {
	store := ctx.KVStore(k.storeKey)

	err := types.ValidateHRP(hrp)
	if err != nil {
		return err
	}

	store.Set(types.NativeHrpKey, []byte(hrp))
	return nil
}

// ValidateFeeToken validates that a fee token record is valid
// It checks:
// - The HRP is valid
// - The HRP is not for the chain's native prefix
// - Check that IBC channels and ports are real
func (k Keeper) ValidateHrpIbcRecord(ctx sdk.Context, record types.HrpIbcRecord) error {
	err := types.ValidateHRP(record.HRP)
	if err != nil {
		return err
	}

	nativeHrp, err := k.GetNativeHRP(ctx)
	if err != nil {
		return err
	}

	if record.HRP == nativeHrp {
		return sdkerrors.Wrap(types.ErrInvalidHRP, "cannot set a record for the chain's native prefix")
	}

	//TODO: Validate IBC channel ID exists
	return nil
}

// GetHrpIbcRecord returns the hrp ibc record for a specific hrp
func (k Keeper) GetHrpSourceChannel(ctx sdk.Context, hrp string) (string, error) {
	record, err := k.GetHrpIbcRecord(ctx, hrp)
	if err != nil {
		return "", nil
	}

	return record.SourceChannel, nil
}

// GetHrpIbcRecord returns the hrp ibc record for a specific hrp
func (k Keeper) GetHrpIbcRecord(ctx sdk.Context, hrp string) (types.HrpIbcRecord, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.HrpIBCRecordStorePrefix)
	if !prefixStore.Has([]byte(hrp)) {
		return types.HrpIbcRecord{}, types.ErrRecordNotFound
	}
	bz := prefixStore.Get([]byte(hrp))

	record := types.HrpIbcRecord{}
	err := proto.Unmarshal(bz, &record)
	if err != nil {
		return types.HrpIbcRecord{}, err
	}

	return record, nil
}

// setHrpIbcRecord sets a new hrp ibc record for a specific denom
func (k Keeper) setHrpIbcRecord(ctx sdk.Context, hrpIbcRecord types.HrpIbcRecord) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.HrpIBCRecordStorePrefix)

	if hrpIbcRecord.SourceChannel == "" {
		if prefixStore.Has([]byte(hrpIbcRecord.HRP)) {
			prefixStore.Delete([]byte(hrpIbcRecord.HRP))
		}
		return nil
	}

	bz, err := proto.Marshal(&hrpIbcRecord)
	if err != nil {
		return err
	}

	prefixStore.Set([]byte(hrpIbcRecord.HRP), bz)
	return nil
}

func (k Keeper) GetHrpIbcRecords(ctx sdk.Context) (HrpIbcRecords []types.HrpIbcRecord) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.HrpIBCRecordStorePrefix)

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	records := []types.HrpIbcRecord{}

	for ; iterator.Valid(); iterator.Next() {

		record := types.HrpIbcRecord{}

		err := proto.Unmarshal(iterator.Value(), &record)
		if err != nil {
			panic(err)
		}

		records = append(records, record)
	}
	return records
}

func (k Keeper) setHrpIbcRecords(ctx sdk.Context, hrpIbcRecords []types.HrpIbcRecord) {
	for _, record := range hrpIbcRecords {
		k.setHrpIbcRecord(ctx, record)
	}
}
