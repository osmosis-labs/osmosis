package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) GetAllIntermediaryAccounts() []types.SuperfluidIntermediaryAccount {
	// TODO: implement
	return []types.SuperfluidIntermediaryAccount{}
}

func (k Keeper) SetIntermediaryAccount(acc types.SuperfluidIntermediaryAccount) {
	// TODO: implement
}

func (k Keeper) DeleteIntermediaryAccount(acc types.SuperfluidIntermediaryAccount) {
	// TODO: implement
}

func (k Keeper) SetSyntheticLockupOwner(acc types.SuperfluidIntermediaryAccount, synthLock lockuptypes.SyntheticLock) {
	// TODO: this might be not be useful since synthetic lockup is already supporting this
	// addr := acc.GetAddress()
	// set on superfluid storage | `{address}{lockid}{suffix}` => synthLock
}

func (k Keeper) GetOwnedSyntheticLockups(acc types.SuperfluidIntermediaryAccount) []lockuptypes.SyntheticLock {
	// TODO: this might be not be useful since synthetic lockup is already supporting this
	// addr := acc.GetAddress()
	// TODO: read from iterator for `{address}` prefix
	return []lockuptypes.SyntheticLock{}
}

func (k Keeper) GetIntermediaryAccountOSMODelegation(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// addr := acc.GetAddress()
	// k.sk.GetDelegation(addr, acc.ValAddr) * ...
	return sdk.OneInt()
}

func (k Keeper) GetLPTokenAmount(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// addr := acc.GetAddress()
	// k.bk.GetBalance(addr)
	return sdk.OneInt()
}

func (k Keeper) GetMintedOSMOAmount(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// TODO: read from own superfluid storage
	return sdk.OneInt()
}

func (k Keeper) SetMintedOSMOAmount(acc types.SuperfluidIntermediaryAccount, amount sdk.Int) {
	// TODO: write on own superfluid storage
}

func (k Keeper) GetSlashedOSMOAmount(acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// Slashed amount = Minted OSMO amount - Delegation amount
	return k.GetMintedOSMOAmount(acc).Sub(k.GetIntermediaryAccountOSMODelegation(acc))
}

// TODO: On TWAP change, set target OSMO mint amount to TWAP * LP
