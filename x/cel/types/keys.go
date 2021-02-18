package types

import (
	"encoding/binary"
)

const (
	// ModuleName defines the module name
	ModuleName = "cel"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"
)

func KeyCell(cellID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, cellID)
	return append([]byte{0x00}, bz...)
}

func KeyExpr(cellID, exprID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, exprID)
	return append(KeyCell(cellID), bz...)
}
