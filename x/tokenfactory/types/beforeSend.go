package types

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

type SudoMsg struct {
	TrackBeforeSend TrackBeforeSendMsg `json:"track_before_send"`
	BlockBeforeSend BlockBeforeSendMsg `json:"block_before_send"`
}

type TrackBeforeSendMsg struct {
	From   string           `json:"from"`
	To     string           `json:"to"`
	Amount wasmvmtypes.Coin `json:"amount"`
}

type BlockBeforeSendMsg struct {
	From   string           `json:"from"`
	To     string           `json:"to"`
	Amount wasmvmtypes.Coin `json:"amount"`
}
