package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/osmosis-labs/osmosis/v13/simulation/simtypes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	valsetkeeper "github.com/osmosis-labs/osmosis/v13/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

func RandomMsgSetValSetPreference(k valsetkeeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgSetValidatorSetPreference, error) {
	var sk types.StakingInterface
	// Generate random digit from 0-1 with
	randWeight := RandomWeight()

	// Generate random validators
	randValidator := RandomValidator(ctx, sk)
	if randValidator == nil {
		return nil, fmt.Errorf("No validator")
	}

	return &types.MsgSetValidatorSetPreference{
		Preferences: []types.ValidatorPreference{
			{
				Weight:         randWeight,
				ValOperAddress: randValidator.OperatorAddress,
			},
		},
	}, nil
}

func RandomMsgDelegateToValSet(k valsetkeeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgDelegateToValidatorSet, error) {
	return nil, nil
}

func RandomMsgUnDelegateToValSet(k valsetkeeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgUndelegateFromValidatorSet, error) {
	return nil, nil
}

func RandomMsgWithdrawDelReward(k valsetkeeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgWithdrawDelegationRewards, error) {
	return nil, nil
}

// TODO: move this in simulator folder account.go
func RandomValidator(ctx sdk.Context, sk types.StakingInterface) *stakingtypes.Validator {
	var r *rand.Rand
	validators := sk.GetAllValidators(ctx)
	if len(validators) == 0 {
		return nil
	}
	return &validators[r.Intn(len(validators))]
}

// Random float point from 0-1
func RandomWeight() sdk.Dec {
	rand.Seed(time.Now().UnixNano())
	valWeightStr := fmt.Sprintf("%.2f", rand.Float64())

	return sdk.MustNewDecFromStr(valWeightStr)
}
