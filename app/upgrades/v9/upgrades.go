package v9

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	"github.com/osmosis-labs/osmosis/v7/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ExecuteProp214(ctx, keepers.GAMMKeeper)

		// Add Interchain Accounts host module
		// set the ICS27 consensus version so InitGenesis is not run
		fromVM[icatypes.ModuleName] = icamodule.ConsensusVersion()

		// create ICS27 Controller submodule params, controller module not enabled.
		controllerParams := icacontrollertypes.Params{}

		// create ICS27 Host submodule params
		hostParams := icahosttypes.Params{
			HostEnabled:   true,
			AllowMessages: []string{"/cosmos.bank.v1beta1.MsgSend"},
		}

		// initialize ICS27 module
		icamodule.InitModule(ctx, controllerParams, hostParams)
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
