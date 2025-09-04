package poolstransformer_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	sqscosmwasmpool "github.com/osmosis-labs/osmosis/v30/ingest/types/cosmwasmpool"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v30/ingest/common/domain"
	poolstransformer "github.com/osmosis-labs/osmosis/v30/ingest/sqs/pools/transformer"
)

type mockWasmKeeper struct {
	QueryRawFn             func(ctx context.Context, contractAddress sdk.AccAddress, key []byte) []byte
	QuerySmartFn           func(ctx context.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error)
	IterateContractStateFn func(ctx context.Context, contractAddress sdk.AccAddress, cb func(key []byte, value []byte) bool)
}

// QueryRaw implements commondomain.WasmKeeper.
func (m *mockWasmKeeper) QueryRaw(ctx context.Context, contractAddress sdk.AccAddress, key []byte) []byte {
	if m.QueryRawFn != nil {
		return m.QueryRawFn(ctx, contractAddress, key)
	}
	panic("unimplemented")
}

// QuerySmart implements commondomain.WasmKeeper.
func (m *mockWasmKeeper) QuerySmart(ctx context.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error) {
	if m.QuerySmartFn != nil {
		return m.QuerySmartFn(ctx, contractAddress, req)
	}
	panic("unimplemented")
}

// IterateContractState implements commondomain.WasmKeeper.
func (m *mockWasmKeeper) IterateContractState(ctx context.Context, contractAddress sdk.AccAddress, cb func(key []byte, value []byte) bool) {
	if m.IterateContractStateFn != nil {
		m.IterateContractStateFn(ctx, contractAddress, cb)
		return
	}
	panic("unimplemented")
}

var _ commondomain.WasmKeeper = &mockWasmKeeper{}

