package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "cosmwasmpool"

	StoreKey = ModuleName

	RouterKey = ModuleName
)

// PoolsKey defines the store key for pools.
var PoolsKey = []byte{0x01}

func FormatPoolsPrefix(poolId uint64) []byte {
	return append(PoolsKey, sdk.Uint64ToBigEndian(poolId)...)
}
