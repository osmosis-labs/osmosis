package domain

import "github.com/osmosis-labs/osmosis/osmomath"

type Asset struct {
	// TBD
}

type TokenSupply struct {
	Denom  string
	Supply osmomath.Int
}
