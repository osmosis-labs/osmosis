package observer

import (
	"context"

	"cosmossdk.io/math"
)

type ChainId string

const (
	ChainId_OSMO    ChainId = "osmosis"
	ChainId_BITCOIN ChainId = "bitcoin"
)

type OutboundTransfer struct {
	DstChain ChainId
	Id       string
	Height   uint64
	Sender   string
	To       string
	Asset    string
	Amount   math.Uint
}

type InboundTransfer struct {
	SrcChain ChainId
	Id       string
	Height   uint64
	Sender   string
	To       string
	Asset    string
	Amount   math.Uint
}

type Chain interface {
	// Sends inbound transfer tx to the chain
	SignalInboundTransfer(ctx context.Context, in InboundTransfer) error
	// Returns receive-only channel with outbound transfer txs from the chain
	ListenOutboundTransfer() <-chan OutboundTransfer

	Start(ctx context.Context) error
	Stop() error
	// Returns current height of the chain
	Height() (uint64, error)
	// Returns number of the required tx confirmations
	ConfirmationsRequired() (uint64, error)
}
