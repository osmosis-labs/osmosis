package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	osmo_simulation "github.com/osmosis-labs/osmosis/x/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgSuperfluidDelegate          int = 100
	DefaultWeightMsgSuperfluidUndelegate        int = 50
	DefaultWeightMsgSuperfluidRedelegate        int = 50
	DefaultWeightSetSuperfluidAssetsProposal    int = 5
	DefaultWeightRemoveSuperfluidAssetsProposal int = 2

	OpWeightMsgSuperfluidDelegate   = "op_weight_msg_superfluid_delegate"
	OpWeightMsgSuperfluidUndelegate = "op_weight_msg_superfluid_undelegate"
	OpWeightMsgSuperfluidRedelegate = "op_weight_msg_superfluid_redelegate"
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
	bk stakingtypes.BankKeeper, sk types.StakingKeeper, lk types.LockupKeeper, k keeper.Keeper,
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
			SimulateMsgSuperfluidDelegate(ak, bk, sk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgSuperfluidUndelegate,
			SimulateMsgSuperfluidUndelegate(ak, bk, lk),
		),
		simulation.NewWeightedOperation(
			weightMsgSuperfluidRedelegate,
			SimulateMsgSuperfluidRedelegate(ak, bk, sk, lk),
		),
	}
}

// SimulateMsgSuperfluidDelegate generates a MsgSuperfluidDelegate with random values
func SimulateMsgSuperfluidDelegate(ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, sk types.StakingKeeper, lk types.LockupKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// select random validator
		validator := RandomValidator(ctx, r, sk)
		if validator == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "No validator"), nil, nil
		}

		// select random lockup
		lock, simAccount := RandomLockAndAccount(ctx, r, lk, accs)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidDelegate, "Account have no period lock"), nil, nil
		}

		msg := types.MsgSuperfluidDelegate{
			Sender:  lock.Owner,
			LockId:  lock.ID,
			ValAddr: validator.OperatorAddress,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func SimulateMsgSuperfluidUndelegate(ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, lk types.LockupKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		lock, simAccount := RandomLockAndAccount(ctx, r, lk, accs)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidUndelegate, "Account have no period lock"), nil, nil
		}

		msg := types.MsgSuperfluidUndelegate{
			Sender: simAccount.Address.String(),
			LockId: lock.ID,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)

	}
}

func SimulateMsgSuperfluidRedelegate(ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, sk types.StakingKeeper, lk types.LockupKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// select random validator
		validator := RandomValidator(ctx, r, sk)
		if validator == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "No validator"), nil, nil
		}

		lock, simAccount := RandomLockAndAccount(ctx, r, lk, accs)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "Account have no period lock"), nil, nil
		}

		msg := types.MsgSuperfluidRedelegate{
			Sender:     lock.Owner,
			LockId:     lock.ID,
			NewValAddr: validator.OperatorAddress,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func RandomLockAndAccount(ctx sdk.Context, r *rand.Rand, lk types.LockupKeeper, accs []simtypes.Account) (*lockuptypes.PeriodLock, simtypes.Account) {
	simAccount, _ := simtypes.RandomAcc(r, accs)
	locks, err := lk.GetPeriodLocks(ctx)
	if err != nil {
		return nil, simAccount
	}
	if len(locks) == 0 {
		return nil, simAccount
	}

	lock := locks[r.Intn(len(locks))]
	for _, acc := range accs {
		if acc.Address.String() == lock.Owner {
			return &lock, acc
		}
	}
	return &lock, simAccount
}

func RandomAccountLock(ctx sdk.Context, r *rand.Rand, lk types.LockupKeeper, addr sdk.AccAddress) *lockuptypes.PeriodLock {
	locks := lk.GetAccountPeriodLocks(ctx, addr)
	if len(locks) == 0 {
		return nil
	}
	return &locks[r.Intn(len(locks))]
}

func RandomValidator(ctx sdk.Context, r *rand.Rand, sk types.StakingKeeper) *stakingtypes.Validator {
	validators := sk.GetAllValidators(ctx)
	if len(validators) == 0 {
		return nil
	}
	return &validators[r.Intn(len(validators))]
}
