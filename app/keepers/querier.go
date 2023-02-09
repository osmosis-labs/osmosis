package keepers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// QuerierWrapper is a local wrapper around BaseApp that exports only the Queryable interface.
// This is used to pass the baseApp to Async ICQ without exposing all methods
type QuerierWrapper struct {
	querier sdk.Queryable
}

var _ sdk.Queryable = QuerierWrapper{}

func NewQuerierWrapper(querier sdk.Queryable) QuerierWrapper {
	return QuerierWrapper{querier: querier}
}

func (q QuerierWrapper) Query(req abci.RequestQuery) abci.ResponseQuery {
	return q.querier.Query(req)
}
