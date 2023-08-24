package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrBadMetadataFormatMsg = "wasm metadata not properly formatted for: '%v'. %s"
	ErrBadExecutionMsg      = "cannot execute contract: %v"

	ErrMsgValidation       = errorsmod.Register("wasm-hooks", 2, "error in wasmhook message validation")
	ErrMarshaling          = errorsmod.Register("wasm-hooks", 3, "cannot marshal the ICS20 packet")
	ErrInvalidPacket       = errorsmod.Register("wasm-hooks", 4, "invalid packet data")
	ErrBadResponse         = errorsmod.Register("wasm-hooks", 5, "cannot create response")
	ErrWasmError           = errorsmod.Register("wasm-hooks", 6, "wasm error")
	ErrBadSender           = errorsmod.Register("wasm-hooks", 7, "bad sender")
	ErrAckFromContract     = errorsmod.Register("wasm-hooks", 8, "contract returned error ack")
	ErrAsyncAckNotAllowed  = errorsmod.Register("wasm-hooks", 9, "contract not allowed to send async acks")
	ErrAckPacketMismatch   = errorsmod.Register("wasm-hooks", 10, "packet does not match the expected packet")
	ErrInvalidContractAddr = errorsmod.Register("wasm-hooks", 11, "invalid contract address")
)
