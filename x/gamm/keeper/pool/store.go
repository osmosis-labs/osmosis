package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/c-osmosis/osmosis/x/gamm/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
)

type Store interface {
	GetNextPoolNumber(sdk.Context) uint64

	StorePool(sdk.Context, types.Pool)
	FetchPool(sdk.Context, uint64) (types.Pool, error)
	FetchAllPools(sdk.Context) ([]types.Pool, error)
	DeletePool(sdk.Context, uint64)
}

type poolStore struct {
	cdc      codec.BinaryMarshaler
	storeKey sdk.StoreKey
}

func NewStore(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey) Store {
	return poolStore{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

func (ps poolStore) getStore(ctx sdk.Context) prefix.Store {
	return prefix.NewStore(ctx.KVStore(ps.storeKey), types.PoolPrefix)
}

func (ps poolStore) GetNextPoolNumber(ctx sdk.Context) uint64 {
	var poolNumber uint64
	store := ctx.KVStore(ps.storeKey)

	bz := store.Get(types.GlobalPoolNumber)
	if bz == nil {
		// initialize the account numbers
		poolNumber = 0
	} else {
		val := gogotypes.UInt64Value{}

		err := ps.cdc.UnmarshalBinaryBare(bz, &val)
		if err != nil {
			panic(err)
		}

		poolNumber = val.GetValue()
	}

	bz = ps.cdc.MustMarshalBinaryBare(&gogotypes.UInt64Value{Value: poolNumber + 1})
	store.Set(types.GlobalPoolNumber, bz)

	return poolNumber
}

func (ps poolStore) StorePool(ctx sdk.Context, pool types.Pool) {
	store := ps.getStore(ctx)
	bz := ps.cdc.MustMarshalBinaryBare(&pool)
	store.Set(utils.Uint64ToBytes(pool.Id), bz)
}

func (ps poolStore) FetchPool(ctx sdk.Context, poolId uint64) (types.Pool, error) {
	store := ps.getStore(ctx)
	bz := store.Get(utils.Uint64ToBytes(poolId))
	if bz == nil {
		return types.Pool{}, types.ErrPoolNotFound
	}

	var pool types.Pool
	ps.cdc.MustUnmarshalBinaryBare(bz, &pool)
	return pool, nil
}

func (ps poolStore) FetchAllPools(ctx sdk.Context) ([]types.Pool, error) {
	store := ps.getStore(ctx)
	var pools []types.Pool
	iter := store.Iterator([]byte{0}, []byte{255})
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		if bz == nil {
			return nil, types.ErrPoolNotFound
		}

		var pool types.Pool
		ps.cdc.MustUnmarshalBinaryBare(bz, &pool)

		pools = append(pools, pool)
	}

	if len(pools) == 0 {
		return nil, types.ErrPoolNotFound
	}

	return pools, nil
}

func (ps poolStore) DeletePool(ctx sdk.Context, poolId uint64) {
	store := ps.getStore(ctx)
	store.Delete(utils.Uint64ToBytes(poolId))
}
