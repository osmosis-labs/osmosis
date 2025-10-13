package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	cltypes "github.com/osmosis-labs/osmosis/v30/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v30/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v30/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v30/x/gamm/types/migration"
	lockuptypes "github.com/osmosis-labs/osmosis/v30/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v30/x/superfluid/types"
)

var (
	STAKE            = "stake"
	DefaultAmt0      = osmomath.NewInt(1000000)
	DefaultCoin0     = sdk.NewCoin(STAKE, DefaultAmt0)
	USDC             = "usdc"
	DefaultAmt1      = osmomath.NewInt(5000000000)
	DefaultCoin1     = sdk.NewCoin(USDC, DefaultAmt1)
	DefaultCoins     = sdk.NewCoins(DefaultCoin0, DefaultCoin1)
	DefaultLowerTick = int64(30545000)
	DefaultUpperTick = int64(31500000)
)

func (s *KeeperTestSuite) SetupMigrationTest(ctx sdk.Context, superfluidDelegated, superfluidUndelegating, unlocking, noLock bool, percentOfSharesToMigrate osmomath.Dec) (joinPoolAmt sdk.Coins, balancerIntermediaryAcc types.SuperfluidIntermediaryAccount, balancerLock *lockuptypes.PeriodLock, poolCreateAcc, poolJoinAcc sdk.AccAddress, balancerPooId, clPoolId uint64, balancerPoolShareOut sdk.Coin, valAddr sdk.ValAddress) { //nolint:revive // TODO: refactor this function
	bankKeeper := s.App.BankKeeper
	gammKeeper := s.App.GAMMKeeper
	superfluidKeeper := s.App.SuperfluidKeeper
	lockupKeeper := s.App.LockupKeeper
	stakingKeeper := s.App.StakingKeeper
	poolmanagerKeeper := s.App.PoolManagerKeeper

	fullRangeCoins := sdk.NewCoins(defaultPoolAssets[0].Token, defaultPoolAssets[1].Token)

	// Generate and fund two accounts.
	// Account 1 will be the account that creates the pool.
	// Account 2 will be the account that joins the pool.
	delAddrs := CreateRandomAccounts(2)
	poolCreateAcc = delAddrs[0]
	poolJoinAcc = delAddrs[1]
	for _, acc := range delAddrs {
		err := testutil.FundAccount(ctx, bankKeeper, acc, defaultAcctFunds)
		s.Require().NoError(err)
	}

	// Set up a single validator.
	valAddr = s.SetupValidator(stakingtypes.Bonded)

	// Create a balancer pool of "stake" and "foo".
	msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.NewDec(0),
	}, defaultPoolAssets, defaultFutureGovernor)
	balancerPooId, err := poolmanagerKeeper.CreatePool(ctx, msg)
	s.Require().NoError(err)

	// Join the balancer pool.
	// Note the account balance before and after joining the pool.
	balanceBeforeJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)
	_, _, err = gammKeeper.JoinPoolNoSwap(ctx, poolJoinAcc, balancerPooId, gammtypes.OneShare.MulRaw(50), sdk.Coins{})
	s.Require().NoError(err)
	balanceAfterJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)

	// The balancer join pool amount is the difference between the account balance before and after joining the pool.
	joinPoolAmt, _ = balanceBeforeJoin.SafeSub(balanceAfterJoin...)

	// Determine the balancer pool's LP token denomination.
	balancerPoolDenom := gammtypes.GetPoolShareDenom(balancerPooId)

	// Register the balancer pool's LP token as a superfluid asset
	err = superfluidKeeper.AddNewSuperfluidAsset(ctx, types.SuperfluidAsset{
		Denom:     balancerPoolDenom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})
	s.Require().NoError(err)

	// Note how much of the balancer pool's LP token the account that joined the pool has.
	balancerPoolShareOut = bankKeeper.GetBalance(ctx, poolJoinAcc, balancerPoolDenom)

	// Create a cl pool with the same underlying assets as the balancer pool.
	clPool := s.PrepareCustomConcentratedPool(poolCreateAcc, defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, osmomath.ZeroDec())
	clPoolId = clPool.GetId()

	// Add a gov sanctioned link between the balancer and concentrated liquidity pool.
	migrationRecord := gammmigration.MigrationRecords{BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
		{BalancerPoolId: balancerPooId, ClPoolId: clPoolId},
	}}
	err = gammKeeper.OverwriteMigrationRecords(ctx, migrationRecord)
	s.Require().NoError(err)

	// The unbonding duration is the same as the staking module's unbonding duration.
	stakingParams, err := stakingKeeper.GetParams(ctx)
	unbondingDuration := stakingParams.UnbondingTime

	// Lock the LP tokens for the duration of the unbonding period.
	originalGammLockId := uint64(0)
	if !noLock {
		originalGammLockId = s.LockTokens(poolJoinAcc, sdk.NewCoins(balancerPoolShareOut), unbondingDuration)
	}

	// Superfluid delegate the balancer lock if the test case requires it.
	// Note the intermediary account that was created.
	if superfluidDelegated {
		err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), originalGammLockId, valAddr.String())
		s.Require().NoError(err)
		intermediaryAccConnection := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
		balancerIntermediaryAcc = superfluidKeeper.GetIntermediaryAccount(ctx, intermediaryAccConnection)
	}

	// Superfluid undelegate the lock if the test case requires it.
	if superfluidUndelegating {
		err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), originalGammLockId)
		s.Require().NoError(err)
	}

	// Unlock the balancer lock if the test case requires it.
	if unlocking {
		// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
		// we need to unlock lock via `SuperfluidUnbondLock`
		if superfluidUndelegating {
			err = superfluidKeeper.SuperfluidUnbondLock(ctx, originalGammLockId, poolJoinAcc.String())
			s.Require().NoError(err)
		} else {
			lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			s.Require().NoError(err)
			_, err = lockupKeeper.BeginUnlock(ctx, originalGammLockId, lock.Coins)
			s.Require().NoError(err)
		}
	}

	balancerLock = &lockuptypes.PeriodLock{}
	if !noLock {
		balancerLock, err = lockupKeeper.GetLockByID(ctx, originalGammLockId)
		s.Require().NoError(err)
	}

	// Create a full range position in the concentrated liquidity pool.
	// This is to have a spot price and liquidity value to work off when migrating.
	s.CreateFullRangePosition(clPool, fullRangeCoins)

	// Register the CL full range LP tokens as a superfluid asset.
	clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     clPoolDenom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})

	s.Require().NoError(err)
	return joinPoolAmt, balancerIntermediaryAcc, balancerLock, poolCreateAcc, poolJoinAcc, balancerPooId, clPoolId, balancerPoolShareOut, valAddr
}

