package ibc_rate_limit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v31/x/ibc-rate-limit/types"
)

// TestWrapContractError pins the classification of contract errors: only the
// contract's own quota rejection becomes ErrRateLimitExceeded; anything else
// the contract fails with is ErrContractError. The message is kept in both.
func TestWrapContractError(t *testing.T) {
	quotaRejection := errors.New("execute wasm contract failed: IBC Rate Limit exceeded for any/ibc/ABCD. Tried to transfer 100 which exceeds the percentage capacity on the 'DAY-1' quota (0/50). Try again after Timestamp(1)")
	wrapped := wrapContractError(quotaRejection)
	require.ErrorIs(t, wrapped, types.ErrRateLimitExceeded)
	require.False(t, errors.Is(wrapped, types.ErrContractError))
	require.Contains(t, wrapped.Error(), "'DAY-1' quota")

	for _, msg := range []string{
		"execute wasm contract failed: Generic error: Querier contract error: codespace: sdk, code: 3",
		"execute wasm contract failed: Error parsing into type rate_limiter::msg::SudoMsg: unknown variant",
		"rate limit exceeded", // the module's own text is not the contract's marker
	} {
		wrapped := wrapContractError(errors.New(msg))
		require.ErrorIs(t, wrapped, types.ErrContractError, msg)
		require.False(t, errors.Is(wrapped, types.ErrRateLimitExceeded), msg)
		require.Contains(t, wrapped.Error(), msg)
	}
}
