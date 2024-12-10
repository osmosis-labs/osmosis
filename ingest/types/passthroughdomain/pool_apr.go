package sqspassthroughdomain

import (
	"fmt"
	"strconv"

	"github.com/osmosis-labs/sqs/sqsdomain/json"
)

// PoolDataRange represents the range of the pool APR data.
type PoolDataRange struct {
	Lower float64 `json:"lower,omitempty"`
	Upper float64 `json:"upper,omitempty"`
}

// PoolAPR represents the APR data of the pool.
type PoolAPR struct {
	// PoolID represents the pool ID.
	PoolID uint64 `json:"-"`
	// Swap Fees represents the swap fees.
	SwapFees PoolDataRange `json:"swap_fees,omitempty"`
	// Superfluid APR represents the superfluid APR.
	SuperfluidAPR PoolDataRange `json:"superfluid,omitempty"`
	// Osmosis APR represents the osmosis APR.
	OsmosisAPR PoolDataRange `json:"osmosis,omitempty"`
	// Boost APR represents the boosted APR.
	BoostAPR PoolDataRange `json:"boost,omitempty"`
	// Total APR represents the total APR.
	TotalAPR PoolDataRange `json:"total_apr,omitempty"`
}

// UnmarshalJSON custom unmarshal method to handle PoolID as uint64.
func (p *PoolAPR) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to unmarshal the data into.
	type Alias PoolAPR
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

	// Convert the PoolID from string to uint64.
	id, err := strconv.ParseUint(temp.PoolID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid pool_id: %w", err)
	}
	p.PoolID = id

	return nil
}
