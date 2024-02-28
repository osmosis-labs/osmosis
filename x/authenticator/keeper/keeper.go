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
	UsedAuthenticators   *authenticator.UsedAuthenticators
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
		UsedAuthenticators:   authenticator.NewUsedAuthenticators(),
	}
}

// GetAuthenticatorDataForAccount gets all authenticators AccAddressFromBech32 with an account
// from the store, the data is  prefixed by 2|<accAddr|
func (k Keeper) GetAuthenticatorDataForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]*types.AccountAuthenticator, error) {
	// unmarshalFn is used to unmarshal the AccountAuthenticator from the store
	unmarshalFn := func(bz []byte) (*types.AccountAuthenticator, error) {
		var accountAuthenticator types.AccountAuthenticator
		err := k.cdc.Unmarshal(bz, &accountAuthenticator)
		if err != nil {
			return &types.AccountAuthenticator{}, err
		}
		return &accountAuthenticator, nil
	}

	accountAuthenticators, err := osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey),
		types.KeyAccount(account),
		unmarshalFn,
	)
	if err != nil {
		return nil, err
	}

	return accountAuthenticators, nil
}

// GetAuthenticatorsForAccount returns all the authenticators for the account
// this function relies in GetAuthenticationDataForAccount, this function calls
// Initialise on each authenticator
func (k Keeper) GetAuthenticatorsForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]authenticator.InitializedAuthenticator, error) {
	// Get the authenticator data from the store
	allAuthenticatorData, err := k.GetAuthenticatorDataForAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	authenticators := make([]authenticator.InitializedAuthenticator, len(allAuthenticatorData))

	// Iterate over all authenticator data from the store
	for i, authenticatorData := range allAuthenticatorData {
		// Ensure that the authenticators added to an account have been registered in the manager
		uninitializedAuthenticator := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorData.Type)
		if uninitializedAuthenticator == nil {
			return nil, fmt.Errorf("authenticator %d failed to initialize, authenticator not registered manager", i)
		}

		// Ensure that initialization of each authenticator works as expected
		initializedAuthenticator, err := uninitializedAuthenticator.Initialize(authenticatorData.Data)
		if err != nil || initializedAuthenticator == nil {
			return nil, fmt.Errorf("authenticator %d failed to initialize", authenticatorData.Id)
		}

		authenticators[i] = authenticator.InitializedAuthenticator{
			Id:            authenticatorData.Id,
			Authenticator: initializedAuthenticator,
		}
	}
	return authenticators, nil
}

// GetAuthenticatorsForAccountOrDefault returns the authenticators for the account if there allRecords
// authenticators in the store, or the default if there is no authenticator associated with an account,
// this would be the case if there is an account with authenticators
// This function relies in GetAuthenticationsForAccount
func (k Keeper) GetAuthenticatorsForAccountOrDefault(ctx sdk.Context, account sdk.AccAddress) ([]authenticator.InitializedAuthenticator, error) {
	authenticators, err := k.GetAuthenticatorsForAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	if len(authenticators) == 0 {
		return []authenticator.InitializedAuthenticator{{
			Id:            0,
			Authenticator: k.AuthenticatorManager.GetDefaultAuthenticator(),
		}}, nil
	}

	return authenticators, nil
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

// GetNextAuthenticatorIdAndIncrement returns the next authenticator id and increments it.
func (k Keeper) GetNextAuthenticatorIdAndIncrement(ctx sdk.Context) uint64 {
	nextAuthenticatorId := k.GetNextAuthenticatorId(ctx)
	k.SetNextAuthenticatorId(ctx, nextAuthenticatorId+1)
	return nextAuthenticatorId
}

// AddAuthenticator adds an authenticator to an account, this function is used to add multiple
// authenticators such as SignatureVerificationAuthenticators and AllOfAuthenticators
func (k Keeper) AddAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorType string, data []byte) error {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorType)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}

	nextId := k.GetNextAuthenticatorId(ctx)
	stringId := strconv.FormatInt(int64(nextId), 10)
	err := impl.OnAuthenticatorAdded(ctx, account, data, stringId)

	if err != nil {
		return err
	}

	k.SetNextAuthenticatorId(ctx, nextId+1)

	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, nextId),
		&types.AccountAuthenticator{
			Id:   nextId,
			Type: authenticatorType,
			Data: data,
		})
	return nil
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
