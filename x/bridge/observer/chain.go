package observer

import (
	"context"

	"cosmossdk.io/math"
)

type ChainId string
type Denom string

const (
	ChainIdOsmosis ChainId = "osmosis"
	ChainIdBitcoin ChainId = "bitcoin"

	DenomBitcoin Denom = "btc"
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

type Chain interface {
	// SignalInboundTransfer sends inbound transfer tx to the chain
	SignalInboundTransfer(context.Context, Transfer) error
	// ListenOutboundTransfer returns receive-only channel
	// with outbound transfer txs from the chain
	ListenOutboundTransfer() <-chan Transfer

	Start(context.Context) error
	Stop(context.Context) error
	// Height returns current height of the chain
	Height() (uint64, error)
	// ConfirmationsRequired returns number of the required confirmations
	// for the given asset
	ConfirmationsRequired() (uint64, error)
}
