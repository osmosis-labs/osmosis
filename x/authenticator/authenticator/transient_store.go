package authenticator

import (
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TransientStore struct {
	storeKey     storetypes.StoreKey
	transientCtx sdk.Context
}

func NewTransientStore(storeKey storetypes.StoreKey, ctx sdk.Context) *TransientStore {
	return &TransientStore{
		storeKey:     storeKey,
		transientCtx: ctx,
	}
}

func (as *TransientStore) ResetTransientContext(ctx sdk.Context) sdk.Context {
	as.transientCtx, _ = ctx.CacheContext()
	return as.transientCtx
}

func (as *TransientStore) GetKvStore() store.KVStore {
	return as.transientCtx.KVStore(as.storeKey)
}

func (as *TransientStore) GetTransientContext() sdk.Context {
	return as.transientCtx
}

func (as *TransientStore) GetTransientContextWithGasMeter(gasMeter sdk.GasMeter) sdk.Context {
	as.transientCtx = as.transientCtx.WithGasMeter(gasMeter)
	return as.transientCtx
}

func (as *TransientStore) WriteInto(ctx sdk.Context) {
	if as.transientCtx.IsZero() {
		panic("Transient context not set")
	}
	srcStore := as.transientCtx.KVStore(as.storeKey)
	destStore := ctx.KVStore(as.storeKey)
	syncStores(srcStore, destStore, true)
}

func (as *TransientStore) WriteCosmWasmAuthenticatorStateInto(ctx sdk.Context, cwa *CosmwasmAuthenticator) {
	srcStore := cwa.GetContractPrefixStore(as.transientCtx)
	destStore := cwa.GetContractPrefixStore(ctx)
	syncStores(srcStore, destStore, true)
}

func (as *TransientStore) UpdateFrom(ctx sdk.Context) {
	if as.transientCtx.IsZero() {
		as.ResetTransientContext(ctx)
	}
	srcStore := ctx.KVStore(as.storeKey)
	destStore := as.transientCtx.KVStore(as.storeKey)
	syncStores(srcStore, destStore, true)
}

func syncStores(srcStore, destStore sdk.KVStore, clearDest bool) {
	// TODO: is there a cleaner way to do this? Ideally we'd just replace the entire store
	if clearDest {
		iterDest := destStore.Iterator(nil, nil)
		defer iterDest.Close()
		for ; iterDest.Valid(); iterDest.Next() {
			destStore.Delete(iterDest.Key())
		}
	}

	iterSrc := srcStore.Iterator(nil, nil)
	defer iterSrc.Close()
	for ; iterSrc.Valid(); iterSrc.Next() {
		destStore.Set(iterSrc.Key(), iterSrc.Value())
	}
}
