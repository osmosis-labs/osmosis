package types

import (
	"fmt"
	"sort"

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

	// DenomTradePairPrefix defines prefix to store denom trade pair for taker fee.
	DenomTradePairPrefix = []byte{0x04}
)

// ModuleRouteToBytes serializes moduleRoute to bytes.
func FormatModuleRouteKey(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", SwapModuleRouterPrefix, poolId))
}

// FormatDenomTradePairKey serializes denom trade pair to bytes.
// Denom trade pair is automatically sorted lexicographically.
func FormatDenomTradePairKey(denom0, denom1 string) []byte {
	denoms := []string{denom0, denom1}
	sort.Strings(denoms)
	return []byte(fmt.Sprintf("%s%s%s%s%s", DenomTradePairPrefix, KeySeparator, denoms[0], KeySeparator, denoms[1]))
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
