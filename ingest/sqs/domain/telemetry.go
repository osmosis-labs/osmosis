package domain

var (
	// sqs_sync_check_error
	//
	// counter that is increased if node sync check fails when processing the first block
	//
	// Has the following labels:
	// * err - the error returned
	// * height - the height of the block being processed
	SQSNodeSyncCheckErrorMetricName = "sqs_sync_check_error"

	// sqs_process_block_error
	//
	// counter that is increased if ingest process block fails with error.
	//
	// Has the following labels:
	// * msg - the error returned
	// * height - the height of the block being processed
	SQSProcessBlockErrorMetricName = "sqs_process_block_error"

	// sqs_process_block_panic
	//
	// counter that is increased if ingest process block fails with panic.
	//
	// Has the following labels:
	// * msg - the error returned
	// * height - the height of the block being processed
	SQSProcessBlockPanicMetricName = "sqs_process_block_panic"

	// sqs_process_block_duration
	//
	// histogram that measures the duration of processing a block
	//
	// Has the following labels:
	// * height - the height of the block being processed
	SQSProcessBlockDurationMetricName = "sqs_process_block_duration"

	// sqs_grpc_connection_error
	//
	// counter that is increased if grpc connection fails
	//
	// Has the following labels:
	// * err - the error returned
	// * height - the height of the block being processed
	SQSGRPCConnectionErrorMetricName = "sqs_grpc_connection_error"
)
