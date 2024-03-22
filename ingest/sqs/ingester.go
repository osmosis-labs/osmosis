package sqs

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/osmosis-labs/sqs/sqsdomain"
	"github.com/osmosis-labs/sqs/sqsdomain/json"
	prototypes "github.com/osmosis-labs/sqs/sqsdomain/proto/types"

	"github.com/osmosis-labs/osmosis/v23/ingest"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
)

const sqsIngesterName = "sidecar-query-server"

var _ ingest.Ingester = &sqsIngester{}

const (
	// 10 KB
	maxCallMsgSize = 10 * 1024 * 1024
	// 5 KB
	payloadThreshold = maxCallMsgSize / 2

	// synchingThresholdSeconds is the threshold for synching detection.
	// If the difference between the last block time and the current time is less than this threshold,
	// the block processing is skipped.
	synchingThresholdSeconds = time.Second

	// blockProcessTimeout is the timeout for block processing.
	// Given that block times are 4.5 seconds, this is quiote generous.
	blockProcessTimeout = time.Second * 30
)

// sqsIngester is a sidecar query server (SQS) implementation of Ingester.
// It encapsulates all individual SQS ingesters.
type sqsIngester struct {
	poolsIngester domain.BlockProcessor
	host          string
	port          string
	grpcConn      *grpc.ClientConn
	appCodec      codec.Codec
	mx            *sync.Mutex
	lastBlockTime time.Time

	isProcessingBlock atomic.Bool
}

// NewSidecarQueryServerIngester creates a new sidecar query server ingester.
// poolsRepository is the storage for pools.
// gammKeeper is the keeper for Gamm pools.
func NewSidecarQueryServerIngester(poolsIngester domain.BlockProcessor, host, port string, appCodec codec.Codec) ingest.Ingester {
	return &sqsIngester{
		poolsIngester: poolsIngester,
		appCodec:      appCodec,

		mx:            &sync.Mutex{},
		lastBlockTime: time.Unix(0, 0),

		isProcessingBlock: atomic.Bool{},
	}
}

type IngestProcessBlockArgs struct {
	Pools []sqsdomain.PoolI
}

// ProcessBlock implements ingest.Ingester.
func (i *sqsIngester) ProcessBlock(ctx sdk.Context) (err error) {
	// This mechanism ensures that we do not process the next block while synching.
	if i.isProcessingBlock.Load() {
		return nil
	}

	// Set processing block to true
	i.isProcessingBlock.Store(true)

	// Note, that this has to be async as SQS might query the node for CW contract estimates.
	// As a result, SQS data ingestion must not be blocked by this.
	go func() (err error) {
		defer func() {
			// Set processing block to false
			i.isProcessingBlock.Store(false)

			// Measure block processing time
			telemetry.MeasureSince(i.lastBlockTime, "sqs_block_process_time")

			// Update last block time for synching detection
			i.lastBlockTime = time.Now()

			var panicErr any
			if panicErr = recover(); panicErr != nil {
				telemetry.IncrCounterWithLabels([]string{"sqs_block_process_error"}, 1, []metrics.Label{
					{
						Name:  "err",
						Value: fmt.Sprintf("panic: %v", panicErr),
					},
				})
			}

			if err != nil {
				telemetry.IncrCounterWithLabels([]string{"sqs_block_process_error"}, 1, []metrics.Label{
					{
						Name:  "err",
						Value: err.Error(),
					},
				})
			}

			// If error or panic occurred, close the grpc connection
			// so that it can be attempted to be re-established in the next block processing
			if (panicErr != nil || err != nil) && i.grpcConn != nil {
				i.grpcConn.Close()
				i.grpcConn = nil
			}
		}()

		// Synching detecting
		timeSinceLastBlock := time.Since(i.lastBlockTime)
		if timeSinceLastBlock < synchingThresholdSeconds {
			ctx.Logger().Info("Synching detected, skipping block processing", "height", ctx.BlockHeight(), "time_since_last_block", timeSinceLastBlock)
			return nil
		}

		startTime := time.Now()
		ctx.Logger().Info("Processing block", "height", ctx.BlockHeight())

		if i.grpcConn == nil {
			// TODO: move to config
			i.grpcConn, err = grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxCallMsgSize)), grpc.WithDisableRetry())
			if err != nil {
				return err
			}
		}

		ingesterClient := prototypes.NewSQSIngesterClient(i.grpcConn)

		// Process block by reading and writing data and ingesting data into sinks
		pools, takerFeeMap, err := i.poolsIngester.Process(ctx)
		if err != nil {
			return err
		}

		// Serialize taker fee map
		takerFeeMapBz, err := takerFeeMap.MarshalJSON()
		if err != nil {
			return err
		}

		// TODO: move this to config
		goCtx := sdk.WrapSDKContext(ctx)
		goCtx, cancel := context.WithTimeout(goCtx, blockProcessTimeout)
		defer cancel()

		// Start processing block by sending block height and taker fee updates
		if _, err := ingesterClient.StartBlockProcess(goCtx, &prototypes.StartBlockProcessRequest{
			BlockHeight:  uint64(ctx.BlockHeight()),
			TakerFeesMap: takerFeeMapBz,
		}); err != nil {
			return err
		}

		// Stream pools
		if err := i.streamPools(ctx, ingesterClient, pools); err != nil {
			return err
		}

		// End block processing
		if _, err := ingesterClient.EndBlockProcess(sdk.UnwrapSDKContext(ctx), &prototypes.EndBlockProcessRequest{
			BlockHeight: uint64(ctx.BlockHeight()),
		}); err != nil {
			return err
		}

		ctx.Logger().Info("Processed block", "height", ctx.BlockHeight(), "time", time.Since(startTime))

		return nil
	}()

	return nil
}

