package sqs

import (
	"strings"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// Config defines the config for the sidecar query server.
type Config struct {
	// IsEnabled defines if the sidecar query server is enabled.
	IsEnabled bool `mapstructure:"enabled"`
	// GRPCIngestAddress defines the gRPC address of the sidecar query server ingest.
	GRPCIngestAddress []string `mapstructure:"grpc-ingest-address"`
	// GRPCIngestMaxCallSizeBytes defines the maximum size of a gRPC ingest call in bytes.
	GRPCIngestMaxCallSizeBytes int `mapstructure:"grpc-ingest-max-call-size-bytes"`
}

const (
	groupOptName = "osmosis-sqs"

	// This is the pool ID that is used for converting between UOSMO and USDC
	// for liquidity pricing.
	// https://app.osmosis.zone/pool/1263
	DefaultUSDCUOSMOPool = 1263
)

// DefaultConfig defines the default config for the sidecar query server.
var DefaultConfig = Config{
	IsEnabled: false,
	// Default gRPC address is localhost:50051
	GRPCIngestAddress: []string{"localhost:50051"},
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

	grpcIngestAddress := strings.Split(osmoutils.ParseString(opts, groupOptName, "grpc-ingest-address"), ",")

	grpcIngestMaxCallSizeBytes := osmoutils.ParseInt(opts, groupOptName, "grpc-ingest-max-call-size-bytes")

	return Config{
		IsEnabled:                  isEnabled,
		GRPCIngestAddress:          grpcIngestAddress,
		GRPCIngestMaxCallSizeBytes: grpcIngestMaxCallSizeBytes,
	}
}
