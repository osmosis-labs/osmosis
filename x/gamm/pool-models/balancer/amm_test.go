package balancer_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/internal/test_helpers"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

type BalancerTestSuite struct {
	test_helpers.CfmmCommonTestSuite
}

func TestBalancerTestSuite(t *testing.T) {
	suite.Run(t, new(BalancerTestSuite))
}

func TestBalancerPoolParams(t *testing.T) {
	// Tests that creating a pool with the given pair of spread factor and exit fee
	// errors or succeeds as intended. Furthermore, it checks that
	// NewPool panics in the error case.
	tests := []struct {
		SpreadFactor sdk.Dec
		ExitFee      sdk.Dec
		shouldErr    bool
	}{
		// Should work
		{defaultSpreadFactor, defaultZeroExitFee, noErr},
		// Can't set the spread factor as negative
		{sdk.NewDecWithPrec(-1, 2), defaultZeroExitFee, wantErr},
		// Can't set the spread factor as 1
		{sdk.NewDec(1), defaultZeroExitFee, wantErr},
		// Can't set the spread factor above 1
		{sdk.NewDecWithPrec(15, 1), defaultZeroExitFee, wantErr},
		// Can't set the exit fee as negative
		{defaultSpreadFactor, sdk.NewDecWithPrec(-1, 2), wantErr},
		// Can't set the exit fee as 1
		{defaultSpreadFactor, sdk.NewDec(1), wantErr},
		// Can't set the exit fee above 1
		{defaultSpreadFactor, sdk.NewDecWithPrec(15, 1), wantErr},
	}

	for i, params := range tests {
		PoolParams := balancer.PoolParams{
			SwapFee: params.SpreadFactor,
			ExitFee: params.ExitFee,
		}
		err := PoolParams.Validate(dummyPoolAssets)
		if params.shouldErr {
			require.Error(t, err, "unexpected lack of error, tc %v", i)
			// Check that these are also caught if passed to the underlying pool creation func
			_, err = balancer.NewBalancerPool(1, PoolParams, dummyPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
			require.Error(t, err)
		} else {
			require.NoError(t, err, "unexpected error, tc %v", i)
		}
	}
}

func (s *KeeperTestSuite) TestEnsureDenomInPool() {
	tests := map[string]struct {
		poolAssets  []balancer.PoolAsset
		tokensIn    sdk.Coins
		expectPass  bool
		expectedErr error
	}{
		"all of tokensIn is in pool asset map": {
			poolAssets:  []balancer.PoolAsset{defaultOsmoPoolAsset, defaultAtomPoolAsset},
			tokensIn:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.OneInt())),
			expectPass:  true,
			expectedErr: nil,
		},
		"one of tokensIn is in pool asset map": {
			poolAssets:  []balancer.PoolAsset{defaultOsmoPoolAsset, defaultAtomPoolAsset},
			tokensIn:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.OneInt()), sdk.NewCoin("foo", sdk.OneInt())),
			expectPass:  false,
			expectedErr: types.ErrDenomNotFoundInPool,
		},
		"none of tokensIn is in pool asset map": {
			poolAssets:  []balancer.PoolAsset{defaultOsmoPoolAsset, defaultAtomPoolAsset},
			tokensIn:    sdk.NewCoins(sdk.NewCoin("foo", sdk.OneInt())),
			expectPass:  false,
			expectedErr: types.ErrDenomNotFoundInPool,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.ResetTest()

			poolAssetsByDenom, err := balancer.GetPoolAssetsByDenom(tc.poolAssets)
			s.Require().NoError(err, "test: %s", name)

			err = balancer.EnsureDenomInPool(poolAssetsByDenom, tc.tokensIn)

			if tc.expectPass {
				s.Require().NoError(err, "test: %s", name)
			} else {
				s.Require().ErrorIs(err, tc.expectedErr, "test: %s", name)
			}
		})
	}
}
