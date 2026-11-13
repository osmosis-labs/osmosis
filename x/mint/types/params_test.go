package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/x/mint/types"
)

// TestValidateDistributionProportions checks that the proportions must sum to one
// and that the deprecated developer rewards proportion must be zero.
func TestValidateDistributionProportions(t *testing.T) {
	tests := map[string]struct {
		proportions types.DistributionProportions
		expectErr   bool
	}{
		"valid": {
			proportions: types.DistributionProportions{
				Staking:          osmomath.NewDecWithPrec(4, 1),
				PoolIncentives:   osmomath.NewDecWithPrec(3, 1),
				DeveloperRewards: osmomath.ZeroDec(),
				CommunityPool:    osmomath.NewDecWithPrec(3, 1),
			},
		},
		"non-zero developer rewards": {
			proportions: types.DistributionProportions{
				Staking:          osmomath.NewDecWithPrec(4, 1),
				PoolIncentives:   osmomath.NewDecWithPrec(3, 1),
				DeveloperRewards: osmomath.NewDecWithPrec(2, 1),
				CommunityPool:    osmomath.NewDecWithPrec(1, 1),
			},
			expectErr: true,
		},
		"sum not one": {
			proportions: types.DistributionProportions{
				Staking:          osmomath.NewDecWithPrec(4, 1),
				PoolIncentives:   osmomath.NewDecWithPrec(3, 1),
				DeveloperRewards: osmomath.ZeroDec(),
				CommunityPool:    osmomath.NewDecWithPrec(2, 1),
			},
			expectErr: true,
		},
		"negative staking": {
			proportions: types.DistributionProportions{
				Staking:          osmomath.NewDecWithPrec(-1, 1),
				PoolIncentives:   osmomath.NewDecWithPrec(8, 1),
				DeveloperRewards: osmomath.ZeroDec(),
				CommunityPool:    osmomath.NewDecWithPrec(3, 1),
			},
			expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			params := types.DefaultParams()
			params.DistributionProportions = tc.proportions

			err := params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
