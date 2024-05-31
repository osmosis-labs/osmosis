package simulation

// DONTCOVER

import (
	"math/rand"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v23/x/mint/types"

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
		DeveloperRewards: osmomath.NewDecWithPrec(25, 2),
		CommunityPool:    osmomath.NewDecWithPrec(0o5, 2),
	}
	weightedDevRewardReceivers = []types.WeightedAddress{
		{
			Address: "symphony1u7ryvx794sy5yqwezfryygsce84q287ts98n66",
			Weight:  osmomath.NewDecWithPrec(2887, 4),
		},
		{
			Address: "symphony1zrmuw4xux344w4k9pw93qs8d0d7kc0fnhxw4wd",
			Weight:  osmomath.NewDecWithPrec(229, 3),
		},
		{
			Address: "symphony1t9vjrxn6cwdkuf990sncq7akqsz26feaz5euxt",
			Weight:  osmomath.NewDecWithPrec(1625, 4),
		},
		{
			Address: "symphony172qywhy2qxcnkvr6vcal23ntz645h20qe5880r",
			Weight:  osmomath.NewDecWithPrec(109, 3),
		},
		{
			Address: "symphony195ds5rrxcqcwflj692e6gmykhl9vu0r0qs7tt5",
			Weight:  osmomath.NewDecWithPrec(995, 3).Quo(osmomath.NewDec(10)), // 0.0995
		},
		{
			Address: "symphony1f2jp2q4qq0f8nlmp0v3ah96h3kqjj0vheprf7q",
			Weight:  osmomath.NewDecWithPrec(6, 1).Quo(osmomath.NewDec(10)), // 0.06
		},
		{
			Address: "symphony1k27t46ehr7y80ktrtmn9grmc9wkw27ds9hq005",
			Weight:  osmomath.NewDecWithPrec(15, 2).Quo(osmomath.NewDec(10)), // 0.015
		},
		{
			Address: "symphony1dhtgp9726rx5zv9079xz2wz43pec484akwktn5",
			Weight:  osmomath.NewDecWithPrec(1, 1).Quo(osmomath.NewDec(10)), // 0.01
		},
		{
			Address: "symphony1fqqucy9y2adaapyjze5g0hv40vp6rt2kt0cjts",
			Weight:  osmomath.NewDecWithPrec(75, 2).Quo(osmomath.NewDec(100)), // 0.0075
		},
		{
			Address: "symphony192953mpz44nn76vgmknt75vspsnv2k6d9dyc4w",
			Weight:  osmomath.NewDecWithPrec(7, 1).Quo(osmomath.NewDec(100)), // 0.007
		},
		{
			Address: "symphony1jcchx5enuex05al39y25gl6hyerwj74unntaqx",
			Weight:  osmomath.NewDecWithPrec(5, 1).Quo(osmomath.NewDec(100)), // 0.005
		},
		{
			Address: "symphony1pt2knp6s8exw7j28gjgmwr2wvw4suc3w8ncunl",
			Weight:  osmomath.NewDecWithPrec(25, 2).Quo(osmomath.NewDec(100)), // 0.0025
		},
		{
			Address: "symphony1c4zx9pmtn3j4a2eus2mmpclpllpqzgzezte7yz",
			Weight:  osmomath.NewDecWithPrec(25, 2).Quo(osmomath.NewDec(100)), // 0.0025
		},
		{
			Address: "symphony1d6fwytjdlwzg7hg26zpzrl4y3f5ykft9xetlmk",
			Weight:  osmomath.NewDecWithPrec(1, 1).Quo(osmomath.NewDec(100)), // 0.001
		},
		{
			Address: "symphony1gmyrqx37tvpmqpkvga6ex4jtv0920hfa3pndqz",
			Weight:  osmomath.NewDecWithPrec(8, 1).Quo(osmomath.NewDec(1000)), // 0.0008
		},
	}
)

// RandomizedGenState generates a random GenesisState for mint.
func RandomizedGenState(simState *module.SimulationState) {
	var epochProvisions osmomath.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, epochProvisionsKey, &epochProvisions, simState.Rand,
		func(r *rand.Rand) { epochProvisions = genEpochProvisions(r) },
	)

	var reductionFactor osmomath.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, reductionFactorKey, &reductionFactor, simState.Rand,
		func(r *rand.Rand) { reductionFactor = genReductionFactor(r) },
	)

	var reductionPeriodInEpochs int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, reductionPeriodInEpochsKey, &reductionPeriodInEpochs, simState.Rand,
		func(r *rand.Rand) { reductionPeriodInEpochs = genReductionPeriodInEpochs(r) },
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
