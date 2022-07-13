package simulation

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

func noopTxBuilder() func(ctx sdk.Context, msg sdk.Msg) (sdk.Tx, error) {
	return func(sdk.Context, sdk.Msg) (sdk.Tx, error) { return nil, errors.New("unimplemented") }
}

func genAndDeliverTxWithRandFees(
	sim *SimCtx,
	txGen client.TxConfig,
	msg legacytx.LegacyMsg,
	coinsSpentInMsg sdk.Coins,
	ctx sdk.Context,
	simAccount simulation.Account,
	moduleName string,
) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	account := sim.App.GetAccountKeeper().GetAccount(ctx, simAccount.Address)
	spendable := sim.App.GetBankKeeper().SpendableCoins(ctx, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(coinsSpentInMsg)
	if hasNeg {
		return simulation.NoOpMsg(moduleName, msg.Type(), "message doesn't leave room for fees"), nil, err
	}

	// Only allow fees in "uosmo"
	coins = sdk.NewCoins(sdk.NewCoin("uosmo", coins.AmountOf("uosmo")))

	fees, err = simulation.RandomFees(sim.GetRand(), ctx, coins)
	if err != nil {
		return simulation.NoOpMsg(moduleName, msg.Type(), "unable to generate fees"), nil, err
	}
	return genAndDeliverTx(sim, txGen, msg, fees, ctx, simAccount, moduleName)
}

func genAndDeliverTx(
	sim *SimCtx,
	txGen client.TxConfig,
	msg legacytx.LegacyMsg,
	fees sdk.Coins,
	ctx sdk.Context,
	simAccount simulation.Account,
	moduleName string,
) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	account := sim.App.GetAccountKeeper().GetAccount(ctx, simAccount.Address)
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
		return simulation.NoOpMsg(moduleName, msg.Type(), "unable to generate mock tx"), nil, err
	}

	_, _, err = sim.App.GetBaseApp().Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return simulation.NoOpMsg(moduleName, msg.Type(), "unable to deliver tx"), nil, err
	}

	return simulation.NewOperationMsg(msg, true, "", nil), nil, nil
}
