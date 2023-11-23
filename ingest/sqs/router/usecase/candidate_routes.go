package usecase

import (
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/pools"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/route"
)

// GetCandidateRoutes returns candidate routes from tokenInDenom to tokenOutDenom using BFS.
func (r Router) GetCandidateRoutes(tokenInDenom, tokenOutDenom string) ([]route.RouteImpl, error) {
	var routes []route.RouteImpl
	var visited = make(map[uint64]bool)

	queue := make([]route.RouteImpl, 0)
	queue = append(queue, route.RouteImpl{Pools: []domain.RoutablePool{}})

	for len(queue) > 0 && len(routes) < r.maxRoutes {
		currentRoute := queue[0]
		queue = queue[1:]

		lastPoolID := uint64(0)
		currenTokenInDenom := tokenInDenom
		if len(currentRoute.Pools) > 0 {
			lastPool := currentRoute.Pools[len(currentRoute.Pools)-1]
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
				if len(currentRoute.Pools) > 0 && denom == tokenInDenom {
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
					newPath := route.RouteImpl{
						Pools: make([]domain.RoutablePool, len(currentRoute.Pools), len(currentRoute.Pools)+1),
					}
					copy(newPath.Pools, currentRoute.Pools)

					takerFee, err := r.takerFeeMap.GetTakerFee(currenTokenInDenom, denom)
					if err != nil {
						return nil, err
					}

					newPath.Pools = append(newPath.Pools, pools.NewRoutablePool(pool, denom, takerFee))

					if len(newPath.Pools) <= r.maxHops {
						if hasTokenOut {
							routes = append(routes, newPath)
							break
						} else {
							queue = append(queue, newPath)
						}
					}
				}
			}
		}

		for _, pool := range currentRoute.Pools {
			visited[pool.GetId()] = true
		}
	}

	return r.validateAndFilterRoutes(routes, tokenInDenom)
}

// Pool represents a pool in the decentralized exchange.
type Pool struct {
	ID       int
	TokenIn  string
	TokenOut string
}
