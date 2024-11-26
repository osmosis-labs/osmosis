package app

import (
	"github.com/osmosis-labs/osmosis/v27/app/keepers"
	"github.com/osmosis-labs/osmosis/v27/app/params"

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
	keepers.AppModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