func (s *PoolTransformerTestSuite) TestUpdateAlloyedTransmuterPool() {
	tests := []struct {
		name         string
		contractName string
	}{
		{
			name:         "transmuter v3",
			contractName: apptesting.TransmuterV3ContractName,
		},
		{
			name:         "transmuter v4",
			contractName: apptesting.TransmuterV4ContractName,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
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
			}, tc.contractName)

			// Add asset groups to the pool
			assetGroup := map[string]sqscosmwasmpool.AssetGroup{}
			rebalancingConfigs := map[string]sqscosmwasmpool.RebalancingConfig{}
			if tc.contractName == apptesting.TransmuterV4ContractName {
				createAssetGroupMsg := `{"create_asset_group":{"label":"group1","denoms":["` + apptesting.DefaultTransmuterDenomA + `","` + apptesting.DefaultTransmuterDenomB + `"]}}`
				_, err := s.App.ContractKeeper.Execute(s.Ctx, pool.GetAddress(), s.TestAccs[0], []byte(createAssetGroupMsg), sdk.NewCoins())
				s.Require().NoError(err)

				assetGroup = map[string]sqscosmwasmpool.AssetGroup{
					"group1": {Denoms: []string{apptesting.DefaultTransmuterDenomA, apptesting.DefaultTransmuterDenomB}, IsCorrupted: false},
				}

				// add rebalancing config for asset group
				addRebalancingConfigMsg := `{"add_rebalancing_config":{"scope":{"type":"asset_group","value":"group1"},"rebalancing_config":{"ideal_upper":"0.8","ideal_lower":"0.2","critical_upper":"0.9","critical_lower":"0.1","limit":"0.05","adjustment_rate_strained":"0.1","adjustment_rate_critical":"0.2"}}}`
				_, err = s.App.ContractKeeper.Execute(s.Ctx, pool.GetAddress(), s.TestAccs[0], []byte(addRebalancingConfigMsg), sdk.NewCoins())
				s.Require().NoError(err)

				// add rebalancing config for denom
				addDenomRebalancingConfigMsg := `{"add_rebalancing_config":{"scope":{"type":"denom","value":"` + apptesting.DefaultTransmuterDenomA + `"},"rebalancing_config":{"ideal_upper":"0.7","ideal_lower":"0.3","critical_upper":"0.85","critical_lower":"0.15","limit":"0.1","adjustment_rate_strained":"0.05","adjustment_rate_critical":"0.15"}}}`
				_, err = s.App.ContractKeeper.Execute(s.Ctx, pool.GetAddress(), s.TestAccs[0], []byte(addDenomRebalancingConfigMsg), sdk.NewCoins())
				s.Require().NoError(err)

				rebalancingConfigs = map[string]sqscosmwasmpool.RebalancingConfig{
					"asset_group::group1": {
						IdealUpper:             "0.8",
						IdealLower:             "0.2",
						CriticalUpper:          "0.9",
						CriticalLower:          "0.1",
						Limit:                  "0.05",
						AdjustmentRateStrained: "0.1",
						AdjustmentRateCritical: "0.2",
					},
					"denom::" + apptesting.DefaultTransmuterDenomA: {
						IdealUpper:             "0.7",
						IdealLower:             "0.3",
						CriticalUpper:          "0.85",
						CriticalLower:          "0.15",
						Limit:                  "0.1",
						AdjustmentRateStrained: "0.05",
						AdjustmentRateCritical: "0.15",
					},
				}
			}

			if tc.contractName == apptesting.TransmuterV3ContractName {
				// register limiter
				registerLimiterMsg := `{"register_limiter":{"denom":"` + apptesting.DefaultTransmuterDenomA + `","label":"limiter1","limiter_params":{"static_limiter":{"upper_limit":"0.2"}}}}`
				_, err := s.App.ContractKeeper.Execute(s.Ctx, pool.GetAddress(), s.TestAccs[0], []byte(registerLimiterMsg), sdk.NewCoins())
				s.Require().NoError(err)

				rebalancingConfigs = map[string]sqscosmwasmpool.RebalancingConfig{
					"denom::" + apptesting.DefaultTransmuterDenomA: {
						Limit: "0.2",
					},
				}
			}

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
					RebalancingConfigs: rebalancingConfigs,
					AssetGroups:        assetGroup,
				},
			}, cosmWasmPoolModel.Data)

			s.Equal([]string{
				apptesting.DefaultTransmuterDenomA,
				apptesting.DefaultTransmuterDenomB,
				alloyedDenom,
			}, poolDenoms)
		})
	}

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
				QuerySmartFn: func(ctx context.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error) {
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

func (s *PoolTransformerTestSuite) TestAlloyedTransmuterListAssetGroups() {
	tests := []struct {
		name            string
		poolID          uint64
		contractAddress string
		mockQueryResult []byte
		mockQueryError  error
		expectedResult  map[string]sqscosmwasmpool.AssetGroup
		expectedError   bool
	}{
		{
			name:            "Success with multiple groups",
			poolID:          1,
			contractAddress: "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3",
			mockQueryResult: []byte(`{
  "asset_groups": {
    "group1": {"denoms": ["denom1", "denom2"], "is_corrupted": false},
    "group2": {"denoms": ["denom3"], "is_corrupted": false}
  }
}`),
			expectedResult: map[string]sqscosmwasmpool.AssetGroup{
				"group1": {Denoms: []string{"denom1", "denom2"}, IsCorrupted: false},
				"group2": {Denoms: []string{"denom3"}, IsCorrupted: false},
			},
		},
		{
			name:            "Empty groups returns empty",
			poolID:          2,
			contractAddress: "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3",
			mockQueryResult: []byte(`{"asset_groups": {}}`),
			expectedResult:  map[string]sqscosmwasmpool.AssetGroup{},
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
				QuerySmartFn: func(ctx context.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error) {
					return tc.mockQueryResult, tc.mockQueryError
				},
			}

			contractAddr, err := sdk.AccAddressFromBech32(tc.contractAddress)
			require.NoError(t, err)

			result, err := poolstransformer.AlloyedTransmuterListAssetGroups(s.Ctx, mockWasmKeeper, tc.poolID, contractAddr)

			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResult, result)
			}
		})
	}
}
