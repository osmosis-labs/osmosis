package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IncentiveRecord is the high-level struct we use to deal with an independent incentive being distributed on a pool.
// Note that PoolId, Denom, and MinUptime are included in the key so we avoid storing them in state, hence the distinction
// between IncentiveRecord and IncentiveRecordBody.
type IncentiveRecord struct {
	PoolId uint64

	// incentive_denom is the denom of the token being distributed as part of this incentive record
	IncentiveDenom string

	// remaining_amount is the total amount of incentives to be distributed
	RemainingAmount sdk.Dec

	// emission_rate is the incentive emission rate per second
	EmissionRate sdk.Dec

	// start_time is the time when the incentive starts distributing
	StartTime time.Time

	// min_uptime is the minimum uptime required for liquidity to qualify for this incentive.
	// It should be always be one of the supported uptimes in types.SupportedUptimes
	MinUptime time.Duration
}
