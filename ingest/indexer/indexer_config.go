package indexer

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v30/ingest/indexer/domain"
	service "github.com/osmosis-labs/osmosis/v30/ingest/indexer/service/client"
)

// Config defines the config for the indexer.
type Config struct {
	IsEnabled                bool   `mapstructure:"enabled"`
	MaxPublishDelay          int    `mapstructure:"max-publish-delay"`
	GCPProjectId             string `mapstructure:"gcp-project-id"`
	BlockTopicId             string `mapstructure:"block-topic-id"`
	TransactionTopicId       string `mapstructure:"transaction-topic-id"`
	PoolTopicId              string `mapstructure:"pool-topic-id"`
	TokenSupplyTopicId       string `mapstructure:"token-supply-topic-id"`
	TokenSupplyOffsetTopicId string `mapstructure:"token-supply-offset-topic-id"`
	PairTopicId              string `mapstructure:"pair-offset-topic-id"`
}

// groupOptName is the name of the indexer options group.
const (
	groupOptName = "osmosis-indexer"
)

// DefaultConfig defines the default config for the indexer client.
var DefaultConfig = Config{
	IsEnabled:                false,
	MaxPublishDelay:          4,
	GCPProjectId:             "",
	BlockTopicId:             "",
	TransactionTopicId:       "",
	PoolTopicId:              "",
	TokenSupplyTopicId:       "",
	TokenSupplyOffsetTopicId: "",
}

// NewConfigFromOptions returns a new indexer config from the given options.
func NewConfigFromOptions(opts servertypes.AppOptions) Config {
	isEnabled := osmoutils.ParseBool(opts, groupOptName, "is-enabled", false)

	if !isEnabled {
		return Config{
			IsEnabled: false,
		}
	}

	maxPublishDelay := osmoutils.ParseInt(opts, groupOptName, "max-publish-delay")
	gcpProjectId := osmoutils.ParseString(opts, groupOptName, "gcp-project-id")
	blockTopicId := osmoutils.ParseString(opts, groupOptName, "block-topic-id")
	transactionTopicId := osmoutils.ParseString(opts, groupOptName, "transaction-topic-id")
	poolTopicId := osmoutils.ParseString(opts, groupOptName, "pool-topic-id")
	tokenSupplyTopicId := osmoutils.ParseString(opts, groupOptName, "token-supply-topic-id")
	tokenSupplyOffsetTopicId := osmoutils.ParseString(opts, groupOptName, "token-supply-offset-topic-id")
	pairTopicID := osmoutils.ParseString(opts, groupOptName, "pair-topic-id")

	return Config{
		IsEnabled:                isEnabled,
		MaxPublishDelay:          maxPublishDelay,
		GCPProjectId:             gcpProjectId,
		BlockTopicId:             blockTopicId,
		TransactionTopicId:       transactionTopicId,
		PoolTopicId:              poolTopicId,
		TokenSupplyTopicId:       tokenSupplyTopicId,
		TokenSupplyOffsetTopicId: tokenSupplyOffsetTopicId,
		PairTopicId:              pairTopicID,
	}
}

// Initialize initializes the indexer by creating a new PubSubClient and returning a new IndexerIngester.
func (c Config) Initialize() domain.Publisher {
	pubSubClient := service.NewPubSubCLient(c.MaxPublishDelay, c.GCPProjectId, c.BlockTopicId, c.TransactionTopicId, c.PoolTopicId, c.TokenSupplyTopicId, c.TokenSupplyOffsetTopicId, c.PairTopicId)
	return NewIndexerPublisher(*pubSubClient)
}
