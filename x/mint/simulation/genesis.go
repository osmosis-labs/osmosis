package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/osmosis-labs/osmosis/v7/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants.
const (
	epochProvisionsKey                        = "epoch_provisions"
	epochIdentifierKey                        = "epoch_identifier"
	reductionFactorKey                        = "reduction_factor"
	reductionPeriodInEpochsKey                = "reduction_period_in_epochs"
	stakingDistributionProportionKey          = "staking_distribution_proportion"
	poolIncentivesDistributionProportionKey   = "pool_incentives_distribution_proportion"
	developerRewardsDistributionProportionKey = "developer_rewards_distribution_proportion"
	communityPoolDistributionProportionKey    = "community_pool_distribution_proportion"
	weightedDevRewardReceiversKey             = "weighted_dev_reward_receivers"
	mintingRewardsDistributionStartEpochKey   = "minting_rewards_distribution_start_epoch"
	reductionStartedEpochKey                  = "reduction_started_epoch"

	maxInt64 = int(^uint(0) >> 1)
)

var (
	epochIdentifierOptions = []string{"day", "week"}
)

// RandomizedGenState generates a random GenesisState for mint.
func RandomizedGenState(simState *module.SimulationState) {
	var epochProvisions sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, epochProvisionsKey, &epochProvisions, simState.Rand,
		func(r *rand.Rand) { epochProvisions = genEpochProvisions(r) },
	)

	var epochIdentifier string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, epochIdentifierKey, &epochIdentifier, simState.Rand,
		func(r *rand.Rand) { epochIdentifier = genEpochIdentifier(r) },
	)

	var reductionFactor sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, reductionFactorKey, &reductionFactor, simState.Rand,
		func(r *rand.Rand) { reductionFactor = genReductionFactor(r) },
	)

	var reductionPeriodInEpochs int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, reductionPeriodInEpochsKey, &reductionPeriodInEpochs, simState.Rand,
		func(r *rand.Rand) { reductionPeriodInEpochs = genReductionPeriodInEpochs(r) },
	)

	randomDisitributionProportions := genProportionsAddingUpToOne(simState.Rand, 4)

	var stakingDistributionProportion sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, stakingDistributionProportionKey, &stakingDistributionProportion, simState.Rand,
		func(r *rand.Rand) { stakingDistributionProportion = randomDisitributionProportions[0] },
	)

	var poolIncentivesDistributionProportion sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, poolIncentivesDistributionProportionKey, &poolIncentivesDistributionProportion, simState.Rand,
		func(r *rand.Rand) { poolIncentivesDistributionProportion = randomDisitributionProportions[1] },
	)

	var developerRewardsDistributionProportion sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, developerRewardsDistributionProportionKey, &developerRewardsDistributionProportion, simState.Rand,
		func(r *rand.Rand) { developerRewardsDistributionProportion = randomDisitributionProportions[2] },
	)

	var communityPoolDistributionProportion sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, communityPoolDistributionProportionKey, &communityPoolDistributionProportion, simState.Rand,
		func(r *rand.Rand) { communityPoolDistributionProportion = randomDisitributionProportions[3] },
	)

	var weightedDevRewardReceivers []types.WeightedAddress
	simState.AppParams.GetOrGenerate(
		simState.Cdc, weightedDevRewardReceiversKey, &weightedDevRewardReceivers, simState.Rand,
		func(r *rand.Rand) {
			addressCount := r.Intn(5)
			randomDevRewardProportions := genProportionsAddingUpToOne(simState.Rand, addressCount)

			for i := 0; i < addressCount; i++ {
				weightedDevRewardReceivers = append(weightedDevRewardReceivers, types.WeightedAddress{
					Address: fmt.Sprintf("address_%d", i),
					Weight:  randomDevRewardProportions[i],
				})
			}
		},
	)

	var mintintRewardsDistributionStartEpoch int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, mintingRewardsDistributionStartEpochKey, &mintintRewardsDistributionStartEpoch, simState.Rand,
		func(r *rand.Rand) { mintintRewardsDistributionStartEpoch = genMintintRewardsDistributionStartEpoch(r) },
	)

	var reductionStartedEpoch int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, reductionStartedEpochKey, &reductionStartedEpoch, simState.Rand,
		func(r *rand.Rand) { reductionStartedEpoch = genReductionStartedEpoch(r) },
	)

	mintDenom := sdk.DefaultBondDenom
	params := types.NewParams(
		mintDenom,
		epochProvisions,
		epochIdentifier,
		reductionFactor,
		reductionPeriodInEpochs,
		types.DistributionProportions{
			Staking:          stakingDistributionProportion,
			PoolIncentives:   poolIncentivesDistributionProportion,
			DeveloperRewards: developerRewardsDistributionProportion,
			CommunityPool:    communityPoolDistributionProportion,
		},
		weightedDevRewardReceivers,
		mintintRewardsDistributionStartEpoch)

	minter := types.NewMinter(epochProvisions)

	mintGenesis := types.NewGenesisState(minter, params, reductionStartedEpoch)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected pseudo-randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}

func genEpochProvisions(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(int64(r.Intn(maxInt64)))
}

func genEpochIdentifier(r *rand.Rand) string {
	return epochIdentifierOptions[rand.Intn(len(epochIdentifierOptions))]
}

func genReductionFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(10)), 1)
}

func genReductionPeriodInEpochs(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func genStakingDistributionProportion(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(5)), 1)
}

func genPoolIncentivesDistributionProportion(r *rand.Rand, limitRatio sdk.Dec) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(int(limitRatio.MulInt64(10).TruncateInt64()))), 1)
}

func genDeveloperRewardsDistributionProportion(r *rand.Rand, limitRatio sdk.Dec) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(int(limitRatio.MulInt64(10).TruncateInt64()))), 1)
}

// genProportionsAddingUpToOne reurns a slice with numberOfProportions that add up to 1.
func genProportionsAddingUpToOne(r *rand.Rand, numberOfProportions int) []sdk.Dec {
	proportions := make([]sdk.Dec, numberOfProportions)

	// We start by estimating the first proportion with a limit of 1.
	// Then, subtract the first proportion from 1 to esimate
	// the remaining ratio to be used as upper bound for next randomization.
	// Next, repeat the randomization process for the remaining proportions.
	remainingRatio := sdk.OneDec()
	for i := 0; i < numberOfProportions-1; i++ {
		nextProportion := sdk.NewDecWithPrec(int64(r.Intn(int(remainingRatio.MulInt64(10).TruncateInt64()))), 1)
		proportions[i] = nextProportion
		remainingRatio = remainingRatio.Sub(nextProportion)
	}
	proportions[numberOfProportions-1] = remainingRatio
	return proportions
}

func genMintintRewardsDistributionStartEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func genReductionStartedEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}
