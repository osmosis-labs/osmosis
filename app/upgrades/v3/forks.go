package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/osmosis-labs/osmosis/v8/app/keepers"
)

// RunForkLogic executes height-gated on-chain fork logic for the Osmosis v3
// upgrade.
func RunForkLogic(ctx sdk.Context, keepers *keepers.AppKeepers) {
	ctx.Logger().Info("Applying Osmosis v3 upgrade." +
		" Fixing governance deposit so proposals can be voted upon," +
		" and fixing validator min commission rate.")
	FixMinDepositDenom(ctx, keepers.GovKeeper)
	FixMinCommisionRate(ctx, keepers.StakingKeeper)
}

// Fixes an error where minimum deposit was set to "500 osmo". This denom does
// not exist, which makes it impossible for a proposal to go to a vote.
func FixMinDepositDenom(ctx sdk.Context, gov *govkeeper.Keeper) {
	params := gov.GetDepositParams(ctx)
	params.MinDeposit = sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(500000000)))
	gov.SetDepositParams(ctx, params)
}

// Fixes an error where validators can be created with a commission rate less
// than the network minimum rate.
func FixMinCommisionRate(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	// Upgrade every validators min-commission rate
	validators := staking.GetAllValidators(ctx)
	minCommissionRate := staking.GetParams(ctx).MinCommissionRate
	for _, v := range validators {
		if v.Commission.Rate.LT(minCommissionRate) {
			comm, err := staking.MustUpdateValidatorCommission(ctx, v, minCommissionRate)
			if err != nil {
				panic(err)
			}

			v.Commission = comm

			// call the before-modification hook since we're about to update the commission
			staking.BeforeValidatorModified(ctx, v.GetOperator())
			staking.SetValidator(ctx, v)
		}
	}
}