// streamPools streams pools to the SQS server in chunks.
// The process is serial today but can be parallelized in the future.
// Returns an error if any.
func (i *sqsIngester) streamPools(ctx sdk.Context, ingesterClient prototypes.SQSIngesterClient, pools []sqsdomain.PoolI) error {
	proceccChainPoolsClient, err := ingesterClient.ProcessChainPools(ctx)
	if err != nil {
		return err
	}

	chunk := &prototypes.ChainPoolsDataChunk{
		Pools: make([]*prototypes.PoolData, 0),
	}

	ctx.Logger().Info("begin sending pools", "height", ctx.BlockHeight(), "num_pools", len(pools))
	byteCount := 0
	for j := 0; j < len(pools); j++ {

		// Serialize chain pool
		chainPoolBz, err := i.appCodec.MarshalInterfaceJSON(pools[j].GetUnderlyingPool())
		if err != nil {
			return err
		}

		byteCount += len(chainPoolBz)

		// Serialize sqs pool model
		sqsPoolBz, err := json.Marshal(pools[j].GetSQSPoolModel())
		if err != nil {
			return err
		}

		byteCount += len(sqsPoolBz)

		// if the pool is concentrated, serialize tick model
		var tickModelBz []byte
		if pools[j].GetType() == poolmanagertypes.Concentrated {
			tickModel, err := pools[j].GetTickModel()
			if err != nil {
				return err
			}

			tickModelBz, err = json.Marshal(tickModel)
			if err != nil {
				return err
			}

			byteCount += len(tickModelBz)
		}

		// Append pool data to chunk
		chunk.Pools = append(chunk.Pools, &prototypes.PoolData{
			ChainModel: chainPoolBz,
			SqsModel:   sqsPoolBz,
			TickModel:  tickModelBz,
		})

		// Send chunk if it exceeds the threshold or if it is the last chunk
		shouldSendChunk := byteCount > payloadThreshold || j == len(pools)-1
		if shouldSendChunk {
			if err := proceccChainPoolsClient.Send(chunk); err != nil {
				return err
			}

			byteCount = 0

			// Initialize new chunk
			chunk = &prototypes.ChainPoolsDataChunk{
				Pools: make([]*prototypes.PoolData, 0),
			}
		}
	}

	// Close and receive the response
	if _, err := proceccChainPoolsClient.CloseAndRecv(); err != nil {
		return err
	}

	ctx.Logger().Info("finished sending pools", "height", ctx.BlockHeight())

	return nil
}

// GetName implements ingest.Ingester.
func (*sqsIngester) GetName() string {
	return sqsIngesterName
}
