package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	osmo_simulation "github.com/osmosis-labs/osmosis/v3/x/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v3/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v3/x/lockup/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgLockTokens        int = 10
	DefaultWeightMsgBeginUnlockingAll int = 10
	DefaultWeightMsgUnlockTokens      int = 10
	DefaultWeightMsgBeginUnlocking    int = 10
	DefaultWeightMsgUnlockPeriodLock  int = 10
	OpWeightMsgLockTokens                 = "op_weight_msg_create_lockup"
	OpWeightMsgBeginUnlockingAll          = "op_weight_msg_begin_unlocking_all"
	OpWeightMsgUnlockTokens               = "op_weight_msg_unlock_tokens"
	OpWeightMsgBeginUnlocking             = "op_weight_msg_begin_unlocking"
	OpWeightMsgUnlockPeriodLock           = "op_weight_msg_unlock_period_lock"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgLockTokens        int
		weightMsgBeginUnlockingAll int
		weightMsgUnlockTokens      int
		weightMsgBeginUnlocking    int
		weightMsgUnlockPeriodLock  int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgLockTokens, &weightMsgLockTokens, nil,
		func(_ *rand.Rand) {
			weightMsgLockTokens = DefaultWeightMsgLockTokens
			weightMsgBeginUnlockingAll = DefaultWeightMsgBeginUnlockingAll
			weightMsgUnlockTokens = DefaultWeightMsgUnlockTokens
			weightMsgBeginUnlocking = DefaultWeightMsgBeginUnlocking
			weightMsgUnlockPeriodLock = DefaultWeightMsgUnlockPeriodLock
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLockTokens,
			SimulateMsgLockTokens(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBeginUnlockingAll,
			SimulateMsgBeginUnlockingAll(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUnlockTokens,
			SimulateMsgUnlockTokens(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBeginUnlocking,
			SimulateMsgBeginUnlocking(ak, bk, k),
		), simulation.NewWeightedOperation(
			weightMsgUnlockPeriodLock,
			SimulateMsgUnlockPeriodLock(ak, bk, k),
		),
	}
}

func genLockTokens(r *rand.Rand, acct simtypes.Account, coins sdk.Coins) (res sdk.Coins) {
	numCoins := 1 + r.Intn(Min(coins.Len(), 6))
	denomIndices := r.Perm(numCoins)
	for i := 0; i < numCoins; i++ {
		denom := coins[denomIndices[i]].Denom
		amt, _ := simtypes.RandPositiveInt(r, coins[i].Amount)
		res = append(res, sdk.Coin{Denom: denom, Amount: amt})
	}

	res.Sort()
	return
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// SimulateMsgLockTokens generates a MsgLockTokens with random values
func SimulateMsgLockTokens(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgLockTokens, "Account have no coin"), nil, nil
		}
		lockTokens := genLockTokens(r, simAccount, simCoins)

		durationSecs := r.Intn(1 * 60 * 60 * 24 * 7) // range of 1 week
		duration := time.Duration(durationSecs) * time.Second

		msg := types.MsgLockTokens{
			Owner:    simAccount.Address.String(),
			Duration: duration,
			Coins:    lockTokens,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, lockTokens, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func SimulateMsgBeginUnlockingAll(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgBeginUnlockingAll, "Account have no coin"), nil, nil
		}

		msg := types.MsgBeginUnlockingAll{
			Owner: simAccount.Address.String(),
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func SimulateMsgUnlockTokens(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgUnlockTokens, "Account have no coin"), nil, nil
		}

		msg := types.MsgUnlockTokens{
			Owner: simAccount.Address.String(),
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func SimulateMsgBeginUnlocking(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgBeginUnlocking, "Account have no coin"), nil, nil
		}

		lock := RandomAccountLock(ctx, r, k, simAccount.Address)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgBeginUnlocking, "Account have no period lock"), nil, nil
		}

		msg := types.MsgBeginUnlocking{
			Owner: simAccount.Address.String(),
			ID:    lock.ID,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func SimulateMsgUnlockPeriodLock(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgUnlockPeriodLock, "Account have no coin"), nil, nil
		}

		lock := RandomAccountLock(ctx, r, k, simAccount.Address)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgUnlockPeriodLock, "Account have no period lock"), nil, nil
		}

		// TODO: Switch this to instead be a future op on locking
		if k.GetAccountUnlockableCoins(ctx, simAccount.Address).Empty() {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgUnlockPeriodLock, "Account hasn't started unlocking"), nil, nil
		}
		msg := types.MsgUnlockPeriodLock{
			Owner: simAccount.Address.String(),
			ID:    lock.ID,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func RandomAccountLock(ctx sdk.Context, r *rand.Rand, k keeper.Keeper, addr sdk.AccAddress) *types.PeriodLock {
	locks := k.GetAccountPeriodLocks(ctx, addr)
	if len(locks) == 0 {
		return nil
	}
	return &locks[r.Intn(len(locks))]
}
