package stableswap

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
)

var (
	defaultSwapFee            = sdk.MustNewDecFromStr("0.025")
	defaultExitFee            = sdk.ZeroDec()
	defaultPoolId             = uint64(1)
	defaultStableswapPoolParams = PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultTwoAssetScalingFactors = []uint64{1, 1}
	defaultFutureGovernor = ""

	twoStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("foo", 1000000000),
		sdk.NewInt64Coin("bar", 1000000000),
	)
)

func TestGetScaledPoolAmts(t *testing.T) {

	tests := map[string]struct {
		denoms     []string
		expReserves []sdk.Dec
		expErr  error
	}{
		// should this error?
		"pass in no denoms": {
			denoms: []string{},
			expReserves: []sdk.Dec{},
			expErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := NewStableswapPool(
				defaultPoolId,
				defaultStableswapPoolParams,
				twoStablePoolAssets,
				defaultTwoAssetScalingFactors,
				defaultFutureGovernor,
			)
			require.NoError(t, err, "test: %s", name)

			reserves, err := p.getScaledPoolAmts(tc.denoms...)

			if tc.expErr != nil {
				require.Error(t, err, "test: %s", name)
				require.Equal(t, tc.expErr, err, "test: %s", name)
				require.Equal(t, tc.expReserves, reserves, "test: %s", name)
				return
			}
			require.NoError(t, err, "test: %s", name)
			require.Equal(t, tc.expReserves, reserves)
		})
	}
}

func TestGetDescaledPoolAmts(t *testing.T) {

	tests := map[string]struct {
		denom     string
		amount 	  osmomath.BigDec
		expResult osmomath.BigDec
		expErr  error
	}{
		"pass in no denoms": {
			denom: "",
			amount: osmomath.ZeroDec(),
			expResult: osmomath.ZeroDec(),
			expErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := NewStableswapPool(
				defaultPoolId,
				defaultStableswapPoolParams,
				twoStablePoolAssets,
				defaultTwoAssetScalingFactors,
				defaultFutureGovernor,
			)
			require.NoError(t, err, "test: %s", name)

			reserves := p.getDescaledPoolAmt(tc.denom, tc.amount)

			/* TODO: consider adding error return to getDescaledPoolAmt
			if tc.expErr != nil {
				require.Error(t, err, "test: %s", name)
				require.Equal(t, tc.expErr, err, "test: %s", name)
				require.Equal(t, tc.expResult, reserves, "test: %s", name)
				return
			}
			*/
			require.NoError(t, err, "test: %s", name)
			require.Equal(t, tc.expResult, reserves)
		})
	}
}
