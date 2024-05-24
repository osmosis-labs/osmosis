package app

import (
	"github.com/osmosis-labs/osmosis/v25/app/keepers"
	"github.com/osmosis-labs/osmosis/v25/app/params"

	"github.com/cosmos/cosmos-sdk/std"
)

var encodingConfig params.EncodingConfig = MakeEncodingConfig()

func GetEncodingConfig() params.EncodingConfig {
	return encodingConfig
}

// MakeEncodingConfig creates an EncodingConfig.
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// UNFORKING v2 TODO: Verify that we no longer need to register legacy amino codec
	// keepers.AppModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	keepers.AppModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
