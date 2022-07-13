package simulation

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v7/app/params"
)

func noopTxBuilder() func(ctx sdk.Context, msg sdk.Msg) (sdk.Tx, error) {
	return func(sdk.Context, sdk.Msg) (sdk.Tx, error) { return nil, errors.New("unimplemented") }
}

func genAndDeliverTxWithRandFees(
	sim *SimCtx,
	msg sdk.Msg,
	msgName string,
	coinsSpentInMsg sdk.Coins,
	ctx sdk.Context,
) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	account, found := sim.FindAccount(msg.GetSigners()[0])
	if !found {
		return simulation.NoOpMsg(msgName, msgName, "unable to generate mock tx"), nil, errors.New("sim acct not found")
	}
	spendable := sim.App.GetBankKeeper().SpendableCoins(ctx, account.Address)

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(coinsSpentInMsg)
	if hasNeg {
		return simulation.NoOpMsg(msgName, msgName, "message doesn't leave room for fees"), nil, err
	}

	// Only allow fees in "uosmo"
	coins = sdk.NewCoins(sdk.NewCoin("uosmo", coins.AmountOf("uosmo")))

	fees, err = simulation.RandomFees(sim.GetRand(), ctx, coins)
	if err != nil {
		return simulation.NoOpMsg(msgName, msgName, "unable to generate fees"), nil, err
	}
	return sim.genAndDeliverTx(msg, msgName, fees, ctx)
}

// TODO: Comeback and clean this up to not suck
func (sim *SimCtx) genAndDeliverTx(
	msg sdk.Msg,
	msgName string, // TODO fix
	fees sdk.Coins,
	ctx sdk.Context,
) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	account, found := sim.FindAccount(msg.GetSigners()[0])
	if !found {
		return simulation.NoOpMsg(msgName, msgName, "unable to generate mock tx"), nil, errors.New("sim acct not found")
	}
	authAcc := sim.App.GetAccountKeeper().GetAccount(ctx, account.Address)
	txConfig := params.MakeEncodingConfig().TxConfig // TODO: unhardcode
	tx, err := helpers.GenTx(
		txConfig,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{authAcc.GetAccountNumber()},
		[]uint64{authAcc.GetSequence()},
		account.PrivKey,
	)
	if err != nil {
		return simulation.NoOpMsg(msgName, msgName, "unable to generate mock tx"), nil, err
	}
	return sim.deliverTx(tx, msg, msgName)
}

// TODO: Fix these args
func (sim *SimCtx) deliverTx(tx sdk.Tx, msg sdk.Msg, msgName string) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	txConfig := params.MakeEncodingConfig().TxConfig // TODO: unhardcode
	_, _, err := sim.App.GetBaseApp().Deliver(txConfig.TxEncoder(), tx)
	if err != nil {
		return simulation.NoOpMsg(msgName, msgName, "unable to deliver tx"), nil, err
	}

	return simulation.NewOperationMsg(msg, true, "", nil), nil, nil
}
