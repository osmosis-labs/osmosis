package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	InboundTransfersKey   = []byte{0x01}
	FinalizedTransfersKey = []byte{0x02}
	LastHeightsKey        = []byte{0x03}

	KeySeparator = "|"
)

// InboundTransferKey returns the store prefix key where all the data
// associated with a specific InboundTransfer is stored
func InboundTransferKey(externalID string, externalHeight uint64) []byte {
	externalIDPrefix := append(InboundTransfersKey, []byte(externalID+KeySeparator)...)
	return binary.BigEndian.AppendUint64(externalIDPrefix, externalHeight)
}

// FinalizedTransferKey returns the store prefix key where all the data
// associated with a specific InboundTransfer is stored
func FinalizedTransferKey(externalID string) []byte {
	return append(FinalizedTransfersKey, []byte(externalID)...)
}

// LastHeightKey returns the store prefix key where all the data
// associated with a specific InboundTransfer is stored
func LastHeightKey(assetID AssetID) []byte {
	return append(LastHeightsKey, []byte(assetID.Name())...)
}
