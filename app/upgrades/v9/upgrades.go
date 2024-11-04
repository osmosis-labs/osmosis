package v9

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"

	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	"github.com/osmosis-labs/osmosis/v27/app/keepers"
	"github.com/osmosis-labs/osmosis/v27/app/upgrades"
)

const preUpgradeAppVersion = 8

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ExecuteProp214(ctx, keepers.GAMMKeeper)

		// We set the app version to pre-upgrade because it will be incremented by one
		// after the upgrade is applied by the handler.
		versionSetter := keepers.UpgradeKeeper.GetVersionSetter()
		versionSetter.SetProtocolVersion(preUpgradeAppVersion)

		// save oldIcaVersion, so we can skip icahost.InitModule in longer term tests.
		oldIcaVersion := fromVM[icatypes.ModuleName]

		// Add Interchain Accounts host module
		// set the ICS27 consensus version so InitGenesis is not run

		// ConsensusVersion no longer exists, keeping commented for historical purposes
		// fromVM[icatypes.ModuleName] = mm.Modules[icatypes.ModuleName].ConsensusVersion()

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
				sdk.MsgTypeURL(&govtypesv1.MsgVote{}),
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

		// skip InitModule in upgrade tests after the upgrade has gone through.
		if oldIcaVersion != fromVM[icatypes.ModuleName] {
			icamodule.InitModule(ctx, controllerParams, hostParams)
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
