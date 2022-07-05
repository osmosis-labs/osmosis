package balancer_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

func TestValidateWeight(t *testing.T) {
	testCases := []struct {
		name string

		poolAsset balancer.PoolAsset

		expErr bool
	}{
		{
			name: "negative pool asset",
			poolAsset: balancer.PoolAsset{
				Weight: sdk.NewIntFromBigInt(big.NewInt(-1)),
			},
			expErr: true,
		},
		{
			name: "zero pool asset",
			poolAsset: balancer.PoolAsset{
				Weight: sdk.NewIntFromBigInt(big.NewInt(0)),
			},
			expErr: true,
		},
		{
			name: "pool asset grater than 2 ^ 32",
			poolAsset: balancer.PoolAsset{
				Weight: sdk.NewIntFromUint64((1 << 32) + 1),
			},
			expErr: true,
		},
		{
			name: "pool asset equal to 2 ^ 32",
			poolAsset: balancer.PoolAsset{
				Weight: sdk.NewIntFromUint64(1 << 32),
			},
			expErr: true,
		},
		{
			name: "pool asset smaller than 2^32",
			poolAsset: balancer.PoolAsset{
				Weight: sdk.NewIntFromUint64((1 << 32) - 1),
			},
			expErr: false,
		},
		{
			name: "pool asset equal to 1000",
			poolAsset: balancer.PoolAsset{
				Weight: sdk.NewIntFromUint64(1000),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.poolAsset.ValidateWeight()
			if tc.expErr == true {
				require.Error(t, err, "invalid PoolAsset, there should be an error")
			} else {
				require.NoError(t, err, "valid PoolAsset, no error expected")
			}
		})
	}
}
