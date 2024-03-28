package sqs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/sqsdomain/repository"
	chaininforedisrepo "github.com/osmosis-labs/sqs/sqsdomain/repository/redis/chaininfo"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
)

var _ domain.Ingester = &sqsIngester{}

// sqsIngester is a sidecar query server (SQS) implementation of Ingester.
// It encapsulates all individual SQS ingesters.
type sqsIngester struct {
	txManager     repository.TxManager
	poolsIngester domain.PoolIngester
	chainInfoRepo chaininforedisrepo.ChainInfoRepository
	keepers       domain.SQSIngestKeepers
}

// NewSidecarQueryServerIngester creates a new sidecar query server ingester.
// poolsRepository is the storage for pools.
// gammKeeper is the keeper for Gamm pools.
func NewSidecarQueryServerIngester(poolsIngester domain.PoolIngester, chainInfoIngester chaininforedisrepo.ChainInfoRepository, txManager repository.TxManager, keepers domain.SQSIngestKeepers) domain.Ingester {
	return &sqsIngester{
		txManager:     txManager,
		chainInfoRepo: chainInfoIngester,
		poolsIngester: poolsIngester,
		keepers:       keepers,
	}
}

// ProcessAllBlockData implements ingest.Ingester.
func (i *sqsIngester) ProcessAllBlockData(ctx sdk.Context) error {
	// Start atomic transaction
	tx := i.txManager.StartTx()

	// Concentrated pools

	concentratedPools, err := i.keepers.ConcentratedKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// CFMM pools

	cfmmPools, err := i.keepers.GammKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// CosmWasm pools

	cosmWasmPools, err := i.keepers.CosmWasmPoolKeeper.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return err
	}

	blockPools := domain.BlockPools{
		ConcentratedPools: concentratedPools,
		CosmWasmPools:     cosmWasmPools,
		CFMMPools:         cfmmPools,
	}

	// Process block by reading and writing data and ingesting data into sinks
	if err := i.poolsIngester.ProcessPoolState(ctx, tx, blockPools); err != nil {
		return err
	}

	return i.storeLatestHeight(ctx, tx)
}

// ProcessChangedBlockData implements ingest.Ingester.
func (i *sqsIngester) ProcessChangedBlockData(ctx sdk.Context, changedPools domain.BlockPools) error {
	// Start atomic transaction
	tx := i.txManager.StartTx()

	concentratedPoolIDTickChange := changedPools.ConcentratedPoolIDTickChange

	// Copy over the pools that were changed in the block
	for _, pool := range changedPools.ConcentratedPools {
		changedPools.ConcentratedPoolIDTickChange[pool.GetId()] = struct{}{}
	}

	// Update concentrated pools
	for poolID := range concentratedPoolIDTickChange {
		// Skip if the pool if it is already tracked
		if _, ok := changedPools.ConcentratedPoolIDTickChange[poolID]; ok {
			continue
		}

		pool, err := i.keepers.ConcentratedKeeper.GetConcentratedPoolById(ctx, poolID)
		if err != nil {
			return err
		}

		changedPools.ConcentratedPools = append(changedPools.ConcentratedPools, pool)

		changedPools.ConcentratedPoolIDTickChange[poolID] = struct{}{}
	}

	// Process block by reading and writing data and ingesting data into sinks
	if err := i.poolsIngester.ProcessPoolState(ctx, tx, changedPools); err != nil {
		return err
	}

	return i.storeLatestHeight(ctx, tx)
}

// ProcessBlock implements ingest.Ingester.
func (i *sqsIngester) storeLatestHeight(ctx sdk.Context, tx repository.Tx) error {
	goCtx := sdk.WrapSDKContext(ctx)

	height := ctx.BlockHeight()

	ctx.Logger().Info("ingesting latest blockchain height", "height", height)

	err := i.chainInfoRepo.StoreLatestHeight(sdk.WrapSDKContext(ctx), tx, uint64(height))
	if err != nil {
		ctx.Logger().Error("failed to ingest latest blockchain height", "error", err)
		return err
	}

	// Flush all writes atomically
	return tx.Exec(goCtx)
}
