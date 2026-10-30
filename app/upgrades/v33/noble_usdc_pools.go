package v33

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	"github.com/osmosis-labs/osmosis/v32/x/gamm/pool-models/balancer"
	txfeestypes "github.com/osmosis-labs/osmosis/v32/x/txfees/types"
)

// MaxNobleUSDCPositionsPerPool bounds the positions withdrawn from one drained pool. A pool with more
// positions than this is skipped.
const MaxNobleUSDCPositionsPerPool = 20_000

// errPoolNotHoldingDenom marks a listed pool that no longer holds the migrated denom. It is skipped.
var errPoolNotHoldingDenom = errors.New("pool does not hold the migrated denom")

// NobleUSDCPoolMigration configures MigrateNobleUSDCPools.
type NobleUSDCPoolMigration struct {
	// Transmuter is the alloy transmuter that converts Denom into AlloyedDenom 1:1.
	Transmuter string
	// Denom is the denom replaced in the pools.
	Denom string
	// AlloyedDenom is the denom Transmuter mints, which replaces Denom in converted pools.
	AlloyedDenom string
	// CLConvertPools are concentrated liquidity pools converted in place.
	CLConvertPools []uint64
	// BalancerConvertPools are balancer pools converted in place.
	BalancerConvertPools []uint64
	// CLDrainPools are concentrated liquidity pools whose positions are withdrawn to their owners.
	CLDrainPools []uint64
	// MaxPositionsPerPool bounds the positions withdrawn from one drained pool.
	MaxPositionsPerPool int
}

// NobleUSDCPoolMigrationResult summarizes MigrateNobleUSDCPools.
type NobleUSDCPoolMigrationResult struct {
	// Converted is the number of pools whose Denom reserve became AlloyedDenom.
	Converted int
	// Drained is the number of pools whose positions were withdrawn. Kept positions may remain.
	Drained int
	// SkippedPools is the number of listed pools that were left unchanged.
	SkippedPools int
	// PositionsWithdrawn is the number of positions withdrawn to their owners.
	PositionsWithdrawn int
	// PositionsKept is the number of positions left in drained pools: positions owned by contracts,
	// and positions whose withdrawal failed (for example an unmatured underlying lock).
	PositionsKept int
	// ConvertedAmount is the amount of Denom converted out of pool reserves.
	ConvertedAmount osmomath.Int
	// FeeTokenRemoved reports whether Denom was removed as a fee token.
	FeeTokenRemoved bool
}

// mainnetNobleUSDCPoolMigration returns the osmosis-1 pool migration (see noble_usdc_pool_ids.go).
func mainnetNobleUSDCPoolMigration() NobleUSDCPoolMigration {
	return NobleUSDCPoolMigration{
		Transmuter:           AllUSDCTransmuterAddress,
		Denom:                NobleUSDCDenom,
		AlloyedDenom:         AllUSDCDenom,
		CLConvertPools:       nobleUSDCCLConvertPools,
		BalancerConvertPools: nobleUSDCBalancerConvertPools,
		CLDrainPools:         nobleUSDCCLDrainPools,
		MaxPositionsPerPool:  MaxNobleUSDCPositionsPerPool,
	}
}

// migrateNobleUSDCPoolsAndLog runs the mainnet pool migration and logs the outcome. A failure is
// logged instead of halting the upgrade.
func migrateNobleUSDCPoolsAndLog(ctx sdk.Context, keepers *keepers.AppKeepers) {
	m := mainnetNobleUSDCPoolMigration()
	result, err := MigrateNobleUSDCPools(ctx, keepers, m)
	if err != nil {
		ctx.Logger().Error("v33 upgrade: Noble USDC pool migration skipped", "error", err)
		return
	}
	ctx.Logger().Info("v33 upgrade: Noble USDC pool migration finished",
		"converted", result.Converted, "drained", result.Drained, "skipped_pools", result.SkippedPools,
		"positions_withdrawn", result.PositionsWithdrawn, "positions_kept", result.PositionsKept,
		"converted_amount", sdk.NewCoin(m.Denom, result.ConvertedAmount).String(),
		"fee_token_removed", result.FeeTokenRemoved)
}

