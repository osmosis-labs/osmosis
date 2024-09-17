package service

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/cosmos/cosmos-sdk/codec"
	"google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/sqsdomain"

	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"

	prototypes "github.com/osmosis-labs/sqs/sqsdomain/proto/types"
)

type pubSubClient struct {
	appCodec          codec.Codec
	lastFullBlockTime time.Time
}

const (
	// TODO: move to config
	projectID = "osmosis-ingest"
	topicName = "sqs-pools"

	fullBlockRefetchInterval uint64 = 45
)

var (
	_ domain.SQSGRPClient = &pubSubClient{}
)

func NewPubSubCLient(appCodec codec.Codec) *pubSubClient {
	return &pubSubClient{
		appCodec:          appCodec,
		lastFullBlockTime: time.Time{},
	}
}

// PushData implements domain.GracefulSQSGRPClient.
func (g *pubSubClient) PushData(ctx context.Context, height uint64, pools []sqsdomain.PoolI, takerFeesMap sqsdomain.TakerFeeMap) (err error) {
	defer func() {
		// If there is an error, we will return it
		if err == nil {
			// If there is no error, we will check if we need to refetch the block
			//
			// Every fullBlockRefetchInterval blocks, we will signal to refetch all pool data from the chain
			// This is so that newly joined nodes can catch up.
			if height%fullBlockRefetchInterval == 1 {

				err = fmt.Errorf("pub-sub signal to refetch the block at height %d with interval %d", height, fullBlockRefetchInterval)
			}
		}
	}()

	// Marshal pools
	// Marshal pools
	poolData, err := marshalPools(g.appCodec, pools)
	if err != nil {
		return err
	}

	// Marshal taker fees
	takerFeesBz, err := takerFeesMap.MarshalJSON()
	if err != nil {
		return err
	}

	req := &prototypes.ProcessBlockRequest{
		BlockHeight:  height,
		TakerFeesMap: takerFeesBz,
		Pools:        poolData,
	}

	protoBz, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	fmt.Println("Publishing to pubsub", "height", height, "topic", topicName, "bytes", len(protoBz))

	return g.publish(ctx, projectID, topicName, protoBz, height)
}

// publish publishes a message to a pubsub topic
func (g *pubSubClient) publish(ctx context.Context, projectID, topicID string, msg []byte, height uint64) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub: NewClient: %w", err)
	}
	defer client.Close()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	isFullBlockRefetch := height%fullBlockRefetchInterval == 0
	if isFullBlockRefetch {
		g.lastFullBlockTime = blockTime
	}

	t := client.Topic(topicID)
	t.EnableMessageOrdering = true
	result := t.Publish(ctx, &pubsub.Message{
		Data: msg,
		Attributes: map[string]string{
			"is_full_block_refetch": fmt.Sprintf("%t", isFullBlockRefetch),
			"last_full_block_time":  g.lastFullBlockTime.Format(time.RFC3339),
		},
		OrderingKey: fmt.Sprintf("%d", height),
		PublishTime: blockTime,
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	return nil
}
