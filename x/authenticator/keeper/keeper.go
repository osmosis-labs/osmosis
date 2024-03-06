package keeper

import (
	"fmt"
	"strconv"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/osmoutils"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	AuthenticatorManager *authenticator.AuthenticatorManager
}

func NewKeeper(
	cdc codec.BinaryCodec,
	managerStoreKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	authenticatorManager *authenticator.AuthenticatorManager,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:             managerStoreKey,
		cdc:                  cdc,
		paramSpace:           ps,
		AuthenticatorManager: authenticatorManager,
	}
}

// unmarshalAccountAuthenticator is used to unmarshal the AccountAuthenticator from the store
func (k Keeper) unmarshalAccountAuthenticator(bz []byte) (*types.AccountAuthenticator, error) {
	var accountAuthenticator types.AccountAuthenticator
	err := k.cdc.Unmarshal(bz, &accountAuthenticator)
	if err != nil {
		return &types.AccountAuthenticator{}, err
	}
	return &accountAuthenticator, nil
}

// GetAuthenticatorDataForAccount gets all authenticators AccAddressFromBech32 with an account
// from the store, the data is  prefixed by 2|<accAddr|
func (k Keeper) GetAuthenticatorDataForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]*types.AccountAuthenticator, error) {
	accountAuthenticators, err := osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey),
		types.KeyAccount(account),
		k.unmarshalAccountAuthenticator,
	)
	if err != nil {
		return nil, err
	}

	return accountAuthenticators, nil
}

// GetSelectedAuthenticatorDataForAccount gets all authenticators from an account
// from the store, the data is  prefixed by 2|<accAddr|<keyId>
func (k Keeper) GetSelectedAuthenticatorData(
	ctx sdk.Context,
	account sdk.AccAddress,
	selectedAuthenticator int,
) (*types.AccountAuthenticator, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.KeyAccountId(account, uint64(selectedAuthenticator)))
	authenticatorFromStore, err := k.unmarshalAccountAuthenticator(bz)
	if err != nil {
		return &types.AccountAuthenticator{}, err
	}

	return authenticatorFromStore, nil
}

// GetSelectedAuthenticatorForAccountFromStore returns a single authenticator for the account
// this function relies in GetAuthenticationDataForAccount, this function calls
// Initialise on the specific authenticator
func (k Keeper) GetInitializedAuthenticatorForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
	selectedAuthenticator int,
) (authenticator.InitializedAuthenticator, error) {
	// Get the authenticator data from the store
	authenticatorFromStore, err := k.GetSelectedAuthenticatorData(ctx, account, selectedAuthenticator)
	if err != nil {
		return authenticator.InitializedAuthenticator{}, err
	}

	// Return the default authenticator here if there is nothing in the store
	if authenticatorFromStore.Type == "" {
		return authenticator.InitializedAuthenticator{
			Id:            0,
			Authenticator: k.AuthenticatorManager.GetDefaultAuthenticator(),
		}, nil
	}

	uninitializedAuthenticator := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorFromStore.Type)
	if uninitializedAuthenticator == nil {
		return authenticator.InitializedAuthenticator{},
			fmt.Errorf(
				"authenticator %d failed to initialize, authenticator not registered in manager",
				selectedAuthenticator,
			)
	}
	// Ensure that initialization of each authenticator works as expected
	// NOTE: Always return a concrete authenticator not a pointer, do not modify in place
	// NOTE: The authenticator manager returns a struct that is reused
	initializedAuthenticator, err := uninitializedAuthenticator.Initialize(authenticatorFromStore.Data)
	if err != nil || initializedAuthenticator == nil {
		return authenticator.InitializedAuthenticator{},
			fmt.Errorf(
				"authenticator %d failed to initialize",
				selectedAuthenticator,
			)
	}

	finalAuthenticator := authenticator.InitializedAuthenticator{
		Id:            authenticatorFromStore.Id,
		Authenticator: initializedAuthenticator,
	}

	return finalAuthenticator, nil
}

const FirstAuthenticatorId = 1

// GetNextAuthenticatorId returns the next authenticator id
func (k Keeper) GetNextAuthenticatorId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextAuthenticatorId := gogotypes.UInt64Value{}
	found, err := osmoutils.Get(store, types.KeyNextAccountAuthenticatorId(), &nextAuthenticatorId)
	if err != nil {
		panic(err)
	}
	if !found {
		k.SetNextAuthenticatorId(ctx, FirstAuthenticatorId)
		return FirstAuthenticatorId
	}
	return nextAuthenticatorId.Value
}

// SetNextAuthenticatorId sets next authenticator id.
func (k Keeper) SetNextAuthenticatorId(ctx sdk.Context, authenticatorId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextAccountAuthenticatorId(), &gogotypes.UInt64Value{Value: authenticatorId})
}

// AddAuthenticator adds an authenticator to an account, this function is used to add multiple
// authenticators such as SignatureVerificationAuthenticators and AllOfAuthenticators
func (k Keeper) AddAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorType string, data []byte) (uint64, error) {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorType)
	if impl == nil {
		return 0, fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}

	// Get the next global id value for authenticators from the store
	id := k.GetNextAuthenticatorId(ctx)
	stringId := strconv.FormatUint(id, 10)

	// Each authenticator has a custom OnAuthenticatorAdded function
	err := impl.OnAuthenticatorAdded(ctx, account, data, stringId)
	if err != nil {
		return 0, err
	}

	k.SetNextAuthenticatorId(ctx, id+1)

	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, id),
		&types.AccountAuthenticator{
			Id:   id,
			Type: authenticatorType,
			Data: data,
		})
	return id, nil
}

// RemoveAuthenticator removes an authenticator from an account
func (k Keeper) RemoveAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorId uint64) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyAccountId(account, authenticatorId)

	var existing types.AccountAuthenticator
	found, err := osmoutils.Get(store, key, &existing)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("authenticator with id %d does not exist for account %s", authenticatorId, account)
	}
	impl := k.AuthenticatorManager.GetAuthenticatorByType(existing.Type)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", existing.Type)
	}

	stringId := strconv.FormatInt(int64(authenticatorId), 10)

	// Authenticators can prevent removal. This should be used sparingly
	err = impl.OnAuthenticatorRemoved(ctx, account, existing.Data, stringId)
	if err != nil {
		return err
	}

	store.Delete(key)
	return nil
}

// GetAuthenticatorExtension unpacks the extension for the transaction, this is used with transactions specify
// an authenticator to use
func (k Keeper) GetAuthenticatorExtension(exts []*codectypes.Any) types.AuthenticatorTxOptions {
	var authExtension types.AuthenticatorTxOptions
	for _, ext := range exts {
		err := k.cdc.UnpackAny(ext, &authExtension)
		if err == nil {
			return authExtension
		}
	}
	return nil
}
