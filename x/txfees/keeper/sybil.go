package keeper

import (
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

func (s Sybil) SetGasPrice(gasPrice sdk.Dec) Sybil {
	return Sybil{GasPrice: gasPrice, FeesPaid: s.FeesPaid}
}
func (s Sybil) SetFeesPaid(feesPaid sdk.Coin) Sybil {
	return Sybil{GasPrice: s.GasPrice, FeesPaid: feesPaid}
}
func (s Sybil) AddToFeesPaid(feesPaid sdk.Coin) Sybil {
	return Sybil{GasPrice: s.GasPrice, FeesPaid: s.FeesPaid.Add(feesPaid)}
}
