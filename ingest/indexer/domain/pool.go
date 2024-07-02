package domain

import (
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

type Pool struct {
	ChainModel poolmanagertypes.PoolI `json:"chain_model"`
}
