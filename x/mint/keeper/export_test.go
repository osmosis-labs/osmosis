package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/mint/types"
)

type (
	ErrInvalidRatio                  = invalidRatioError
	ErrInsufficientDevVestingBalance = insufficientDevVestingBalanceError
)

const (
	EmptyWeightedAddressReceiver = emptyWeightedAddressReceiver
	DeveloperVestingAmount       = developerVestingAmount
)

var (
	ErrAmountCannotBeNilOrZero               = errAmountCannotBeNilOrZero
	ErrDevVestingModuleAccountAlreadyCreated = errDevVestingModuleAccountAlreadyCreated
	ErrDevVestingModuleAccountNotCreated     = errDevVestingModuleAccountNotCreated

	GetProportions = getProportions
)

func (k Keeper) DistributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.Coin, proportion sdk.Dec) (sdk.Int, error) {
	return k.distributeToModule(ctx, recipientModule, mintedCoin, proportion)
}

func (k Keeper) DistributeDeveloperRewards(ctx sdk.Context, totalMintedCoin sdk.Coin, developerRewardsProportion sdk.Dec, developerRewardsReceivers []types.WeightedAddress) (sdk.Int, error) {
	return k.distributeDeveloperRewards(ctx, totalMintedCoin, developerRewardsProportion, developerRewardsReceivers)
}
