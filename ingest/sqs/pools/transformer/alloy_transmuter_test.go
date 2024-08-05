package poolstransformer_test

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sqscosmwasmpool "github.com/osmosis-labs/sqs/sqsdomain/cosmwasmpool"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	poolstransformer "github.com/osmosis-labs/osmosis/v25/ingest/sqs/pools/transformer"
)

type mockWasmKeeper struct {
	QueryRawFn   func(ctx sdk.Context, contractAddress sdk.AccAddress, key []byte) []byte
	QuerySmartFn func(ctx sdk.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error)
}

// QueryRaw implements commondomain.WasmKeeper.
func (m *mockWasmKeeper) QueryRaw(ctx sdk.Context, contractAddress sdk.AccAddress, key []byte) []byte {
	if m.QueryRawFn != nil {
		return m.QueryRawFn(ctx, contractAddress, key)
	}
	panic("unimplemented")
}

// QuerySmart implements commondomain.WasmKeeper.
func (m *mockWasmKeeper) QuerySmart(ctx sdk.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error) {
	if m.QuerySmartFn != nil {
		return m.QuerySmartFn(ctx, contractAddress, req)
	}
	panic("unimplemented")
}

var _ commondomain.WasmKeeper = &mockWasmKeeper{}

func (s *PoolTransformerTestSuite) TestUpdateAlloyedTransmuterPool() {
	s.Setup()

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(
		sdk.NewCoin(apptesting.DefaultTransmuterDenomA, osmomath.NewInt(100000000)),
		sdk.NewCoin(apptesting.DefaultTransmuterDenomB, osmomath.NewInt(100000000)),
	))

	pool := s.PrepareAlloyTransmuterPool(s.TestAccs[0], apptesting.AlloyTransmuterInstantiateMsg{
		PoolAssetConfigs:                []apptesting.AssetConfig{{Denom: apptesting.DefaultTransmuterDenomA, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomANormFactor)}, {Denom: apptesting.DefaultTransmuterDenomB, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomBNormFactor)}},
		AlloyedAssetSubdenom:            apptesting.DefaultAlloyedSubDenom,
		AlloyedAssetNormalizationFactor: osmomath.NewInt(apptesting.DefaultAlloyedDenomNormFactor),
		Admin:                           s.TestAccs[0].String(),
		Moderator:                       s.TestAccs[1].String(),
	})

	// Create OSMO / USDC pool
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

	// Initialize the pool ingester
	poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

	cosmWasmPoolModel := sqscosmwasmpool.CosmWasmPoolModel{}
	poolDenoms := []string{apptesting.DefaultTransmuterDenomA, apptesting.DefaultTransmuterDenomB}

	poolIngester.UpdateAlloyTransmuterInfo(s.Ctx, pool.GetId(), pool.GetAddress(), &cosmWasmPoolModel, &poolDenoms)

	alloyedDenom := fmt.Sprintf("factory/%s/alloyed/%s", pool.GetAddress(), apptesting.DefaultAlloyedSubDenom)

	// Check if the pool has been updated
	s.Equal(sqscosmwasmpool.CosmWasmPoolData{
		AlloyTransmuter: &sqscosmwasmpool.AlloyTransmuterData{
			AlloyedDenom: alloyedDenom,
			AssetConfigs: []sqscosmwasmpool.TransmuterAssetConfig{
				{Denom: apptesting.DefaultTransmuterDenomA, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomANormFactor)},
				{Denom: apptesting.DefaultTransmuterDenomB, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomBNormFactor)},
				{Denom: alloyedDenom, NormalizationFactor: osmomath.NewInt(apptesting.DefaultAlloyedDenomNormFactor)}},
			RateLimiterConfig: sqscosmwasmpool.AlloyedRateLimiter{
				StaticLimiterByDenomMap: map[string]sqscosmwasmpool.StaticLimiter{},
				ChangeLimiterByDenomMap: map[string]sqscosmwasmpool.ChangeLimiter{},
			},
		},
	}, cosmWasmPoolModel.Data)

	s.Equal([]string{
		apptesting.DefaultTransmuterDenomA,
		apptesting.DefaultTransmuterDenomB,
		alloyedDenom,
	}, poolDenoms)
}

func (s *PoolTransformerTestSuite) TestAlloyTransmuterListLimiters() {
	tests := []struct {
		name            string
		poolID          uint64
		contractAddress string
		mockQueryResult []byte
		mockQueryError  error
		expectedResult  sqscosmwasmpool.AlloyedRateLimiter
		expectedError   bool
	}{
		{
			name:            "Success with static limiters",
			poolID:          1,
			contractAddress: "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3",
			mockQueryResult: []byte(`{
    "limiters": [
        [
            ["denom1", "Static Limiter 1"],
            {"static_limiter": {"upper_limit": "0.2"}}
        ],
        [
            ["denom2", "Static Limiter 2"],
            {"static_limiter": {"upper_limit": "0.3"}}
        ]
    ]
}`),
			expectedResult: sqscosmwasmpool.AlloyedRateLimiter{
				StaticLimiterByDenomMap: map[string]sqscosmwasmpool.StaticLimiter{
					"denom1": {UpperLimit: "0.2"},
					"denom2": {UpperLimit: "0.3"},
				},
				ChangeLimiterByDenomMap: map[string]sqscosmwasmpool.ChangeLimiter{},
			},
		},
		{
			name:            "Success with change limiters",
			poolID:          2,
			contractAddress: "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3",
			mockQueryResult: []byte(`{
    "limiters": [
        [
            ["denom3", "Change Limiter 1"],
            {"change_limiter": {
                "latest_value": "0.1",
                "window_config": {"window_size": 1000, "division_count": 10},
                "boundary_offset": "0.05"
            }}
        ]
    ]
}`),
			expectedResult: sqscosmwasmpool.AlloyedRateLimiter{
				StaticLimiterByDenomMap: map[string]sqscosmwasmpool.StaticLimiter{},
				ChangeLimiterByDenomMap: map[string]sqscosmwasmpool.ChangeLimiter{
					"denom3": {
						LatestValue:    "0.1",
						WindowConfig:   sqscosmwasmpool.WindowConfig{WindowSize: 1000, DivisionCount: 10},
						BoundaryOffset: "0.05",
					},
				},
			},
		},
		{
			name:            "Query error",
			poolID:          3,
			contractAddress: "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3",
			mockQueryError:  errors.New("query failed"),
			expectedError:   true,
		},
		{
			name:            "Unmarshal error",
			poolID:          4,
			contractAddress: "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3",
			mockQueryResult: []byte(`invalid json`),
			expectedError:   true,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {

			s.Setup()

			mockWasmKeeper := &mockWasmKeeper{
				QuerySmartFn: func(ctx sdk.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error) {
					return tc.mockQueryResult, tc.mockQueryError
				},
			}

			contractAddr, err := sdk.AccAddressFromBech32(tc.contractAddress)
			require.NoError(t, err)

			result, err := poolstransformer.AlloyTransmuterListLimiters(s.Ctx, mockWasmKeeper, tc.poolID, contractAddr)

			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResult, result)
			}
		})
	}
}
