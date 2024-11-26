package domain

import (
	"strings"
	"time"
)

// Pair represents a pair of tokens in a pool and message to be published to PubSub
type Pair struct {
	PoolID               uint64    `json:"pool_id"`
	MultiAsset           bool      `json:"multi_asset"`
	Denom0               string    `json:"denom_0"`
	IdxDenom0            uint8     `json:"idx_denom_0"`
	Denom1               string    `json:"denom_1"`
	IdxDenom1            uint8     `json:"idx_denom_1"`
	FeeBps               uint64    `json:"fee_bps"`
	IngestedAt           time.Time `json:"ingested_at"`
	PairCreatedAt        time.Time `json:"pair_created_at"`
	PairCreatedAtHeight  uint64    `json:"pair_created_at_height"`
	PairCreatedAtTxnHash string    `json:"pair_created_at_txn_hash"`
}

// ShouldFilterDenom returns true if the given denom should be filtered out.
func ShouldFilterDenom(denom string) bool {
	return denom == "" || strings.Contains(denom, "cl/pool") || strings.Contains(denom, "gamm/pool")
}

// IsMultiDenom returns true if the given denoms has >2 denoms
func IsMultiDenom(denoms []string) bool {
	return len(denoms) > 2
}
