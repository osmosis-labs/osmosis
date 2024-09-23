package keepers

import (
	storetypes "cosmossdk.io/store/types"
)

// QuerierWrapper is a local wrapper around BaseApp that exports only the Queryable interface.
// This is used to pass the baseApp to Async ICQ without exposing all methods
type QuerierWrapper struct {
	querier storetypes.Queryable
}

var _ storetypes.Queryable = QuerierWrapper{}

func NewQuerierWrapper(querier storetypes.Queryable) QuerierWrapper {
	return QuerierWrapper{querier: querier}
}

func (q QuerierWrapper) Query(req *storetypes.RequestQuery) (*storetypes.ResponseQuery, error) {
	return q.querier.Query(req)
}
