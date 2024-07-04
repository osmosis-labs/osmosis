package domain

import "time"

type Pool struct {
	IngestedAt time.Time `json:"ingested_at"`
}
