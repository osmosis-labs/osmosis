package domain

import "time"

// Transaction represents a transaction in the block
// Events is a list of events that occurred in the transaction
// Different event types have different structures and attributes so we use interface{}
// TO DO: TxHash, TxnIndex, EventIndex to be added
type Transaction struct {
	Height     uint64        `json:"height"`
	BlockTime  time.Time     `json:"timestamp"`
	Events     []interface{} `json:"events"`
	IngestedAt time.Time     `json:"ingested_at"`
}
