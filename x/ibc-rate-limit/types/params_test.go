package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateContractAddress(t *testing.T) {
	testCases := map[string]struct {
		addr     interface{}
		expected bool
	}{
		// ToDo: Why do tests expect the bech32 prefix to be cosmos?
		"valid_addr": {
			addr:     "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			expected: true,
		},
		"invalid_addr": {
			addr:     "cosmos1234",
			expected: false,
		},
		"invalid parameter type": {
			addr:     123456,
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validateContractAddress(tc.addr)

			// Assertions.
			if !tc.expected {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateParams(t *testing.T) {
	testCases := map[string]struct {
		addr     interface{}
		expected bool
	}{
		// ToDo: Why do tests expect the bech32 prefix to be cosmos?
		"valid_addr": {
			addr:     "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			expected: true,
		},
		"invalid_addr": {
			addr:     "cosmos1234",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			params := Params{
				ContractAddress: tc.addr.(string),
			}
			err := params.Validate()

			// Assertions.
			if !tc.expected {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
