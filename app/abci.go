package app

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

// Query implements the ABCI interface. It delegates to CommitMultiStore if it
// implements Queryable.
func (app *OsmosisApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	cacheKey := GetCacheKey(req)
	if len(cacheKey) > 0 {
		cached, found := GetCachedValue(cacheKey)
		if found {
			return cached
		}
	}

	res = app.BaseApp.Query(req)

	if len(cacheKey) > 0 {
		SetCache(req.Path, cacheKey, res)
	}

	return res
}