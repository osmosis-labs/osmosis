package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

// Simulation parameter constants
const (
	votePeriodKey               = "vote_period"
	voteThresholdKey            = "vote_threshold"
	rewardBandKey               = "reward_band"
	rewardDistributionWindowKey = "reward_distribution_window"
	slashFractionKey            = "slash_fraction"
	slashWindowKey              = "slash_window"
	minValidPerWindowKey        = "min_valid_per_window"
)

// GenVotePeriod randomized VotePeriod
func GenVotePeriod(r *rand.Rand) uint64 {
	return uint64(1 + r.Intn(100))
}

// GenVoteThreshold randomized VoteThreshold
func GenVoteThreshold(r *rand.Rand) osmomath.Dec {
	return osmomath.NewDecWithPrec(333, 3).Add(osmomath.NewDecWithPrec(int64(r.Intn(333)), 3))
}

// GenRewardBand randomized RewardBand
func GenRewardBand(r *rand.Rand) osmomath.Dec {
	return osmomath.ZeroDec().Add(osmomath.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// GenRewardDistributionWindow randomized RewardDistributionWindow
func GenRewardDistributionWindow(r *rand.Rand) uint64 {
	return uint64(100 + r.Intn(100000))
}

// GenSlashFraction randomized SlashFraction
func GenSlashFraction(r *rand.Rand) osmomath.Dec {
	return osmomath.ZeroDec().Add(osmomath.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// GenSlashWindow randomized SlashWindow
func GenSlashWindow(r *rand.Rand) uint64 {
	return uint64(100 + r.Intn(100000))
}

// GenMinValidPerWindow randomized MinValidPerWindow
func GenMinValidPerWindow(r *rand.Rand) osmomath.Dec {
	return osmomath.ZeroDec().Add(osmomath.NewDecWithPrec(int64(r.Intn(500)), 3))
}

// RandomizedGenState generates a random GenesisState for oracle
func RandomizedGenState(simState *module.SimulationState) {
	//var votePeriod uint64
	//simState.AppParams.GetOrGenerate(
	//	votePeriodKey, &votePeriod, simState.Rand,
	//	func(r *rand.Rand) { votePeriod = GenVotePeriod(r) },
	//)

	var voteThreshold osmomath.Dec
	simState.AppParams.GetOrGenerate(
		voteThresholdKey, &voteThreshold, simState.Rand,
		func(r *rand.Rand) { voteThreshold = GenVoteThreshold(r) },
	)

	var rewardBand osmomath.Dec
	simState.AppParams.GetOrGenerate(
		rewardBandKey, &rewardBand, simState.Rand,
		func(r *rand.Rand) { rewardBand = GenRewardBand(r) },
	)

	var rewardDistributionWindow uint64
	simState.AppParams.GetOrGenerate(
		rewardDistributionWindowKey, &rewardDistributionWindow, simState.Rand,
		func(r *rand.Rand) { rewardDistributionWindow = GenRewardDistributionWindow(r) },
	)

	var slashFraction osmomath.Dec
	simState.AppParams.GetOrGenerate(
		slashFractionKey, &slashFraction, simState.Rand,
		func(r *rand.Rand) { slashFraction = GenSlashFraction(r) },
	)

	var slashWindow uint64
	simState.AppParams.GetOrGenerate(
		slashWindowKey, &slashWindow, simState.Rand,
		func(r *rand.Rand) { slashWindow = GenSlashWindow(r) },
	)

	var minValidPerWindow osmomath.Dec
	simState.AppParams.GetOrGenerate(
		minValidPerWindowKey, &minValidPerWindow, simState.Rand,
		func(r *rand.Rand) { minValidPerWindow = GenMinValidPerWindow(r) },
	)

	oracleGenesis := types.NewGenesisState(
		types.Params{
			VotePeriodEpochIdentifier:  types.DefaultVotePeriodEpochIdentifier,
			VoteThreshold:              voteThreshold,
			RewardBand:                 rewardBand,
			RewardDistributionWindow:   rewardDistributionWindow,
			Whitelist:                  types.DenomList{},
			SlashFraction:              slashFraction,
			SlashWindowEpochIdentifier: types.DefaultSlashWindowEpochIdentifier,
			MinValidPerWindow:          minValidPerWindow,
		},
		[]types.ExchangeRateTuple{},
		[]types.FeederDelegation{},
		[]types.MissCounter{},
		[]types.AggregateExchangeRatePrevote{},
		[]types.AggregateExchangeRateVote{},
		[]types.TobinTax{},
	)

	bz, err := json.MarshalIndent(&oracleGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated oracle parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(oracleGenesis)
}
