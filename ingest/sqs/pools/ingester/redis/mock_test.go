package redis_test

import (
	"context"

	"cosmossdk.io/math"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type redisPoolRepositoryMock struct {
	cfmmPools         []domain.PoolI
	concentratedPools []domain.PoolI
	cosmwasmPools     []domain.PoolI
}

var _ domain.PoolsRepository = &redisPoolRepositoryMock{}

// GetAllCFMM implements domain.PoolsRepository.
func (r *redisPoolRepositoryMock) GetAllCFMM(context.Context) ([]domain.PoolI, error) {
	return r.cfmmPools, nil
}

// GetAllConcentrated implements domain.PoolsRepository.
func (r *redisPoolRepositoryMock) GetAllConcentrated(context.Context) ([]domain.PoolI, error) {
	return r.concentratedPools, nil
}

// GetAllCosmWasm implements domain.PoolsRepository.
func (r *redisPoolRepositoryMock) GetAllCosmWasm(context.Context) ([]domain.PoolI, error) {
	return r.cosmwasmPools, nil
}

// GetAllPools implements domain.PoolsRepository.
func (r *redisPoolRepositoryMock) GetAllPools(context.Context) ([]domain.PoolI, error) {
	allPools := make([]domain.PoolI, 0, len(r.cfmmPools)+len(r.concentratedPools)+len(r.cosmwasmPools))
	allPools = append(allPools, r.cfmmPools...)
	allPools = append(allPools, r.concentratedPools...)
	allPools = append(allPools, r.cosmwasmPools...)
	return allPools, nil
}

// StorePools implements domain.PoolsRepository.
func (r *redisPoolRepositoryMock) StorePools(ctx context.Context, tx domain.Tx, cfmmPools []domain.PoolI, concentratedPools []domain.PoolI, cosmwasmPools []domain.PoolI) error {
	r.cfmmPools = cfmmPools
	r.concentratedPools = concentratedPools
	r.cosmwasmPools = cosmwasmPools
	return nil
}

type redisRouterRepositoryMock struct {
	takerFeeMap domain.TakerFeeMap
}

// GetAllTakerFees implements domain.RouterRepository.
func (r *redisRouterRepositoryMock) GetAllTakerFees(ctx context.Context) (domain.TakerFeeMap, error) {
	return r.takerFeeMap, nil
}

// GetTakerFee implements domain.RouterRepository.
func (r *redisRouterRepositoryMock) GetTakerFee(ctx context.Context, denom0 string, denom1 string) (math.LegacyDec, error) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	return r.takerFeeMap[domain.DenomPair{Denom0: denom0, Denom1: denom1}], nil
}

// SetTakerFee implements domain.RouterRepository.
func (r *redisRouterRepositoryMock) SetTakerFee(ctx context.Context, tx domain.Tx, denom0 string, denom1 string, takerFee math.LegacyDec) error {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	r.takerFeeMap[domain.DenomPair{Denom0: denom0, Denom1: denom1}] = takerFee
	return nil
}

var _ domain.RouterRepository = &redisRouterRepositoryMock{}
