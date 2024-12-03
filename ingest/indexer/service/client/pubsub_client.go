package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"cloud.google.com/go/pubsub"

	indexerdomain "github.com/osmosis-labs/osmosis/v28/ingest/indexer/domain"
)

// PubSubClient is a client for publishing messages to a PubSub topic.
type PubSubClient struct {
	maxPublishDelay          int
	projectId                string
	blockTopicId             string
	transactionTopicId       string
	poolTopicId              string
	tokenSupplyTopicId       string
	tokenSupplyOffsetTopicId string
	pairTopicId              string
	pubsubClient             *pubsub.Client
}

// NewPubSubCLient creates a new PubSubClient.
func NewPubSubCLient(maxPublishDelay int, projectId, blockTopicId, transactionTopicId, poolTopicId, tokenSupplyTopicId, tokenSupplyOffsetTopicId, pairTopicID string) *PubSubClient {
	return &PubSubClient{
		maxPublishDelay:          maxPublishDelay,
		projectId:                projectId,
		blockTopicId:             blockTopicId,
		transactionTopicId:       transactionTopicId,
		poolTopicId:              poolTopicId,
		tokenSupplyTopicId:       tokenSupplyTopicId,
		tokenSupplyOffsetTopicId: tokenSupplyOffsetTopicId,
		pairTopicId:              pairTopicID,
	}
}

// publish publishes a message to the PubSub topic.
func (p *PubSubClient) publish(ctx context.Context, message any, topicId string) error {
	// Create PubSub client if it doesn't exist
	if p.pubsubClient == nil {
		client, err := pubsub.NewClient(ctx, p.projectId)
		if err != nil {
			return err
		}
		p.pubsubClient = client
	}

	// Marshal message to bytes
	msgBytes, err := p.marshal(message)
	if err != nil {
		return err
	}

	// Publish message to the topic. When the message publishing rate is very low, messages may remain pending and stale within the Pub/Sub SDK.
	// For example, if only one message is published over a span of several minutes, the default DelayThreshold and CountThreshold values
	// are high enough that the message may seem undelivered or lost.
	// To mitigate this, it's essential to reduce the DelayThreshold to a lower value, such as 4 seconds, to ensure timely delivery.
	topic := p.pubsubClient.Topic(topicId)
	topic.PublishSettings.DelayThreshold = time.Duration(p.maxPublishDelay) * time.Second
	topic.Publish(ctx, &pubsub.Message{
		Data: msgBytes,
	})

	return nil
}

// Publish implements PubSubClient.PublishBlock
func (p *PubSubClient) PublishBlock(ctx context.Context, block indexerdomain.Block) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.blockTopicId == "" {
		return errors.New("project id and block topic id must be set")
	}
	block.IngestedAt = time.Now().UTC()
	return p.publish(ctx, block, p.blockTopicId)
}

// PublishTransaction implements PubSubClient.PublishTransaction
func (p *PubSubClient) PublishTransaction(ctx context.Context, txn indexerdomain.Transaction) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.transactionTopicId == "" {
		return errors.New("project id and transaction topic id must be set")
	}
	txn.IngestedAt = time.Now().UTC()
	return p.publish(ctx, txn, p.transactionTopicId)
}

// PublishTokenSupply implements domain.PubSubClient.
func (p *PubSubClient) PublishTokenSupply(ctx context.Context, tokenSupply indexerdomain.TokenSupply) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.tokenSupplyTopicId == "" {
		return errors.New("project id and token supply topic id must be set")
	}
	tokenSupply.IngestedAt = time.Now().UTC()
	return p.publish(ctx, tokenSupply, p.tokenSupplyTopicId)
}

// PublishTokenSupplyOffset implements domain.PubSubClient.
func (p *PubSubClient) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset indexerdomain.TokenSupplyOffset) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.tokenSupplyOffsetTopicId == "" {
		return errors.New("project id and token supply offset topic id must be set")
	}
	tokenSupplyOffset.IngestedAt = time.Now().UTC()
	return p.publish(ctx, tokenSupplyOffset, p.tokenSupplyOffsetTopicId)
}

// PublishPair implements PubSubClient.PublishPair
func (p *PubSubClient) PublishPair(ctx context.Context, pair indexerdomain.Pair) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.pairTopicId == "" {
		return errors.New("project id and pool topic id must be set")
	}

	pair.IngestedAt = time.Now().UTC()
	return p.publish(ctx, pair, p.pairTopicId)
}

// marshal marshals a message to bytes.
func (p *PubSubClient) marshal(message any) ([]byte, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return data, nil
}
