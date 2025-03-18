package passthroughdomain

// PoolAPRDataStatusWrap is a wrapper for PoolAPRData that includes status flags.
type PoolAPRDataStatusWrap struct {
	PoolAPR
	IsStale bool `json:"is_stale,omitempty"`
	IsError bool `json:"is_error,omitempty"`
}

// PoolFeesDataStatusWrap is a wrapper for PoolFeesData that includes status flags.
type PoolFeesDataStatusWrap struct {
	PoolFee
	IsStale bool `json:"is_stale,omitempty"`
	IsError bool `json:"is_error,omitempty"`
}
