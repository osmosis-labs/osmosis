package keeper

import (
	"encoding/json"

	"github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetActivities get activites of users for genesis export
func (k Keeper) GetActivities(ctx sdk.Context) []types.UserActivities {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ActionKey)))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	actionsByUser := make(map[string]types.Actions)
	for ; iterator.Valid(); iterator.Next() {
		value := types.UserActivity{}
		err := json.Unmarshal(iterator.Value(), &value)
		if err != nil {
			panic(err)
		}
		actions := actionsByUser[value.User]
		actions = append(actions, value.Action)
		actionsByUser[value.User] = actions
	}

	activities := []types.UserActivities{}
	for user, actions := range actionsByUser {
		address, err := sdk.AccAddressFromBech32(user)
		if err != nil {
			panic(err)
		}
		withdrawn := k.GetWithdrawnActions(ctx, address)
		activities = append(activities, types.UserActivities{
			User:      user,
			Actions:   actions,
			Withdrawn: withdrawn,
		})
	}
	return activities
}

func (k Keeper) SetUserActions(ctx sdk.Context, address sdk.AccAddress, actions []types.Action) {
	for _, action := range actions {
		k.SetUserAction(ctx, address, action)
	}
}

func (k Keeper) GetUserActions(ctx sdk.Context, address sdk.AccAddress) []types.Action {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.ActionKey), address.Bytes()...))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	actions := []types.Action{}
	for ; iterator.Valid(); iterator.Next() {
		value := types.UserActivity{}
		err := json.Unmarshal(iterator.Value(), &value)
		if err != nil {
			panic(err)
		}
		actions = append(actions, value.Action)
	}

	return actions
}

func (k Keeper) SetUserAction(ctx sdk.Context, address sdk.AccAddress, action types.Action) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.ActionKey))
	key := append(address, sdk.Uint64ToBigEndian(uint64(action))...)
	value := types.UserActivity{
		User:   address.String(),
		Action: action,
	}
	valueBz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	prefixStore.Set(key, valueBz)
}

// GetClaimablePercentageByActivity returns completed action percentage from user's activity
func (k Keeper) GetClaimablePercentageByActivity(ctx sdk.Context, address sdk.AccAddress) sdk.Dec {
	numActions := len(k.GetUserActions(ctx, address))
	numWithdrawnActions := len(k.GetWithdrawnActions(ctx, address))
	numTotalActions := len(types.Action_name)
	return sdk.NewDec(int64(numActions - numWithdrawnActions)).QuoInt64(int64(numTotalActions))
}

func (k Keeper) SetUserWithdrawnActions(ctx sdk.Context, address sdk.AccAddress, actions []types.Action) {
	for _, action := range actions {
		k.SetUserWithdrawnAction(ctx, address, action)
	}
}

func (k Keeper) GetWithdrawnActions(ctx sdk.Context, address sdk.AccAddress) []types.Action {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, append([]byte(types.WithdrawnActionKey), address.Bytes()...))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	actions := []types.Action{}
	for ; iterator.Valid(); iterator.Next() {
		value := types.UserActivity{}
		err := json.Unmarshal(iterator.Value(), &value)
		if err != nil {
			panic(err)
		}
		actions = append(actions, value.Action)
	}

	return actions
}

func (k Keeper) SetUserWithdrawnAction(ctx sdk.Context, address sdk.AccAddress, action types.Action) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(types.WithdrawnActionKey))
	key := append(address, sdk.Uint64ToBigEndian(uint64(action))...)
	value := types.UserActivity{
		User:   address.String(),
		Action: action,
	}
	valueBz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	prefixStore.Set(key, valueBz)
}
