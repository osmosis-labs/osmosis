package keeper

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, _ int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) {
	params := k.GetParams(ctx)
	if epochIdentifier == params.RefreshEpochIdentifier {
		// cref [#830](https://github.com/osmosis-labs/osmosis/issues/830),
		// the supplied epoch number is wrong at time of commit. hence we get from the info.
		epochNumber := k.ek.GetEpochInfo(ctx, epochIdentifier).CurrentEpoch

		// Slash all module accounts' LP token based on slash amount before twap update
		for _, asset := range k.GetAllSuperfluidAssets(ctx) {
			err := k.updateEpochTwap(ctx, asset, epochNumber)
			if err != nil {
				// TODO: Revisit what we do here. (halt all distr, only skip this asset)
				// Since at MVP of feature, we only have one pool of superfluid staking,
				// we can punt this question.
				// each of the errors feels like significant misconfig
				return
			}
		}

		// Move delegation rewards to perpetual gauge
		k.MoveSuperfluidDelegationRewardToGauges(ctx)

		// Refresh intermediary accounts' delegation amounts
		k.RefreshIntermediaryDelegationAmounts(ctx)
	}
}

func (k Keeper) updateEpochTwap(ctx sdk.Context, asset types.SuperfluidAsset, epochNumber int64) error {
	if asset.AssetType == types.SuperfluidAssetTypeLPShare {
		// LP_token_Osmo_equivalent = OSMO_amount_on_pool / LP_token_supply
		poolId := gammtypes.MustGetPoolIdFromShareDenom(asset.Denom)
		pool, err := k.gk.GetPool(ctx, poolId)
		if err != nil {
			// Pool has been unexpectedly deleted
			k.Logger(ctx).Error(err.Error())
			k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
			return err
		}

		// get OSMO amount
		bondDenom := k.sk.BondDenom(ctx)
		osmoPoolAsset, err := pool.GetPoolAsset(bondDenom)
		if err != nil {
			// Pool has unexpectedly removed Osmo from its assets.
			k.Logger(ctx).Error(err.Error())
			k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
			return err
		}

		twap := osmoPoolAsset.Token.Amount.ToDec().Quo(pool.GetTotalShares().Amount.ToDec())
		k.SetEpochOsmoEquivalentTWAP(ctx, epochNumber, asset.Denom, twap)
	} else if asset.AssetType == types.SuperfluidAssetTypeNative {
		// TODO: should get twap price from gamm module and use the price
		// which pool should it use to calculate native token price?
		k.Logger(ctx).Error("unsupported superfluid asset type")
		return errors.New("SuperfluidAssetTypeNative is unspported")
	}
	return nil
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

// lockup hooks
func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	// undelegate automatically when start unlocking if superfluid staking is available
	intermediaryAccAddr := h.k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if !intermediaryAccAddr.Empty() {
		// superfluid delegate for additional amount
		err := h.k.SuperfluidDelegateMore(ctx, lockID, amount)
		if err != nil {
			h.k.Logger(ctx).Error(err.Error())
		}
	}
}

func (h Hooks) OnStartUnlock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	// undelegate automatically when start unlocking if superfluid staking is available
	intermediaryAccAddr := h.k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if !intermediaryAccAddr.Empty() {
		_, err := h.k.SuperfluidUndelegate(ctx, address.String(), lockID)
		if err != nil {
			h.k.Logger(ctx).Error(err.Error())
			// TODO: If not panic, there could be the case user get infinite amount of rewards without actual lockup
			panic(err)
		}
	}
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {

}

func (h Hooks) OnTokenSlashed(ctx sdk.Context, lockID uint64, amount sdk.Coins) {

}

// staking hooks
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)   {}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	h.k.SlashLockupsForValidatorSlash(ctx, valAddr, fraction)
}
func (h Hooks) BeforeSlashingUnbondingDelegation(ctx sdk.Context, unbondingDelegation stakingtypes.UnbondingDelegation,
	infractionHeight int64, slashFactor sdk.Dec) {
	h.k.SlashLockupsForUnbondingDelegationSlash(ctx, unbondingDelegation.DelegatorAddress, unbondingDelegation.ValidatorAddress, slashFactor)
}

func (h Hooks) BeforeSlashingRedelegation(ctx sdk.Context, srcValidator stakingtypes.Validator, redelegation stakingtypes.Redelegation,
	infractionHeight int64, slashFactor sdk.Dec) {
}
