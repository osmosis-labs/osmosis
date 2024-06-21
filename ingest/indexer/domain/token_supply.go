package domain

import "github.com/osmosis-labs/osmosis/osmomath"

type TokenSupply struct {
	Denom  string       `json:"denom"`
	Supply osmomath.Int `json:"supply"`
}

type TokenSupplyOffset struct {
	Denom        string       `json:"denom"`
	SupplyOffset osmomath.Int `json:"supply_offset"`
}
