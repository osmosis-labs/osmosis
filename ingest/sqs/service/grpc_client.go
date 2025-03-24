package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	ingesttypes "github.com/osmosis-labs/osmosis/v29/ingest/types"
	prototypes "github.com/osmosis-labs/osmosis/v29/ingest/types/proto/types"

	"github.com/osmosis-labs/osmosis/v29/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v29/x/poolmanager/types"
)

type GRPCClient struct {
	grpcAddress          string
	grpcMaxCallSizeBytes int
	grpcConn             *grpc.ClientConn
	appCodec             codec.Codec

	// If an error during connection occurs, we will attempt to reconnect every `blocksBeforeRetryConnection` blocks
	// as to avoid the node falling behind due to the retry logic of grpc.
	blocksBeforeRetryConnection int64
}

const (
	// initialBlocksBeforeRetryConnection is the number of blocks to wait before attempting to reconnect
	// to the SQS service in case of an error.
	initialBlocksBeforeRetryConnection = 5
)

var (
	_ domain.SQSGRPClient = &GRPCClient{}
)

func NewGRPCCLient(grpcAddress string, grpxMaxCallSizeBytes int, appCodec codec.Codec) *GRPCClient {
	return &GRPCClient{
		grpcAddress:                 grpcAddress,
		grpcMaxCallSizeBytes:        grpxMaxCallSizeBytes,
		appCodec:                    appCodec,
		blocksBeforeRetryConnection: 0,
	}
}

// PushData implements domain.GracefulSQSGRPClient.
func (g *GRPCClient) PushData(ctx context.Context, height uint64, pools []ingesttypes.PoolI, takerFeesMap ingesttypes.TakerFeeMap) (err error) {
	// If sqs service is unavailable, we should reset the connection
	// and attempt to reconnect during the next block.
	var shouldResetConnection bool

	defer func() {
		if shouldResetConnection {
			if g.grpcConn != nil {
				g.grpcConn.Close()
				g.grpcConn = nil
			}

			// Increase the counter for the grpc connection error
			telemetry.IncrCounterWithLabels([]string{domain.SQSGRPCConnectionErrorMetricName}, 1, []metrics.Label{
				telemetry.NewLabel("height", fmt.Sprintf("%d", height)),
				telemetry.NewLabel("err", err.Error()),
			})
		}
	}()

	if g.grpcConn == nil {
		// If we are in the retry mode, we should not attempt to reconnect
		// for every block to avoid the node falling behind.
		// Instead, we will attempt to reconnect every `blocksBeforeRetryConnection` blocks.
		if g.blocksBeforeRetryConnection > 0 {
			g.blocksBeforeRetryConnection--
			return fmt.Errorf("grpc connection is not ready")
		}

		// Note: we disable retries since we have a custom logic to repeat retries in the next block.
		// Using the built-in GRPC retry back-off logic is likely to halt the serial system.
		// As a result, we opt in for simply continuing to attempting to process the next block
		// and retrying the connection and ingest
		g.grpcConn, err = grpc.NewClient(
			g.grpcAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(g.grpcMaxCallSizeBytes)),
			grpc.WithDisableRetry(),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
			grpc.WithConnectParams(grpc.ConnectParams{
				// Note: this exists as to avoid node falling behind when SQS is not running
				// but config enables it. In that case, every block, the node will attempt to reconnect
				// and retry with exponential backoff. By setting max delay under block time (~1.++ second)
				// we ensure that the node will not fall behind.
				Backoff: backoff.Config{
					BaseDelay:  100 * time.Millisecond,
					Jitter:     0.2,
					Multiplier: 2,

					MaxDelay: 400 * time.Millisecond,
				},
			}),
		)
		if err != nil {
			shouldResetConnection = true

			g.blocksBeforeRetryConnection = initialBlocksBeforeRetryConnection

			return err
		}

		// Attempt to connect to the SQS service
		// Note: this operation is non-blocking
		// As a result, we have an exponential backoff below to validate if the connection is ready
		g.grpcConn.Connect()

		initialSleep := time.Millisecond * 100
		shouldFail := true
		for i := 0; i < 3; i++ {
			// If the connection is ready, we can break and continue
			if g.grpcConn.GetState() == connectivity.Ready {
				shouldFail = false
				break
			}

			// If the connection is not ready, we should sleep and retry
			// with an exponential backoff
			time.Sleep(initialSleep)
			initialSleep *= 2
		}

		// If we failed to connect to the SQS service after several attempts, we should return an error
		if shouldFail {
			shouldResetConnection = true
			g.blocksBeforeRetryConnection = initialBlocksBeforeRetryConnection
			return fmt.Errorf("grpc connection is not ready")
		}
	}

	// Marshal pools
	poolData, err := g.marshalPools(pools)
	if err != nil {
		return err
	}

	// Marshal taker fees
	takerFeesBz, err := takerFeesMap.MarshalJSON()
	if err != nil {
		return err
	}

	ingesterClient := prototypes.NewSQSIngesterClient(g.grpcConn)

	req := prototypes.ProcessBlockRequest{
		BlockHeight:  height,
		TakerFeesMap: takerFeesBz,
		Pools:        poolData,
	}

	_, err = ingesterClient.ProcessBlock(ctx, &req)
	if err != nil {
		status, ok := status.FromError(err)

		// If the connection is unavailable, we should reset the connection
		// and attempt to reconnect during the next block.
		// On any other error, we assume that the connection is still valid so we
		// do no attempt to recreate it. However, we still return the error to the caller.
		shouldResetConnection = ok && status.Code() == codes.Unavailable

		return err
	}

	return nil
}

// marshalPools marshals pools into a format that can be sent over gRPC.
func (g *GRPCClient) marshalPools(pools []ingesttypes.PoolI) ([]*prototypes.PoolData, error) {
	// Marshal pools
	poolData := make([]*prototypes.PoolData, 0, len(pools))
	for _, pool := range pools {
		// Serialize chain pool
		chainPoolBz, err := g.appCodec.MarshalInterfaceJSON(pool.GetUnderlyingPool())
		if err != nil {
			return nil, err
		}

		// Serialize sqs pool model
		sqsPoolBz, err := json.Marshal(pool.GetSQSPoolModel())
		if err != nil {
			return nil, err
		}

		// if the pool is concentrated, serialize tick model
		var tickModelBz []byte
		if pool.GetType() == poolmanagertypes.Concentrated {
			tickModel, err := pool.GetTickModel()
			if err != nil {
				return nil, err
			}

			tickModelBz, err = json.Marshal(tickModel)
			if err != nil {
				return nil, err
			}
		}

		// Append pool data to chunk
		poolData = append(poolData, &prototypes.PoolData{
			ChainModel: chainPoolBz,
			SqsModel:   sqsPoolBz,
			TickModel:  tickModelBz,
		})
	}
	return poolData, nil
}
