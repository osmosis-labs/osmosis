package simtypes

import (
	"math/rand"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/osmosis-labs/osmosis/v21/app/params"
)

// TODO: Must delete
func GenAndDeliverTxWithRandFees(
	r *rand.Rand,
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	msg legacytx.LegacyMsg,
	coinsSpentInMsg sdk.Coins,
	ctx sdk.Context,
	simAccount simulation.Account,
	ak AccountKeeper,
	bk BankKeeper,
	moduleName string,
) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(coinsSpentInMsg...)
	if hasNeg {
		return simulation.NoOpMsg(moduleName, msg.Type(), "message doesn't leave room for fees"), nil, err
	}

	// Only allow fees in "uosmo"
	coins = sdk.NewCoins(sdk.NewCoin("uosmo", coins.AmountOf("uosmo")))

	fees, err = simulation.RandomFees(r, ctx, coins)
	if err != nil {
		return simulation.NoOpMsg(moduleName, msg.Type(), "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(app, txGen, msg, fees, ctx, simAccount, ak, moduleName)
}

// TODO: Must delete
func GenAndDeliverTx(
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	msg legacytx.LegacyMsg,
	fees sdk.Coins,
	ctx sdk.Context,
	simAccount simulation.Account,
	ak AccountKeeper,
	moduleName string,
) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	tx, err := genTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		sims.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)
	if err != nil {
		return simulation.NoOpMsg(moduleName, msg.Type(), "unable to generate mock tx"), nil, err
	}

	txConfig := params.MakeEncodingConfig().TxConfig
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return simulation.OperationMsg{}, nil, err
	}

	app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

	return simulation.NewOperationMsg(msg, true, "", nil), nil, nil
}
