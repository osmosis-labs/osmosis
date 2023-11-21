package usecase

import (
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/pools"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/route"
)

// GetCandidateRoutesBFS returns candidate routes from tokenInDenom to tokenOutDenom using BFS.
func (r Router) GetCandidateRoutesBFS(tokenInDenom, tokenOutDenom string) ([]domain.Route, error) {
	var routes [][]domain.RoutablePool
	var visited = make(map[uint64]bool)

	queue := make([]Route, 0)
	queue = append(queue, Route{Path: []domain.RoutablePool{}})

	for len(queue) > 0 && len(routes) < r.maxRoutes {
		currentRoute := queue[0]
		queue = queue[1:]

		lastPoolID := uint64(0)
		currenTokenInDenom := tokenInDenom
		if len(currentRoute.Path) > 0 {
			lastPool := currentRoute.Path[len(currentRoute.Path)-1]
			lastPoolID = lastPool.GetId()
			currenTokenInDenom = lastPool.GetTokenOutDenom()
		}

		for i := 0; i < len(r.sortedPools) && len(routes) < r.maxRoutes; i++ {
			pool := r.sortedPools[i]

			if visited[pool.GetId()] {
				continue
			}

			poolDenoms := pool.GetPoolDenoms()
			hasTokenIn := false
			hasTokenOut := false
			shouldSkipPool := false
			for _, denom := range poolDenoms {
				if denom == currenTokenInDenom {
					hasTokenIn = true
				}
				if denom == tokenOutDenom {
					hasTokenOut = true
				}

				// Avoid going through pools that has the initial token in denom twice.
				if len(currentRoute.Path) > 0 && denom == tokenInDenom {
					shouldSkipPool = true
					break
				}
			}

			if shouldSkipPool {
				continue
			}

			if !hasTokenIn {
				continue
			}

			currentPoolID := pool.GetId()
			for _, denom := range poolDenoms {
				if denom == currenTokenInDenom {
					continue
				}
				if hasTokenOut && denom != tokenOutDenom {
					continue
				}

				if lastPoolID == uint64(0) || lastPoolID != currentPoolID {
					newPath := append([]domain.RoutablePool{}, currentRoute.Path...)

					takerFee, err := r.takerFeeMap.GetTakerFee(currenTokenInDenom, denom)
					if err != nil {
						return nil, err
					}

					newPath = append(newPath, pools.NewRoutablePool(pool, denom, takerFee))

					if len(newPath) <= r.maxHops {
						if hasTokenOut {
							routes = append(routes, newPath)
							break
						} else {
							queue = append(queue, Route{Path: newPath})
						}
					}
				}
			}
		}

		for _, pool := range currentRoute.Path {
			visited[pool.GetId()] = true
		}
	}

	result := make([]domain.Route, 0, len(routes))
	for _, currentRoute := range routes {
		result = append(result, &route.RouteImpl{Pools: currentRoute})
	}

	return result, nil
}

// Pool represents a pool in the decentralized exchange.
type Pool struct {
	ID       int
	TokenIn  string
	TokenOut string
}

// Route represents a route between token-in and token-out denominations.
type Route struct {
	Path []domain.RoutablePool // IDs of pools in the route
}
