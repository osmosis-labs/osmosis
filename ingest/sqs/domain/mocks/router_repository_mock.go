package mocks

import (
	"context"

	"cosmossdk.io/math"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
)

type RedisRouterRepositoryMock struct {
	TakerFees domain.TakerFeeMap
}

// GetAllTakerFees implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) GetAllTakerFees(ctx context.Context) (domain.TakerFeeMap, error) {
	return r.TakerFees, nil
}

// GetTakerFee implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) GetTakerFee(ctx context.Context, denom0 string, denom1 string) (math.LegacyDec, error) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	return r.TakerFees[domain.DenomPair{Denom0: denom0, Denom1: denom1}], nil
}

// SetTakerFee implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) SetTakerFee(ctx context.Context, tx domain.Tx, denom0 string, denom1 string, takerFee math.LegacyDec) error {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	r.TakerFees[domain.DenomPair{Denom0: denom0, Denom1: denom1}] = takerFee
	return nil
}

var _ domain.RouterRepository = &RedisRouterRepositoryMock{}
