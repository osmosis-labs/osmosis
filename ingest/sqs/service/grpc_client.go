package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	ingesttypes "github.com/osmosis-labs/osmosis/v28/ingest/types"
	prototypes "github.com/osmosis-labs/osmosis/v28/ingest/types/proto/types"

	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v28/x/poolmanager/types"
)

type GRPCClient struct {
	grpcAddress          string
	grpcMaxCallSizeBytes int
	grpcConn             *grpc.ClientConn
	appCodec             codec.Codec

	grpcConnMx sync.RWMutex
}

var (
	_ domain.SQSGRPClient = &GRPCClient{}
)

func NewGRPCCLient(grpcAddress string, grpxMaxCallSizeBytes int, appCodec codec.Codec) *GRPCClient {
	return &GRPCClient{
		grpcAddress:          grpcAddress,
		grpcMaxCallSizeBytes: grpxMaxCallSizeBytes,
		appCodec:             appCodec,
		grpcConnMx:           sync.RWMutex{},
	}
}

// PushData implements domain.GracefulSQSGRPClient.
func (g *GRPCClient) PushData(ctx context.Context, height uint64, pools []ingesttypes.PoolI, takerFeesMap ingesttypes.TakerFeeMap) (err error) {
	readyConn, err := g.returnReadyConnection()
	if err != nil {
		return err
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

	ingesterClient := prototypes.NewSQSIngesterClient(readyConn)

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

// StartConnectionAsync starts a background goroutine that attempts to reconnect to the SQS service.
// It sleeps for a minute if the connection is ready and revalidates it, and a second if the connection is not ready
// attempting to regain connection.
func (g *GRPCClient) StartConnectionAsync() {
	go func() {
		var err error

		for {
			g.grpcConnMx.Lock()
			if g.grpcConn == nil {
				g.grpcConn, err = grpc.NewClient(
					g.grpcAddress,
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(g.grpcMaxCallSizeBytes)),
					grpc.WithDisableRetry(),
					grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
					grpc.WithConnectParams(grpc.ConnectParams{
						// Arbitrary choices based on brute-force testing.
						Backoff: backoff.Config{
							BaseDelay:  50 * time.Millisecond,
							Jitter:     0.2,
							Multiplier: 2,
							MaxDelay:   10 * time.Second,
						},
					}),
				)
				if err != nil {
					g.grpcConnMx.Unlock()
					continue
				}
			}

			isConnected := g.grpcConn.GetState() == connectivity.Ready
			if !isConnected {
				g.grpcConn.Connect()
			}
			g.grpcConnMx.Unlock()

			// If the connection is ready, we should sleep for a minute
			// to avoid unhelpful retries.
			// Otherwise, we should attempt to reconnect sooner, with a second-long wait.
			if isConnected {
				time.Sleep(time.Minute)
			} else {
				time.Sleep(time.Second)
			}
		}
	}()
}

// IsConnected returns true if the gRPC connection is ready.
func (g *GRPCClient) IsConnected() bool {
	g.grpcConnMx.RLock()
	defer g.grpcConnMx.RUnlock()

	return g.grpcConn.GetState() == connectivity.Ready
}

// returnReadyConnection returns a ready gRPC connection.
// If the connection is not ready, it returns an error.
func (g *GRPCClient) returnReadyConnection() (*grpc.ClientConn, error) {
	g.grpcConnMx.RLock()
	defer g.grpcConnMx.RUnlock()
	grpcConn := g.grpcConn

	if grpcConn != nil {
		if grpcConn.GetState() != connectivity.Ready {
			return nil, fmt.Errorf("sqs grpc connection is not ready yet. Continuing to attempt connection in the background")
		}
	}

	return grpcConn, nil
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
