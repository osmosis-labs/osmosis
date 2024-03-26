package types

import (
	"cosmossdk.io/math"
)

func NewInboundTransfer(
	externalID string,
	externalHeight uint64,
	destAddr string,
	assetID AssetID,
	amount math.Int,
) InboundTransfer {
	return InboundTransfer{
		ExternalId:     externalID,
		ExternalHeight: externalHeight,
		DestAddr:       destAddr,
		AssetId:        assetID,
		Amount:         amount,
		Voters:         make([]string, 0),
		Finalized:      false,
	}
}
