package poolmanager

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	alloyedpooltypes "github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/cosmwasm/msg/v3"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

var IntMaxValue = intMaxValue

func (k Keeper) GetNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPoolIdAndIncrement(ctx)
}

// SetPoolRoutesUnsafe sets the given routes to the poolmanager keeper
// to allow routing from a pool type to a certain swap module.
// For example, balancer -> gamm.
// This utility function is only exposed for testing and should not be moved
// outside of the _test.go files.
func (k *Keeper) SetPoolRoutesUnsafe(routes map[types.PoolType]types.PoolModuleI) {
	k.routes = routes
	k.cachedPoolModules = &sync.Map{}
}

// SetPoolModulesUnsafe sets the given modules to the poolmanager keeper.
// This utility function is only exposed for testing and should not be moved
// outside of the _test.go files.
func (k *Keeper) SetPoolModulesUnsafe(poolModules []types.PoolModuleI) {
	k.poolModules = poolModules
	k.cachedPoolModules = &sync.Map{}
}

func (k Keeper) GetAllPoolRoutes(ctx sdk.Context) []types.ModuleRoute {
	return k.getAllPoolRoutes(ctx)
}

func (k Keeper) ValidateCreatedPool(ctx sdk.Context, poolId uint64, pool types.PoolI) error {
	return k.validateCreatedPool(poolId, pool)
}

func (k Keeper) CreateMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) ([]osmomath.Int, error) {
	return k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
}

func (k Keeper) TrackVolume(ctx sdk.Context, poolId uint64, volumeGenerated sdk.Coin) {
	k.trackVolume(ctx, poolId, volumeGenerated)
}

func (k Keeper) ChargeTakerFee(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, sender sdk.AccAddress, exactIn bool) (sdk.Coin, sdk.Coin, error) {
	return k.chargeTakerFee(ctx, tokenIn, tokenOutDenom, sender, exactIn)
}

func (k Keeper) QueryAndCheckAlloyedDenom(ctx sdk.Context, contractAddr sdk.AccAddress) (string, error) {
	return k.queryAndCheckAlloyedDenom(ctx, contractAddr)
}

func (k Keeper) SnapshotTakerFeeShareAlloyComposition(ctx sdk.Context, contractAddr sdk.AccAddress) ([]types.TakerFeeShareAgreement, error) {
	return k.snapshotTakerFeeShareAlloyComposition(ctx, contractAddr)
}

func (k Keeper) RecalculateAndSetTakerFeeShareAlloyComposition(ctx sdk.Context, poolId uint64) error {
	return k.recalculateAndSetTakerFeeShareAlloyComposition(ctx, poolId)
}

func (k Keeper) GetCachedTrackers() (map[string]types.TakerFeeShareAgreement, map[string]types.AlloyContractTakerFeeShareState, []uint64) {
	return k.getCacheTrackers()
}

func (k *Keeper) SetCacheTrackers(takerFeeShareAgreement map[string]types.TakerFeeShareAgreement, registeredAlloyPoolToState map[string]types.AlloyContractTakerFeeShareState, registeredAlloyedPoolId []uint64) {
	k.setCacheTrackers(takerFeeShareAgreement, registeredAlloyPoolToState, registeredAlloyedPoolId)
}

func (k Keeper) GetAlloyedDenomFromPoolId(ctx sdk.Context, poolId uint64) (string, error) {
	return k.getAlloyedDenomFromPoolId(ctx, poolId)
}

func (k Keeper) GetTakerFeeShareAgreements(denomsInvolvedInRoute []string) ([]types.TakerFeeShareAgreement, []types.TakerFeeShareAgreement) {
	return k.getTakerFeeShareAgreements(denomsInvolvedInRoute)
}

func (k Keeper) ProcessDenomShareAgreements(ctx sdk.Context, denomShareAgreements []types.TakerFeeShareAgreement, totalTakerFees sdk.Coins) error {
	return k.processDenomShareAgreements(ctx, denomShareAgreements, totalTakerFees)
}

func (k Keeper) ProcessAlloyedAssetShareAgreements(ctx sdk.Context, alloyedAssetShareAgreements []types.TakerFeeShareAgreement, totalTakerFees sdk.Coins) error {
	return k.processAlloyedAssetShareAgreements(ctx, alloyedAssetShareAgreements, totalTakerFees)
}

func (k Keeper) ValidatePercentage(percentage osmomath.Dec) error {
	return k.validatePercentage(percentage)
}

func (k Keeper) CreateNormalizationFactorsMap(assetConfigs []alloyedpooltypes.AssetConfig) (map[string]osmomath.Dec, error) {
	return k.createNormalizationFactorsMap(assetConfigs)
}

func (k Keeper) CalculateTakerFeeShareAgreements(ctx sdk.Context, totalPoolLiquidity []sdk.Coin, normalizationFactors map[string]osmomath.Dec) ([]types.TakerFeeShareAgreement, error) {
	return k.calculateTakerFeeShareAgreements(totalPoolLiquidity, normalizationFactors)
}

func (k *Keeper) SetRegisteredAlloyedPool(ctx sdk.Context, poolId uint64) error {
	return k.setRegisteredAlloyedPool(ctx, poolId)
}

func (k *Keeper) SetTakerFeeShareAgreementsMapCached(ctx sdk.Context) error {
	return k.setTakerFeeShareAgreementsMapCached(ctx)
}

func (k Keeper) GetAllTakerFeeShareAgreementsMap(ctx sdk.Context) (map[string]types.TakerFeeShareAgreement, error) {
	return k.getAllTakerFeeShareAgreementsMap(ctx)
}

func (k Keeper) IncreaseTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, takerFeeShareDenom string, takerFeeChargedDenom string, additiveValue osmomath.Int) error {
	return k.increaseTakerFeeShareDenomsToAccruedValue(ctx, takerFeeShareDenom, takerFeeChargedDenom, additiveValue)
}

func (k Keeper) GetAllRegisteredAlloyedPoolsByDenomMap(ctx sdk.Context) (map[string]types.AlloyContractTakerFeeShareState, error) {
	return k.getAllRegisteredAlloyedPoolsByDenomMap(ctx)
}

func (k *Keeper) SetAllRegisteredAlloyedPoolsByDenomCached(ctx sdk.Context) error {
	return k.setAllRegisteredAlloyedPoolsByDenomCached(ctx)
}

func (k Keeper) GetAllRegisteredAlloyedPoolsIdArray(ctx sdk.Context) ([]uint64, error) {
	return k.getAllRegisteredAlloyedPoolsIdArray(ctx)
}

func (k *Keeper) SetAllRegisteredAlloyedPoolIdArrayCached(ctx sdk.Context) error {
	return k.setAllRegisteredAlloyedPoolIdArrayCached(ctx)
}
