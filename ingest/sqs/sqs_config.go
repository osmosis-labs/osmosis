package sqs

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v23/ingest"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
	poolsingester "github.com/osmosis-labs/osmosis/v23/ingest/sqs/pools/ingester"
)

// Config defines the config for the sidecar query server.
type Config struct {
	// IsEnabled defines if the sidecar query server is enabled.
	IsEnabled bool `mapstructure:"enabled"`

	// Storage defines the storage host and port.
	StorageHost string `mapstructure:"db-host"`
	StoragePort string `mapstructure:"db-port"`
}

const (
	groupOptName = "osmosis-sqs"

	noRoutesCacheExpiry = 0
)

// DefaultConfig defines the default config for the sidecar query server.
var DefaultConfig = Config{

	IsEnabled: false,

	StorageHost: "localhost",
	StoragePort: "6379",
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

		StorageHost: osmoutils.ParseString(opts, groupOptName, "db-host"),
		StoragePort: osmoutils.ParseString(opts, groupOptName, "db-port"),
	}
}

// Initialize initializes the sidecar query server and returns the ingester.
func (c Config) Initialize(appCodec codec.Codec, keepers domain.SQSIngestKeepers) (ingest.Ingester, error) {
	// Create redis client and ensure that it is up.
	redisAddress := fmt.Sprintf("%s:%s", c.StorageHost, c.StoragePort)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	redisStatus := redisClient.Ping(context.Background())
	_, err := redisStatus.Result()
	if err != nil {
		return nil, err
	}

	// Create pools ingester
	poolsIngester := poolsingester.NewPoolIngester(keepers)

	// Create sqs ingester that encapsulates all ingesters.
	sqsIngester := NewSidecarQueryServerIngester(poolsIngester, c.StorageHost, c.StoragePort, appCodec)

	return sqsIngester, nil
}
