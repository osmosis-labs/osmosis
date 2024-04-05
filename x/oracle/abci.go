package oracle

import (
	"time"

	appparams "github.com/osmosis-labs/osmosis/v23/app/params"

	"github.com/osmosis-labs/osmosis/v23/x/oracle/keeper"

	"github.com/osmosis-labs/osmosis/v23/x/oracle/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	params := k.GetParams(ctx)
	if appparams.IsPeriodLastBlock(ctx, params.VotePeriod) {

		// Build claim map over all validators in active set
		validatorClaimMap := make(map[string]types.Claim)

		maxValidators := k.StakingKeeper.MaxValidators(ctx)
		iterator := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
		defer iterator.Close()

		powerReduction := k.StakingKeeper.PowerReduction(ctx)

		i := 0
		for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
			validator := k.StakingKeeper.Validator(ctx, iterator.Value())

			// Exclude not bonded validator
			if validator.IsBonded() {
				valAddr := validator.GetOperator()
				validatorClaimMap[valAddr.String()] = types.NewClaim(validator.GetConsensusPower(powerReduction), 0, 0, valAddr)
				i++
			}
		}

		// Denom-TobinTax map
		voteTargets := make(map[string]sdk.Dec)
		k.IterateTobinTaxes(ctx, func(denom string, tobinTax sdk.Dec) bool {
			voteTargets[denom] = tobinTax
			return false
		})

		// Clear all exchange rates
		k.IterateLunaExchangeRates(ctx, func(denom string, _ sdk.Dec) (stop bool) {
			k.DeleteOsmoExchangeRate(ctx, denom)
			return false
		})

		// Organize votes to ballot by denom
		// NOTE: **Filter out inactive or jailed validators**
		// NOTE: **Make abstain votes to have zero vote power**
		voteMap := k.OrganizeBallotByDenom(ctx, validatorClaimMap)

		if referenceTerra := PickReferenceTerra(ctx, k, voteTargets, voteMap); referenceTerra != "" {
			// make voteMap of Reference Terra to calculate cross exchange rates
			ballotRT := voteMap[referenceTerra]
			voteMapRT := ballotRT.ToMap()
			exchangeRateRT := ballotRT.WeightedMedian()

			// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
			for denom, ballot := range voteMap {

				// Convert ballot to cross exchange rates
				if denom != referenceTerra {
					ballot = ballot.ToCrossRateWithSort(voteMapRT)
				}

				// Get weighted median of cross exchange rates
				exchangeRate := Tally(ballot, params.RewardBand, validatorClaimMap)

				// Transform into the original form uluna/stablecoin
				if denom != referenceTerra {
					exchangeRate = exchangeRateRT.Quo(exchangeRate)
				}

				// Set the exchange rate, emit ABCI event
				k.SetOsmoExchangeRateWithEvent(ctx, denom, exchangeRate)
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

		// Distribute rewards to ballot winners
		k.RewardBallotWinners(
			ctx,
			(int64)(params.VotePeriod),
			(int64)(params.RewardDistributionWindow),
			voteTargets,
			validatorClaimMap,
		)

		// Clear the ballot
		k.ClearBallots(ctx, params.VotePeriod)

		// Update vote targets and tobin tax
		k.ApplyWhitelist(ctx, params.Whitelist, voteTargets)
	}

	// Do slash who did miss voting over threshold and
	// reset miss counters of all validators at the last block of slash window
	if appparams.IsPeriodLastBlock(ctx, params.SlashWindow) {
		k.SlashAndResetMissCounters(ctx)
	}
}
