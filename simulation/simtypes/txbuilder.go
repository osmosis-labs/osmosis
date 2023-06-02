package simtypes

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/osmosis-labs/osmosis/v16/app/params"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v16/x/tokenfactory/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
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
	authAcc := sim.AccountKeeper().GetAccount(ctx, account.Address)
	txConfig := params.MakeEncodingConfig().TxConfig // TODO: unhardcode
	// TODO: Consider making a default tx builder that charges some random fees
	// Low value for amount of work right now though.
	fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 25000))

	gas := getGas(msg)

	tx, err := genTx(
		txConfig,
		[]sdk.Msg{msg},
		fees,
		gas,
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
func (sim *SimCtx) deliverTx(tx sdk.Tx, msg sdk.Msg, msgName string) (simulation.OperationMsg, []simulation.FutureOperation, []byte, error) {
	txConfig := params.MakeEncodingConfig().TxConfig // TODO: unhardcode
	gasInfo, results, err := sim.BaseApp().Deliver(txConfig.TxEncoder(), tx)
	if err != nil {
		return simulation.NoOpMsg(msgName, msgName, fmt.Sprintf("unable to deliver tx. \nreason: %v\n results: %v\n msg: %s\n tx: %s", err, results, msg, tx)), nil, nil, err
	}

	opMsg := simulation.NewOperationMsg(msg, true, "", gasInfo.GasWanted, gasInfo.GasUsed, nil)
	opMsg.Route = msgName
	opMsg.Name = msgName

	return opMsg, nil, results.Data, nil
}

// GenTx generates a signed mock transaction.
// TODO: Surely theres proper API's in the SDK for this?
// (This was copied from SDK simapp, and deleted the egregiously non-deterministic memo handling)
func genTx(gen client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))
	memo := "sample_memo"
	signMode := gen.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	txBuilder := gen.NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmt)
	txBuilder.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sig, err := tx.SignWithPrivKey(signMode, signerData, txBuilder, p, gen, accSeqs[i])
		if err != nil {
			panic(err)
		}

		sigs[i] = sig
		err = txBuilder.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return txBuilder.GetTx(), nil
}

// special cases some messages that require higher gas limits
func getGas(msg sdk.Msg) uint64 {
	_, ok := msg.(*tokenfactorytypes.MsgCreateDenom)
	if ok {
		return uint64(tokenfactorytypes.DefaultCreationGasFee + helpers.DefaultGenTxGas)
	}
	return uint64(helpers.DefaultGenTxGas)
}
