# Events

Section describes the module events

The module emits the following proto-events

| Source type | Source name          | Protobuf  reference                                                                  |
| ----------- | -------------------- |--------------------------------------------------------------------------------------|
| Message     | `MsgRequestCallback` | [CallbackRegisteredEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L12)       |
| Message     | `MsgCancelCallback`  | [CallbackCancelledEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L28)        |
| Module      | `EndBlocker`         | [CallbackExecutedSuccessEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L46)  |
| Module      | `EndBlocker`         | [CallbackExecutedFailedEvent](../../../proto/osmosis/callback/v1beta1/events.proto#L59)   |