// MigrateNobleUSDCPools moves Denom out of the configured pools:
//
//   - A converted pool exchanges its whole Denom balance for AlloyedDenom by joining the transmuter as
//     the pool account, then has Denom renamed to AlloyedDenom in its state. Positions, shares,
//     liquidity and price are unchanged, and TWAP records are started for the new denom pair.
//   - A drained pool has every position withdrawn to its owner, except positions owned by contracts,
//     which track their own positions. The owners receive Denom, which the holder migration that runs
//     after this one converts for user accounts.
//
// Each pool, and each withdrawn position, runs in its own cache context: a failure is logged and
// rolled back without affecting the others. Finally, Denom is removed as a fee token if its fee pool
// was converted or drained.
//
// An error is returned, with no state changed, if the config is invalid or the transmuter fails the
// same pre-flight check as the holder migration.
func MigrateNobleUSDCPools(ctx sdk.Context, keepers *keepers.AppKeepers, m NobleUSDCPoolMigration) (NobleUSDCPoolMigrationResult, error) {
	result := NobleUSDCPoolMigrationResult{ConvertedAmount: osmomath.ZeroInt()}
	if m.Denom == "" || m.AlloyedDenom == "" || m.Denom == m.AlloyedDenom || m.MaxPositionsPerPool <= 0 {
		return result, errors.New("invalid Noble USDC pool migration config")
	}
	transmuter, err := sdk.AccAddressFromBech32(m.Transmuter)
	if err != nil {
		return result, fmt.Errorf("transmuter address: %w", err)
	}
	if err := checkTransmuter(ctx, keepers, transmuter, m.Denom, m.AlloyedDenom); err != nil {
		return result, err
	}

	for _, poolId := range m.CLConvertPools {
		convertPoolAndCount(ctx, &result, poolId, "concentrated", func(cacheCtx sdk.Context) (osmomath.Int, error) {
			return convertCLPool(cacheCtx, keepers, m, transmuter, poolId)
		})
	}
	for _, poolId := range m.BalancerConvertPools {
		convertPoolAndCount(ctx, &result, poolId, "balancer", func(cacheCtx sdk.Context) (osmomath.Int, error) {
			return convertBalancerPool(cacheCtx, keepers, m, transmuter, poolId)
		})
	}
	for _, poolId := range m.CLDrainPools {
		withdrawn, kept, err := drainCLPool(ctx, keepers, m, poolId)
		if err != nil {
			logSkippedPool(ctx, poolId, "drain", err)
			result.SkippedPools++
			continue
		}
		result.Drained++
		result.PositionsWithdrawn += withdrawn
		result.PositionsKept += kept
	}

	result.FeeTokenRemoved = removeFeeTokenIfPoolMigrated(ctx, keepers, m)
	return result, nil
}

// convertPoolAndCount runs convert in a cache context and records the outcome in result.
func convertPoolAndCount(ctx sdk.Context, result *NobleUSDCPoolMigrationResult, poolId uint64, kind string, convert func(sdk.Context) (osmomath.Int, error)) {
	var amount osmomath.Int
	err := osmoutils.ApplyFuncIfNoErrorLogToDebug(ctx, func(cacheCtx sdk.Context) error {
		var err error
		amount, err = convert(cacheCtx)
		return err
	})
	if err != nil {
		logSkippedPool(ctx, poolId, "convert "+kind, err)
		result.SkippedPools++
		return
	}
	result.Converted++
	result.ConvertedAmount = result.ConvertedAmount.Add(amount)
}

