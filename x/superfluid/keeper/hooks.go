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
			// TODO: should we enable all the locks for specific lp token
			// or only locks that people want to participiate in superfluid staking within those locks?
			totalAmt := k.lk.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         asset.Denom,
				Duration:      time.Second,
			})
			k.SetSuperfluidAssetInfo(ctx, types.SuperfluidAssetInfo{
				Denom:                      asset.Denom,
				TotalStakedAmount:          totalAmt,
				RiskAdjustedOsmoEquivalent: k.GetRiskAdjustedOsmoValue(ctx, asset, totalAmt),
			})
		}

		// TODO: how to get GAMM pool ids to fetch price info - do all ids?
		// TODO: should get twap price from gamm module and use the price
		k.SetEpochTwapPrice(ctx, epochNumber, 1, sdk.NewDec(1))
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
