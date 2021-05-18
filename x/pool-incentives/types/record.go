package types

func (r DistrRecord) Validate() error {
	if !r.Weight.IsPositive() {
		return ErrDistrRecordNotPositiveWeight
	}
	return nil
}
