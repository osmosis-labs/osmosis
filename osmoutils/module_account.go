package osmoutils

import (
	"errors"
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// OsmoUtilsExtraAccountTypes is a map of extra account types that can be overridden.
// This is defined as a global variable so it can be modified in the chain's app.go and used here without
// having to import the chain. Specifically, this is used for compatibility with Osmosis' Cosmos SDK fork
var OsmoUtilsExtraAccountTypes map[reflect.Type]struct{}

type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
}

// CanCreateModuleAccountAtAddr tells us if we can safely make a module account at
// a given address. By collision resistance of the address (given API safe construction),
// the only way for an account to be already be at this address is if its claimed by the same
// pre-image from the correct module,
// or some SDK command breaks assumptions and creates an account at designated address.
// This function checks if there is an account at that address, and runs some safety checks
// to be extra-sure its not a user account (e.g. non-zero sequence, pubkey, of fore-seen account types).
// If there is no account, or if we believe its not a user-spendable account, we allow module account
// creation at the address.
// else, we do not.
//
// TODO: This is generally from an SDK design flaw
// code based off wasmd code: https://github.com/CosmWasm/wasmd/pull/996
// Its _mandatory_ that the caller do the API safe construction to generate a module account addr,
// namely, address.Module(ModuleName, {key})
func CanCreateModuleAccountAtAddr(ctx sdk.Context, ak AccountKeeper, addr sdk.AccAddress) error {
	existingAcct := ak.GetAccount(ctx, addr)
	if existingAcct == nil {
		return nil
	}
	if existingAcct.GetSequence() != 0 || existingAcct.GetPubKey() != nil {
		return fmt.Errorf("cannot create module account %s, "+
			"due to an account at that address already existing & having sent txs", addr)
	}
	overrideAccountTypes := map[reflect.Type]struct{}{
		reflect.TypeOf(&authtypes.BaseAccount{}):                 {},
		reflect.TypeOf(&vestingtypes.DelayedVestingAccount{}):    {},
		reflect.TypeOf(&vestingtypes.ContinuousVestingAccount{}): {},
		reflect.TypeOf(&vestingtypes.BaseVestingAccount{}):       {},
		reflect.TypeOf(&vestingtypes.PeriodicVestingAccount{}):   {},
		reflect.TypeOf(&vestingtypes.PermanentLockedAccount{}):   {},
	}
	for extraAccountType := range OsmoUtilsExtraAccountTypes {
		overrideAccountTypes[extraAccountType] = struct{}{}
	}

	if _, clear := overrideAccountTypes[reflect.TypeOf(existingAcct)]; clear {
		return nil
	}
	return errors.New("cannot create module account %s, " +
		"due to an account at that address already existing & not being an overrideable type")
}

// CreateModuleAccount creates a module account at the provided address.
// It overrides an account if it exists at that address, with a non-zero sequence number & pubkey
// Contract: addr is derived from `address.Module(ModuleName, key)`
func CreateModuleAccount(ctx sdk.Context, ak AccountKeeper, addr sdk.AccAddress) error {
	err := CanCreateModuleAccountAtAddr(ctx, ak, addr)
	if err != nil {
		return err
	}

	acc := ak.NewAccount(
		ctx,
		authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(addr),
			addr.String(),
		),
	)
	ak.SetAccount(ctx, acc)
	return nil
}
