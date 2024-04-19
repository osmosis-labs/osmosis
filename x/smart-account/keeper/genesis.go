package keeper

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v24/x/smart-account/types"
)

// GetAllAuthenticatorData is used in genesis export to export all the authenticator for all accounts
func (k Keeper) GetAllAuthenticatorData(ctx sdk.Context) ([]types.AuthenticatorData, error) {
	var accountAuthenticators []types.AuthenticatorData

	parse := func(key []byte, value []byte) error {
		var authenticator types.AccountAuthenticator
		err := k.cdc.Unmarshal(value, &authenticator)
		if err != nil {
			return err
		}

		// The authenticator store key looks like "2|osmo1<address>|<authenticator_id>" we need the address to
		// successfully import and export the authenticator module
		accountAddr := strings.Split(string(key), "|")[1]

		// Find existing AuthenticatorData for the account address
		var found bool
		for i := range accountAuthenticators {
			if accountAuthenticators[i].Address == accountAddr {
				accountAuthenticators[i].Authenticators = append(accountAuthenticators[i].Authenticators, authenticator)
				found = true
				break
			}
		}

		// If AuthenticatorData doesn't exist, create a new one
		if !found {
			accountAuthenticators = append(accountAuthenticators, types.AuthenticatorData{
				Address:        accountAddr,
				Authenticators: []types.AccountAuthenticator{authenticator},
			})
		}

		return nil
	}

	iterator := sdk.KVStorePrefixIterator(
		ctx.KVStore(k.storeKey),
		types.KeyAccountAuthenticatorsPrefixId(),
	)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		err := parse(iterator.Key(), iterator.Value())
		if err != nil {
			return nil, err
		}
	}

	return accountAuthenticators, nil
}

// AddAuthenticatorWithId adds an authenticator to an account, this function is used in genesis import
func (k Keeper) AddAuthenticatorWithId(
	ctx sdk.Context,
	account sdk.AccAddress,
	authenticatorType string,
	data []byte,
	id uint64,
) error {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorType)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}
	cacheCtx, _ := ctx.CacheContext()
	err := impl.OnAuthenticatorAdded(cacheCtx, account, data, strconv.FormatUint(id, 10))
	if err != nil {
		return err
	}
	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, id),
		&types.AccountAuthenticator{
			Id:   id,
			Type: authenticatorType,
			Data: data,
		})
	return nil
}
