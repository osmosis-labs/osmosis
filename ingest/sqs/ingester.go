package sqs

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/osmosis-labs/osmosis/v24/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v24/x/poolmanager/types"
)

const ingesterName = "sqs-ingester"

var _ domain.Ingester = &sqsIngester{}

func (i sqsIngester) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", i.GetName())
}

// sqsIngester is a sidecar query server (SQS) implementation of Ingester.
// It encapsulates all individual SQS ingesters.
type sqsIngester struct {
	poolsTransformer domain.PoolsTransformer
	keepers          domain.SQSIngestKeepers
	sqsGRPCClients   []domain.SQSGRPClient
	logger           log.Logger
}

func (i sqsIngester) GetName() string {
	return ingesterName
}

// NewSidecarQueryServerIngester creates a new sidecar query server ingester.
// poolsRepository is the storage for pools.
// gammKeeper is the keeper for Gamm pools.
func NewSidecarQueryServerIngester(poolsIngester domain.PoolsTransformer, appCodec codec.Codec, keepers domain.SQSIngestKeepers, sqsGRPCClients []domain.SQSGRPClient) domain.Ingester {
	return &sqsIngester{
		poolsTransformer: poolsIngester,
		keepers:          keepers,
		sqsGRPCClients:   sqsGRPCClients,
	}
}

// ProcessAllBlockData implements ingest.Ingester.
func (i *sqsIngester) ProcessAllBlockData(ctx sdk.Context) ([]poolmanagertypes.PoolI, error) {
	// Initialize logger if it is nil
	if i.logger == nil {
		i.logger = i.Logger(ctx)
	}

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

	for _, client := range i.sqsGRPCClients {
		err = client.PushData(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
		if err != nil {
			return nil, err
		}
	}

	return cosmWasmPools, nil
}

// ProcessChangedBlockData implements ingest.Ingester.
func (i *sqsIngester) ProcessChangedBlockData(ctx sdk.Context, changedPools domain.BlockPools) error {
	// Initialize logger if it is nil
	if i.logger == nil {
		i.logger = i.Logger(ctx)
	}

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

	// Loop to push data to each SQS instances in a separate thread
	var wg sync.WaitGroup
	for _, client := range i.sqsGRPCClients {
		wg.Add(1)
		go func(c domain.SQSGRPClient, wg *sync.WaitGroup) {
			defer wg.Done()
			err := c.PushData(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
			if err != nil {
				i.logger.Error("Failed to push data to SQS", "error", err.Error())
			}
		}(client, &wg)
	}
	wg.Wait()

	return nil
}