func (s *KeeperTestSuite) SlashAndValidateResult(ctx sdk.Context, gammLockId, concentratedLockId, poolIdEntering uint64, percentOfSharesToMigrate osmomath.Dec, valAddr sdk.ValAddress, balancerLock lockuptypes.PeriodLock, expectSlash bool) {
	// Retrieve the concentrated lock and gamm lock prior to slashing.
	concentratedLockPreSlash, err := s.App.LockupKeeper.GetLockByID(ctx, concentratedLockId)
	s.Require().NoError(err)
	gammLockPreSlash, err := s.App.LockupKeeper.GetLockByID(ctx, gammLockId)
	if percentOfSharesToMigrate.LT(osmomath.OneDec()) {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
	}

	// Slash the validator.
	slashFactor := osmomath.NewDecWithPrec(5, 2)
	s.App.SuperfluidKeeper.SlashLockupsForValidatorSlash(
		ctx,
		valAddr,
		slashFactor)

	// Retrieve the concentrated lock and gamm lock after slashing.
	concentratedLockPostSlash, err := s.App.LockupKeeper.GetLockByID(ctx, concentratedLockId)
	s.Require().NoError(err)
	gammLockPostSlash, err := s.App.LockupKeeper.GetLockByID(ctx, gammLockId)
	if percentOfSharesToMigrate.LT(osmomath.OneDec()) {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
	}

	// Check if the concentrated lock was slashed.
	clDenom := cltypes.GetConcentratedLockupDenomFromPoolId(poolIdEntering)
	slashAmtCL := concentratedLockPreSlash.Coins.AmountOf(clDenom).ToLegacyDec().Mul(slashFactor).TruncateInt()
	if !expectSlash {
		slashAmtCL = osmomath.ZeroInt()
	}
	s.Require().Equal(concentratedLockPreSlash.Coins.AmountOf(clDenom).Sub(slashAmtCL).String(), concentratedLockPostSlash.Coins.AmountOf(clDenom).String())

	// Check if the gamm lock was slashed.
	// We only check if the gamm lock was slashed if the lock was not migrated entirely.
	// Otherwise, there would be no newly created gamm lock to check.
	if percentOfSharesToMigrate.LT(osmomath.OneDec()) {
		gammDenom := balancerLock.Coins[0].Denom
		slashAmtGamm := gammLockPreSlash.Coins.AmountOf(gammDenom).ToLegacyDec().Mul(slashFactor).TruncateInt()
		if !expectSlash {
			slashAmtGamm = osmomath.ZeroInt()
		}
		s.Require().Equal(gammLockPreSlash.Coins.AmountOf(gammDenom).Sub(slashAmtGamm).String(), gammLockPostSlash.Coins.AmountOf(gammDenom).String())
	}
}

