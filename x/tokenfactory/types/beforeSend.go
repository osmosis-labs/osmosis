package types

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

type SudoMsg struct {
	BeforeSend BeforeSendMsg `json:"before_send"`
}

type BeforeSendMsg struct {
	From   string            `json:"from"`
	To     string            `json:"to"`
	Amount wasmvmtypes.Coins `json:"amount"`
}
