package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func GenAndDeliverTxWithRandFees(
	r *rand.Rand,
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	msg legacytx.LegacyMsg,
	coinsSpentInMsg sdk.Coins,
	ctx sdk.Context,
	simAccount simtypes.Account,
	ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper,
	moduleName string,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(coinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(moduleName, msg.Type(), "message doesn't leave room for fees"), nil, err
	}

	// Only allow fees in "uosmo"
	coins = sdk.NewCoins(sdk.NewCoin("uosmo", coins.AmountOf("uosmo")))

	fees, err = simtypes.RandomFees(r, ctx, coins)
	if err != nil {
		return simtypes.NoOpMsg(moduleName, msg.Type(), "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(app, txGen, msg, fees, ctx, simAccount, ak, moduleName)
}

func GenAndDeliverTx(
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	msg legacytx.LegacyMsg,
	fees sdk.Coins,
	ctx sdk.Context,
	simAccount simtypes.Account,
	ak stakingTypes.AccountKeeper,
	moduleName string,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)
	if err != nil {
		return simtypes.NoOpMsg(moduleName, msg.Type(), "unable to generate mock tx"), nil, err
	}

	_, _, err = app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(moduleName, msg.Type(), "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
}
