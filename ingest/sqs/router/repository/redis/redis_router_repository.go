package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/json"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
)

type redisRouterRepo struct {
	repositoryManager        mvc.TxManager
	routerCacheExpirySeconds uint64
}

const (
	keySeparator = "-"

	routerPrefix   = "r" + keySeparator
	takerFeePrefix = routerPrefix + "tf" + keySeparator
	routesPrefix   = routerPrefix + "r" + keySeparator
)

var (
	_ mvc.RouterRepository = &redisRouterRepo{}
)

// NewRedisRouterRepo will create an implementation of pools.Repository
func NewRedisRouterRepo(repositoryManager mvc.TxManager, routesCacheExpirySeconds uint64) mvc.RouterRepository {
	return &redisRouterRepo{
		repositoryManager:        repositoryManager,
		routerCacheExpirySeconds: routesCacheExpirySeconds,
	}
}

// GetAllTakerFees implements mvc.RouterRepository.
func (r *redisRouterRepo) GetAllTakerFees(ctx context.Context) (domain.TakerFeeMap, error) {
	tx := r.repositoryManager.StartTx()

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return nil, err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return nil, err
	}

	result := pipeliner.HGetAll(ctx, takerFeePrefix)

	_, err = pipeliner.Exec(ctx)
	if err != nil {
		return nil, err
	}

	resultMap, err := result.Result()
	if err != nil {
		return nil, err
	}

	// Parse taker fee map
	takerFeeMap := make(domain.TakerFeeMap, len(resultMap))
	for denomPairStr, takerFeeStr := range resultMap {
		takerFee, err := osmomath.NewDecFromStr(takerFeeStr)
		if err != nil {
			return nil, err
		}

		denoms := strings.Split(denomPairStr, keySeparator)

		if len(denoms) != 2 {
			return nil, fmt.Errorf("invalid denom pair string key %s. must have 2 denoms, had (%d)", denomPairStr, len(denoms))
		}

		if denoms[0] > denoms[1] {
			return nil, fmt.Errorf("invalid denom pair string key %s. must be in increasing lexicographic order", denomPairStr)
		}

		takerFeeMap[domain.DenomPair{
			Denom0: denoms[0],
			Denom1: denoms[1],
		}] = takerFee
	}

	return takerFeeMap, nil
}

// GetTakerFee implements mvc.RouterRepository.
func (r *redisRouterRepo) GetTakerFee(ctx context.Context, denom0 string, denom1 string) (osmomath.Dec, error) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	tx := r.repositoryManager.StartTx()

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return osmomath.Dec{}, err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return osmomath.Dec{}, err
	}

	result := pipeliner.HGet(ctx, takerFeePrefix, denom0+keySeparator+denom1)

	_, err = pipeliner.Exec(ctx)
	if err != nil {
		return osmomath.Dec{}, err
	}

	resultStr, err := result.Result()
	if err != nil {
		return osmomath.Dec{}, err
	}

	return osmomath.NewDecFromStr(resultStr)
}

// SetTakerFee sets taker fee for a denom pair.
func (r *redisRouterRepo) SetTakerFee(ctx context.Context, tx mvc.Tx, denom0, denom1 string, takerFee osmomath.Dec) error {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return err
	}
	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return err
	}

	cmd := pipeliner.HSet(ctx, takerFeePrefix, denom0+keySeparator+denom1, takerFee.String())
	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

// SetRoutesTx implements mvc.RouterRepository.
func (r *redisRouterRepo) SetRoutesTx(ctx context.Context, tx mvc.Tx, denom0, denom1 string, routes route.CandidateRoutes) error {
	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return err
	}
	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return err
	}

	routesStr, err := json.Marshal(routes)
	if err != nil {
		return err
	}

	routeCacheExpiryDuration := time.Second * time.Duration(r.routerCacheExpirySeconds)

	cmd := pipeliner.Set(ctx, getRoutesPrefixByDenoms(denom0, denom1), routesStr, routeCacheExpiryDuration)
	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

// SetRoutes implements mvc.RouterRepository.
func (r *redisRouterRepo) SetRoutes(ctx context.Context, denom0, denom1 string, routes route.CandidateRoutes) error {
	// Create transaction
	tx := r.repositoryManager.StartTx()

	// Set routes
	if err := r.SetRoutesTx(ctx, tx, denom0, denom1, routes); err != nil {
		return err
	}

	// Execute transaction.
	if err := tx.Exec(ctx); err != nil {
		return err
	}

	return nil
}

// GetRoutes implements mvc.RouterRepository.
func (r *redisRouterRepo) GetRoutes(ctx context.Context, denom0, denom1 string) (route.CandidateRoutes, error) {
	// Create transaction
	tx := r.repositoryManager.StartTx()

	redisTx, err := tx.AsRedisTx()
	if err != nil {
		return route.CandidateRoutes{}, err
	}

	pipeliner, err := redisTx.GetPipeliner(ctx)
	if err != nil {
		return route.CandidateRoutes{}, err
	}

	// Create command to retrieve results.
	getCmd := pipeliner.Get(ctx, getRoutesPrefixByDenoms(denom0, denom1))

	_, err = pipeliner.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return route.CandidateRoutes{}, nil
		}
		return route.CandidateRoutes{}, err
	}

	// Parse routes
	var routes route.CandidateRoutes
	err = json.Unmarshal([]byte(getCmd.Val()), &routes)
	if err != nil {
		return route.CandidateRoutes{}, err
	}

	return routes, nil
}

func getRoutesPrefixByDenoms(denom0, denom1 string) string {
	return routesPrefix + denom0 + keySeparator + denom1
}