// logSkippedPool logs a listed pool left unchanged. A pool that no longer holds the denom is expected
// churn and is logged at info level.
func logSkippedPool(ctx sdk.Context, poolId uint64, action string, err error) {
	if errors.Is(err, errPoolNotHoldingDenom) {
		ctx.Logger().Info("v33 upgrade: Noble USDC pool skipped", "pool_id", poolId, "action", action, "reason", err)
		return
	}
	ctx.Logger().Error("v33 upgrade: Noble USDC pool migration failed", "pool_id", poolId, "action", action, "error", err)
}

// convertCLPool converts the Denom reserve of a concentrated liquidity pool and renames Denom to
// AlloyedDenom in the pool. It returns the amount converted.
func convertCLPool(ctx sdk.Context, keepers *keepers.AppKeepers, m NobleUSDCPoolMigration, transmuter sdk.AccAddress, poolId uint64) (osmomath.Int, error) {
	pool, err := keepers.ConcentratedLiquidityKeeper.GetConcentratedPoolById(ctx, poolId)
	if err != nil {
		return osmomath.Int{}, err
	}
	if pool.GetToken0() != m.Denom && pool.GetToken1() != m.Denom {
		return osmomath.Int{}, errPoolNotHoldingDenom
	}
	amount, err := convertPoolReserve(ctx, keepers, transmuter, pool.GetAddress(), m.Denom, m.AlloyedDenom)
	if err != nil {
		return osmomath.Int{}, err
	}
	if err := keepers.ConcentratedLiquidityKeeper.ReplacePoolDenomUnsafe(ctx, poolId, m.Denom, m.AlloyedDenom); err != nil {
		return osmomath.Int{}, err
	}
	// Start TWAP records for the new denom pair, as on pool creation. Without them the pool's TWAP
	// records would stop updating.
	keepers.TwapKeeper.ConcentratedLiquidityListener().AfterConcentratedPoolCreated(ctx, nil, poolId)
	return amount, nil
}

// convertBalancerPool converts the Denom reserve of a balancer pool and renames Denom to AlloyedDenom
// in the pool. It returns the amount converted.
func convertBalancerPool(ctx sdk.Context, keepers *keepers.AppKeepers, m NobleUSDCPoolMigration, transmuter sdk.AccAddress, poolId uint64) (osmomath.Int, error) {
	pool, err := keepers.GAMMKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return osmomath.Int{}, err
	}
	balancerPool, ok := pool.(*balancer.Pool)
	if !ok {
		return osmomath.Int{}, fmt.Errorf("pool %d is %T, not a balancer pool", poolId, pool)
	}
	recorded, err := balancerPool.GetPoolAsset(m.Denom)
	if err != nil {
		return osmomath.Int{}, errPoolNotHoldingDenom
	}
	amount, err := convertPoolReserve(ctx, keepers, transmuter, balancerPool.GetAddress(), m.Denom, m.AlloyedDenom)
	if err != nil {
		return osmomath.Int{}, err
	}
	// The pool's recorded amount is what its LPs and swaps rely on, so the bank balance must cover it.
	if amount.LT(recorded.Token.Amount) {
		return osmomath.Int{}, fmt.Errorf("pool %d holds %s %s but records %s", poolId, amount, m.Denom, recorded.Token.Amount)
	}
	if err := keepers.GAMMKeeper.ReplaceBalancerPoolDenomUnsafe(ctx, poolId, m.Denom, m.AlloyedDenom); err != nil {
		return osmomath.Int{}, err
	}
	keepers.TwapKeeper.GammHooks().AfterCFMMPoolCreated(ctx, nil, poolId)
	return amount, nil
}

// convertPoolReserve joins the transmuter as pool with its whole denom balance and checks that the
// pool received the same amount of alloyedDenom. It returns the amount converted, zero if the pool
// holds no denom.
func convertPoolReserve(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter, pool sdk.AccAddress, denom, alloyedDenom string) (osmomath.Int, error) {
	coin := keepers.BankKeeper.GetBalance(ctx, pool, denom)
	if coin.IsZero() {
		return osmomath.ZeroInt(), nil
	}
	if err := joinAsHolder(ctx, keepers, transmuter, pool, coin, alloyedDenom); err != nil {
		return osmomath.Int{}, err
	}
	return coin.Amount, nil
}

