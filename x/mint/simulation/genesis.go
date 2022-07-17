package simulation

// DONTCOVER

import (
	"math/rand"

	"github.com/osmosis-labs/osmosis/v10/x/mint/types"

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
		Staking:          sdk.NewDecWithPrec(25, 2),
		PoolIncentives:   sdk.NewDecWithPrec(45, 2),
		DeveloperRewards: sdk.NewDecWithPrec(25, 2),
		CommunityPool:    sdk.NewDecWithPrec(0o5, 2),
	}
	weightedDevRewardReceivers = []types.WeightedAddress{
		{
			Address: "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
			Weight:  sdk.NewDecWithPrec(2887, 4),
		},
		{
			Address: "osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
			Weight:  sdk.NewDecWithPrec(229, 3),
		},
		{
			Address: "osmo13lt0hzc6u3htsk7z5rs6vuurmgg4hh2ecgxqkf",
			Weight:  sdk.NewDecWithPrec(1625, 4),
		},
		{
			Address: "osmo1kvc3he93ygc0us3ycslwlv2gdqry4ta73vk9hu",
			Weight:  sdk.NewDecWithPrec(109, 3),
		},
		{
			Address: "osmo19qgldlsk7hdv3ddtwwpvzff30pxqe9phq9evxf",
			Weight:  sdk.NewDecWithPrec(995, 3).Quo(sdk.NewDec(10)), // 0.0995
		},
		{
			Address: "osmo19fs55cx4594een7qr8tglrjtt5h9jrxg458htd",
			Weight:  sdk.NewDecWithPrec(6, 1).Quo(sdk.NewDec(10)), // 0.06
		},
		{
			Address: "osmo1ssp6px3fs3kwreles3ft6c07mfvj89a544yj9k",
			Weight:  sdk.NewDecWithPrec(15, 2).Quo(sdk.NewDec(10)), // 0.015
		},
		{
			Address: "osmo1c5yu8498yzqte9cmfv5zcgtl07lhpjrj0skqdx",
			Weight:  sdk.NewDecWithPrec(1, 1).Quo(sdk.NewDec(10)), // 0.01
		},
		{
			Address: "osmo1yhj3r9t9vw7qgeg22cehfzj7enwgklw5k5v7lj",
			Weight:  sdk.NewDecWithPrec(75, 2).Quo(sdk.NewDec(100)), // 0.0075
		},
		{
			Address: "osmo18nzmtyn5vy5y45dmcdnta8askldyvehx66lqgm",
			Weight:  sdk.NewDecWithPrec(7, 1).Quo(sdk.NewDec(100)), // 0.007
		},
		{
			Address: "osmo1z2x9z58cg96ujvhvu6ga07yv9edq2mvkxpgwmc",
			Weight:  sdk.NewDecWithPrec(5, 1).Quo(sdk.NewDec(100)), // 0.005
		},
		{
			Address: "osmo1tvf3373skua8e6480eyy38avv8mw3hnt8jcxg9",
			Weight:  sdk.NewDecWithPrec(25, 2).Quo(sdk.NewDec(100)), // 0.0025
		},
		{
			Address: "osmo1zs0txy03pv5crj2rvty8wemd3zhrka2ne8u05n",
			Weight:  sdk.NewDecWithPrec(25, 2).Quo(sdk.NewDec(100)), // 0.0025
		},
		{
			Address: "osmo1djgf9p53n7m5a55hcn6gg0cm5mue4r5g3fadee",
			Weight:  sdk.NewDecWithPrec(1, 1).Quo(sdk.NewDec(100)), // 0.001
		},
		{
			Address: "osmo1488zldkrn8xcjh3z40v2mexq7d088qkna8ceze",
			Weight:  sdk.NewDecWithPrec(8, 1).Quo(sdk.NewDec(1000)), // 0.0008
		},
	}
)

// RandomizedGenState generates a random GenesisState for mint.
func RandomizedGenState(simState *module.SimulationState) {
	var epochProvisions sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, epochProvisionsKey, &epochProvisions, simState.Rand,
		func(r *rand.Rand) { epochProvisions = genEpochProvisions(r) },
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

func genEpochProvisions(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(int64(r.Intn(maxInt64)))
}

func genReductionFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(10)), 1)
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
