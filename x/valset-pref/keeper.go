package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	"github.com/osmosis-labs/osmosis/v12/x/valset-pref/types"
)

type Keeper struct {
	storeKey      sdk.StoreKey
	paramSpace    paramtypes.Subspace
	stakingKeeper types.StakingInterface
	bankKeeper    types.BankInterface
	distrKeeper   types.DistrInterface
}

func NewKeeper(storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	stakingKeeper types.StakingInterface,
	bankKeeper types.BankInterface,
	distrKeeper types.DistrInterface,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		stakingKeeper: stakingKeeper,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetValidatorSetPreferences(ctx sdk.Context, delegator string, validators types.ValidatorSetPreferences) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, []byte(delegator), &validators)
}

func (k Keeper) GetValidatorSetPreference(ctx sdk.Context, delegator string) (types.ValidatorSetPreferences, bool) {
	validatorSet := types.ValidatorSetPreferences{}

	store := ctx.KVStore(k.storeKey)
	b := store.Get([]byte(delegator))
	if b == nil {
		return types.ValidatorSetPreferences{}, false
	}

	err := proto.Unmarshal(b, &validatorSet)
	if err != nil {
		return types.ValidatorSetPreferences{}, false
	}

	return validatorSet, true
}

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
