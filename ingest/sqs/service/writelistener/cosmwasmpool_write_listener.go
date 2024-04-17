package writelistener

import (
	"bytes"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v24/ingest/sqs/domain"
	cosmwasmpoolmodel "github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/model"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v24/x/poolmanager/types"
)

// const (
// 	transmuterAddress = "osmo10c8y69yylnlwrhu32ralf08ekladhfknfqrjsy9yqc9ml8mlxpqq2sttzk"
// )

var _ storetypes.WriteListener = (*cosmwasmPoolWriteListener)(nil)

type cosmwasmPoolWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
}

func NewCosmwasmPool(poolTracker domain.BlockPoolUpdateTracker) storetypes.WriteListener {
	return &cosmwasmPoolWriteListener{
		poolTracker: poolTracker,
	}
}

// OnWrite implements types.WriteListener
//
// NOTE: This only detects cwpools that have been created or migrated. It does not detect changes in balances (i.e. swaps / position creation / withdraws)
func (s *cosmwasmPoolWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Track the cw pool that was just created/migrated
	if len(key) > 0 && bytes.Equal(cosmwasmpooltypes.PoolsKey, key[:1]) {
		var pool cosmwasmpoolmodel.CosmWasmPool
		if err := pool.Unmarshal(value); err != nil {
			return err
		}

		s.poolTracker.TrackCosmWasm(&pool)

		// Add the pool in the address to pool map
		// This is used to track balance changes for cwpools
		var poolI poolmanagertypes.PoolI = &pool
		s.poolTracker.TrackCosmWasmPoolsAddressToPoolMap(poolI)
	}
	return nil
}

type cosmwasmPoolBalanceWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
}

func NewCosmwasmPoolBalance(poolTracker domain.BlockPoolUpdateTracker) storetypes.WriteListener {
	return &cosmwasmPoolBalanceWriteListener{
		poolTracker: poolTracker,
	}
}

// OnWrite implements types.WriteListener
// Tracks balance changes for cwpools (i.e. swaps / position creation / withdraws)
func (s *cosmwasmPoolBalanceWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Check if the key is a balance change for any address
	if len(key) > 0 && key[0] == banktypes.BalancesPrefix[0] {
		// The key is a balance change. Check if the address in question is a cwpool address
		fmt.Println("key[0]", key[0])
		fmt.Println("banktypes.BalancesPrefix[0]", banktypes.BalancesPrefix[0])

		// We expect the key to be of the form:
		// <prefix> (length 1)
		// <address_length> (length 1)
		// <address> (length address_length)
		addressLength := key[1]
		addressBytes := key[1+1 : 1+addressLength+1]
		fmt.Println("addressBytes", addressBytes)
		address, err := sdk.AccAddressFromBech32(string(addressBytes))
		if err != nil {
			// The address is not a valid bech32 address. Ignore it
			return nil
		}
		addressStr := string(address)
		fmt.Println("addressStr", addressStr)

		cwPoolMap := s.poolTracker.GetCosmWasmPoolsAddressToIDMap()
		if pool, ok := cwPoolMap[addressStr]; ok {
			// The address is a cwpool address. Add the pool to the pool tracker
			fmt.Println("track")
			s.poolTracker.TrackCosmWasm(pool)
		}
	}
	return nil
}