// drainCLPool withdraws every position of a concentrated liquidity pool to its owner, apart from
// positions owned by contracts. Each withdrawal runs in its own cache context. It returns the number
// of positions withdrawn and kept, or an error, with no state changed, if the pool does not hold Denom
// or has more than MaxPositionsPerPool positions.
func drainCLPool(ctx sdk.Context, keepers *keepers.AppKeepers, m NobleUSDCPoolMigration, poolId uint64) (int, int, error) {
	clKeeper := keepers.ConcentratedLiquidityKeeper
	pool, err := clKeeper.GetConcentratedPoolById(ctx, poolId)
	if err != nil {
		return 0, 0, err
	}
	if pool.GetToken0() != m.Denom && pool.GetToken1() != m.Denom {
		return 0, 0, errPoolNotHoldingDenom
	}
	positionIds, err := clKeeper.GetPositionIdsForPool(ctx, poolId, m.MaxPositionsPerPool)
	if err != nil {
		return 0, 0, err
	}

	withdrawn, kept := 0, 0
	for _, positionId := range positionIds {
		if err := withdrawPositionToOwner(ctx, keepers, positionId); err != nil {
			ctx.Logger().Info("v33 upgrade: Noble USDC pool position kept", "pool_id", poolId, "position_id", positionId, "reason", err)
			kept++
			continue
		}
		withdrawn++
	}
	return withdrawn, kept, nil
}

// withdrawPositionToOwner withdraws all of a position's liquidity, with its spread rewards and
// incentives, to its owner in a cache context. Positions owned by contracts are not withdrawn.
func withdrawPositionToOwner(ctx sdk.Context, keepers *keepers.AppKeepers, positionId uint64) error {
	position, err := keepers.ConcentratedLiquidityKeeper.GetPosition(ctx, positionId)
	if err != nil {
		return err
	}
	owner, err := sdk.AccAddressFromBech32(position.Address)
	if err != nil {
		return fmt.Errorf("owner address: %w", err)
	}
	if keepers.WasmKeeper.HasContractInfo(ctx, owner) {
		return errors.New("owned by a contract")
	}
	return osmoutils.ApplyFuncIfNoErrorLogToDebug(ctx, func(cacheCtx sdk.Context) error {
		_, _, err := keepers.ConcentratedLiquidityKeeper.WithdrawPosition(cacheCtx, owner, positionId, position.Liquidity)
		return err
	})
}

// removeFeeTokenIfPoolMigrated removes Denom as a fee token if its fee pool is one of the migrated
// pools: a converted pool no longer holds Denom, and a drained pool has no liquidity to price it.
// It reports whether the fee token was removed.
func removeFeeTokenIfPoolMigrated(ctx sdk.Context, keepers *keepers.AppKeepers, m NobleUSDCPoolMigration) bool {
	feeToken, err := keepers.TxFeesKeeper.GetFeeToken(ctx, m.Denom)
	if err != nil {
		return false
	}
	if !containsPool(m.CLConvertPools, feeToken.PoolID) && !containsPool(m.BalancerConvertPools, feeToken.PoolID) && !containsPool(m.CLDrainPools, feeToken.PoolID) {
		return false
	}
	// A fee token with pool ID 0 is deleted.
	if err := keepers.TxFeesKeeper.SetFeeTokens(ctx, []txfeestypes.FeeToken{{Denom: m.Denom, PoolID: 0}}); err != nil {
		ctx.Logger().Error("v33 upgrade: removing Noble USDC fee token failed", "error", err)
		return false
	}
	return true
}

// containsPool reports whether poolId is in pools.
func containsPool(pools []uint64, poolId uint64) bool {
	for _, id := range pools {
		if id == poolId {
			return true
		}
	}
	return false
}
