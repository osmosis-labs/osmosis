package sqs

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v24/ingest/sqs/domain"
)

var _ domain.Ingester = &sqsIngester{}

// sqsIngester is a sidecar query server (SQS) implementation of Ingester.
// It encapsulates all individual SQS ingesters.
type sqsIngester struct {
	poolsTransformer domain.PoolsTransformer
	keepers          domain.SQSIngestKeepers
	sqsGRPCClient    domain.SQSGRPClient
}

// NewSidecarQueryServerIngester creates a new sidecar query server ingester.
// poolsRepository is the storage for pools.
// gammKeeper is the keeper for Gamm pools.
func NewSidecarQueryServerIngester(poolsIngester domain.PoolsTransformer, appCodec codec.Codec, keepers domain.SQSIngestKeepers, sqsGRPCClient domain.SQSGRPClient) domain.Ingester {
	return &sqsIngester{
		poolsTransformer: poolsIngester,
		keepers:          keepers,
		sqsGRPCClient:    sqsGRPCClient,
	}
}

// ProcessAllBlockData implements ingest.Ingester.
func (i *sqsIngester) ProcessAllBlockData(ctx sdk.Context) error {
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
	pools, takerFeesMap, err := i.poolsTransformer.Transform(ctx, blockPools)
	if err != nil {
		return err
	}

	return i.sqsGRPCClient.PushData(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
}

// ProcessChangedBlockData implements ingest.Ingester.
func (i *sqsIngester) ProcessChangedBlockData(ctx sdk.Context, changedPools domain.BlockPools) error {
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

	// NOTE: we are failing to detect CW pool store updates which were noticed post-release.
	// As a result, we push all of them every block.
	// https://linear.app/osmosis/issue/STABI-103/push-updated-pools-into-sqs-instead-of-all-every-block
	cosmWasmPools, err := i.keepers.CosmWasmPoolKeeper.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return err
	}

	changedPools.CosmWasmPools = cosmWasmPools

	// Process block by reading and writing data and ingesting data into sinks
	pools, takerFeesMap, err := i.poolsTransformer.Transform(ctx, changedPools)
	if err != nil {
		return err
	}

	return i.sqsGRPCClient.PushData(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
}
