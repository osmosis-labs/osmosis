package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	if epochIdentifier == params.RefreshEpochIdentifier {
		for _, asset := range k.GetAllSuperfluidAssets(ctx) {
			// TODO: should include unlocking asset as well
			totalAmt := k.lk.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         asset.Denom,
				Duration:      time.Second,
			})
			k.SetSuperfluidAssetInfo(ctx, types.SuperfluidAssetInfo{
				Denom:                      asset.Denom,
				TotalStakedAmount:          totalAmt,
				RiskAdjustedOsmoEquivalent: k.GetRiskAdjustedOsmoValue(ctx, asset),
			})
		}

		// slashing
		// 	Currently for double signs, we iterate over all unbondings and all redelegations. We handle slashing delegated tokens, via a “rebase” factor.
		// 	Meaning, that if we have a 10% slash say, we just alter the conversion rate between “delegation pool shares” and “osmo” when withdrawing your stake.
		// 	Now in our case, we currently don’t have a method for a “rebase” factor in synthetic lockups.
		// 	Eugen: We can add this rebase factor to our Superfluid module, to be executed upon MsgUnbondStake or w/e its called
		// 	Dev: I don’t think we need to worry about deferring iteration

		// staking
		// 	Iterate module accounts
		// 	Need to decide between Module account per LP token & module account per (LP token, validator token pair)
		// 	per LP token
		// 	Then we have to iterate over every delegator. (Potentially millions of entries)
		// 	per (LP token, validator addr pair)
		// 	at 200 superfluid enabled LP tokens, 100 validators, this is 20k module accounts. Very quick to iterate over.
		// 	Gauge rewards are once per denom
		// 	Decided, one module account & one denom per (LP token, validator addr pair)
		// 	Move delegation rewards to perpetual gauge per (LP token, validator addr pair)
		// 	Update the module accounts’ delegation amount based on changed TWAP
		// 	We will need add something to staking, so that we can do this, without triggering unbonding logic.
		// 	We do this via keeper method, not via message
		// 	Just needs to be added, and ensure no weird unforeseen edge cases
	}
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
