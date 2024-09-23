package domain

import (
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
)

type TokenSupply struct {
	Denom      string       `json:"denom"`
	Supply     osmomath.Int `json:"supply"`
	IngestedAt time.Time    `json:"ingested_at"`
}

type TokenSupplyOffset struct {
	Denom        string       `json:"denom"`
	SupplyOffset osmomath.Int `json:"supply_offset"`
	IngestedAt   time.Time    `json:"ingested_at"`
}
