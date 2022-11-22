package simcli

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/appconfig"

	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	_ "github.com/osmosis-labs/osmosis/v13/client/docs/statik"
	epochstypes "github.com/osmosis-labs/osmosis/v13/x/epochs/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v13/x/lockup/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v13/x/superfluid/types"
	twaptypes "github.com/osmosis-labs/osmosis/v13/x/twap/types"

	lockupapi "github.com/osmosis-labs/osmosis/v13/api/osmosis/lockup/module/v1"
)

var (
	// application configuration (used by depinject)
	AppConfig = appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: "runtime",
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName: "simulation",

					BeginBlockers: []string{
						upgradetypes.ModuleName,
						epochstypes.ModuleName,
						capabilitytypes.ModuleName,
						distrtypes.ModuleName,
						slashingtypes.ModuleName,
						evidencetypes.ModuleName,
						stakingtypes.ModuleName,
						superfluidtypes.ModuleName,
						ibchost.ModuleName,
					},

					EndBlockers: []string{
						govtypes.ModuleName,
						stakingtypes.ModuleName,
						twaptypes.ModuleName,
					},

					InitGenesis: []string{
						// capabilitytypes.ModuleName,
						// authtypes.ModuleName,
						// banktypes.ModuleName,
						// distrtypes.ModuleName,
						// stakingtypes.ModuleName,
						// slashingtypes.ModuleName,
						// govtypes.ModuleName,
						// minttypes.ModuleName,
						// crisistypes.ModuleName,
						// ibchost.ModuleName,
						// icatypes.ModuleName,
						// gammtypes.ModuleName,
						// twaptypes.ModuleName,
						// txfeestypes.ModuleName,
						// genutiltypes.ModuleName,
						// evidencetypes.ModuleName,
						// paramstypes.ModuleName,
						// upgradetypes.ModuleName,
						// vestingtypes.ModuleName,
						// ibctransfertypes.ModuleName,
						// poolincentivestypes.ModuleName,
						// superfluidtypes.ModuleName,
						// tokenfactorytypes.ModuleName,
						// valsetpreftypes.ModuleName,
						// incentivestypes.ModuleName,
						// epochstypes.ModuleName,
						// lockuptypes.ModuleName,
						// authz.ModuleName,
						// // wasm after ibc transfer
						// wasm.ModuleName,
						// // ibc_hooks after auth keeper
						// ibc_hooks.ModuleName,
					},
					// OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{
					// 	{
					// 		ModuleName: authtypes.ModuleName,
					// 		KvStoreKey: "acc",
					// 	},
					// },
				}),
			},
			{
				Name:   lockuptypes.ModuleName,
				Config: appconfig.WrapAny(&lockupapi.Module{}),
			},
		},
	})
)
