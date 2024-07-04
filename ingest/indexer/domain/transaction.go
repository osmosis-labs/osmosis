package domain

import "time"

type Transaction struct {
	Height     uint64        `json:"height"`
	BlockTime  time.Time     `json:"timestamp"`
	Events     []interface{} `json:"events"`
	IngestedAt time.Time     `json:"ingested_at"`
}
