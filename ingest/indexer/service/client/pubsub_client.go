package service

import (
	"context"
	"encoding/json"
	"errors"

	"cloud.google.com/go/pubsub"

	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

// PubSubClient is a client for publishing messages to a PubSub topic.
type PubSubClient struct {
	projectId                string
	blockTopicId             string
	transactionTopicId       string
	poolTopicId              string
	tokenSupplyTopicId       string
	tokenSupplyOffsetTopicId string
	pubsubClient             *pubsub.Client
}

// NewPubSubCLient creates a new PubSubClient.
func NewPubSubCLient(projectId string, blockTopicId string, transactionTopicId string, poolTopicId string, tokenSupplyTopicId string, tokenSupplyOffsetTopicId string) *PubSubClient {
	return &PubSubClient{
		projectId:                projectId,
		blockTopicId:             blockTopicId,
		transactionTopicId:       transactionTopicId,
		poolTopicId:              poolTopicId,
		tokenSupplyTopicId:       tokenSupplyTopicId,
		tokenSupplyOffsetTopicId: tokenSupplyOffsetTopicId,
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

	// Publish message to topic
	topic := p.pubsubClient.Topic(topicId)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: msgBytes,
	})

	// Block until message is published
	_, err = result.Get(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Publish implements PubSubClient.PublishBlock
func (p *PubSubClient) PublishBlock(ctx context.Context, block indexerdomain.Block) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.blockTopicId == "" {
		return errors.New("project id and block topic id must be set")
	}
	return p.publish(ctx, block, p.blockTopicId)
}

// PublishTransaction implements PubSubClient.PublishTransaction
func (p *PubSubClient) PublishTransaction(ctx context.Context, txn indexerdomain.Transaction) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.transactionTopicId == "" {
		return errors.New("project id and transaction topic id must be set")
	}
	return p.publish(ctx, txn, p.transactionTopicId)
}

// PublishPool implements PubSubClient.PublishPool
func (p *PubSubClient) PublishPool(ctx context.Context, pool indexerdomain.Pool) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.poolTopicId == "" {
		return errors.New("project id and pool topic id must be set")
	}
	return p.publish(ctx, pool, p.poolTopicId)
}

// PublishTokenSupply implements domain.PubSubClient.
func (p *PubSubClient) PublishTokenSupply(ctx context.Context, tokenSupply indexerdomain.TokenSupply) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.tokenSupplyTopicId == "" {
		return errors.New("project id and token supply topic id must be set")
	}
	return p.publish(ctx, tokenSupply, p.tokenSupplyTopicId)
}

// PublishTokenSupplyOffset implements domain.PubSubClient.
func (p *PubSubClient) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset indexerdomain.TokenSupplyOffset) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.tokenSupplyOffsetTopicId == "" {
		return errors.New("project id and token supply offset topic id must be set")
	}
	return p.publish(ctx, tokenSupplyOffset, p.tokenSupplyOffsetTopicId)
}

// marshal marshals a message to bytes.
func (p *PubSubClient) marshal(message any) ([]byte, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return data, nil
}
