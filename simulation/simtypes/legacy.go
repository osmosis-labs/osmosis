package simtypes

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenerateAndDeliverTx generates a random fee (or with zero fees if set), then generates a
// signed mock tx and delivers the tx to the app for simulated operations.
func GenerateAndDeliverTx(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	chainId string,
	cdc *codec.ProtoCodec,
	ak AccountKeeper,
	bk BankKeeper,
	account simtypes.Account,
	moduleName string,
	msg sdk.Msg,
	msgType string,
	withZeroFees bool,
) (simtypes.OperationMsg, error) {
	simAccount := ak.GetAccount(ctx, account.Address)
	spendable := bk.SpendableCoins(ctx, simAccount.GetAddress())

	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           tx.NewTxConfig(cdc, tx.DefaultSignModes),
		Cdc:             cdc,
		Msg:             msg,
		Context:         ctx,
		SimAccount:      account,
		AccountKeeper:   ak,
		Bankkeeper:      bk,
		ModuleName:      moduleName,
		CoinsSpentInMsg: spendable,
	}

	var opMsg simtypes.OperationMsg
	var err error
	if withZeroFees {
		opMsg, _, err = simulation.GenAndDeliverTx(txCtx, sdk.Coins{})
	} else {
		opMsg, _, err = simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
	if err != nil || !opMsg.OK {
		return opMsg, fmt.Errorf("failed to generate and deliver tx: %w", err)
	}

	return opMsg, nil
}

// GenerateAndCheckTx generates a random fee (or with zero fees if set), then generates a signed
// mock tx and calls `CheckTx` for simulated operations. This heavily matches the cosmos sdk
// `util.go` GenAndDeliverTx method from the simulation module.
func GenerateAndCheckTx(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	chainId string,
	cdc *codec.ProtoCodec,
	ak AccountKeeper,
	bk BankKeeper,
	account simtypes.Account,
	moduleName string,
	msg sdk.Msg,
	msgType string,
	withZeroFees bool,
) (simtypes.OperationMsg, error) {
	// TODO(DEC-1174): Root-cause CheckTx failing on Block Height 1 and remove this workaround.
	if ctx.BlockHeight() == 1 {
		return simtypes.NoOpMsg(moduleName, msgType, "CheckTx skipped for block height 1"), nil
	}

	// Workaround: cosmos-sdk Simulation hard-codes to a deliver context. Generate and use a new
	// check context (with the same headers) specifically for CheckTx.
	checkTxCtx := app.NewContextLegacy(true, ctx.BlockHeader())

	simAccount := ak.GetAccount(checkTxCtx, account.Address)
	spendable := bk.SpendableCoins(checkTxCtx, simAccount.GetAddress())

	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           tx.NewTxConfig(cdc, tx.DefaultSignModes),
		Cdc:             cdc,
		Msg:             msg,
		Context:         checkTxCtx,
		SimAccount:      account,
		AccountKeeper:   ak,
		Bankkeeper:      bk,
		ModuleName:      moduleName,
		CoinsSpentInMsg: spendable,
	}

	var err error

	var fees sdk.Coins
	if withZeroFees {
		fees = sdk.Coins{}
	} else {
		coins, hasNeg := spendable.SafeSub(txCtx.CoinsSpentInMsg...)
		if hasNeg {
			return simtypes.NoOpMsg(txCtx.ModuleName, msgType, "message doesn't leave room for fees"), nil
		}

		fees, err = simtypes.RandomFees(txCtx.R, txCtx.Context, coins)
		if err != nil {
			return simtypes.NoOpMsg(txCtx.ModuleName, msgType, "unable to generate fees"), err
		}
	}

	tx, err := simtestutil.GenSignedMockTx(
		txCtx.R,
		txCtx.TxGen,
		[]sdk.Msg{txCtx.Msg},
		fees,
		simtestutil.DefaultGenTxGas,
		txCtx.Context.ChainID(),
		[]uint64{simAccount.GetAccountNumber()},
		[]uint64{simAccount.GetSequence()},
		txCtx.SimAccount.PrivKey,
	)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, msgType, "unable to generate mock tx"), err
	}

	_, _, err = txCtx.App.SimCheck(txCtx.TxGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, msgType, "unable to check tx"), err
	}

	return simtypes.NewOperationMsg(txCtx.Msg, true, ""), nil
}
