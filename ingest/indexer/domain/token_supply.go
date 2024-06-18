package domain

import "github.com/osmosis-labs/osmosis/osmomath"

type TokenSupply struct {
	Denom  string
	Supply osmomath.Int
}

type TokenSupplyOffset struct {
	Denom        string
	SupplyOffset osmomath.Int
}
