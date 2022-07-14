package simulation

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v7/app/params"
)

//nolint:deadcode,unused
func noopTxBuilder() func(ctx sdk.Context, msg sdk.Msg) (sdk.Tx, error) {
	return func(sdk.Context, sdk.Msg) (sdk.Tx, error) { return nil, errors.New("unimplemented") }
}

// TODO: Comeback and clean this up to not suck
func (sim *SimCtx) defaultTxBuilder(
	ctx sdk.Context,
	msg sdk.Msg,
	msgName string, // TODO fix
) (sdk.Tx, error) {
	account, found := sim.FindAccount(msg.GetSigners()[0])
	if !found {
		return nil, errors.New("unable to generate mock tx: sim acct not found")
	}
	authAcc := sim.App.GetAccountKeeper().GetAccount(ctx, account.Address)
	txConfig := params.MakeEncodingConfig().TxConfig // TODO: unhardcode
	// TODO: Consider making a default tx builder that charges some random fees
	// Low value for amount of work right now though.
	fees := sdk.Coins{}
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
		return nil, fmt.Errorf("unable to generate mock tx %v", err)
	}
	return tx, nil
}

// TODO: Fix these args
func (sim *SimCtx) deliverTx(tx sdk.Tx, msg sdk.Msg, msgName string) (simulation.OperationMsg, []simulation.FutureOperation, error) {
	txConfig := params.MakeEncodingConfig().TxConfig // TODO: unhardcode
	_, results, err := sim.App.GetBaseApp().Deliver(txConfig.TxEncoder(), tx)
	if err != nil {
		return simulation.NoOpMsg(msgName, msgName, fmt.Sprintf("unable to deliver tx. \nreason: %v\n results: %v\n msg: %s\n tx: %s", err, results, msg, tx)), nil, err
	}

	return simulation.NewOperationMsg(msg, true, "", nil), nil, nil
}
