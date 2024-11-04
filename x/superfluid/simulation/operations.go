package simulation

import (
	"math/rand"

	osmosimtypes "github.com/osmosis-labs/osmosis/v27/simulation/simtypes"

	"github.com/cosmos/cosmos-sdk/baseapp"

	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants.
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

var (
	typeMsgSuperfluidDelegate   = sdk.MsgTypeURL(&types.MsgSuperfluidDelegate{})
	typeMsgSuperfluidUndelegate = sdk.MsgTypeURL(&types.MsgSuperfluidUndelegate{})
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak stakingtypes.AccountKeeper,
	bk osmosimtypes.BankKeeper, sk types.StakingKeeper, lk types.LockupKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgSuperfluidDelegate   int
		weightMsgSuperfluidUndelegate int
		// weightMsgSuperfluidRedelegate int
	)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	protoCdc := codec.NewProtoCodec(interfaceRegistry)

	appParams.GetOrGenerate(OpWeightMsgSuperfluidDelegate, &weightMsgSuperfluidDelegate, nil,
		func(_ *rand.Rand) {
			weightMsgSuperfluidDelegate = DefaultWeightMsgSuperfluidDelegate
		},
	)

	appParams.GetOrGenerate(OpWeightMsgSuperfluidUndelegate, &weightMsgSuperfluidUndelegate, nil,
		func(_ *rand.Rand) {
			weightMsgSuperfluidUndelegate = DefaultWeightMsgSuperfluidUndelegate
		},
	)

	// appParams.GetOrGenerate(OpWeightMsgSuperfluidRedelegate, &weightMsgSuperfluidRedelegate, nil,
	// 	func(_ *rand.Rand) {
	// 		weightMsgSuperfluidRedelegate = DefaultWeightMsgSuperfluidRedelegate
	// 	},
	// )

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSuperfluidDelegate,
			SimulateMsgSuperfluidDelegate(protoCdc, ak, bk, sk, lk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgSuperfluidUndelegate,
			SimulateMsgSuperfluidUndelegate(protoCdc, ak, bk, lk, k),
		),
		// simulation.NewWeightedOperation(
		// 	weightMsgSuperfluidRedelegate,
		// 	SimulateMsgSuperfluidRedelegate(ak, bk, sk, lk, k),
		// ),
	}
}

// SimulateMsgSuperfluidDelegate generates a MsgSuperfluidDelegate with random values.
func SimulateMsgSuperfluidDelegate(cdc *codec.ProtoCodec, ak stakingtypes.AccountKeeper, bk osmosimtypes.BankKeeper, sk types.StakingKeeper, lk types.LockupKeeper, k keeper.Keeper) simtypes.Operation {
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

		multiplier := k.GetOsmoEquivalentMultiplier(ctx, lock.Coins[0].Denom)
		if multiplier.IsZero() {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidDelegate, "not able to do superfluid staking if asset Multiplier is zero"), nil, nil
		}

		if !k.GetLockIdIntermediaryAccountConnection(ctx, lock.ID).Empty() {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidDelegate, "Lock is already used for superfluid staking"), nil, nil
		}

		msg := types.MsgSuperfluidDelegate{
			Sender:  lock.Owner,
			LockId:  lock.ID,
			ValAddr: validator.OperatorAddress,
		}

		opMsg, err := osmosimtypes.GenerateAndDeliverTx(r, app, ctx, chainID, cdc, ak, bk, simAccount, types.ModuleName, &msg, typeMsgSuperfluidDelegate, false)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSuperfluidDelegate, "unable to generate and deliver tx"), nil, err
		}

		return opMsg, nil, nil
	}
}

func SimulateMsgSuperfluidUndelegate(cdc *codec.ProtoCodec, ak stakingtypes.AccountKeeper, bk osmosimtypes.BankKeeper, lk types.LockupKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		lock, simAccount := RandomLockAndAccount(ctx, r, lk, accs)
		if lock == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidUndelegate, "Account have no period lock"), nil, nil
		}

		if k.GetLockIdIntermediaryAccountConnection(ctx, lock.ID).Empty() {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSuperfluidUndelegate, "Lock is not used for superfluid staking"), nil, nil
		}

		msg := types.MsgSuperfluidUndelegate{
			Sender: simAccount.Address.String(),
			LockId: lock.ID,
		}

		opMsg, err := osmosimtypes.GenerateAndDeliverTx(r, app, ctx, chainID, cdc, ak, bk, simAccount, types.ModuleName, &msg, typeMsgSuperfluidUndelegate, false)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSuperfluidUndelegate, "unable to generate and deliver tx"), nil, err
		}

		return opMsg, nil, nil
	}
}

// func SimulateMsgSuperfluidRedelegate(ak stakingtypes.AccountKeeper, bk osmosimtypes.BankKeeper, sk types.StakingKeeper, lk types.LockupKeeper, k keeper.Keeper) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
// 	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
// 		simAccount, _ := simtypes.RandomAcc(r, accs)

// 		// select random validator
// 		validator := RandomValidator(ctx, r, sk)
// 		if validator == nil {
// 			return simtypes.NoOpMsg(
// 				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "No validator"), nil, nil
// 		}

// 		lock, simAccount := RandomLockAndAccount(ctx, r, lk, accs)
// 		if lock == nil {
// 			return simtypes.NoOpMsg(
// 				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "Account have no period lock"), nil, nil
// 		}

// 		if k.GetLockIdIntermediaryAccountConnection(ctx, lock.ID).Empty() {
// 			return simtypes.NoOpMsg(
// 				types.ModuleName, types.TypeMsgSuperfluidRedelegate, "Lock is not used for superfluid staking"), nil, nil
// 		}

// 		msg := types.MsgSuperfluidRedelegate{
// 			Sender:     lock.Owner,
// 			LockId:     lock.ID,
// 			NewValAddr: validator.OperatorAddress,
// 		}

// 		txGen := testutil.MakeTestEncodingConfig().TxConfig
// 		return osmosimtypes.GenAndDeliverTxWithRandFees(
// 			r, app, txGen, &msg, nil, ctx, simAccount, ak, bk, types.ModuleName)
// 	}
// }

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
	validators, err := sk.GetAllValidators(ctx)
	if err != nil {
		return nil
	}
	if len(validators) == 0 {
		return nil
	}
	return &validators[r.Intn(len(validators))]
}
