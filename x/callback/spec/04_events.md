# Events

Section describes the module events

The module emits the following proto-events

| Source type | Source name          | Protobuf  reference                                                                  |
| ----------- | -------------------- |--------------------------------------------------------------------------------------|
| Message     | `MsgRequestCallback` | [CallbackRegisteredEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L11)       |
| Message     | `MsgCancelCallback`  | [CallbackCancelledEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L25)        |
| Module      | `EndBlocker`         | [CallbackExecutedSuccessEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L39)  |
| Module      | `EndBlocker`         | [CallbackExecutedFailedEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L53)   |
