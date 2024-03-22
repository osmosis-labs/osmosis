package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestParamsEqual(t *testing.T) {
	p1 := DefaultParams()
	err := p1.Validate()
	require.NoError(t, err)

	// invalid base pool
	p1.BasePool = sdk.NewDec(-1)
	err = p1.Validate()
	require.Error(t, err)

	// invalid pool recovery period
	p3 := DefaultParams()
	p3.PoolRecoveryPeriod = 0
	err = p3.Validate()
	require.Error(t, err)

	// invalid min spread
	p4 := DefaultParams()
	p4.MinStabilitySpread = sdk.NewDecWithPrec(-1, 2)
	err = p4.Validate()
	require.Error(t, err)

	p5 := DefaultParams()
	require.NotNil(t, p5.ParamSetPairs())
	require.NotNil(t, p5.String())
}
