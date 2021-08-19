package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sdksimulation "github.com/cosmos/cosmos-sdk/x/simulation"
)

type SimulationContext struct {
	R          *rand.Rand
	SdkCtx     sdk.Context
	App        *baseapp.BaseApp
	Accs       []simtypes.Account
	simAccount *simtypes.Account
}

func NewSimulationContext(r *rand.Rand, ctx sdk.Context, app *baseapp.BaseApp, accs []simtypes.Account) SimulationContext {
	return SimulationContext{r, ctx, app, accs, nil}
}

func (ctx *SimulationContext) GetMsgSender() simtypes.Account {
	if ctx.simAccount == nil {
		sel := ctx.R.Intn(len(ctx.Accs))
		ctx.simAccount = &ctx.Accs[sel]
	}
	return *ctx.simAccount
}

func GenAndDeliverTxWithRandFees(txCtx sdksimulation.OperationInput) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	spendable := txCtx.Bankkeeper.SpendableCoins(txCtx.Context, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(txCtx.CoinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "message doesn't leave room for fees"), nil, err
	}

	fees, err = simtypes.RandomFees(txCtx.R, txCtx.Context, coins)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(txCtx, fees)
}

func GenAndDeliverTx(txCtx sdksimulation.OperationInput, fees sdk.Coins) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	tx, err := helpers.GenTx(
		txCtx.TxGen,
		[]sdk.Msg{txCtx.Msg},
		fees,
		helpers.DefaultGenTxGas,
		txCtx.Context.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		txCtx.SimAccount.PrivKey,
	)

	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate mock tx"), nil, err
	}

	_, _, err = txCtx.App.Deliver(txCtx.TxGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(txCtx.Msg, true, "", txCtx.Cdc), nil, nil

}
