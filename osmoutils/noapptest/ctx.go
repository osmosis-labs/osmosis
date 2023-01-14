package noapptest

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	dbm "github.com/tendermint/tm-db"
)

func CtxWithStoreKeys(keys []sdk.StoreKey, header tmproto.Header, isCheckTx bool) sdk.Context {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	cms := store.NewCommitMultiStore(db, logger)
	for _, key := range keys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	return sdk.NewContext(cms, header, isCheckTx, logger)
}

func DefaultCtxWithStoreKeys(storeKeys []sdk.StoreKey) sdk.Context {
	header := tmproto.Header{Height: 1, ChainID: "osmoutils-test-1", Time: time.Now().UTC()}
	deliverTx := false
	return CtxWithStoreKeys(storeKeys, header, deliverTx)
}
