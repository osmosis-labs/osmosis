package ibc_rate_limit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v31/x/ibc-rate-limit/types"
)

// TestWrapContractError pins the mapping between what the rate limiter
// contract returns and the module error the chain reports. Only the
// contract's RateLimitExceded variant may surface as ErrRateLimitExceeded;
// anything else is a contract fault and must say so, otherwise a query
// failure inside the contract is shown to users as a rate limit.
func TestWrapContractError(t *testing.T) {
	quotaRejection := errors.New("execute wasm contract failed: IBC Rate Limit exceeded for any/ibc/ABCD. Tried to transfer 100 which exceeds the percentage capacity on the 'DAY-1' quota (0/50). Try again after Timestamp(1)")
	supplyFailure := errors.New("execute wasm contract failed: Generic error: Querier system error: Cannot parse request: invalid denom: ")
	channelBlocked := errors.New("execute wasm contract failed: Channel channel-1 has been blocked for denom transfer/channel-1/usat")

	cases := []struct {
		name       string
		in         error
		wantIs     error
		wantNotIs  error
		wantSubstr string
	}{
		{
			name:       "quota rejection maps to rate limit exceeded",
			in:         quotaRejection,
			wantIs:     types.ErrRateLimitExceeded,
			wantNotIs:  types.ErrContractError,
			wantSubstr: "'DAY-1' quota",
		},
		{
			name:       "failed supply query is a contract error, not a rate limit",
			in:         supplyFailure,
			wantIs:     types.ErrContractError,
			wantNotIs:  types.ErrRateLimitExceeded,
			wantSubstr: "invalid denom",
		},
		{
			name:       "denom restriction is a contract error, not a rate limit",
			in:         channelBlocked,
			wantIs:     types.ErrContractError,
			wantNotIs:  types.ErrRateLimitExceeded,
			wantSubstr: "has been blocked",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := wrapContractError(tc.in)
			require.ErrorIs(t, got, tc.wantIs)
			require.NotErrorIs(t, got, tc.wantNotIs)
			// The contract's own message must survive so the event is diagnosable.
			require.Contains(t, got.Error(), tc.wantSubstr)
		})
	}
}
