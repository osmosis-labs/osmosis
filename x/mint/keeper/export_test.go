package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

type ErrInvalidRatio = invalidRatioError

const (
	DeveloperVestingAmount = developerVestingAmount
)

var (
	ErrAmountCannotBeNilOrZero               = errAmountCannotBeNilOrZero
	ErrDevVestingModuleAccountAlreadyCreated = errDevVestingModuleAccountAlreadyCreated
	ErrDevVestingModuleAccountNotCreated     = errDevVestingModuleAccountNotCreated

	GetProportions = getProportions
)

func (k Keeper) DistributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.Coin, proportion sdk.Dec) (sdk.Coin, error) {
	return k.distributeToModule(ctx, recipientModule, mintedCoin, proportion)
}
