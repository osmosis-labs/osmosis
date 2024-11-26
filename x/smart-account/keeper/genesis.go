package keeper

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"

	storetypes "cosmossdk.io/store/types"
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

		// Extract account address from key
		accountAddr := strings.Split(string(key), "|")[1]

		// Check if this entry is for a new address or the same as the last one processed
		if len(accountAuthenticators) == 0 ||
			accountAuthenticators[len(accountAuthenticators)-1].Address != accountAddr {
			// If it's a new address, create a new AuthenticatorData entry
			accountAuthenticators = append(accountAuthenticators, types.AuthenticatorData{
				Address:        accountAddr,
				Authenticators: []types.AccountAuthenticator{authenticator},
			})
		} else {
			// If it's the same address, append the authenticator to the last entry in the list
			lastIndex := len(accountAuthenticators) - 1
			accountAuthenticators[lastIndex].Authenticators = append(accountAuthenticators[lastIndex].Authenticators, authenticator)
		}

		return nil
	}

	// Iterate over all entries in the store using a prefix iterator
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyAccountAuthenticatorsPrefixId())
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
func (k Keeper) AddAuthenticatorWithId(ctx sdk.Context, account sdk.AccAddress, authenticatorType string, config []byte, id uint64) error {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorType)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}
	cacheCtx, _ := ctx.CacheContext()
	err := impl.OnAuthenticatorAdded(cacheCtx, account, config, strconv.FormatUint(id, 10))
	if err != nil {
		return err
	}
	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, id),
		&types.AccountAuthenticator{
			Id:     id,
			Type:   authenticatorType,
			Config: config,
		})
	return nil
}
