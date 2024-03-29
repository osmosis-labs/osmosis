package observer

import (
	"context"

	"cosmossdk.io/math"

	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

type ChainId string

const (
	ChainIdOsmosis ChainId = "osmosis"
	ChainIdBitcoin ChainId = "bitcoin"
)

type Transfer struct {
	SrcChain ChainId
	DstChain ChainId
	Id       string
	Height   uint64
	Sender   string
	To       string
	Asset    string
	Amount   math.Uint
}

type Client interface {
	// SignalInboundTransfer sends inbound transfer tx to the chain
	SignalInboundTransfer(context.Context, Transfer) error
	// ListenOutboundTransfer returns receive-only channel
	// with outbound transfer txs from the chain
	ListenOutboundTransfer() <-chan Transfer

	Start(context.Context) error
	Stop(context.Context) error
	// Height returns current height of the chain
	Height(context.Context) (uint64, error)
	// ConfirmationsRequired returns number of the required confirmations
	// for the given asset
	ConfirmationsRequired(context.Context, bridgetypes.AssetID) (uint64, error)
}
