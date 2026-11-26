package simulation

// DONTCOVER

import (
	"math/rand"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants.
const (
	epochProvisionsKey         = "genesis_epoch_provisions"
	reductionFactorKey         = "reduction_factor"
	reductionPeriodInEpochsKey = "reduction_period_in_epochs"

	mintingRewardsDistributionStartEpochKey = "minting_rewards_distribution_start_epoch"

	epochIdentifier = "day"
	maxInt64        = int(^uint(0) >> 1)
)

var (
	// Taken from: // https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
	distributionProportions = types.DistributionProportions{
		Staking:          osmomath.NewDecWithPrec(25, 2),
		PoolIncentives:   osmomath.NewDecWithPrec(45, 2),
		DeveloperRewards: osmomath.ZeroDec(), // deprecated, must be 0
		CommunityPool:    osmomath.NewDecWithPrec(30, 2),
	}
	weightedDevRewardReceivers = []types.WeightedAddress{}
)

// RandomizedGenState generates a random GenesisState for mint.
func RandomizedGenState(simState *module.SimulationState) {
	var epochProvisions osmomath.Dec
	simState.AppParams.GetOrGenerate(
		epochProvisionsKey, &epochProvisions, simState.Rand,
		func(r *rand.Rand) { epochProvisions = genEpochProvisions(r) },
	)

	var reductionFactor osmomath.Dec
	simState.AppParams.GetOrGenerate(
		reductionFactorKey, &reductionFactor, simState.Rand,
		func(r *rand.Rand) { reductionFactor = genReductionFactor(r) },
	)

	var reductionPeriodInEpochs int64
	simState.AppParams.GetOrGenerate(
		reductionPeriodInEpochsKey, &reductionPeriodInEpochs, simState.Rand,
		func(r *rand.Rand) { reductionPeriodInEpochs = genReductionPeriodInEpochs(r) },
	)

	var mintintRewardsDistributionStartEpoch int64
	simState.AppParams.GetOrGenerate(
		mintingRewardsDistributionStartEpochKey, &mintintRewardsDistributionStartEpoch, simState.Rand,
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

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}

func genEpochProvisions(r *rand.Rand) osmomath.Dec {
	return osmomath.NewDec(int64(r.Intn(maxInt64)))
}

func genReductionFactor(r *rand.Rand) osmomath.Dec {
	return osmomath.NewDecWithPrec(int64(r.Intn(10)), 1)
}

func genReductionPeriodInEpochs(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func genMintintRewardsDistributionStartEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func genReductionStartedEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}
