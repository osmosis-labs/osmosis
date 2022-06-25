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

		expectedStakingDistributionProportion          = "0.51"
		expectedPoolIncentivesDistributionProportion   = "0.01"
		expectedDeveloperRewardsDistributionProportion = "0.31"
		expectedCommunityPoolDistributionProportion    = "0.17"

		expectedMintintRewardsDistributionStartEpoch = int64(8326275384461735988)

		expectedReductionStartedEpoch = int64(8272964973000937025)

		expectedNextEpochProvisionsStr = "3956524194470336578"
		expectedDenom                  = sdk.DefaultBondDenom
	)

	var expectedWeightedAddresses []types.WeightedAddress = []types.WeightedAddress{
		{
			Address: "osmo10h0yjph5cs87jlrn0d7g7u4dmftufg57mute7qhn6zla86ha3ems0yhsdm",
			Weight:  sdk.NewDecWithPrec(21, 2),
		},
		{
			Address: "osmo1fcjs4czqdfcm8vpx0835kvxgh30hx8s4p8fk53",
			Weight:  sdk.NewDecWithPrec(21, 2),
		},
		{
			Address: "osmo1npjxgta2789ju4v063ewwuyv6wpz84x8gscg7hx4l9xlxwr0pgdqzzez3k",
			Weight:  sdk.NewDecWithPrec(31, 2),
		},
		{
			Address: "osmo15y8quyqq24d2xlkm6lekg5x3uqxcvf362jlq4yuayy6c4y3e9nysq8gmgg",
			Weight:  sdk.NewDecWithPrec(27, 2),
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
