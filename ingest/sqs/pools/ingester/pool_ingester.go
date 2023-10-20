package ingester

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/common"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/ingester/parser"
)

type poolIngester struct {
	poolsRepository    domain.PoolsRepository
	gammKeeper         common.PoolKeeper
	concentratedKeeper common.PoolKeeper
	cosmWasmKeeper     common.CosmWasmPoolKeeper
	bankKeeper         common.BankKeeper
}

// NewPoolIngester returns a new pool ingester.
func NewPoolIngester(poolsRepository domain.PoolsRepository, gammKeeper common.PoolKeeper, concentratedKeeper common.PoolKeeper, cosmwasmKeeper common.CosmWasmPoolKeeper, bankKeeper common.BankKeeper) ingest.Ingester {
	return &poolIngester{
		poolsRepository:    poolsRepository,
		gammKeeper:         gammKeeper,
		concentratedKeeper: concentratedKeeper,
		cosmWasmKeeper:     cosmwasmKeeper,
		bankKeeper:         bankKeeper,
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

	// Parse CFMM pool to the standard SQS types.
	cfmmPoolsParsed := make([]domain.PoolI, 0, len(cfmmPools))
	for _, pool := range cfmmPools {
		pool, err := parser.ConvertCFMM(ctx, pool, pi.bankKeeper)
		if err != nil {
			return err
		}

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

	concentratedPoolsParsed := make([]domain.PoolI, 0, len(concentratedPools))
	for _, pool := range concentratedPools {
		// Parse concentrated pool to the standard SQS types.
		parsedPool, err := parser.ConvertConcentrated(ctx, pool, pi.bankKeeper)
		if err != nil {
			return err
		}

		concentratedPoolsParsed = append(concentratedPoolsParsed, parsedPool)
	}

	err = pi.poolsRepository.StoreConcentrated(goCtx, concentratedPoolsParsed)
	if err != nil {
		return err
	}

	// CosmWasm pools

	cosmWasmPools, err := pi.cosmWasmKeeper.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return err
	}

	cosmWasmPoolsParsed := make([]domain.PoolI, 0, len(cosmWasmPools))
	for _, pool := range cosmWasmPools {
		// Parse CosmWasm pools to the standard SQS types.
		pool, err := parser.ConvertCosmWasm(ctx, pool)
		if err != nil {
			return err
		}

		cosmWasmPoolsParsed = append(cosmWasmPoolsParsed, pool)
	}

	err = pi.poolsRepository.StoreCosmWasm(goCtx, cosmWasmPoolsParsed)
	if err != nil {
		return err
	}

	return nil
}
