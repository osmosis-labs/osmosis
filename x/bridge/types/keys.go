package types

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

var InboundTransfersKey = []byte{0x01}

// InboundTransferKey returns the store prefix key where all the data
// associated with a specific InboundTransfer is stored
func InboundTransferKey(externalID string) []byte {
	return append(InboundTransfersKey, []byte(externalID)...)
}
