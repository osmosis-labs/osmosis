package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Sybil struct {
	GasPrice sdk.Dec
	FeesPaid sdk.Coin
}

func NewSybil(gasPrice sdk.Dec, feesPaid sdk.Coin) Sybil {
	return Sybil{
		GasPrice: gasPrice,
		FeesPaid: feesPaid,
	}
}

func (s *Sybil) AddToFeesPaid(feesPaid sdk.Coin) error {
	// Check same denom
	if feesPaid.Denom != s.FeesPaid.Denom {
		return fmt.Errorf("Cannot add %s denom to sybil's %s fees paid denom", feesPaid.Denom, s.FeesPaid.Denom)
	}
	// Add tokens
	s.FeesPaid.Add(feesPaid)
	return nil
}
