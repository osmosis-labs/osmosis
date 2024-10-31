package keeper

import (
	"fmt"
	"strconv"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmoutils"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

type Keeper struct {
	storeKey                storetypes.StoreKey
	cdc                     codec.BinaryCodec
	paramSpace              paramtypes.Subspace
	CircuitBreakerGovernor  sdk.AccAddress
	isSmartAccountActiveBz  []byte
	isSmartAccountActiveVal bool

	AuthenticatorManager *authenticator.AuthenticatorManager
}

func NewKeeper(
	cdc codec.BinaryCodec,
	StoreKey storetypes.StoreKey,
	govModuleAddr sdk.AccAddress,
	ps paramtypes.Subspace,
	authenticatorManager *authenticator.AuthenticatorManager,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:               StoreKey,
		cdc:                    cdc,
		CircuitBreakerGovernor: govModuleAddr,
		paramSpace:             ps,
		AuthenticatorManager:   authenticatorManager,
	}
}

// unmarshalAccountAuthenticator is used to unmarshal the AccountAuthenticator from the store
func (k Keeper) unmarshalAccountAuthenticator(bz []byte) (*types.AccountAuthenticator, error) {
	var accountAuthenticator types.AccountAuthenticator
	err := k.cdc.Unmarshal(bz, &accountAuthenticator)
	if err != nil {
		return &types.AccountAuthenticator{}, errorsmod.Wrap(err, "failed to unmarshal account authenticator")
	}
	return &accountAuthenticator, nil
}

// GetAuthenticatorDataForAccount gets all authenticators AccAddressFromBech32 with an account
// from the store, the data is prefixed by 2|<accAddr|
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

// GetSelectedAuthenticatorData gets all authenticators from an account
// from the store, the data is  prefixed by 2|<accAddr|<keyId>
func (k Keeper) GetSelectedAuthenticatorData(
	ctx sdk.Context,
	account sdk.AccAddress,
	selectedAuthenticator int,
) (*types.AccountAuthenticator, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.KeyAccountId(account, uint64(selectedAuthenticator)))
	if bz == nil {
		return &types.AccountAuthenticator{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("authenticator %d not found for account %s", selectedAuthenticator, account))
	}
	authenticatorFromStore, err := k.unmarshalAccountAuthenticator(bz)
	if err != nil {
		return &types.AccountAuthenticator{}, err
	}

	return authenticatorFromStore, nil
}

// GetInitializedAuthenticatorForAccount returns a single initialized authenticator for the account.
// It fetches the authenticator data from the store, gets the authenticator struct from the manager, then calls initialize on the authenticator data
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

	uninitializedAuthenticator := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorFromStore.Type)
	if uninitializedAuthenticator == nil {
		// This should never happen, but if it does, it means that stored authenticator is not registered
		// or somehow the registered authenticator was removed / malformed
		telemetry.IncrCounter(1, types.CounterKeyMissingRegisteredAuthenticator)
		k.Logger(ctx).Error("account asscoicated authenticator not registered in manager", "type", authenticatorFromStore.Type, "id", selectedAuthenticator)

		return authenticator.InitializedAuthenticator{},
			errorsmod.Wrapf(
				sdkerrors.ErrLogic,
				"authenticator id %d failed to initialize, authenticator type %s not registered in manager",
				selectedAuthenticator, authenticatorFromStore.Type,
			)
	}
	// Ensure that initialization of each authenticator works as expected
	// NOTE: Always return a concrete authenticator not a pointer, do not modify in place
	// NOTE: The authenticator manager returns a struct that is reused
	initializedAuthenticator, err := uninitializedAuthenticator.Initialize(authenticatorFromStore.Config)
	if err != nil || initializedAuthenticator == nil {
		return authenticator.InitializedAuthenticator{},
			errorsmod.Wrapf(err,
				"authenticator %d with type %s failed to initialize",
				selectedAuthenticator, authenticatorFromStore.Type,
			)
	}

	finalAuthenticator := authenticator.InitializedAuthenticator{
		Id:            authenticatorFromStore.Id,
		Authenticator: initializedAuthenticator,
	}

	return finalAuthenticator, nil
}

const FirstAuthenticatorId = 1

// InitializeOrGetNextAuthenticatorId returns the next authenticator id.
// If it is not set, it initializes it to 1.
func (k Keeper) InitializeOrGetNextAuthenticatorId(ctx sdk.Context) uint64 {
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
// authenticators such as SignatureVerifications and AllOfs
func (k Keeper) AddAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorType string, config []byte) (uint64, error) {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorType)
	if impl == nil {
		return 0, fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}

	// Get the next global id value for authenticators from the store
	id := k.InitializeOrGetNextAuthenticatorId(ctx)

	// Each authenticator has a custom OnAuthenticatorAdded function
	err := impl.OnAuthenticatorAdded(ctx, account, config, strconv.FormatUint(id, 10))
	if err != nil {
		return 0, errorsmod.Wrapf(err, "`OnAuthenticatorAdded` failed on authenticator type %s", authenticatorType)
	}

	k.SetNextAuthenticatorId(ctx, id+1)

	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, id),
		&types.AccountAuthenticator{
			Id:     id,
			Type:   authenticatorType,
			Config: config,
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
		return errorsmod.Wrap(err, "failed to get authenticator")
	}
	if !found {
		return fmt.Errorf("authenticator with id %d does not exist for account %s", authenticatorId, account)
	}
	impl := k.AuthenticatorManager.GetAuthenticatorByType(existing.Type)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", existing.Type)
	}

	// Authenticators can prevent removal. This should be used sparingly
	err = impl.OnAuthenticatorRemoved(ctx, account, existing.Config, strconv.FormatUint(authenticatorId, 10))
	if err != nil {
		return errorsmod.Wrapf(err, "`OnAuthenticatorRemoved` failed on authenticator type %s", existing.Type)
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

func (k Keeper) SetActiveState(ctx sdk.Context, active bool) {
	params := k.GetParams(ctx)
	params.IsSmartAccountActive = active
	k.SetParams(ctx, params)
}
