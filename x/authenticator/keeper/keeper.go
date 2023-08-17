package keeper

import (
	"context"
	"cosmossdk.io/collections"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v17/x/authenticator/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

var AuthenticatorsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema         collections.Schema
	Authenticators collections.Map[string, []byte]

	cdc    codec.BinaryCodec
	params paramtypes.Subspace
}

func NewKeeper(cdc codec.BinaryCodec, storeKey *storetypes.KVStoreKey, ps paramtypes.Subspace) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	sb := collections.NewSchemaBuilder(NewKVStoreService(storeKey))

	return Keeper{
		cdc:            cdc,
		params:         ps,
		Authenticators: collections.NewMap(sb, AuthenticatorsPrefix, "authenticator_list", collections.StringKey, collections.BytesValue),
	}
}

func (k Keeper) GetAuthenticatorsForAccount(ctx context.Context) ([]types.Authenticator[any], error) {
	// passing a nil Ranger equals to: iterate over every possible key
	iter, err := k.Authenticators.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	values, err := iter.Values()
	if err != nil {
		return nil, err
	}

	// Unmarshall
	var accounts []types.Authenticator[any]
	for _, account := range values {
		var acc types.Authenticator[any]
		// ToDo: register interface and implementations in the codec
		err := k.cdc.UnmarshalInterface(account, &acc)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, err
}
