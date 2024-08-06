package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestConstraintClamp(t *testing.T) {
	taxPolicy := DefaultTaxPolicy
	prevRate := DefaultTaxRate

	// Case 1: try to update delta > maxUpdateRate
	newRate := prevRate.Add(taxPolicy.ChangeRateMax.MulInt64(2))
	clampedRate := taxPolicy.Clamp(prevRate, newRate)
	require.Equal(t, prevRate.Add(taxPolicy.ChangeRateMax), clampedRate)

	// Case 2: try to update delta > maxUpdateRate in other direction
	newRate = prevRate.Sub(taxPolicy.ChangeRateMax.MulInt64(2))
	clampedRate = taxPolicy.Clamp(prevRate, newRate)
	require.Equal(t, prevRate.Sub(taxPolicy.ChangeRateMax), clampedRate)

	// Case 3: try to update the new rate > maxRate
	prevRate = taxPolicy.RateMax
	newRate = taxPolicy.RateMax.Add(sdk.NewDecWithPrec(1, 3))
	clampedRate = taxPolicy.Clamp(prevRate, newRate)
	require.Equal(t, taxPolicy.RateMax, clampedRate)

	// Case 4: try to update the new rate < minRate
	prevRate = taxPolicy.RateMin
	newRate = taxPolicy.RateMin.Sub(sdk.NewDecWithPrec(1, 3))
	clampedRate = taxPolicy.Clamp(prevRate, newRate)
	require.Equal(t, taxPolicy.RateMin, clampedRate)
}
