package keeper

import (
	"encoding/binary"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v10/x/launchpad/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState types.GenesisState) {
	store := k.saleStore(ctx.KVStore(k.storeKey))
	if err := k.importSales(store, genState.Sales); err != nil {
		panic(err)
	}
	k.setNextSaleID(store, genState.NextSaleId)
	if err := k.importUserPositions(store, genState.UserPositions); err != nil {
		panic(err)
	}
	k.SetParams(ctx, genState.Params)
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

func (k Keeper) importSales(moduleStore prefix.Store, sales []types.Sale) error {
	for _, sale := range sales {
		idBz := storeIntIdKey(sale.Id)
		// TODO: do we need any validation here ?
		k.saveSale(moduleStore, idBz, &sale)
	}
	return nil
}

func (k Keeper) importUserPositions(moduleStore prefix.Store, ups []types.UserPositionKV) error {
	for idx, up := range ups {
		idBz := storeIntIdKey(up.SaleId)
		address, err := sdk.AccAddressFromBech32(up.AccAddress)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress,
				"invalid account address %s in user positions at index %d", up.AccAddress, idx)
		}
		k.saveUserPosition(moduleStore, idBz, address, &up.U)
	}
	return nil
}
