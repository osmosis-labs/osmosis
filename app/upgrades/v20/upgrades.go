package v20

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/v20/app/keepers"
	"github.com/osmosis-labs/osmosis/v20/app/upgrades"
	cltypes "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v20/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v20/x/lockup/types"
	poolincenitvestypes "github.com/osmosis-labs/osmosis/v20/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v20/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v20/x/superfluid/types"

	// SDK v47 modules
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type IncentivizedCFMMDirectWhenMigrationLinkPresentError struct {
	CFMMPoolID         uint64
	ConcentratedPoolID uint64
	CFMMGaugeID        uint64
}

var emptySlice = []string{}

func (e IncentivizedCFMMDirectWhenMigrationLinkPresentError) Error() string {
	return fmt.Sprintf("CFMM gauge ID (%d) incentivized CFMM pool (%d) directly when migration link is present with concentrated pool (%d)", e.CFMMGaugeID, e.CFMMPoolID, e.ConcentratedPoolID)
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// TODO START: Move this to v21
		// UNFORKINGNOTE: Add the remaining modules to this
		fromVM[govtypes.ModuleName] = 1
		baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		// https://github.com/cosmos/cosmos-sdk/pull/12363/files
		// Set param key table for params module migration
		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			subspace := subspace
			fmt.Printf("subspace: %+v\n", subspace.Name())

			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			// sdk
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable() //nolint:staticcheck
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck

			// ibc types
			case ibctransfertypes.ModuleName:
				keyTable = ibctransfertypes.ParamKeyTable() //nolint:staticcheck
			case icahosttypes.SubModuleName:
				keyTable = icahosttypes.ParamKeyTable() //nolint:staticcheck
			case icacontrollertypes.SubModuleName:
				keyTable = icacontrollertypes.ParamKeyTable() //nolint:staticcheck

			// wasm
			case wasmtypes.ModuleName:
				keyTable = wasmtypes.ParamKeyTable() //nolint:staticcheck

			// // POB
			// case buildertypes.ModuleName:
			// 	// already SDK v47
			// 	continue

			// osmosis modules
			case protorevtypes.ModuleName:
				keyTable = protorevtypes.ParamKeyTable() //nolint:staticcheck
			case superfluidtypes.ModuleName:
				keyTable = superfluidtypes.ParamKeyTable() //nolint:staticcheck

			default:
				continue
			}

			// 			// "Normal" keepers
			// AccountKeeper                *authkeeper.AccountKeeper
			// BankKeeper                   bankkeeper.BaseKeeper
			// AuthzKeeper                  *authzkeeper.Keeper
			// StakingKeeper                *stakingkeeper.Keeper
			// DistrKeeper                  *distrkeeper.Keeper
			// DowntimeKeeper               *downtimedetector.Keeper
			// SlashingKeeper               *slashingkeeper.Keeper
			// IBCKeeper                    *ibckeeper.Keeper
			// IBCHooksKeeper               *ibchookskeeper.Keeper
			// ICAHostKeeper                *icahostkeeper.Keeper
			// ICQKeeper                    *icqkeeper.Keeper
			// TransferKeeper               *ibctransferkeeper.Keeper
			// EvidenceKeeper               *evidencekeeper.Keeper
			// GAMMKeeper                   *gammkeeper.Keeper
			// TwapKeeper                   *twap.Keeper
			// LockupKeeper                 *lockupkeeper.Keeper
			// EpochsKeeper                 *epochskeeper.Keeper
			// IncentivesKeeper             *incentiveskeeper.Keeper
			// ProtoRevKeeper               *protorevkeeper.Keeper
			// MintKeeper                   *mintkeeper.Keeper
			// PoolIncentivesKeeper         *poolincentiveskeeper.Keeper
			// TxFeesKeeper                 *txfeeskeeper.Keeper
			// SuperfluidKeeper             *superfluidkeeper.Keeper
			// GovKeeper                    *govkeeper.Keeper
			// WasmKeeper                   *wasm.Keeper
			// ContractKeeper               *wasmkeeper.PermissionedKeeper
			// TokenFactoryKeeper           *tokenfactorykeeper.Keeper
			// PoolManagerKeeper            *poolmanager.Keeper
			// ValidatorSetPreferenceKeeper *valsetpref.Keeper
			// ConcentratedLiquidityKeeper  *concentratedliquidity.Keeper
			// CosmwasmPoolKeeper           *cosmwasmpool.Keeper

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Migrate Tendermint consensus parameters from x/params module to a deprecated x/consensus module.
		// The old params module is required to still be imported in your app.go in order to handle this migration.
		baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)
		// TODO END: Move this to v21

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Initialize the newly created param
		keepers.ConcentratedLiquidityKeeper.SetParam(ctx, cltypes.KeyUnrestrictedPoolCreatorWhitelist, emptySlice)

		// Initialize the new params in incentives for group creation.
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyGroupCreationFee, incentivestypes.DefaultGroupCreationFee)
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyCreatorWhitelist, emptySlice)

		// Initialize new param in the poolmanager module with a whitelist allowing to bypass taker fees.
		keepers.PoolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyReducedTakerFeeByWhitelist, emptySlice)

		// Converts pool incentive distribution records from concentrated gauges to group gauges.
		err = createGroupsForIncentivePairs(ctx, keepers)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// createGroupsForIncentivePairs converts pool incentive distribution records from concentrated gauges to group gauges.
