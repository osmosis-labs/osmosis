package domain

import (
	"time"
)

type Pair struct {
	PoolID     uint64    `json:"pool_id"`
	Denom0     string    `json:"denom_0"`
	IdxDenom0  uint8     `json:"idx_denom_0"`
	Denom1     string    `json:"denom_1"`
	IdxDenom1  uint8     `json:"idx_denom_1"`
	FeeBps     uint64    `json:"fee_bps"`
	IngestedAt time.Time `json:"ingested_at"`
}
