package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
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

		for _, asset := range k.GetAllSuperfluidAssets(ctx) {
			priceMultiplier := sdk.NewInt(10000)
			twap := sdk.NewDecFromInt(priceMultiplier)
			if asset.AssetType == types.SuperfluidAssetTypeLPShare {
				// LP_token_Osmo_equivalent = OSMO_amount_on_pool / LP_token_supply
				poolId := gammtypes.MustGetPoolIdFromShareDenom(asset.Denom)
				pool, err := k.gk.GetPool(ctx, poolId)
				if err != nil {
					panic(err)
				}
				// get OSMO amount
				osmoPoolAsset, err := pool.GetPoolAsset(appparams.BaseCoinUnit)
				if err != nil {
					panic(err)
				}

				twap = sdk.Dec(osmoPoolAsset.Token.Amount.Mul(priceMultiplier).Quo(pool.GetTotalShares().Amount))
			} else if asset.AssetType == types.SuperfluidAssetTypeNative {
				// TODO: should get twap price from gamm module and use the price
				// which pool should it use to calculate native token price?
			}
			k.SetEpochOsmoEquivalentTWAP(ctx, epochNumber, asset.Denom, twap)
		}
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
