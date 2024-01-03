package sqs

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v21/ingest"
	redischaininfoingester "github.com/osmosis-labs/osmosis/v21/ingest/sqs/chain_info/ingester/redis"
	redischaininforepository "github.com/osmosis-labs/osmosis/v21/ingest/sqs/chain_info/repository/redis"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/common"
	redispoolsingester "github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/ingester/redis"

	redispoolsrepository "github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/repository/redis"
	redisrouterrepository "github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/repository/redis"
)

// Config defines the config for the sidecar query server.
type Config struct {
	// IsEnabled defines if the sidecar query server is enabled.
	IsEnabled bool `mapstructure:"enabled"`

	// Storage defines the storage host and port.
	StorageHost string `mapstructure:"db-host"`
	StoragePort string `mapstructure:"db-port"`
}

const groupOptName = "osmosis-sqs"

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
func (c Config) Initialize(appCodec codec.Codec, keepers common.SQSIngestKeepers) (ingest.Ingester, error) {
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

	txManager := domain.NewTxManager(redisClient)

	redisPoolsRepository := redispoolsrepository.NewRedisPoolsRepo(appCodec, txManager)

	redisRouterRepository := redisrouterrepository.NewRedisRouterRepo(txManager)

	// Create pools ingester
	poolsIngester := redispoolsingester.NewPoolIngester(redisPoolsRepository, redisRouterRepository, txManager, domain.NewAssetListGetter(), keepers)

	// Create chain info ingester
	chainInfoRepository := redischaininforepository.NewChainInfoRepo(txManager)
	chainInfoingester := redischaininfoingester.NewChainInfoIngester(chainInfoRepository, txManager)

	// Create sqs ingester that encapsulates all ingesters.
	sqsIngester := NewSidecarQueryServerIngester(poolsIngester, chainInfoingester, txManager)

	return sqsIngester, nil
}
