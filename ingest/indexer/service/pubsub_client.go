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
	projectId    string
	topicId      string
	pubsubClient *pubsub.Client
}

// NewPubSubCLient creates a new PubSubClient.
func NewPubSubCLient(projectId string, topicId string) *PubSubClient {
	return &PubSubClient{
		projectId: projectId,
		topicId:   topicId,
	}
}

// publish publishes a message to the PubSub topic.
func (p *PubSubClient) publish(ctx context.Context, message any) error {
	// Check if project id and topic id are set
	if p.projectId == "" || p.topicId == "" {
		return errors.New("project id and topic id must be set")
	}

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
	topic := p.pubsubClient.Topic(p.topicId)
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
	return p.publish(ctx, block)
}

// PublishTransaction implements PubSubClient.PublishTransaction
func (p *PubSubClient) PublishTransaction(ctx context.Context, txn indexerdomain.Transaction) error {
	return p.publish(ctx, txn)
}

// PublishAsset implements PubSubClient.PublishAsset
func (p *PubSubClient) PublishAsset(ctx context.Context, asset indexerdomain.Asset) error {
	return p.publish(ctx, asset)
}

// PublishPool implements PubSubClient.PublishPool
func (p *PubSubClient) PublishPool(ctx context.Context, pool indexerdomain.Pool) error {
	return p.publish(ctx, pool)
}

// marshal marshals a message to bytes.
func (p *PubSubClient) marshal(message any) ([]byte, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return data, nil
}
