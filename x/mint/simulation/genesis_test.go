package simulation_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/mint/simulation"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abnormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(5)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: 1000,
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var mintGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &mintGenesis)

	const (
		expectedEpochProvisionsStr      = "7913048388940673156"
		expectedEpochIdentifier         = "week"
		expectedReductionFactorStr      = "0.6"
		expectedReductionPeriodInEpochs = int64(9171281239991390334)

		expectedStakingDistributionProportion          = "0.8"
		expectedPoolIncentivesDistributionProportion   = "0"
		expectedDeveloperRewardsDistributionProportion = "0.1"
		expectedCommunityPoolDistributionProportion    = "0.1"

		expectedMintintRewardsDistributionStartEpoch = int64(3800660657951843153)

		expectedReductionStartedEpoch = int64(1187669907065521396)

		expectedNextEpochProvisionsStr = "3956524194470336578"
		expectedDenom                  = sdk.DefaultBondDenom
	)

	var expectedWeightedAddresses []types.WeightedAddress = []types.WeightedAddress{
		{
			Address: "address_0",
			Weight:  sdk.NewDecWithPrec(4, 1),
		},
		{
			Address: "address_1",
			Weight:  sdk.NewDecWithPrec(0, 1),
		},
		{
			Address: "address_2",
			Weight:  sdk.NewDecWithPrec(2, 1),
		},
		{
			Address: "address_3",
			Weight:  sdk.NewDecWithPrec(4, 1),
		},
	}

	// Epoch provisions from Minter.
	epochProvisionsDec, err := sdk.NewDecFromStr(expectedEpochProvisionsStr)
	require.NoError(t, err)
	require.Equal(t, epochProvisionsDec, mintGenesis.Minter.EpochProvisions)

	// Epoch identifier.
	require.Equal(t, expectedEpochIdentifier, mintGenesis.Params.EpochIdentifier)

	// Reduction factor.
	reductionFactorDec, err := sdk.NewDecFromStr(expectedReductionFactorStr)
	require.NoError(t, err)
	require.Equal(t, reductionFactorDec, mintGenesis.Params.ReductionFactor)

	// Reduction perion in epochs.
	require.Equal(t, expectedReductionPeriodInEpochs, mintGenesis.Params.ReductionPeriodInEpochs)

	// Staking rewards distribution proportion.
	stakingDistributionProportionDec, err := sdk.NewDecFromStr(expectedStakingDistributionProportion)
	require.NoError(t, err)
	require.Equal(t, stakingDistributionProportionDec, mintGenesis.Params.DistributionProportions.Staking)

	// Pool incentives distribution proportion.
	poolIncentivesDistributionProportionDec, err := sdk.NewDecFromStr(expectedPoolIncentivesDistributionProportion)
	require.NoError(t, err)
	require.Equal(t, poolIncentivesDistributionProportionDec, mintGenesis.Params.DistributionProportions.PoolIncentives)

	// Developer rewards distribution proportion.
	developerRewardsDistributionProportionDec, err := sdk.NewDecFromStr(expectedDeveloperRewardsDistributionProportion)
	require.NoError(t, err)
	require.Equal(t, developerRewardsDistributionProportionDec, mintGenesis.Params.DistributionProportions.DeveloperRewards)

	// Community pool distribution proportion.
	communityPoolDistributionProportionDec, err := sdk.NewDecFromStr(expectedCommunityPoolDistributionProportion)
	require.NoError(t, err)
	require.Equal(t, communityPoolDistributionProportionDec, mintGenesis.Params.DistributionProportions.CommunityPool)

	// Weighted developer rewards receivers.
	require.Equal(t, expectedWeightedAddresses, mintGenesis.Params.WeightedDeveloperRewardsReceivers)

	// Minting rewards distribution start epoch
	require.Equal(t, expectedMintintRewardsDistributionStartEpoch, mintGenesis.Params.MintingRewardsDistributionStartEpoch)

	// Reduction started epoch.
	require.Equal(t, expectedReductionStartedEpoch, mintGenesis.ReductionStartedEpoch)

	// Next epoch provisions.
	nextEpochProvisionsDec := epochProvisionsDec.Mul(reductionFactorDec)
	require.NoError(t, err)
	require.Equal(t, nextEpochProvisionsDec, mintGenesis.Minter.NextEpochProvisions(mintGenesis.Params))

	// Denom and Epoch provisions from Params.
	require.Equal(t, expectedDenom, mintGenesis.Params.MintDenom)
	require.Equal(t, fmt.Sprintf("%s%s", expectedEpochProvisionsStr, expectedDenom), mintGenesis.Minter.EpochProvision(mintGenesis.Params).String())
}

// TestRandomizedGenState_Invalid tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState_Invalid(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)
	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}
