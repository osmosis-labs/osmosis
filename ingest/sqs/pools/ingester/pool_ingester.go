package ingester

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type poolIngester struct {
	poolsRepository    domain.PoolsRepository
	gammKeeper         PoolKeeper
	concentratedKeeper PoolKeeper
	cosmWasmKeeper     PoolKeeper
}

// PoolKeeper is an interface for getting pools from a keeper.
type PoolKeeper interface {
	GetPools(ctx sdk.Context) ([]poolmanagertypes.PoolI, error)
}

// NewPoolIngester returns a new pool ingester.
func NewPoolIngester(poolsRepository domain.PoolsRepository, gammKeeper PoolKeeper, concentratedKeeper PoolKeeper, cosmwasmKeeper PoolKeeper) ingest.Ingester {
	return &poolIngester{
		poolsRepository:    poolsRepository,
		gammKeeper:         gammKeeper,
		concentratedKeeper: concentratedKeeper,
		cosmWasmKeeper:     cosmwasmKeeper,
	}
}

// ProcessBlock implements ingest.Ingester.
func (pi *poolIngester) ProcessBlock(ctx sdk.Context) error {
	return pi.updatePoolState(ctx)
}

var _ ingest.Ingester = &poolIngester{}

func (pi *poolIngester) updatePoolState(ctx sdk.Context) error {

	goCtx := sdk.WrapSDKContext(ctx)

	// CFMM pools

	cfmmPools, err := pi.gammKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// TODO: parse pools to the appropriate SQS types
	cfmmPoolsParsed := make([]domain.CFMMPoolI, 0, len(cfmmPools))
	for _, pool := range cfmmPools {
		cfmmPoolsParsed = append(cfmmPoolsParsed, pool)
	}

	err = pi.poolsRepository.StoreCFMM(goCtx, cfmmPoolsParsed)
	if err != nil {
		return err
	}

	// Concentrated pools

	concentratedPools, err := pi.concentratedKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// TODO: parse pools to the appropriate SQS types
	concentratedPoolsParsed := make([]domain.ConcentratedPoolI, 0, len(cfmmPools))
	for _, pool := range concentratedPools {
		concentratedPoolsParsed = append(concentratedPoolsParsed, pool)
	}

	err = pi.poolsRepository.StoreConcentrated(goCtx, concentratedPoolsParsed)
	if err != nil {
		return err
	}

	// CosmWasm pools

	cosmWasmPools, err := pi.cosmWasmKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// TODO: parse pools to the appropriate SQS types
	cosmWasmPoolsParsed := make([]domain.CosmWasmPoolI, 0, len(cfmmPools))
	for _, pool := range cosmWasmPools {
		cosmWasmPoolsParsed = append(cosmWasmPoolsParsed, pool)
	}

	err = pi.poolsRepository.StoreCosmWasm(goCtx, cosmWasmPoolsParsed)
	if err != nil {
		return err
	}

	return nil
}
