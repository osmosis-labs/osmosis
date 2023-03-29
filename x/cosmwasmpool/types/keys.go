package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "cosmwasmpool"

	StoreKey = ModuleName

	RouterKey = ModuleName
)

var (
	// PoolKey defines the store key for pools.
	PoolsKey = []byte{0x01}
)

func FormatPoolsPrefix(poolId uint64) []byte {
	return append(PoolsKey, sdk.Uint64ToBigEndian(poolId)...)
}
