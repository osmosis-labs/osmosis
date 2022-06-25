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
	epochProvisionsKey         = "genesis_epoch_provisions"
	epochIdentifierKey         = "epoch_identifier"
	reductionFactorKey         = "reduction_factor"
	reductionPeriodInEpochsKey = "reduction_period_in_epochs"

	distributionProportionsKey = "distribution_proportions"

	stakingDistributionProportionKey          = "staking_distribution_proportion"
	poolIncentivesDistributionProportionKey   = "pool_incentives_distribution_proportion"
	developerRewardsDistributionProportionKey = "developer_rewards_distribution_proportion"
	communityPoolDistributionProportionKey    = "community_pool_distribution_proportion"
	weightedDevRewardReceiversKey             = "weighted_developer_rewards_receivers"
	mintingRewardsDistributionStartEpochKey   = "minting_rewards_distribution_start_epoch"

	maxInt64 = int(^uint(0) >> 1)
)

var (
	epochIdentifierOptions    = []string{"day", "week"}
	possibleBech32AddrLengths = []uint8{20, 32}
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

	var distributionProportions types.DistributionProportions
	simState.AppParams.GetOrGenerate(
		simState.Cdc, distributionProportionsKey, &distributionProportions, simState.Rand,
		func(r *rand.Rand) { distributionProportions = genDistributionProportions(r) },
	)

	var weightedDevRewardReceivers []types.WeightedAddress
	simState.AppParams.GetOrGenerate(
		simState.Cdc, weightedDevRewardReceiversKey, &weightedDevRewardReceivers, simState.Rand,
		func(r *rand.Rand) { weightedDevRewardReceivers = genWeightedDevRewardReceivers(simState.Rand) },
	)

	var mintintRewardsDistributionStartEpoch int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, mintingRewardsDistributionStartEpochKey, &mintintRewardsDistributionStartEpoch, simState.Rand,
		func(r *rand.Rand) { mintintRewardsDistributionStartEpoch = genMintintRewardsDistributionStartEpoch(r) },
	)

	reductionStartedEpoch := genReductionStartedEpoch(simState.Rand)

	mintDenom := sdk.DefaultBondDenom
	params := types.NewParams(
		mintDenom,
		epochProvisions,
		epochIdentifier,
		reductionFactor,
		reductionPeriodInEpochs,
		distributionProportions,
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
	return "day"
}

func genReductionFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(10)), 1)
}

func genReductionPeriodInEpochs(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

// genProportionsAddingUpToOne reurns a slice with numberOfProportions that add up to 1.
func genProportionsAddingUpToOne(r *rand.Rand, numberOfProportions int) []sdk.Dec {
	if numberOfProportions < 0 {
		panic("numberOfProportions must be greater than or equal to 1")
	}

	proportions := make([]sdk.Dec, numberOfProportions)

	// We start by estimating the first proportion with a limit of 1.
	// Then, subtract the first proportion from 1 to esimate
	// the remaining ratio to be used as upper bound for next randomization.
	// Next, repeat the randomization process for the remaining proportions.
	remainingRatio := sdk.OneDec()
	for i := 0; i < numberOfProportions-1; i++ {
		// We add 0.01 to make sure that zero is never returned because a proportion of 0 is deemed invalid.
		nextProportion := sdk.MustNewDecFromStr("0.01").Add(sdk.NewDecWithPrec(int64(r.Intn(int(remainingRatio.MulInt64(9).TruncateInt64()))), 1))
		proportions[i] = nextProportion
		remainingRatio = remainingRatio.Sub(nextProportion)
	}
	proportions[numberOfProportions-1] = remainingRatio
	return proportions
}

func genDistributionProportions(r *rand.Rand) types.DistributionProportions {
	distributionProportions := types.DistributionProportions{}
	randomDisitributionProportions := genProportionsAddingUpToOne(r, 4)
	distributionProportions.Staking = randomDisitributionProportions[0]
	distributionProportions.PoolIncentives = randomDisitributionProportions[1]
	distributionProportions.DeveloperRewards = randomDisitributionProportions[2]
	distributionProportions.CommunityPool = randomDisitributionProportions[3]
	return distributionProportions
}

func genWeightedDevRewardReceivers(r *rand.Rand) []types.WeightedAddress {
	var weightedDevRewardReceivers []types.WeightedAddress
	addressCount := max(1, r.Intn(5))
	randomDevRewardProportions := genProportionsAddingUpToOne(r, addressCount)

	for i := 0; i < addressCount; i++ {
		addressLength := possibleBech32AddrLengths[r.Intn(len(possibleBech32AddrLengths))]
		addressRandBytes, err := randBytes(r, int(addressLength))
		if err != nil {
			panic(err)
		}
		address, err := sdk.Bech32ifyAddressBytes("osmo", addressRandBytes)
		if err != nil {
			panic(err)
		}
		weightedDevRewardReceivers = append(weightedDevRewardReceivers, types.WeightedAddress{
			Address: address,
			Weight:  randomDevRewardProportions[i],
		})
	}
	return weightedDevRewardReceivers
}

func genMintintRewardsDistributionStartEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func genReductionStartedEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func randBytes(r *rand.Rand, length int) ([]byte, error) {
	result := make([]byte, length)
	n, err := r.Read(result)
	if n != length {
		return nil, fmt.Errorf("did not read enough bytes, read: %d, expected: %d", n, length)
	}
	return result, err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
