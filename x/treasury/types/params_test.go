package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestParams(t *testing.T) {
	params := DefaultParams()
	require.NoError(t, params.Validate())

	params = DefaultParams()
	params.TaxPolicy.RateMax = sdk.ZeroDec()
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.TaxPolicy.RateMin = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.TaxPolicy.Cap = sdk.Coin{Denom: "foo", Amount: sdk.NewInt(-1)}
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.TaxPolicy.ChangeRateMax = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.RewardPolicy.RateMax = sdk.ZeroDec()
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.RewardPolicy.ChangeRateMax = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.SeigniorageBurdenTarget = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.MiningIncrement = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.WindowLong = 0
	require.Error(t, params.Validate())

	params = DefaultParams()
	params.RewardPolicy.RateMin = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	require.NotNil(t, params.ParamSetPairs())
	require.NotNil(t, params.String())
}
