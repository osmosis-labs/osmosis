package types

const (
	CounterKeyMissingRegisteredAuthenticator = "authenticator_missing_registered_authenticator"
	CounterKeyTrackFailed                    = "authenticator_track_failed"

	MeasureKeyAnteHandler = "authenticator_ante_handle"
	MeasureKeyPostHandler = "authenticator_post_handle"

	GaugeKeyAnteHandlerGasConsumed = "authenticator_ante_handler_gas_consumed"
	GaugeKeyPostHandlerGasConsumed = "authenticator_post_handler_gas_consumed"
)
