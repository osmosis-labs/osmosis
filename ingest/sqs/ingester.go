package sqs

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/sqsdomain"
	prototypes "github.com/osmosis-labs/sqs/sqsdomain/proto/types"
	"github.com/osmosis-labs/sqs/sqsdomain/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
)

var _ domain.Ingester = &sqsIngester{}

const maxCallMsgSize = 50 * 1024 * 1024

// sqsIngester is a sidecar query server (SQS) implementation of Ingester.
// It encapsulates all individual SQS ingesters.
type sqsIngester struct {
	txManager        repository.TxManager
	poolsTransformer domain.PoolsTransformer
	keepers          domain.SQSIngestKeepers
	grpcConn         *grpc.ClientConn
	appCodec         codec.Codec
}

// NewSidecarQueryServerIngester creates a new sidecar query server ingester.
// poolsRepository is the storage for pools.
// gammKeeper is the keeper for Gamm pools.
func NewSidecarQueryServerIngester(poolsIngester domain.PoolsTransformer, appCodec codec.Codec, keepers domain.SQSIngestKeepers) domain.Ingester {
	return &sqsIngester{
		poolsTransformer: poolsIngester,
		appCodec:         appCodec,
		keepers:          keepers,
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

	return i.pushDataGRPC(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
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

	return i.pushDataGRPC(ctx, uint64(ctx.BlockHeight()), pools, takerFeesMap)
}

func (i *sqsIngester) pushDataGRPC(ctx context.Context, height uint64, pools []sqsdomain.PoolI, takerFeesMap sqsdomain.TakerFeeMap) (err error) {
	if i.grpcConn == nil {
		// TODO: move to config
		i.grpcConn, err = grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxCallMsgSize)), grpc.WithDisableRetry())
		if err != nil {
			return
		}
	}

	// Marshal pools
	poolData := make([]*prototypes.PoolData, len(pools))
	for _, pool := range pools {
		// Serialize chain pool
		chainPoolBz, err := i.appCodec.MarshalInterfaceJSON(pool.GetUnderlyingPool())
		if err != nil {
			return err
		}

		// Serialize sqs pool model
		sqsPoolBz, err := json.Marshal(pool.GetSQSPoolModel())
		if err != nil {
			return err
		}

		// if the pool is concentrated, serialize tick model
		var tickModelBz []byte
		if pool.GetType() == poolmanagertypes.Concentrated {
			tickModel, err := pool.GetTickModel()
			if err != nil {
				return err
			}

			tickModelBz, err = json.Marshal(tickModel)
			if err != nil {
				return err
			}
		}

		// Append pool data to chunk
		poolData = append(poolData, &prototypes.PoolData{
			ChainModel: chainPoolBz,
			SqsModel:   sqsPoolBz,
			TickModel:  tickModelBz,
		})
	}

	// Marshal taker fees
	takerFeesBz, err := takerFeesMap.MarshalJSON()
	if err != nil {
		return
	}
	ingesterClient := prototypes.NewSQSIngesterClient(i.grpcConn)

	req := prototypes.ProcessBlockRequest{
		BlockHeight:  height,
		TakerFeesMap: takerFeesBz,
		Pools:        poolData,
	}

	_, err = ingesterClient.ProcessBlock(ctx, &req)
	if err != nil {
		return err
	}

	return nil
}
