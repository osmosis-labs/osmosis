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
}

func NewKeeper(storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	stakingKeeper types.StakingInterface,
) Keeper {
	return Keeper{
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		stakingKeeper: stakingKeeper,
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
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(delegator))
	if bz == nil {
		return types.ValidatorSetPreferences{}, false
	}
	var valsetPref types.ValidatorSetPreferences
	if err := proto.Unmarshal(bz, &valsetPref); err != nil {
		return types.ValidatorSetPreferences{}, false
	}

	return valsetPref, true
}
