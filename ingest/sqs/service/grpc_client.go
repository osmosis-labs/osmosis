package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
	ingesttypes "github.com/osmosis-labs/osmosis/v28/ingest/types"
	prototypes "github.com/osmosis-labs/osmosis/v28/ingest/types/proto/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v28/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// timeAfterFunc is time.AfterFunc
type timeAfterFunc func(time.Duration) <-chan time.Time

type GRPCClient struct {
	timeAfterFunc     timeAfterFunc
	connCheckInterval time.Duration
	grpcAddress       string
	conn              domain.ClientConn
	appCodec          codec.Codec
}

var _ domain.SQSGRPClient = &GRPCClient{}

// NewGRPCCLient will create a new gRPC client connection to the SQS service and return a GRPCClient instance.
// Underlying connection is being managed in a separate goroutine to handle reconnections.
// Clients should call IsConnected() before using the client to ensure the connection is ready for use and cancel the context
// for graceful shutdown.
func NewGRPCCLient(ctx context.Context, grpcAddress string, grpcMaxCallSizeBytes int, appCodec codec.Codec) (*GRPCClient, error) {
	conn, err := grpc.NewClient(
		grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxCallSizeBytes)),
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
		return nil, err
	}
	client := &GRPCClient{
		timeAfterFunc:     time.After,
		grpcAddress:       grpcAddress,
		conn:              conn,
		appCodec:          appCodec,
		connCheckInterval: time.Second,
	}

	go client.connect(ctx)

	return client, nil
}

// PushData implements domain.GracefulSQSGRPClient.
func (g *GRPCClient) PushData(ctx context.Context, height uint64, pools []ingesttypes.PoolI, takerFeesMap ingesttypes.TakerFeeMap) (err error) {
	if err := g.IsConnected(); err != nil {
		return err
	}

	readyConn, ok := g.conn.(*grpc.ClientConn)
	if !ok {
		return fmt.Errorf("failed to cast g.conn to grpc.ClientConn")
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

// connect manages underlying gRPC connection by checking the connection state and attempting to reconnect when necessary.
func (g *GRPCClient) connect(ctx context.Context) {
	for {
		// Check if the context is done
		if ctx.Err() != nil {
			return
		}

		state := g.conn.GetState()
		if state != connectivity.Ready {
			// Attempt to connect
			g.conn.Connect()

			// Wait for a state change or timeout/cancel
			if !g.conn.WaitForStateChange(ctx, state) {
				return // Context done
			}
		} else {
			select {
			case <-ctx.Done():
				return
			case <-g.timeAfterFunc(g.connCheckInterval):
				// Recheck the connection state after interval
			}
		}
	}
}

// IsConnected returns true if the gRPC connection is ready.
func (g *GRPCClient) IsConnected() error {
	if g.conn.GetState() != connectivity.Ready {
		return fmt.Errorf("SQS gRPC connection to %s is not ready yet", g.grpcAddress)
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
