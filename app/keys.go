package app

import (
	wasm "github.com/CosmWasm/wasmd/x/wasm"
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
	claimtypes "github.com/osmosis-labs/osmosis/v7/x/claim/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func KVStoreKeys() []string {
	return []string{
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		ibchost.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey,
		gammtypes.StoreKey,
		lockuptypes.StoreKey,
		claimtypes.StoreKey,
		incentivestypes.StoreKey,
		epochstypes.StoreKey,
		poolincentivestypes.StoreKey,
		authzkeeper.StoreKey,
		txfeestypes.StoreKey,
		superfluidtypes.StoreKey,
		bech32ibctypes.StoreKey,
		wasm.StoreKey,
	}
}
