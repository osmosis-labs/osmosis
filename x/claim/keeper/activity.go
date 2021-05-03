package keeper

import (
	"encoding/json"

	"github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetActivities get activites of users for genesis export
func (k Keeper) GetActivities(ctx sdk.Context) []types.UserActions {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ActionKey)))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	actionsByUser := make(map[string]types.Actions)
	for ; iterator.Valid(); iterator.Next() {
		value := types.UserAction{}
		err := json.Unmarshal(iterator.Value(), &value)
		if err != nil {
			panic(err)
		}
		actions := actionsByUser[value.User]
		actions = append(actions, value.Action)
		actionsByUser[value.User] = actions
	}

	activities := []types.UserActions{}
	for user, actions := range actionsByUser {
		activities = append(activities, types.UserActions{
			User:    user,
			Actions: actions,
		})
	}
	return activities
}

func (k Keeper) SetUserActions(ctx sdk.Context, address sdk.AccAddress, actions []types.Action) {
	for _, action := range actions {
		k.CheckAndSetUserAction(ctx, address, action)
	}
}

func (k Keeper) GetUserActions(ctx sdk.Context, address sdk.AccAddress) []types.Action {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ActionKey), address.Bytes()...))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	actions := []types.Action{}
	for ; iterator.Valid(); iterator.Next() {
		value := types.UserAction{}
		err := json.Unmarshal(iterator.Value(), &value)
		if err != nil {
			panic(err)
		}
		actions = append(actions, value.Action)
	}

	return actions
}

func (k Keeper) CheckAndSetUserAction(ctx sdk.Context, address sdk.AccAddress, action types.Action) bool {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ActionKey))
	key := append(address, sdk.Uint64ToBigEndian(uint64(action))...)
	if prefixStore.Has(key) {
		return false
	}
	value := types.UserAction{
		User:   address.String(),
		Action: action,
	}
	valueBz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	prefixStore.Set(key, valueBz)
	return true
}
