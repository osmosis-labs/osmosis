package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState types.GenesisState) {
	// TODO setSales, setNextSaleNumber
	//TODO  k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the launchpad module's exported genesis.
func (k Keeper) ExportGenesis(sdkCtx sdk.Context) (*types.GenesisState, error) {
	store := k.saleStore(sdkCtx.KVStore(k.storeKey))
	sales, err := k.exportSales(store)
	if err != nil {
		return nil, err
	}
	ups, err := k.exportUserPositions(store)
	if err != nil {
		return nil, err
	}
	nextSaleId, _ := getNextSaleID(store)
	return &types.GenesisState{
		Sales:         sales,
		UserPositions: ups,
		NextSaleId:    nextSaleId,
		Params:        k.GetParams(sdkCtx),
	}, nil
}

func (k Keeper) exportSales(moduleStore prefix.Store) ([]types.Sale, error) {
	var res = []types.Sale{}
	iter := sdk.KVStorePrefixIterator(moduleStore, storeStoreKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var o types.Sale
		err := k.cdc.Unmarshal(iter.Value(), &o)
		if err != nil {
			return nil, err
		}
		res = append(res, o)
	}

	return res, nil
}

func (k Keeper) exportUserPositions(moduleStore prefix.Store) ([]types.UserPositionKV, error) {
	var res = []types.UserPositionKV{}
	iter := sdk.KVStorePrefixIterator(moduleStore, storeStoreKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key() // <big endian uint64 sale_id><acc_addr>
		sale := binary.BigEndian.Uint64(key[:8])
		addr := sdk.AccAddress(key[8:])
		var o = types.UserPositionKV{AccAddress: addr.String(), SaleId: sale}
		err := k.cdc.Unmarshal(iter.Value(), &o.U)
		if err != nil {
			return nil, err
		}
		res = append(res, o)
	}

	return res, nil
}
