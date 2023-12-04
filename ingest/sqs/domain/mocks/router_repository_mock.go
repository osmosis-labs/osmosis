package mocks

import (
	"context"

	"cosmossdk.io/math"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
)

type RedisRouterRepositoryMock struct {
	TakerFees domain.TakerFeeMap
	Routes    map[domain.DenomPair]route.CandidateRoutes
}

// GetAllTakerFees implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) GetAllTakerFees(ctx context.Context) (domain.TakerFeeMap, error) {
	return r.TakerFees, nil
}

// GetRoutes implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) GetRoutes(ctx context.Context, denom0 string, denom1 string) (route.CandidateRoutes, error) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	routes := r.Routes[domain.DenomPair{Denom0: denom0, Denom1: denom1}]
	return routes, nil
}

// GetTakerFee implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) GetTakerFee(ctx context.Context, denom0 string, denom1 string) (math.LegacyDec, error) {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	return r.TakerFees[domain.DenomPair{Denom0: denom0, Denom1: denom1}], nil
}

// SetRoutes implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) SetRoutes(ctx context.Context, denom0 string, denom1 string, routes route.CandidateRoutes) error {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	r.Routes[domain.DenomPair{Denom0: denom0, Denom1: denom1}] = routes
	return nil
}

// SetRoutesTx implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) SetRoutesTx(ctx context.Context, tx mvc.Tx, denom0 string, denom1 string, routes route.CandidateRoutes) error {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	r.Routes[domain.DenomPair{Denom0: denom0, Denom1: denom1}] = routes
	return nil
}

// SetTakerFee implements domain.RouterRepository.
func (r *RedisRouterRepositoryMock) SetTakerFee(ctx context.Context, tx mvc.Tx, denom0 string, denom1 string, takerFee math.LegacyDec) error {
	// Ensure increasing lexicographic order.
	if denom1 < denom0 {
		denom0, denom1 = denom1, denom0
	}

	r.TakerFees[domain.DenomPair{Denom0: denom0, Denom1: denom1}] = takerFee
	return nil
}

var _ mvc.RouterRepository = &RedisRouterRepositoryMock{}
