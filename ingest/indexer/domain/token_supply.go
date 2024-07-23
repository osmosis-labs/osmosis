package domain

import (
	"strings"
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

// ShouldFilterDenom returns true if the given denom should be filtered out.
func ShouldFilterDenom(denom string) bool {
	return strings.Contains(denom, "cl/pool") || strings.Contains(denom, "gamm/pool")
}

// IsMultiDenom returns true if the given denoms has >2 denoms
func IsMultiDenom(denoms []string) bool {
	return len(denoms) > 2
}
