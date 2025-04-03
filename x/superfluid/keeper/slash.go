package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SlashLockupsForValidatorSlash should be called before the validator at valAddr is slashed.
// This function is responsible for inspecting every intermediate account to valAddr.
// For each intermediate account IA, it slashes every constituent delegation behind IA.
// Furthermore, if the infraction height is sufficiently old, slashes unbondings
// Note: Based on sdk.staking.Slash function review, slashed tokens are burnt not sent to community pool
// we ignore that, and send the underliyng tokens to the community pool anyway.
func (k Keeper) SlashLockupsForValidatorSlash(context context.Context, valAddr sdk.ValAddress, slashFactor osmomath.Dec) {
	// Important note: The SDK slashing for historical heights is wrong.
	// It defines a "slash amount" off of the live staked amount.
	// Then it charges all the unbondings & redelegations at the slash factor.
	// It then creates a new slash factor for the amount remaining to be charged from the slash amount,
	// across all the live accounts.
	// This is the "effectiveSlashFactor".
	//
	// The SDK's design is wack / wrong in our view, and this was a pre Cosmos Hub
	// launch hack that never got remedied.
	// We are not concerned about maximal consistency with the SDK, and instead charge slashFactor to
	// both unbonding and live delegations. Rather than slashFactor to unbonding delegations,
	// and effectiveSlashFactor to new delegations.
	ctx := sdk.UnwrapSDKContext(context)
	accs := k.GetIntermediaryAccountsForVal(ctx, valAddr)

	// for every intermediary account, we first slash the live tokens comprosing delegated to it,
	// and then all of its unbonding delegations.
	// We do these slashes as burns.
	for _, acc := range accs {
		locks := k.lk.GetLocksLongerThanDurationDenom(ctx, acc.Denom, time.Second)
		for _, lock := range locks {
			// slashing only applies to synthetic lockup amount
			synthLock, err := k.lk.GetSyntheticLockup(ctx, lock.ID, stakingSyntheticDenom(acc.Denom, acc.ValAddr))
			// synth lock doesn't exist for bonding
			if err != nil {
				synthLock, err = k.lk.GetSyntheticLockup(ctx, lock.ID, unstakingSyntheticDenom(acc.Denom, acc.ValAddr))
				// synth lock doesn't exist for unbonding
				// => no superfluid staking on this lock ID, so continue
				if err != nil {
					continue
				}
			}

			// slash the lock whether its bonding or unbonding.
			// this overslashes unbondings that started unbonding before the slash infraction,
			// but this seems to be an acceptable trade-off based upon choices taken in the SDK.
			k.slashSynthLock(ctx, synthLock, slashFactor)
		}
	}
}

func (k Keeper) slashSynthLock(ctx sdk.Context, synthLock *lockuptypes.SyntheticLock, slashFactor osmomath.Dec) {
	// Only single token lock is allowed here
	lock, _ := k.lk.GetLockByID(ctx, synthLock.UnderlyingLockId)
	slashAmt := lock.Coins[0].Amount.ToLegacyDec().Mul(slashFactor)
	lockSharesToSlash := sdk.NewCoins(sdk.NewCoin(lock.Coins[0].Denom, slashAmt.TruncateInt()))

	// If the slashCoins contains a cl denom, we need to update the underlying cl position to reflect the slash.
	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		if strings.HasPrefix(lock.Coins[0].Denom, cltypes.ConcentratedLiquidityTokenPrefix) {
			// Run prepare logic to get the underlying coins to slash.
			// We get the pool address here since the underlying coins will be sent directly from the pool to the community pool instead of the lock module account.
			// Additionally, we update the cl position's state entry to reflect the slash in the position's liquidity.
			poolAddress, underlyingCoinsToSlash, err := k.prepareConcentratedLockForSlash(cacheCtx, lock, slashAmt)
			if err != nil {
				return err
			}
			// Run the normal slashing logic, but instead of sending gamm shares to the community pool, we send the underlying coins
			// the cl shares represent to the community pool and burn the cl shares from the lockup module account as well as the lock itself
			_, err = k.lk.SlashTokensFromLockByIDSendUnderlyingAndBurn(cacheCtx, lock.ID, lockSharesToSlash, underlyingCoinsToSlash, poolAddress)
			return err
		} else {
			// These tokens get moved to the community pool.
			_, err := k.lk.SlashTokensFromLockByID(cacheCtx, lock.ID, lockSharesToSlash)
			return err
		}
	})
}

// prepareConcentratedLockForSlash is a helper function that runs pre-slash logic for concentrated lockups. This function:
// 1. Figures out the underlying assets from the liquidity being slashed and creates a coin object this represents
// 2. Sets the cl position's liquidity state entry to reflect the slash
// 3. Returns the pool address that will send the underlying coins as well as the underlying coins to slash
func (k Keeper) prepareConcentratedLockForSlash(ctx sdk.Context, lock *lockuptypes.PeriodLock, slashAmt osmomath.Dec) (sdk.AccAddress, sdk.Coins, error) {
	// Ensure lock is a single coin lock
	if len(lock.Coins) != 1 {
		return sdk.AccAddress{}, sdk.Coins{}, fmt.Errorf("lock must be a single coin lock, got %s", lock.Coins)
	}

	// Get the position ID from the lock denom
	positionID, err := k.clk.GetPositionIdToLockId(ctx, lock.GetID())
	if err != nil {
		return sdk.AccAddress{}, sdk.Coins{}, err
	}

	// Figure out the underlying assets from the liquidity slash
	position, err := k.clk.GetPosition(ctx, positionID)
	if err != nil {
		return sdk.AccAddress{}, sdk.Coins{}, err
	}

	slashAmtNeg := slashAmt.Neg()

	// If slashAmt is not negative, return an error
	if slashAmtNeg.IsPositive() {
		return sdk.AccAddress{}, sdk.Coins{}, fmt.Errorf("slash amount must be negative, got %s", slashAmt)
	}

	// Create new position object from the position being slashed
	// We use this to safely calculate the underlying assets from the liquidity being slashed
	positionForCalculatingUnderlying := position
	positionForCalculatingUnderlying.Liquidity = slashAmt

	concentratedPool, err := k.clk.GetConcentratedPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.AccAddress{}, sdk.Coins{}, err
	}
	asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(ctx, positionForCalculatingUnderlying, concentratedPool)
	if err != nil {
		return sdk.AccAddress{}, sdk.Coins{}, err
	}

	// Create a coins object to be sent to the community pool
	coinsToSlash := sdk.NewCoins(asset0, asset1)

	// Update the cl positions liquidity to the new amount
	_, err = k.clk.UpdatePosition(ctx, position.PoolId, sdk.MustAccAddressFromBech32(position.Address), position.LowerTick, position.UpperTick, slashAmtNeg, position.JoinTime, position.PositionId)
	if err != nil {
		return sdk.AccAddress{}, sdk.Coins{}, err
	}

	return concentratedPool.GetAddress(), coinsToSlash, nil
}
