package keeper

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
	"time"

	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeEpochStart is the epoch start hook.
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	params := k.GetParams(ctx)

	if epochIdentifier == params.VotePeriodEpochIdentifier {
		// Build claim map over all validators in active set
		validatorClaimMap := make(map[string]types.Claim)

		maxValidators, err := k.StakingKeeper.MaxValidators(ctx)
		if err != nil {
			panic("cannot get max validators")
		}
		iterator, err := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
		defer iterator.Close()
		if err != nil {
			panic("cannot get validators power store iterator")
		}

		powerReduction := k.StakingKeeper.PowerReduction(ctx)

		i := 0
		for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
			validator, err := k.StakingKeeper.GetValidator(ctx, iterator.Value())

			// Exclude not bonded validator
			if err == nil && validator.IsBonded() {
				valAddrStr := validator.GetOperator()
				valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
				if err != nil {
					panic("invalid validator address")
				}
				validatorClaimMap[valAddrStr] = types.NewClaim(validator.GetConsensusPower(powerReduction), 0, 0, valAddr)
				i++
			}
		}

		// Denom-TobinTax map
		voteTargets := make(map[string]osmomath.Dec)
		k.IterateTobinTaxes(ctx, func(denom string, tobinTax osmomath.Dec) bool {
			voteTargets[denom] = tobinTax
			return false
		})

		// Clear all exchange rates
		k.IterateNoteExchangeRates(ctx, func(denom string, _ osmomath.Dec) (stop bool) {
			k.DeleteMelodyExchangeRate(ctx, denom)
			return false
		})

		// Organize votes to ballot by denom
		// NOTE: **Filter out inactive or jailed validators**
		// NOTE: **Make abstain votes to have zero vote power**
		voteMap := k.OrganizeBallotByDenom(ctx, validatorClaimMap)

		if referenceSymphony, err := PickReferenceSymphony(ctx, k, voteTargets, voteMap); err == nil && referenceSymphony != "" {
			// make voteMap of Reference Symphony to calculate cross exchange rates
			ballotRT := voteMap[referenceSymphony]
			voteMapRT := ballotRT.ToMap()
			exchangeRateRT := ballotRT.WeightedMedian()

			// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
			for denom, ballot := range voteMap {
				// Convert ballot to cross exchange rates
				if denom != referenceSymphony {
					ballot = ballot.ToCrossRateWithSort(voteMapRT)
				}

				// Get weighted median of cross exchange rates
				exchangeRate := Tally(ballot, params.RewardBand, validatorClaimMap)

				// Transform into the original form unote/stablecoin
				if denom != referenceSymphony {
					exchangeRate = exchangeRateRT.Quo(exchangeRate)
				}

				// Set the exchange rate, emit ABCI event
				k.SetMelodyExchangeRateWithEvent(ctx, denom, exchangeRate)
			}
		}

		//---------------------------
		// Do miss counting & slashing
		voteTargetsLen := len(voteTargets)
		for _, claim := range validatorClaimMap {
			// Skip abstain & valid voters
			if int(claim.WinCount) == voteTargetsLen {
				continue
			}

			// Increase miss counter
			k.SetMissCounter(ctx, claim.Recipient, k.GetMissCounter(ctx, claim.Recipient)+1)
		}

		// TODO: no rewards so far
		// Distribute rewards to ballot winners
		//k.RewardBallotWinners(
		//	ctx,
		//	(int64)(params.VotePeriod),
		//	(int64)(params.RewardDistributionWindow),
		//	voteTargets,
		//	validatorClaimMap,
		//)

		// Clear the ballot
		k.ClearBallots(ctx, uint64(epochNumber))

		// Update vote targets and tobin tax
		k.ApplyWhitelist(ctx, params.Whitelist, voteTargets)
	}

	// Do slash who did miss voting over threshold and
	// reset miss counters of all validators at the last block of slash window
	if params.SlashWindowEpochIdentifier == epochIdentifier {
		// TODO: yurii: enable slashing
		//k.SlashAndResetMissCounters(ctx)
	}

	return nil
}

// ___________________________________________________________________________________________________

// Hooks is the wrapper struct for the incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Hooks returns the hook wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// GetModuleName implements types.EpochHooks.
func (Hooks) GetModuleName() string {
	return types.ModuleName
}

// BeforeEpochStart is the epoch start hook.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

// AfterEpochEnd is the epoch end hook.
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
