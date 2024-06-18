package sqs

import (
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	indexerservice "github.com/osmosis-labs/osmosis/v25/ingest/indexer/service"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	poolstransformer "github.com/osmosis-labs/osmosis/v25/ingest/sqs/pools/transformer"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/service"
)

// Config defines the config for the sidecar query server.
type Config struct {
	// IsEnabled defines if the sidecar query server is enabled.
	IsEnabled bool `mapstructure:"enabled"`
	// GRPCIngestAddress defines the gRPC address of the sidecar query server ingest.
	GRPCIngestAddress string `mapstructure:"grpc-ingest-address"`
	// GRPCIngestMaxCallSizeBytes defines the maximum size of a gRPC ingest call in bytes.
	GRPCIngestMaxCallSizeBytes int `mapstructure:"grpc-ingest-max-call-size-bytes"`
}

const (
	groupOptName = "osmosis-sqs"

	// This is the pool ID that is used for converting between UOSMO and USDC
	// for liquidity pricing.
	// https://app.osmosis.zone/pool/1263
	defaultUSDCUOSMOPool = 1263
)

// DefaultConfig defines the default config for the sidecar query server.
var DefaultConfig = Config{
	IsEnabled: false,
	// Default gRPC address is localhost:50051
	GRPCIngestAddress: "localhost:50051",
	// 50 MB by default. Our pool data is estimated to be at approximately 15MB.
	// During normal operation, we should not approach even 1 MB since we are to stream only
	// modified pools.
	GRPCIngestMaxCallSizeBytes: 50 * 1024 * 1024,
}

// NewConfigFromOptions returns a new sidecar query server config from the given options.
func NewConfigFromOptions(opts servertypes.AppOptions) Config {
	isEnabled := osmoutils.ParseBool(opts, groupOptName, "is-enabled", false)

	if !isEnabled {
		return Config{
			IsEnabled: false,
		}
	}

	grpcIngestAddress := osmoutils.ParseString(opts, groupOptName, "grpc-ingest-address")

	grpcIngestMaxCallSizeBytes := osmoutils.ParseInt(opts, groupOptName, "grpc-ingest-max-call-size-bytes")

	return Config{
		IsEnabled:                  isEnabled,
		GRPCIngestAddress:          grpcIngestAddress,
		GRPCIngestMaxCallSizeBytes: grpcIngestMaxCallSizeBytes,
	}
}

// Initialize initializes the sidecar query server and returns the ingester.
func (c Config) Initialize(appCodec codec.Codec, keepers domain.SQSIngestKeepers) (domain.Ingester, error) {
	// Create pools ingester
	poolsIngester := poolstransformer.NewPoolTransformer(keepers, defaultUSDCUOSMOPool)

	// Create sqs grpc client
	sqsGRPCClient := service.NewGRPCCLient(c.GRPCIngestAddress, c.GRPCIngestMaxCallSizeBytes, appCodec)

	// Create indexers pubsub client
	pubSubClient := indexerservice.NewPubSubCLient("osmosis-data-team", "dev.block-topic")

	// Create sqs ingester that encapsulates all ingesters.
	sqsIngester := NewSidecarQueryServerIngester(poolsIngester, appCodec, keepers, sqsGRPCClient, pubSubClient)

	return sqsIngester, nil
}
