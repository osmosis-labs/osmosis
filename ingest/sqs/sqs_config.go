package sqs

import (
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v23/ingest"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
	poolsingester "github.com/osmosis-labs/osmosis/v23/ingest/sqs/pools/ingester"
)

// Config defines the config for the sidecar query server.
type Config struct {
	// IsEnabled defines if the sidecar query server is enabled.
	IsEnabled bool `mapstructure:"enabled"`
}

const (
	groupOptName = "osmosis-sqs"
)

// DefaultConfig defines the default config for the sidecar query server.
var DefaultConfig = Config{
	IsEnabled: false,
}

// NewConfigFromOptions returns a new sidecar query server config from the given options.
func NewConfigFromOptions(opts servertypes.AppOptions) Config {
	isEnabled := osmoutils.ParseBool(opts, groupOptName, "is-enabled", false)

	if !isEnabled {
		return Config{
			IsEnabled: false,
		}
	}

	return Config{
		IsEnabled: isEnabled,
	}
}

// Initialize initializes the sidecar query server and returns the ingester.
func (c Config) Initialize(appCodec codec.Codec, keepers domain.SQSIngestKeepers) (ingest.Ingester, error) {
	// Create pools ingester
	poolsIngester := poolsingester.NewPoolIngester(keepers)

	// Create sqs ingester that encapsulates all ingesters.
	sqsIngester := NewSidecarQueryServerIngester(poolsIngester, appCodec)

	return sqsIngester, nil
}
