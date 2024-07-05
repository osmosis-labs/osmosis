package domain

import (
	"time"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

type Pool struct {
	ChainModel poolmanagertypes.PoolI `json:"chain_model"`
	IngestedAt time.Time              `json:"ingested_at"`
}
