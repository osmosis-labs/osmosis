package types

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

const (
	ModuleName   = "poolmanager"
	KeySeparator = "|"

	StoreKey = ModuleName

	RouterKey = ModuleName
)

var (
	// KeyNextGlobalPoolId defines key to store the next Pool ID to be used.
	KeyNextGlobalPoolId = []byte{0x01}

	// SwapModuleRouterPrefix defines prefix to store pool id to swap module mappings.
	SwapModuleRouterPrefix = []byte{0x02}

	// KeyPoolVolumePrefix defines prefix to store pool volume.
	KeyPoolVolumePrefix = []byte{0x03}
)

// ModuleRouteToBytes serializes moduleRoute to bytes.
func FormatModuleRouteKey(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", SwapModuleRouterPrefix, poolId))
}

// ParseModuleRouteFromBz parses the raw bytes into ModuleRoute.
// Returns error if fails to parse or if the bytes are empty.
func ParseModuleRouteFromBz(bz []byte) (ModuleRoute, error) {
	moduleRoute := ModuleRoute{}
	err := proto.Unmarshal(bz, &moduleRoute)
	if err != nil {
		return ModuleRoute{}, err
	}
	return moduleRoute, err
}

// KeyPoolVolume returns the key for the pool volume corresponding to the given poolId.
func KeyPoolVolume(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s", KeyPoolVolumePrefix, KeySeparator, poolId, KeySeparator))
}
