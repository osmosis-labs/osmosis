package sqs

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
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
func (i *sqsIngester) ProcessAllBlockData(ctx sdk.Context) ([]poolmanagertypes.PoolI, error) {
	// Concentrated pools

	concentratedPools, err := i.keepers.ConcentratedKeeper.GetPools(ctx)
	if err != nil {
		return nil, err
	}

	// CFMM pools

	cfmmPools, err := i.keepers.GammKeeper.GetPools(ctx)
	if err != nil {
		return nil, err
	}

	// CosmWasm pools

	cosmWasmPools, err := i.keepers.CosmWasmPoolKeeper.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return nil, err
	}

	blockPools := domain.BlockPools{
		ConcentratedPools: concentratedPools,
		CosmWasmPools:     cosmWasmPools,
		CFMMPools:         cfmmPools,
	}

	// Process block by reading and writing data and ingesting data into sinks
	pools, takerFeesMap, err := i.poolsTransformer.Transform(ctx, blockPools)
	if err != nil {
		return nil, err
	}

	err = i.sqsGRPCClient.PushData(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
	if err != nil {
		return nil, err
	}

	return cosmWasmPools, nil
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

	// Process block by reading and writing data and ingesting data into sinks
	pools, takerFeesMap, err := i.poolsTransformer.Transform(ctx, changedPools)
	if err != nil {
		return err
	}

	return i.sqsGRPCClient.PushData(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
}
