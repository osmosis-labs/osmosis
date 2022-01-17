package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	osmo_simulation "github.com/osmosis-labs/osmosis/x/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgSuperfluidDelegate   int = 10
	DefaultWeightMsgSuperfluidUndelegate int = 5
	DefaultWeightMsgSuperfluidRedelegate int = 5
	OpWeightMsgSuperfluidDelegate            = "op_weight_msg_superfluid_delegate"
	OpWeightMsgSuperfluidUndelegate          = "op_weight_msg_superfluid_undelegate"
	OpWeightMsgSuperfluidRedelegate          = "op_weight_msg_superfluid_redelegate"
)

// Steps for superfluid simulation
// - lockup module should create random lockups (ensure simulation already exists for this)
// - SuperfluidDelegate for existing lockup or create new one
// - SuperfluidUndelegate for random lockup
// - SuperfluidRedelegate for random lockup
// - lockup module’s begin unlock for random lockup simulation  (check simulation is already available)
// - Price should be modified as time goes by gamm module  (check simulation is already available)
// - AfterEpochEnd hook should be coming for params.RefreshEpochIdentifier from epoch module (check simulation is already available)
// - BeforeValidatorSlashed and BeforeSlashingUnbondingDelegation hook should be coming from staking module (check simulation is already available)
// - Distribution module should be distributing rewards on simulation (check simulation is already available)
// - Incentives module should distribute superfluid gauge rewards on simulation (check simulation is already available)
// - Time passing for checking automatic unbondings

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak stakingtypes.AccountKeeper,
	bk stakingtypes.BankKeeper, lk superfluidtypes.LockupKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgSuperfluidDelegate   int
		weightMsgSuperfluidUndelegate int
		weightMsgSuperfluidRedelegate int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgSuperfluidDelegate, &weightMsgSuperfluidDelegate, nil,
		func(_ *rand.Rand) {
			weightMsgSuperfluidDelegate = DefaultWeightMsgSuperfluidDelegate
			weightMsgSuperfluidUndelegate = DefaultWeightMsgSuperfluidUndelegate
			weightMsgSuperfluidRedelegate = DefaultWeightMsgSuperfluidRedelegate
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSuperfluidDelegate,
			SimulateMsgSuperfluidDelegate(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgSuperfluidUndelegate,
			SimulateMsgSuperfluidUndelegate(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgSuperfluidRedelegate,
			SimulateMsgSuperfluidRedelegate(ak, bk, k),
		),
	}
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

// SimulateMsgSuperfluidDelegate generates a MsgSuperfluidDelegate with random values
func SimulateMsgSuperfluidDelegate(ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// owner of lockup
		// random existing lock id
		// select random validator - if not exists, use ""

		lock := RandomLock(ctx, r, k, simAccount.Address)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "Account have no period lock"), nil, nil
		}

		msg := types.MsgSuperfluidDelegate{
			Sender:  lock.Owner,
			LockId:  lock.Id,
			ValAddr: valAddr,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, SuperfluidDelegate, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func SimulateMsgSuperfluidUndelegate(ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		msg := types.MsgSuperfluidUndelegate{
			Sender: simAccount.Address.String(),
			LockId: lockId,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)

	}
}

func SimulateMsgSuperfluidRedelegate(ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "Account have no coin"), nil, nil
		}

		lock := RandomLock(ctx, r, k, simAccount.Address)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "Account have no period lock"), nil, nil
		}

		msg := types.MsgSuperfluidRedelegate{
			Sender:     lock.Owner,
			LockId:     lock.Id,
			NewValAddr: newValAddr,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func RandomLock(ctx sdk.Context, r *rand.Rand, k keeper.LockupKeeper) *types.PeriodLock {
	locks := k.GetPeriodLocks(ctx)
	if len(locks) == 0 {
		return nil
	}
	return &locks[r.Intn(len(locks))]
}
