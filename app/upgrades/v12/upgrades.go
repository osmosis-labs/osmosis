package v12

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	"github.com/osmosis-labs/osmosis/v27/app/keepers"
	"github.com/osmosis-labs/osmosis/v27/app/upgrades"
	twaptypes "github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

// We set the app version to pre-upgrade because it will be incremented by one
// after the upgrade is applied by the handler.
const preUpgradeAppVersion = 11

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		// Although the app version was already set during the v9 upgrade, our v10 was a fork and
		// v11 was decided to be limited to the "gauge creation minimum fee" change only:
		// https://github.com/osmosis-labs/osmosis/pull/2202
		//
		// As a result, the upgrade handler was not executed to increment the app version.
		// This change helps to correctly set the app version for v12.
		versionSetter := keepers.UpgradeKeeper.GetVersionSetter()
		versionSetter.SetProtocolVersion(preUpgradeAppVersion)

		// Specifying the whole list instead of adding and removing. Less fragile.
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			AllowMessages: []string{
				// Change: Added MsgTrasnsfer
				sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),

				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				// Change: Added MsgUndelegate
				sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				sdk.MsgTypeURL(&govtypesv1.MsgVote{}),
				// Change: Removed authz messages
				sdk.MsgTypeURL(&gammtypes.MsgJoinPool{}),
				sdk.MsgTypeURL(&gammtypes.MsgExitPool{}),
				sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountIn{}),
				sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountOut{}),
				sdk.MsgTypeURL(&gammtypes.MsgJoinSwapExternAmountIn{}),
				sdk.MsgTypeURL(&gammtypes.MsgJoinSwapShareAmountOut{}),
				sdk.MsgTypeURL(&gammtypes.MsgExitSwapExternAmountOut{}),
				sdk.MsgTypeURL(&gammtypes.MsgExitSwapShareAmountIn{}),
				// Change: Added superfluid unbound
				sdk.MsgTypeURL(&superfluidtypes.MsgSuperfluidUnbondLock{}),
			},
		}
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)

		// Initialize TWAP state
		// N.B.: deprecation nolint
		// nolint: staticcheck
		// Note: left for historic reasons.
		// latestPoolId := keepers.GAMMKeeper.GetNextPoolId(ctx) - 1
		// err := keepers.TwapKeeper.MigrateExistingPools(ctx, latestPoolId)
		// if err != nil {
		// 	return nil, err
		// }

		// Set TWAP parameters to default values.
		keepers.TwapKeeper.SetParams(ctx, twaptypes.DefaultParams())

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
