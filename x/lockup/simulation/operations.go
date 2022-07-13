package simulation

import (
	"math/rand"
	"time"

	osmo_simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

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

func RandomMsgLockTokens(k keeper.Keeper, sim *osmo_simulation.SimCtx, ctx sdk.Context) (*types.MsgLockTokens, error) {
	sender := sim.RandomSimAccount()
	lockCoins := sim.RandExponentialCoin(ctx, sender.Address)
	duration := osmo_simulation.RandSelect(sim, time.Minute, time.Hour, time.Hour*24)
	return &types.MsgLockTokens{
		Owner:    sender.Address.String(),
		Duration: duration,
		Coins:    sdk.Coins{lockCoins},
	}, nil
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

func RandomAccountLock(ctx sdk.Context, r *rand.Rand, k keeper.Keeper, addr sdk.AccAddress) *types.PeriodLock {
	locks := k.GetAccountPeriodLocks(ctx, addr)
	if len(locks) == 0 {
		return nil
	}
	return &locks[r.Intn(len(locks))]
}
