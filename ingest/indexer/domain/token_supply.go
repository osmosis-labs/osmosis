package domain

import (
	"strings"

	"github.com/osmosis-labs/osmosis/osmomath"
)

type TokenSupply struct {
	Denom  string       `json:"denom"`
	Supply osmomath.Int `json:"supply"`
}

type TokenSupplyOffset struct {
	Denom        string       `json:"denom"`
	SupplyOffset osmomath.Int `json:"supply_offset"`
}

// ShouldFilterDenom returns true if the given denom should be filtered out.
func ShouldFilterDenom(denom string) bool {
	return strings.Contains(denom, "cl/pool") || strings.Contains(denom, "gamm/pool")
}
