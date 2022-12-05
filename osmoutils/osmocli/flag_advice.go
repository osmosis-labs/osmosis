package osmocli

type FlagAdvice struct {
	HasPagination bool

	// Map of FieldName -> FlagName
	CustomFlagOverrides map[string]string

	// Tx sender value
	IsTx              bool
	TxSenderFieldName string
	FromValue         string
}