// The expected update is to convert concentrated gauges to group gauges iff
//   - migration record exists for the concentrated pool and another CFMM pool
//   - if migration between concentrated and CFMM exists, then the CFMM pool is not incentivized individually
//
// All other distribution records are not modified.
//
// The updated distribution records are saved in the store.
func createGroupsForIncentivePairs(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	// Create map from CL pools ID to CFMM pools ID
	// from migration records
	migrationInfo, err := keepers.GAMMKeeper.GetAllMigrationInfo(ctx)
	if err != nil {
		return err
	}

	poolIDMigrationRecordMap := make(map[uint64]uint64)
	for _, info := range migrationInfo.BalancerToConcentratedPoolLinks {
		poolIDMigrationRecordMap[info.ClPoolId] = info.BalancerPoolId
		poolIDMigrationRecordMap[info.BalancerPoolId] = info.ClPoolId
	}

	distrInfo := keepers.PoolIncentivesKeeper.GetDistrInfo(ctx)

	// For all incentive distribution records,
	// retrieve the gauge associated with the record
	// If gauge directs incentives to a concentrated pool AND the concentrated pool
	// is linked to balancer via migration map, create
	// a group gauge and replace it in the distribution record.
	// Note that if there is a concentrated pool that is not
	// linked to balancer, nothing is done.
	// Stableswap pools are expected to be silently ignored. We do not
	// expect any stableswap pool to be linked to concentrated.
	for i, distrRecord := range distrInfo.Records {
		gaugeID := distrRecord.GaugeId

		// Gauge with ID zero goes to community pool.
		if gaugeID == poolincenitvestypes.CommunityPoolDistributionGaugeID {
			continue
		}

		gauge, err := keepers.IncentivesKeeper.GetGaugeByID(ctx, gaugeID)
		if err != nil {
			return err
		}

		// At the time of v20 upgrade, we only have concentrated pools
		// that are linked to balancer. Concentrated gauges receive all rewards
		// and then retroactively update balancer.
		// Concentrated pools have NoLock Gauge associated with them.
		// As a result, we look for this specific type here.
		// If type mismatched, this is a CFMM pool gauge. In that case,
		// we continue to the next incentive record after validating
		// that there is no migration record present for this CFMM pool. That is
		// it is not incentivized individually when concentrated pool already retroactively
		// distributed rewards to it.
		if gauge.DistributeTo.LockQueryType != lockuptypes.NoLock {
			// Validate that if there is a migration record pair between a concentrated
			// and a cfmm pool, only concentrated is present in the distribution records.
			longestLockableDuration, err := keepers.PoolIncentivesKeeper.GetLongestLockableDuration(ctx)
			if err != nil {
				return err
			}
			cfmmPoolID, err := keepers.PoolIncentivesKeeper.GetPoolIdFromGaugeId(ctx, gaugeID, longestLockableDuration)
			if err != nil {
				return err
			}

			// If we had a migration record present and a balancer pool is still incentivized individually,
			// something went wrong. This is because the presence of migration record implies retroactive
			// incentive distribution from concentrated to balancer.
			linkedConcentratedPoolID, hasAssociatedConcentratedPoolLinked := poolIDMigrationRecordMap[cfmmPoolID]
			if hasAssociatedConcentratedPoolLinked {
				return IncentivizedCFMMDirectWhenMigrationLinkPresentError{
					CFMMPoolID:         cfmmPoolID,
					ConcentratedPoolID: linkedConcentratedPoolID,
					CFMMGaugeID:        gaugeID,
				}
			}

			// Validation passed. This was an individual CFMM pool with no link to concentrated
			// Silently skip it.
			continue
		}

		// Get PoolID associated with the given NoLock gauge ID
		// NoLock gauges are associated with an incentives epoch duration.
		incentivesEpochDuration := keepers.IncentivesKeeper.GetEpochInfo(ctx).Duration
		concentratedPoolID, err := keepers.PoolIncentivesKeeper.GetPoolIdFromGaugeId(ctx, gaugeID, incentivesEpochDuration)
		if err != nil {
			return err
		}

		associatedGammPoolID, ok := poolIDMigrationRecordMap[concentratedPoolID]
		if !ok {
			// There is no CFMM pool ID for the concentrated pool ID, continue to the next.
			continue
		}

		// Found concentrated and CFMM pools that are linked by
		// migration records. Create a Group for them
		groupedPoolIDs := []uint64{concentratedPoolID, associatedGammPoolID}
		groupGaugeID, err := keepers.IncentivesKeeper.CreateGroupAsIncentivesModuleAcc(ctx, incentivestypes.PerpetualNumEpochsPaidOver, groupedPoolIDs)
		if err != nil {
			return err
		}

		// Replace the gauge ID with the group gauge ID in the distribution records
		distrInfo.Records[i].GaugeId = groupGaugeID
	}

	keepers.PoolIncentivesKeeper.SetDistrInfo(ctx, distrInfo)

	return nil
}
