package domain

import (
	"context"
)

// Tokens represent the token's usecases
type TokensUsecase interface {
	GetDenomPrecisions(ctx context.Context) (map[string]int, error)
}

// Token represents the token's domain model
type Token struct {
	// ChainDenom is the denom used in the chain state.
	ChainDenom string `json:"chain_denom"`
	// HumanDenom is the human readable denom.
	HumanDenom string `json:"human_denom"`
	// Precision is the precision of the token.
	Precision int `json:"precision"`
}
