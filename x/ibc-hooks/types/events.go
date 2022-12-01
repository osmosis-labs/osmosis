package types

const (
	EventErrNotifyingAck    = "ack listener failed to process the ack"
	EventErrUnsubscribing   = "listeners contract failed to unsubscribe all"
	AttributeKeyContract    = "listener_contract"
	AttributeKeyAck         = "acknowledgement"
	AttributeKeyFailureType = "failure_type"
)
