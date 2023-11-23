package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
)

type RedisPoolsRepositoryMock struct {
	cfmmPools         []domain.PoolI
	concentratedPools []domain.PoolI
	cosmwasmPools     []domain.PoolI
}

// ClearAllPools implements domain.PoolsRepository.
func (*RedisPoolsRepositoryMock) ClearAllPools(ctx context.Context, tx mvc.Tx) error {
	panic("unimplemented")
}

var _ mvc.PoolsRepository = &RedisPoolsRepositoryMock{}

// GetAllCFMM implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetAllCFMM(context.Context) ([]domain.PoolI, error) {
	return r.cfmmPools, nil
}

// GetAllConcentrated implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetAllConcentrated(context.Context) ([]domain.PoolI, error) {
	return r.concentratedPools, nil
}

// GetAllCosmWasm implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetAllCosmWasm(context.Context) ([]domain.PoolI, error) {
	return r.cosmwasmPools, nil
}

// GetAllPools implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetAllPools(context.Context) ([]domain.PoolI, error) {
	allPools := make([]domain.PoolI, 0, len(r.cfmmPools)+len(r.concentratedPools)+len(r.cosmwasmPools))
	allPools = append(allPools, r.cfmmPools...)
	allPools = append(allPools, r.concentratedPools...)
	allPools = append(allPools, r.cosmwasmPools...)
	return allPools, nil
}

// StorePools implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) StorePools(ctx context.Context, tx mvc.Tx, cfmmPools []domain.PoolI, concentratedPools []domain.PoolI, cosmwasmPools []domain.PoolI) error {
	r.cfmmPools = cfmmPools
	r.concentratedPools = concentratedPools
	r.cosmwasmPools = cosmwasmPools
	return nil
}
