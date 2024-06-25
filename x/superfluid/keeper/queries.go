package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/osmomath"
)

func (k Keeper) GetTotalSuperfluidDelegations(ctx sdk.Context) (sdkmath.Int, error) {
	totalSuperfluidDelegated := osmomath.NewInt(0)

	intermediaryAccounts := k.GetAllIntermediaryAccounts(ctx)
	for _, intermediaryAccount := range intermediaryAccounts {
		valAddr, err := sdk.ValAddressFromBech32(intermediaryAccount.ValAddr)
		if err != nil {
			return sdkmath.Int{}, err
		}

		val, err := k.sk.GetValidator(ctx, valAddr)
		if err != nil {
			return sdkmath.Int{}, stakingtypes.ErrNoValidatorFound
		}

		delegation, err := k.sk.GetDelegation(ctx, intermediaryAccount.GetAccAddress(), valAddr)
		if err != nil {
			continue
		}

		syntheticOsmoAmt := delegation.Shares.Quo(val.DelegatorShares).MulInt(val.Tokens).RoundInt()
		totalSuperfluidDelegated = totalSuperfluidDelegated.Add(syntheticOsmoAmt)
	}
	return totalSuperfluidDelegated, nil
}
