package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "gamm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	// KeyNextGlobalPoolNumber defines key to store the next Pool ID to be used
	KeyNextGlobalPoolNumber = []byte{0x01}
	// KeyPrefixPools defines prefix to store pools
	KeyPrefixPools = []byte{0x02}
	// KeyTotalLiquidity defines key to store total liquidity
	KeyTotalLiquidity = []byte{0x03}
)

func MustGetPoolIdFromShareDenom(denom string) uint64 {
	numberStr := strings.TrimLeft(denom, "gamm/pool/")
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		panic(err)
	}
	return uint64(number)
}

func ValidatePoolShareDenom(denom string) error {
	numberStr := strings.TrimLeft(denom, "gamm/pool/")
	_, err := strconv.Atoi(numberStr)
	if err != nil {
		return err
	}
	return nil
}

func GetDenomPrefix(denom string) []byte {
	return append(KeyTotalLiquidity, []byte(denom)...)
}

func GetPoolShareDenom(poolId uint64) string {
	return fmt.Sprintf("gamm/pool/%d", poolId)
}

func GetKeyPrefixPools(poolId uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolId)...)
}
