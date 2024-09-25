package noapptest

import (
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
)

func CtxWithStoreKeys(keys []storetypes.StoreKey, header tmproto.Header, isCheckTx bool) sdk.Context {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	cms := store.NewCommitMultiStore(db, logger, storemetrics.NewNoOpMetrics())
	for _, key := range keys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	}
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	return sdk.NewContext(cms, header, isCheckTx, logger)
}

func DefaultCtxWithStoreKeys(storeKeys []storetypes.StoreKey) sdk.Context {
	header := tmproto.Header{Height: 1, ChainID: "osmoutils-test-1", Time: time.Now().UTC()}
	deliverTx := false
	return CtxWithStoreKeys(storeKeys, header, deliverTx)
}
