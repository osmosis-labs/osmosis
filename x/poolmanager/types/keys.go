package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
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

	// KeyTakerFeeStakersProtoRev defines key to store the taker fee for stakers tracker.
	// Deprecated: Now utilizes KeyTakerFeeStakersProtoRevArray.
	KeyTakerFeeStakersProtoRev = []byte{0x05}

	// KeyTakerFeeCommunityPoolProtoRev defines key to store the taker fee for community pool tracker.
	// Deprecated: Now utilizes KeyTakerFeeCommunityPoolProtoRevArray.
	KeyTakerFeeCommunityPoolProtoRev = []byte{0x06}

	// KeyTakerFeeProtoRevAccountingHeight defines key to store the accounting height for the above taker fee trackers.
	KeyTakerFeeProtoRevAccountingHeight = []byte{0x07}

	// KeyTakerFeeStakersProtoRevArray defines key to store the taker fee for stakers tracker coin array.
	KeyTakerFeeStakersProtoRevArray = []byte{0x08}

	// KeyTakerFeeCommunityPoolProtoRevArray defines key to store the taker fee for community pool tracker coin array.
	KeyTakerFeeCommunityPoolProtoRevArray = []byte{0x09}

	// TakerFeeSkimAccrualPrefix defines the prefix to store taker fee skim accrual data.
	TakerFeeSkimAccrualPrefix = []byte{0x0A}

	// KeyTakerFeeShare defines the key to store taker fee share data.
	KeyTakerFeeShare = []byte{0x0B}

	// KeyRegisteredAlloyPool defines the key to store registered alloy pool data.
	KeyRegisteredAlloyPool = []byte{0x0C}

	// KeyRevenueShareUser defines the key to store user revenue share registrations.
	KeyRevenueShareUser = []byte{0x0D}

	// KeyRevenueShareChildCounter defines the key to store user revenue subscriptions.
	KeyRevenueShareChildCounter = []byte{0x0E}
)

// ModuleRouteToBytes serializes moduleRoute to bytes.
func FormatModuleRouteKey(poolId uint64) []byte {
	// Estimate the length of the string representation of poolId
	// 11 is a very safe upper bound, (99,999,999,999) pools, and is a 12 byte allocation
	length := 11
	result := make([]byte, 1, 1+length)
	result[0] = SwapModuleRouterPrefix[0]
	// Write poolId into the byte slice starting after the prefix
	written := strconv.AppendUint(result[1:], poolId, 10)

	// Slice result to the actual length used
	return result[:1+len(written)]
}

// FormatDenomTradePairKey serializes denom trade pair to bytes.
// Denom trade pair order matters.
func FormatDenomTradePairKey(tokenInDenom, tokenOutDenom string) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s%s%s%s%s", DenomTradePairPrefix, KeySeparator, tokenInDenom, KeySeparator, tokenOutDenom)
	return buffer.Bytes()
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

// ParseDenomTradePairKey parses the raw bytes of the DenomTradePairKey into a denom trade pair.
func ParseDenomTradePairKey(key []byte) (tokenInDenom, tokenOutDenom string, err error) {
	keyStr := string(key)
	parts := strings.Split(keyStr, KeySeparator)

	tokenInDenom = parts[1]
	tokenOutDenom = parts[2]

	err = sdk.ValidateDenom(tokenInDenom)
	if err != nil {
		return "", "", err
	}

	err = sdk.ValidateDenom(tokenOutDenom)
	if err != nil {
		return "", "", err
	}

	return tokenInDenom, tokenOutDenom, nil
}

// KeyTakerFeeShareDenomAccrualForTakerFeeChargedDenom generates a key for a specific taker fee share denomination and taker fee charged denomination.
// The key is used to store and retrieve the accrued value of the taker fee denomination for the given taker fee share denom.
func KeyTakerFeeShareDenomAccrualForTakerFeeChargedDenom(takerFeeShareDenom string, takerFeeChargedDenom string) []byte {
	return []byte(fmt.Sprintf("%s%s%s%s%s", TakerFeeSkimAccrualPrefix, KeySeparator, takerFeeShareDenom, KeySeparator, takerFeeChargedDenom))
}

// KeyTakerFeeShareDenomAccrualForAllDenoms generates a key for a specific taker fee share denomination.
// The key is used to store and retrieve the accrued value for all taker fee charged denominations for the given taker fee share denomination.
func KeyTakerFeeShareDenomAccrualForAllDenoms(takerFeeShareDenom string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", TakerFeeSkimAccrualPrefix, KeySeparator, takerFeeShareDenom))
}

// FormatTakerFeeShareAgreementKey generates a key for a specific denomination.
// The key is used to store and retrieve the TakerFeeShareAgreement for the given denomination.
func FormatTakerFeeShareAgreementKey(denom string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", KeyTakerFeeShare, KeySeparator, denom))
}

// FormatRegisteredAlloyPoolKey generates a key for a registered alloy pool with a specific pool ID and alloyed denomination.
// The key is used to store and retrieve the registered alloy pool data.
func FormatRegisteredAlloyPoolKey(poolId uint64, alloyedDenom string) []byte {
	return []byte(fmt.Sprintf("%s%s%d%s%s", KeyRegisteredAlloyPool, KeySeparator, poolId, KeySeparator, alloyedDenom))
}

// FormatRegisteredAlloyPoolKeyPoolIdOnly generates a key for a registered alloy pool with a specific pool ID.
// The key is used to store and retrieve the registered alloy pool data without specifying the alloyed denomination.
func FormatRegisteredAlloyPoolKeyPoolIdOnly(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%s%d", KeyRegisteredAlloyPool, KeySeparator, poolId))
}

func FormatRevenueShareAddressKey(user sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s%s%d", KeyRevenueShareUser, KeySeparator, user.String()))
}

func FormatRevenueShareChildCounterKey(user sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s%s%d", KeyRevenueShareUser, KeySeparator, user.String()))
}
