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
	// PoolsKey defines the store key for pools.
	PoolsKey = []byte{0x01}

	// CodeIdWhiteListKey defines the store key for code id whitelist.
	CodeIdWhiteListKey = []byte{0x02}
)

func FormatPoolsPrefix(poolId uint64) []byte {
	return append(PoolsKey, sdk.Uint64ToBigEndian(poolId)...)
}

func FormatCodeIdWhitelistPrefix(codeId uint64) []byte {
	return append(CodeIdWhiteListKey, sdk.Uint64ToBigEndian(codeId)...)
}
