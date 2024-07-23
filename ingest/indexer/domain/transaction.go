package domain

import (
	"time"

	"github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// map data type is unsupported by the dataflow / apache beam
// thus using struct to represent the data
type EventWrapper struct {
	Index int         `json:"event_index"`
	Event types.Event `json:"event"`
}

type Transaction struct {
	Height             uint64         `json:"height"`
	BlockTime          time.Time      `json:"timestamp"`
	GasWanted          uint64         `json:"gas_wanted"`
	GasUsed            uint64         `json:"gas_used"`
	Fees               sdk.Coins      `json:"fees"`
	MessageType        string         `json:"msg_type"`
	TransactionHash    string         `json:"tx_hash"`
	TransactionIndexId int            `json:"tx_index_id"`
	Events             []EventWrapper `json:"events"`
	IngestedAt         time.Time      `json:"ingested_at"`
}
