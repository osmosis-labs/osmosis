package passthroughdomain

import (
	"github.com/osmosis-labs/osmosis/v28/ingest/types/json" // TODO
)

// PoolFee represents the fees data of a pool.
type PoolFee struct {
	PoolID         string  `json:"-"`
	Volume24h      float64 `json:"volume_24h"`
	Volume7d       float64 `json:"volume_7d"`
	FeesSpent24h   float64 `json:"fees_spent_24h"`
	FeesSpent7d    float64 `json:"fees_spent_7d"`
	FeesPercentage string  `json:"fees_percentage"`
}

// UnmarshalJSON custom unmarshal method to handle PoolID as string.
func (p *PoolFee) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to unmarshal the data into.
	type Alias PoolFee
	temp := &struct {
		PoolID string `json:"pool_id"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	// Unmarshal the data into the temporary struct.
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	p.PoolID = temp.PoolID

	return nil
}

// PoolFees represents the fees data of the pools.
type PoolFees struct {
	LastUpdateAt int64     `json:"last_update_at"`
	Data         []PoolFee `json:"data"`
}
