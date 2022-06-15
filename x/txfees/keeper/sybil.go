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

// unused
// func (s Sybil) SetGasPrice(gasPrice sdk.Dec) Sybil {
// 	return Sybil{GasPrice: gasPrice, FeesPaid: s.FeesPaid}
// }
// func (s Sybil) SetFeesPaid(feesPaid sdk.Coin) Sybil {
// 	return Sybil{GasPrice: s.GasPrice, FeesPaid: feesPaid}
// }

func (s Sybil) AddToFeesPaid(feesPaid sdk.Coin) (Sybil, error) {
	// Check same denom
	if feesPaid.Denom != s.FeesPaid.Denom {
		return Sybil{}, fmt.Errorf("Cannot add %s denom to sybil's %s fees paid denom", feesPaid.Denom, s.FeesPaid.Denom)
	}
	// Add tokens
	fp := s.FeesPaid.Add(feesPaid)
	// Return new sybil with tokens added together & previous gas price
	return Sybil{GasPrice: s.GasPrice, FeesPaid: fp}, nil
}
