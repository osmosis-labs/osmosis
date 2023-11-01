package keeper

import (
	"encoding/json"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v20/x/authenticator/iface"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/osmoutils"

	"github.com/osmosis-labs/osmosis/v20/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v20/x/authenticator/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	AuthenticatorManager *authenticator.AuthenticatorManager
	TransientStore       *authenticator.TransientStore
}

func NewKeeper(
	cdc codec.BinaryCodec,
	managerStoreKey sdk.StoreKey,
	authenticatorStoreKey sdk.StoreKey,
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
		TransientStore:       authenticator.NewTransientStore(authenticatorStoreKey, sdk.Context{}),
	}
}

func (k Keeper) GetAuthenticatorDataForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]*types.AccountAuthenticator, error) {
	accountAuthenticators, err := osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey),
		types.KeyAccount(account),
		func(bz []byte) (*types.AccountAuthenticator, error) {
			// unmarshall the authenticator
			var authenticator types.AccountAuthenticator
			err := k.cdc.Unmarshal(bz, &authenticator)
			if err != nil {
				return &types.AccountAuthenticator{}, err
			}

			return &authenticator, nil
		})
	if err != nil {
		return nil, err
	}

	return accountAuthenticators, nil
}

func (k Keeper) GetAuthenticatorsForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]iface.Authenticator, error) {
	authenticatorData, err := k.GetAuthenticatorDataForAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	authenticators := make([]iface.Authenticator, len(authenticatorData))
	for i, accountAuthenticator := range authenticatorData {
		authenticators[i] = accountAuthenticator.AsAuthenticator(k.AuthenticatorManager)
		if authenticators[i] == nil {
			return nil, fmt.Errorf("authenticator %d failed to initialize", accountAuthenticator.Id)
		}
	}
	return authenticators, nil
}

func (k Keeper) GetAuthenticatorsForAccountOrDefault(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]iface.Authenticator, error) {
	authenticators, err := k.GetAuthenticatorsForAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	if len(authenticators) == 0 {
		authenticators = append(authenticators, k.AuthenticatorManager.GetDefaultAuthenticator())
	}

	return authenticators, nil
}

// GetNextAuthenticatorId returns the next authenticator id.
func (k Keeper) GetNextAuthenticatorId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextAuthenticatorId := gogotypes.UInt64Value{}
	found, err := osmoutils.Get(store, types.KeyNextAccountAuthenticatorId(), &nextAuthenticatorId)
	if err != nil {
		panic(err)
	}
	if !found {
		k.SetNextAuthenticatorId(ctx, 0)
		return 0
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

// AddAuthenticator adds an authenticator to an account
func (k Keeper) AddAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorType string, data []byte) error {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorType)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}
	// Note to reviewer: Here we were passing cacheCtx to ensure that OnAuthenticatorAdded couldn't modify state.
	// I'm removing it so that I can modify state inside spend limit.
	// We were not doing the same thing on RemoveAuthenticator, and these should behave in the same way,
	// I'm removing it here, which I think is ok: the authenticators are responsible of implementing that function properly
	// and if the tx fails the changes will be reverted
	//cacheCtx, _ := ctx.CacheContext()
	err := impl.OnAuthenticatorAdded(ctx, account, data)
	if err != nil {
		return err
	}
	nextId := k.GetNextAuthenticatorIdAndIncrement(ctx)
	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, nextId), // ToDo: will this lead to any concurrency issues?
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

	// Authenticators can prevent removal. This should be used sparingly
	err = impl.OnAuthenticatorRemoved(ctx, account, existing.Data)
	if err != nil {
		return err
	}

	store.Delete(key)
	return nil
}

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

func (k Keeper) GetSpendLimitsForAccountImpl(ctx sdk.Context, account sdk.AccAddress) ([]*types.SpendLimit, error) {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticator.SpendLimitAuthenticator{}.Type())
	spendLimit, ok := impl.(authenticator.SpendLimitAuthenticator)
	if !ok {
		return nil, fmt.Errorf("SpendLimitAuthenticator not properly registered")
	}
	datas, err := spendLimit.GetRegisteredDatas(ctx, account)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get registered datas")
	}

	var limits []*types.SpendLimit
	for _, data := range datas {
		// TODO: maybe not unmarshall in the internal func?
		dataBz, err := json.Marshal(data)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to marshal spend limit data")
		}
		initialized, err := spendLimit.Initialize(dataBz)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to initialize spend limit")
		}
		userLimit, ok := initialized.(authenticator.SpendLimitAuthenticator)
		if !ok {
			return nil, fmt.Errorf("Couldn't cast to SpendLimitAuthenticator")
		}

		limits = append(limits, &types.SpendLimit{
			Account:    account.String(),
			Limit:      data.AllowedDelta,
			QuoteDenom: userLimit.GetQuoteDenom(),
			Period:     string(data.PeriodType),
			// TODO: This can be negative
			Used: userLimit.GetSpentInPeriod(ctx, account, ctx.BlockTime()).Uint64(),
			Id:   data.Id,
		})
	}
	return limits, nil
}
