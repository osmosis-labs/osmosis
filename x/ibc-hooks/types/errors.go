package types

var (
	ErrBadPacketMetadataMsg = "cannot unmarshal metadata: '%v'. %s"
	ErrBadMetadataFormatMsg = "wasm metadata not properly formatted for: '%v'. %s"
	ErrBadExecutionMsg      = "cannot execute contract: %v"
	ErrBadResponse          = "cannot create response: %v"
)
