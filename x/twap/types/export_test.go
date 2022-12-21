package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func (t TwapRecord) Validate() error {
	return t.validate()
}

func TestValidatePeriod(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		period      interface{}
		expectedErr bool
	}{
		"valid_period": {
			period:      time.Hour * 48,
			expectedErr: false,
		},
		"negative_period": {
			period:      -time.Hour,
			expectedErr: true,
		},
		"invalid parameter type": {
			period:      123456,
			expectedErr: true,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := validatePeriod(tc.period)

			// Assertions.
			if tc.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
