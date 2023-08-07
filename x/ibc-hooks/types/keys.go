package types

const (
	ModuleName = "ibchooks"
	RouterKey  = ModuleName
	StoreKey   = "hooks-for-ibc" // not using the module name because of collisions with key "ibc"

	IBCCallbackKey = "ibc_callback"
	IBCAsyncAckKey = "ibc_async_ack"

	MsgEmitAckKey           = "emit_ack"
	AttributeSender         = "sender"
	AttributeChannel        = "channel"
	AttributePacketSequence = "sequence"

	SenderPrefix = "ibc-wasm-hook-intermediary"
)
