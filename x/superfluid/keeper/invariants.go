package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

const totalSuperfluidDelegationInvariantName = "total-superfluid-delegation-invariant-name"

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(types.ModuleName, totalSuperfluidDelegationInvariantName, TotalSuperfluidDelegationInvariant(keeper))
}

// AllInvariants runs all invariants of the gamm module.
func AllInvariants(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return TotalSuperfluidDelegationInvariant(keeper)(ctx)
	}
}

// TotalSuperfluidDelegationInvariant checks the sum of intermediary account delegation is same as sum of individual lockup delegation.
func TotalSuperfluidDelegationInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		accs := keeper.GetAllIntermediaryAccounts(ctx)
		totalSuperfluidDelegationTokens := osmomath.ZeroDec()

		// Compute the total amount delegated from all intermediary accounts
		for _, acc := range accs {
			valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
			if err != nil {
				return sdk.FormatInvariant(types.ModuleName, totalSuperfluidDelegationInvariantName,
					"\tinvalid validator address exists"), true
			}
			validator, err := keeper.sk.GetValidator(ctx, valAddr)
			if err != nil {
				return sdk.FormatInvariant(types.ModuleName, totalSuperfluidDelegationInvariantName,
					"\tvalidator does not exists for specified validator address on intermediary account"), true
			}
			delegation, err := keeper.sk.GetDelegation(ctx, acc.GetAccAddress(), valAddr)
			if err == nil {
				tokens := validator.TokensFromShares(delegation.Shares)
				totalSuperfluidDelegationTokens = totalSuperfluidDelegationTokens.Add(tokens)
			}
		}

		// Compute the total delegation amount expected
		// from every lockID intermediary account connections
		totalExpectedSuperfluidAmount := osmomath.ZeroInt()
		connections := keeper.GetAllLockIdIntermediaryAccountConnections(ctx)
		for _, connection := range connections {
			lockId := connection.LockId
			lock, err := keeper.lk.GetLockByID(ctx, lockId)
			if err != nil || lock == nil {
				return sdk.FormatInvariant(types.ModuleName, totalSuperfluidDelegationInvariantName,
					"\tinvalid superfluid lock id exists with no actual lockup"), true
			}
			if len(lock.Coins) != 1 {
				return sdk.FormatInvariant(types.ModuleName, totalSuperfluidDelegationInvariantName,
					"\tonly single coin lockup is eligible for superfluid staking"), true
			}
			amount, err := keeper.GetSuperfluidOSMOTokens(ctx, lock.Coins[0].Denom, lock.Coins[0].Amount)
			if err != nil {
				return sdk.FormatInvariant(types.ModuleName, totalSuperfluidDelegationInvariantName,
					"\tunderlying LP share no longer elidible for superfluid staking"), true
			}
			totalExpectedSuperfluidAmount = totalExpectedSuperfluidAmount.Add(amount)
		}

		if !totalExpectedSuperfluidAmount.Equal(totalSuperfluidDelegationTokens.TruncateInt()) {
			return sdk.FormatInvariant(types.ModuleName,
					totalSuperfluidDelegationInvariantName,
					fmt.Sprintf("\ttotal superfluid intermediary account delegation amount does not match total sum of lockup delegations: %s != %s\n", totalExpectedSuperfluidAmount.String(), totalSuperfluidDelegationTokens.String())),
				true
		}

		return sdk.FormatInvariant(types.ModuleName, totalSuperfluidDelegationInvariantName,
			"\ttotal superfluid intermediary account delegation amount matches total sum of lockup delegations\n"), false
	}
}
