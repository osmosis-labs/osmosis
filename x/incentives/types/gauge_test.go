package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Validates that the gauge returns true if this is the last distributing epoch of the gauge.
// Assumes that this is called before updating the gauge's state at the end of the epoch.
func TestIsLastNonPerpetualDistribution(t *testing.T) {

	tests := map[string]struct {
		gauge    Gauge
		expected bool
	}{
		"last non-perpetual distribution": {
			gauge: Gauge{
				IsPerpetual:       false,
				FilledEpochs:      99,
				NumEpochsPaidOver: 100,
			},
			expected: true,
		},
		"false because perpetual": {
			gauge: Gauge{
				IsPerpetual:       true,
				FilledEpochs:      99,
				NumEpochsPaidOver: 100,
			},
			expected: false,
		},
		"false because not last": {
			gauge: Gauge{
				IsPerpetual:       false,
				FilledEpochs:      98,
				NumEpochsPaidOver: 100,
			},
			expected: false,
		},
		"true even though over-filled": {
			gauge: Gauge{
				IsPerpetual:       false,
				FilledEpochs:      100,
				NumEpochsPaidOver: 100,
			},
			expected: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			isLastNonPerpetualDistribution := tc.gauge.IsLastNonPerpetualDistribution()

			require.Equal(t, tc.expected, isLastNonPerpetualDistribution)
		})
	}
}
