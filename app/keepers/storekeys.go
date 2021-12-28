package keepers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"
	claimtypes "github.com/osmosis-labs/osmosis/x/claim/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	txfeestypes "github.com/osmosis-labs/osmosis/x/txfees/types"
)

// TODO: Figure out a test to ensure every keeper has its store key set
func AllKVStoreKeys() map[string]*sdk.KVStoreKey {
	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		gammtypes.StoreKey, lockuptypes.StoreKey, claimtypes.StoreKey, incentivestypes.StoreKey,
		epochstypes.StoreKey, poolincentivestypes.StoreKey, authzkeeper.StoreKey, txfeestypes.StoreKey,
		bech32ibctypes.StoreKey,
	)
	return keys
}

func AllTStoreKeys() map[string]*sdk.TransientStoreKey {
	return sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
}

func AllMemStoreKeys() map[string]*sdk.MemoryStoreKey {
	return sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
}
