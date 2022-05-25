package v9

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v9/app/keepers"
)

const preUpgradeAppVersion = 8

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ExecuteProp214(ctx, keepers.GAMMKeeper)
<<<<<<< HEAD
		return mm.RunMigrations(ctx, configurator, vm)
=======
    
         // We set the app version to pre-upgrade because it will be incremented by one
		// after the upgrade is applied by the handler.
		if err := keepers.UpgradeKeeper.SetAppVersion(ctx, preUpgradeAppVersion); err != nil {
			return nil, err
		}

		// Add Interchain Accounts host module
		// set the ICS27 consensus version so InitGenesis is not run
		fromVM[icatypes.ModuleName] = mm.Modules[icatypes.ModuleName].ConsensusVersion()

		// create ICS27 Controller submodule params, controller module not enabled.
		controllerParams := icacontrollertypes.Params{}

		// create ICS27 Host submodule params
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			AllowMessages: []string{
				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				sdk.MsgTypeURL(&govtypes.MsgVote{}),
				sdk.MsgTypeURL(&authz.MsgExec{}),
				sdk.MsgTypeURL(&authz.MsgGrant{}),
				sdk.MsgTypeURL(&authz.MsgRevoke{}),
				sdk.MsgTypeURL(&gammtypes.MsgJoinPool{}),
				sdk.MsgTypeURL(&gammtypes.MsgExitPool{}),
				sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountIn{}),
				sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountOut{}),
				sdk.MsgTypeURL(&gammtypes.MsgJoinSwapExternAmountIn{}),
				sdk.MsgTypeURL(&gammtypes.MsgJoinSwapShareAmountOut{}),
				sdk.MsgTypeURL(&gammtypes.MsgExitSwapExternAmountOut{}),
				sdk.MsgTypeURL(&gammtypes.MsgExitSwapShareAmountIn{}),
			},
		}

		// initialize ICS27 module
		icamodule, correctTypecast := mm.Modules[icatypes.ModuleName].(ica.AppModule)
		if !correctTypecast {
			panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
		}
		icamodule.InitModule(ctx, controllerParams, hostParams)
		return mm.RunMigrations(ctx, configurator, fromVM)
>>>>>>> 1510160 (chore: upgrade sdk with app version fix for state-sync (#1570))
	}
}
