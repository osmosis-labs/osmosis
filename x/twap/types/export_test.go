package types

// validate validates the twap record, returns nil on success, error otherwise.
func (t TwapRecord) Validate() error {
	return t.validate()
}
