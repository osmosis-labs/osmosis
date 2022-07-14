package simulation

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/osmosis-labs/osmosis/v7/app/params"

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
	authAcc := sim.App.GetAccountKeeper().GetAccount(ctx, account.Address)
	txConfig := params.MakeEncodingConfig().TxConfig // TODO: unhardcode
	// TODO: Consider making a default tx builder that charges some random fees
	// Low value for amount of work right now though.
	fees := sdk.Coins{}
	tx, err := genTx(
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

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		//nolint:forcetypeassert
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}
