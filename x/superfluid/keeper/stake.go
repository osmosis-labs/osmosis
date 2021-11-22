package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) SuperfluidDelegate(ctx sdk.Context, lockID uint64, valAddr string) error {
	// Register a synthetic lockup for superfluid staking with `superbonding{valAddr}` suffix
	suffix := fmt.Sprintf("superbonding%s", valAddr)
	k.lk.CreateSyntheticLockup(ctx, lockID, suffix, false)

	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	if lock.Coins.Len() != 1 {
		return fmt.Errorf("multiple coins lock is not supported")
	}

	// prevent unbonding lockups to be not able to be used for superfluid staking
	if lock.IsUnlocking() {
		return fmt.Errorf("unbonding lockup is not allowed to participate in superfluid staking")
	}

	// length check
	if lock.Duration < time.Hour*24*14 { // if less than 2 weeks bonding, skip
		return fmt.Errorf("lockup does not have enough lock duration")
	}

	// create intermediary account that converts LP token to OSMO
	acc := types.SuperfluidIntermediaryAccount{
		Denom:   lock.Coins[0].Denom,
		ValAddr: valAddr,
	}

	mAddr := acc.GetAddress()
	twap := k.GetLastEpochOsmoEquivalentTWAP(ctx, acc.Denom)
	if !twap.EpochTwapPrice.IsZero() {
		// mint OSMO token based on TWAP of locked denom to denom module account
		decAmt := twap.EpochTwapPrice.Mul(sdk.Dec(lock.Coins.AmountOf(acc.Denom)))
		amt := decAmt.RoundInt()
		coins := sdk.Coins{sdk.NewCoin(appparams.BaseCoinUnit, amt)}
		k.bk.MintCoins(ctx, minttypes.ModuleName, coins)
		k.bk.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, mAddr, coins)

		// make delegation from module account to the validator
		valAddress, err := sdk.ValAddressFromBech32(valAddr)
		if err != nil {
			return err
		}
		validator, found := k.sk.GetValidator(ctx, valAddress)
		if !found {
			return fmt.Errorf("validator not found")
		}
		_, err = k.sk.Delegate(ctx, mAddr, amt, stakingtypes.Unbonded, validator, true)
		if err != nil {
			return err
		}

		// create a perpetual gauge to send staking distribution rewards to
		acc.GaugeId, err = k.ik.CreateGauge(ctx, true, mAddr, sdk.Coins{}, lockuptypes.QueryCondition{}, ctx.BlockTime(), 1)
		if err != nil {
			return err
		}

		// connect intermediary account struct to its address
		k.SetIntermediaryAccount(ctx, acc)

		// create connection record between lock id and intermediary account
		k.SetLockIdIntermediaryAccountConnection(ctx, lockID, acc)
	}

	return nil
}

func (k Keeper) SuperfluidUndelegate(ctx sdk.Context, lockID uint64) error {
	// Remove previously created synthetic lockup
	intermediaryAccAddr := k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	intermediaryAcc := k.GetIntermediaryAccount(ctx, intermediaryAccAddr)
	suffix := fmt.Sprintf("superbonding%s", intermediaryAcc.ValAddr)
	err := k.lk.DeleteSyntheticLockup(ctx, lockID, suffix)
	if err != nil {
		return err
	}

	// unbonding synthetic suffix = `unbonding{valAddr}`
	suffix = fmt.Sprintf("superunbonding%s", intermediaryAcc.ValAddr)
	// TODO: synthetic lockup unbonding duration should be different from regular unbonding lockup, should set the duration here
	k.lk.CreateSyntheticLockup(ctx, lockID, suffix, true)

	// TODO: Unbonding amount should be modified for TWAP change or not?
	return nil
}

func (k Keeper) SuperfluidRedelegate(ctx sdk.Context, lockID uint64, newValAddr string) error {
	err := k.SuperfluidUndelegate(ctx, lockID)
	if err != nil {
		return err
	}

	k.SuperfluidDelegate(ctx, lockID, newValAddr)
	return nil
}

func (k Keeper) SuperfluidWithdraw(lockID uint64) {
	// It looks like LP token will be automatically removed by lockup module
	// TODO: If there's any local storage used by superfluid module for each lockID, just clean it up.
	// TODO: automatically done or manually done?
	// TODO: check synthetic suffix = `unbonding{valAddr}`, lockID is matured and removed already on lockup storage
}

// TODO: Implement hook for native lockup unbonding

// TODO: Need to (eventually) override the existing staking messages and queries, for undelegating, delegating, rewards, and redelegating, to all be going through all superfluid module.
// Want integrators to be able to use the same staking queries and messages
// Eugen’s point: Only rewards message needs to be updated. Rest of messages are fine
// Queries need to be updated
// We can do this at the very end though, since it just relates to queries.
