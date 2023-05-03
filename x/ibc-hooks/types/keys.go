package types

const (
	ModuleName     = "ibchooks"
	RouterKey      = ModuleName
	StoreKey       = "hooks-for-ibc" // not using the module name because of collisions with key "ibc"
	IBCCallbackKey = "ibc_callback"
	IBCAsyncAckKey = "ibc_async_ack"
	SenderPrefix   = "ibc-wasm-hook-intermediary"
)
