package types

// ValidateBasic is a basic validation test on recordd distribution gauges' weights.
func (r DistrRecord) ValidateBasic() error {
	if r.Weight.IsNegative() {
		return ErrDistrRecordNotPositiveWeight
	}
	return nil
}
