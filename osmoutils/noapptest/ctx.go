package noapptest

import (
	"time"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dbm "github.com/cometbft/cometbft-db"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

func CtxWithStoreKeys(keys []storetypes.StoreKey, header tmproto.Header, isCheckTx bool) sdk.Context {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	cms := store.NewCommitMultiStore(db)
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
