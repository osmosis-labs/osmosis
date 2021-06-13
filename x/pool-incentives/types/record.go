package types

func (r DistrRecord) ValidateBasic() error {
	if !r.Weight.IsPositive() {
		return ErrDistrRecordNotPositiveWeight
	}
	return nil
}
