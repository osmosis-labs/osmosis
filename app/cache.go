package app

import (
	"encoding/hex"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/patrickmn/go-cache"
)

var (
	cacheMap = map[string]time.Duration{
		// Querying gov tally calculates the all voting of the validators and delegators.
		// It can take a time 20sec ~ 30sec.
		// Correct result is less important compared to how long this query takes.
		// Therefore, to reduce resource usage, cache for 60 minutes.
		"custom/gov/tally": 60 * time.Minute,
	}

	cacher = cache.New(3*time.Second, 1*time.Minute)
)

func GetCacheKey(req abci.RequestQuery) string {
	if _, ok := cacheMap[req.Path]; !ok {
		return ""
	}

	cacheKey := req.Path
	sortedBz, err := sdk.SortJSON(req.Data)
	if err == nil {
		// Case of the custom querier
		cacheKey += hex.EncodeToString(sortedBz)
	} else {
		// Case of the grpc
		cacheKey += hex.EncodeToString(req.Data)
	}

	return cacheKey
}

func GetCachedValue(cacheKey string) (res abci.ResponseQuery, found bool) {
	if len(cacheKey) == 0 {
		return abci.ResponseQuery{}, false
	}

	cached, found := cacher.Get(cacheKey)
	if found {
		res, ok := cached.(abci.ResponseQuery)
		if ok {
			return res, true
		}
	}
	return abci.ResponseQuery{}, false
}

func SetCache(path string, cacheKey string, res abci.ResponseQuery) {
	if len(cacheKey) == 0 {
		return
	}

	expiration, ok := cacheMap[path]
	if !ok {
		return
	}

	cacher.Set(cacheKey, res, expiration)
}