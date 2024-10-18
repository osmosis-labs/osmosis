# Wasm Bindings

The only custom binding the module has is in the callback message which is sent to the contract during the execution of the callback.

```go
// SudoMsg callback message sent to a contract.
// This is encoded as JSON input to the contract when executing the callback
type SudoMsg struct {
	// Callback is the endpoint name at the contract which is called
	Callback *CallbackMsg `json:"callback,omitempty"`
}

// CallbackMsg is the callback message sent to a contract.
type CallbackMsg struct {
	// JobID is the user specified job id
	JobID uint64 `json:"job_id"`
}
```

The above struct is converted into the json encoded string in the following way.

```json
{"callback":{"job_id":1}}
```

## Requesting Callback

The contract can request a callback by using proto msg [MsgRequestCallback](./02_messages.md#msgrequestcallback)

## Cancelling Callback

The contract can cancel an existing callback by using proto msg [MsgCancelCallback](./02_messages.md#msgcancelcallback)
