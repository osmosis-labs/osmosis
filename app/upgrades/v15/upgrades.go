package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icqkeeper "github.com/strangelove-ventures/async-icq/v4/keeper"
	icqtypes "github.com/strangelove-ventures/async-icq/v4/types"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v4/router/types"

	"github.com/osmosis-labs/osmosis/v14/wasmbinding"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/v14/app/keepers"
	appParams "github.com/osmosis-labs/osmosis/v14/app/params"
	"github.com/osmosis-labs/osmosis/v14/app/upgrades"
	gammkeeper "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v14/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		poolmanagerParams := poolmanagertypes.NewParams(keepers.GAMMKeeper.GetParams(ctx).PoolCreationFee)

		keepers.PoolManagerKeeper.SetParams(ctx, poolmanagerParams)
		keepers.PacketForwardKeeper.SetParams(ctx, packetforwardtypes.DefaultParams())
		setICQParams(ctx, keepers.ICQKeeper)

		// N.B: pool id in gamm is to be deprecated in the future
		// Instead,it is moved to poolmanager.
		migrateNextPoolId(ctx, keepers.GAMMKeeper, keepers.PoolManagerKeeper)

		//  N.B.: this is done to avoid initializing genesis for poolmanager module.
		// Otherwise, it would overwrite migrations with InitGenesis().
		// See RunMigrations() for details.
		fromVM[poolmanagertypes.ModuleName] = 0

		// Metadata for uosmo and uion were missing prior to this upgrade.
		// They are added in this upgrade.
		registerOsmoIonMetadata(ctx, keepers.BankKeeper)

		// Stride stXXX/XXX pools are being migrated from the standard balancer curve to the
		// solidly stable curve.
		migrateBalancerPoolsToSolidlyStable(ctx, keepers.GAMMKeeper, keepers.PoolManagerKeeper)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func setICQParams(ctx sdk.Context, icqKeeper *icqkeeper.Keeper) {
	icqparams := icqtypes.DefaultParams()
	icqparams.AllowQueries = wasmbinding.GetStargateWhitelistedPaths()
	// Adding SmartContractState query to allowlist
	icqparams.AllowQueries = append(icqparams.AllowQueries, "/cosmwasm.wasm.v1.Query/SmartContractState")
	icqKeeper.SetParams(ctx, icqparams)
}

func migrateBalancerPoolsToSolidlyStable(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanager.Keeper) {
	// migrate stOSMO_OSMOPoolId, stJUNO_JUNOPoolId, stSTARS_STARSPoolId
	pools := []uint64{stOSMO_OSMOPoolId, stJUNO_JUNOPoolId, stSTARS_STARSPoolId}
	for _, poolId := range pools {
		migrateBalancerPoolToSolidlyStable(ctx, gammKeeper, poolmanagerKeeper, poolId)
	}
}

func migrateBalancerPoolToSolidlyStable(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanager.Keeper, poolId uint64) {
	// fetch the pool with the given poolId
	balancerPool, err := gammKeeper.GetPool(ctx, poolId)
	if err != nil {
		panic(err)
	}

	// initialize the stableswap pool
	stableswapPool, err := stableswap.NewStableswapPool(
		poolId,
		stableswap.PoolParams{SwapFee: balancerPool.GetSwapFee(ctx), ExitFee: balancerPool.GetExitFee(ctx)},
		balancerPool.GetTotalPoolLiquidity(ctx),
		[]uint64{1, 1},
		"osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af", // Stride Foundation 2/3 multisig
		"",
	)
	if err != nil {
		panic(err)
	}

	// ensure the number of stableswap LP shares is the same as the number of balancer LP shares
	totalShares := sdk.NewCoin(
		gammtypes.GetPoolShareDenom(poolId),
		balancerPool.GetTotalShares(),
	)
	stableswapPool.TotalShares = totalShares

	// TODO: check balances
	// balancesBefore := 
	// overwrite the balancer pool with the new stableswap pool
	err = gammKeeper.OverwritePool(ctx, &stableswapPool)
	if err != nil {
		panic(err)
	}
}

func migrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanager.Keeper) {
	// N.B: pool id in gamm is to be deprecated in the future
	// Instead,it is moved to poolmanager.
	// nolint: staticcheck
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	poolmanagerKeeper.SetNextPoolId(ctx, nextPoolId)

	for poolId := uint64(1); poolId < nextPoolId; poolId++ {
		poolType, err := gammKeeper.GetPoolType(ctx, poolId)
		if err != nil {
			panic(err)
		}

		poolmanagerKeeper.SetPoolRoute(ctx, poolId, poolType)
	}
}

func registerOsmoIonMetadata(ctx sdk.Context, bankKeeper bankkeeper.Keeper) {
	uosmoMetadata := banktypes.Metadata{
		Description: "The native token of Osmosis",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    appParams.BaseCoinUnit,
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    appParams.HumanCoinUnit,
				Exponent: appParams.OsmoExponent,
				Aliases:  nil,
			},
		},
		Base:    appParams.BaseCoinUnit,
		Display: appParams.HumanCoinUnit,
	}

	uionMetadata := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uion",
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    "ion",
				Exponent: 6,
				Aliases:  nil,
			},
		},
		Base:    "uion",
		Display: "ion",
	}

	bankKeeper.SetDenomMetaData(ctx, uosmoMetadata)
	bankKeeper.SetDenomMetaData(ctx, uionMetadata)
}
