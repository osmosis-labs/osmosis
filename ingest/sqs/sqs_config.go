package sqs

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v21/ingest"
	redischaininfoingester "github.com/osmosis-labs/osmosis/v21/ingest/sqs/chain_info/ingester/redis"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	sqslog "github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/common"
	redispoolsingester "github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/ingester/redis"
)

// Config defines the config for the sidecar query server.
type Config struct {
	// IsEnabled defines if the sidecar query server is enabled.
	IsEnabled bool `mapstructure:"enabled"`

	// Storage defines the storage host and port.
	StorageHost string `mapstructure:"db-host"`
	StoragePort string `mapstructure:"db-port"`

	// Defines the web server configuration.
	ServerAddress             string `mapstructure:"server-address"`
	ServerTimeoutDurationSecs int    `mapstructure:"timeout-duration-secs"`

	// Defines the logger configuration.
	LoggerFilename     string `mapstructure:"logger-filename"`
	LoggerIsProduction bool   `mapstructure:"logger-is-production"`
	LoggerLevel        string `mapstructure:"logger-level"`

	ChainGRPCGatewayEndpoint string `mapstructure:"grpc-gateway-endpoint"`

	// Router encapsulates the router config.
	Router *domain.RouterConfig `mapstructure:"router"`
}

const groupOptName = "osmosis-sqs"

// DefaultConfig defines the default config for the sidecar query server.
var DefaultConfig = Config{

	IsEnabled: false,

	StorageHost: "localhost",
	StoragePort: "6379",

	ServerAddress:             ":9092",
	ServerTimeoutDurationSecs: 2,

	LoggerFilename:     "sqs.log",
	LoggerIsProduction: true,
	LoggerLevel:        "info",

	ChainGRPCGatewayEndpoint: "http://localhost:26657",

	Router: &domain.RouterConfig{
		PreferredPoolIDs:          []uint64{},
		MaxPoolsPerRoute:          4,
		MaxRoutes:                 5,
		MaxSplitRoutes:            3,
		MaxSplitIterations:        10,
		MinOSMOLiquidity:          10000, // 10_000 OSMO
		RouteUpdateHeightInterval: 0,
		RouteCacheEnabled:         false,
		RouteCacheExpirySeconds:   600, // 10 minutes
	},
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

		ServerAddress:             osmoutils.ParseString(opts, groupOptName, "server-address"),
		ServerTimeoutDurationSecs: osmoutils.ParseInt(opts, groupOptName, "timeout-duration-secs"),

		LoggerFilename:     osmoutils.ParseString(opts, groupOptName, "logger-filename"),
		LoggerIsProduction: osmoutils.ParseBool(opts, groupOptName, "logger-is-production", false),
		LoggerLevel:        osmoutils.ParseString(opts, groupOptName, "logger-level"),

		ChainGRPCGatewayEndpoint: osmoutils.ParseString(opts, groupOptName, "grpc-gateway-endpoint"),

		Router: &domain.RouterConfig{
			PreferredPoolIDs: osmoutils.ParseUint64Slice(opts, groupOptName, "preferred-pool-ids"),

			MaxPoolsPerRoute: osmoutils.ParseInt(opts, groupOptName, "max-pools-per-route"),

			MaxRoutes: osmoutils.ParseInt(opts, groupOptName, "max-routes"),

			MaxSplitRoutes: osmoutils.ParseInt(opts, groupOptName, "max-split-routes"),

			MaxSplitIterations: osmoutils.ParseInt(opts, groupOptName, "max-split-iterations"),

			MinOSMOLiquidity: osmoutils.ParseInt(opts, groupOptName, "min-osmo-liquidity"),

			RouteUpdateHeightInterval: osmoutils.ParseInt(opts, groupOptName, "route-update-height-interval"),

			RouteCacheEnabled: osmoutils.ParseBool(opts, groupOptName, "route-cache-enabled", false),

			RouteCacheExpirySeconds: uint64(osmoutils.ParseInt(opts, groupOptName, "route-cache-expiry-seconds")),
		},
	}
}

// Initialize initializes the sidecar query server and returns the ingester.
func (c Config) Initialize(appCodec codec.Codec, keepers common.SQSIngestKeepers) (ingest.Ingester, error) {
	// logger
	logger, err := sqslog.NewLogger(c.LoggerIsProduction, c.LoggerFilename, c.LoggerLevel)
	if err != nil {
		return nil, fmt.Errorf("error while creating logger: %s", err)
	}
	logger.Info("Starting sidecar query server")

	// Create sidecar query server
	sidecarQueryServer, err := NewSideCarQueryServer(
		appCodec,
		*c.Router,
		c.StorageHost,
		c.StoragePort,
		c.ServerAddress,
		c.ChainGRPCGatewayEndpoint,
		c.ServerTimeoutDurationSecs,
		logger)
	if err != nil {
		return nil, fmt.Errorf("error while creating sidecar query server: %s", err)
	}

	txManager := sidecarQueryServer.GetTxManager()

	// Create pools ingester
	poolsIngester := redispoolsingester.NewPoolIngester(sidecarQueryServer.GetPoolsRepository(), sidecarQueryServer.GetRouterRepository(), sidecarQueryServer.GetTokensUseCase(), txManager, *c.Router, keepers)
	poolsIngester.SetLogger(sidecarQueryServer.GetLogger())

	chainInfoingester := redischaininfoingester.NewChainInfoIngester(sidecarQueryServer.GetChainInfoRepository(), txManager)
	chainInfoingester.SetLogger(sidecarQueryServer.GetLogger())

	// Create sqs ingester that encapsulates all ingesters.
	sqsIngester := NewSidecarQueryServerIngester(poolsIngester, chainInfoingester, txManager)

	return sqsIngester, nil
}