// TODO add user balance pre swap and then add to result

func (s *KeeperTestSuite) ValidateMigrateResult(
	positionId, balancerPooId, poolIdLeaving, clPoolId, poolIdEntering uint64,
	percentOfSharesToMigrate, liquidityMigrated osmomath.Dec,
	balancerLock lockuptypes.PeriodLock,
	joinPoolAmt sdk.Coins,
	balancerPoolShareOut, coinsToMigrate sdk.Coin,
	amount0, amount1 osmomath.Int,
) {
	// Check that the concentrated liquidity and join time match what we expect
	position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(liquidityMigrated, position.Liquidity)
	s.Require().Equal(s.Ctx.BlockTime(), position.JoinTime)

	// Expect the poolIdLeaving to be the balancer pool id
	// Expect the poolIdEntering to be the concentrated liquidity pool id
	s.Require().Equal(balancerPooId, poolIdLeaving)
	s.Require().Equal(clPoolId, poolIdEntering)

	// exitPool has rounding difference.
	// We test if correct amt has been exited and frozen by comparing with rounding tolerance.
	defaultErrorTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: osmomath.NewDec(2),
		RoundingDir:       osmomath.RoundDown,
	}
	osmoassert.Equal(s.T(), defaultErrorTolerance, joinPoolAmt.AmountOf(defaultPoolAssets[0].Token.Denom).ToLegacyDec().Mul(percentOfSharesToMigrate).RoundInt(), amount0)
	osmoassert.Equal(s.T(), defaultErrorTolerance, joinPoolAmt.AmountOf(defaultPoolAssets[1].Token.Denom).ToLegacyDec().Mul(percentOfSharesToMigrate).RoundInt(), amount1)
}

type Positions struct {
	numAccounts                     int
	numBondedSuperfluid             int
	numUnbondingSuperfluidLocked    int
	numUnbondingSuperfluidUnlocking int
	numVanillaLockLocked            int
	numVanillaLockUnlocking         int
	numNoLock                       int
}

type positionInfo struct {
	joinPoolCoins sdk.Coins
	coin          sdk.Coin
	shares        osmomath.Int
	lockId        uint64
}

type PositionType int

const (
	BondedSuperfluid PositionType = iota
	UnbondingSuperfluidLocked
	UnbondingSuperfluidUnlocking
	VanillaLockLocked
	VanillaLockUnlocking
	NoLock
)
