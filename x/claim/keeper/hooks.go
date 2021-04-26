package keeper

import (
	"time"

	"github.com/c-osmosis/osmosis/x/claim/types"
	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) AfterAddLiquidity(ctx sdk.Context, sender sdk.AccAddress) {
	if k.SetUserAction(ctx, sender, types.ActionAddLiquidity) {
		k.ClaimCoins(ctx, sender.String())
	}
}

func (k Keeper) AfterSwap(ctx sdk.Context, sender sdk.AccAddress) {
	if k.SetUserAction(ctx, sender, types.ActionSwap) {
		k.ClaimCoins(ctx, sender.String())
	}
}

func (k Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	if k.SetUserAction(ctx, voterAddr, types.ActionVote) {
		k.ClaimCoins(ctx, voterAddr.String())
	}
}

func (k Keeper) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	if k.SetUserAction(ctx, delAddr, types.ActionDelegateStake) {
		k.ClaimCoins(ctx, delAddr.String())
	}
}

//_________________________________________________________________________________________

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k Keeper
}

var _ gammtypes.GammHooks = Hooks{}
var _ lockuptypes.LockupHooks = Hooks{}
var _ govtypes.GovHooks = Hooks{}
var _ stakingtypes.StakingHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// gamm hooks
func (h Hooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	h.k.AfterAddLiquidity(ctx, sender)
}
func (h Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	h.k.AfterAddLiquidity(ctx, sender)
}
func (h Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
}
func (h Hooks) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	h.k.AfterSwap(ctx, sender)
}

// lockup hooks
func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
}
func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
}

// governance hooks
func (h Hooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {}
func (h Hooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
}

func (h Hooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	h.k.AfterProposalVote(ctx, proposalID, voterAddr)
}

func (h Hooks) AfterProposalInactive(ctx sdk.Context, proposalID uint64) {}
func (h Hooks) AfterProposalActive(ctx sdk.Context, proposalID uint64)   {}

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
	h.k.BeforeDelegationCreated(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {}
