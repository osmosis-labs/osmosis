package domain

import "time"

type Block struct {
	ChainId     string    `json:"chain_id"`
	Height      uint64    `json:"height"`
	BlockTime   time.Time `json:"timestamp"`
	GasConsumed uint64    `json:"gas_consumed"`
	IngestedAt  time.Time `json:"ingested_at"`
}
