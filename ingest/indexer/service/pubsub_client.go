package service

import (
	"context"
	"encoding/json"
	"errors"

	"cloud.google.com/go/pubsub"

	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

type PubSubClient struct {
	projectId    string
	topicId      string
	pubsubClient *pubsub.Client
}

func NewPubSubCLient(projectId string, topicId string) *PubSubClient {
	return &PubSubClient{
		projectId: projectId,
		topicId:   topicId,
	}
}

func (p *PubSubClient) Publish(ctx context.Context, height uint64, block indexerdomain.Block) error {

	if p.projectId == "" || p.topicId == "" {
		return errors.New("project id and topic id must be set")
	}

	if p.pubsubClient == nil {
		client, err := pubsub.NewClient(ctx, p.projectId)
		if err != nil {
			return err
		}
		p.pubsubClient = client
	}

	msgBytes, err := p.marshal(block)
	if err != nil {
		return err
	}

	topic := p.pubsubClient.Topic(p.topicId)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: msgBytes,
	})

	_, err = result.Get(ctx)
	if err != nil {
		return err
	}

	return nil

}

func (p *PubSubClient) marshal(message any) ([]byte, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return data, nil
}
