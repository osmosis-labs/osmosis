package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SybilResistantGas struct {
	GasPrice sdk.Dec
	FeesPaid sdk.Coin
}

func NewSybilResistantGas(gasPrice sdk.Dec, feesPaid sdk.Coin) SybilResistantGas {
	return SybilResistantGas{
		GasPrice: gasPrice,
		FeesPaid: feesPaid,
	}
}

func (s SybilResistantGas) SetGasPrice(gasPrice sdk.Dec) SybilResistantGas {
	return SybilResistantGas{GasPrice: gasPrice, FeesPaid: s.FeesPaid}
}
func (s SybilResistantGas) SetFeesPaid(feesPaid sdk.Coin) SybilResistantGas {
	return SybilResistantGas{GasPrice: s.GasPrice, FeesPaid: feesPaid}
}
func (s SybilResistantGas) AddToFeesPaid(feesPaid sdk.Coin) SybilResistantGas {
	return SybilResistantGas{GasPrice: s.GasPrice, FeesPaid: s.FeesPaid.Add(feesPaid)}
}
