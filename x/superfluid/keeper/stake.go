package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) SuperfluidDelegate(ctx sdk.Context, lockID uint64, valAddr string) error {
	// Register a synthetic lockup for superfluid staking with `superdelegation{valAddr}` suffix
	suffix := fmt.Sprintf("superdelegation%s", valAddr)
	k.lk.CreateSyntheticLockup(ctx, lockID, suffix, false)

	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	if lock.Coins.Len() != 1 {
		return fmt.Errorf("multiple coins lock is not supported")
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
	}
	return nil
}

func (k Keeper) SuperfluidRedelegate(lockID uint64, newValAddr string) {
	// TODO: Delete previous synthetic lockup, should use native lockup module only or just record lockID
	// - shadow pair on superfluid module?
	// Since synthetic lockup could be used in several places, would be better to create matching on own storage
	// TODO: Create unbonding synthetic lockup for previous shadow
	// synthetic suffix = `redelegating{valAddr}`
	// TODO: Register a synthetic lockup, call SuperfluidDelegate?
	// TODO: Unbonding amount should be modified for TWAP change or not?
}

func (k Keeper) SuperfluidUndelegate(lockID uint64) {
	// Create unbonding synthetic lockup
	// TODO: Unbonding amount should be modified for TWAP change or not?
	// synthetic suffix = `unbonding{valAddr}`
}

func (k Keeper) SuperfluidWithdraw(lockID uint64) {
	// It looks like LP token will be automatically removed by lockup module
	// TODO: If there's any local storage used by superfluid module for each lockID, just clean it up.
	// TODO: automatically done or manually done?
	// TODO: check synthetic suffix = `unbonding{valAddr}`, lockID is matured and removed already on lockup storage
}

// TODO: Need to (eventually) override the existing staking messages and queries, for undelegating, delegating, rewards, and redelegating, to all be going through all superfluid module.
// Want integrators to be able to use the same staking queries and messages
// Eugen’s point: Only rewards message needs to be updated. Rest of messages are fine
// Queries need to be updated
// We can do this at the very end though, since it just relates to queries.
