package authenticator

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func GetSigners(
	msgs []sdk.Msg,
) []sdk.AccAddress {
	var signers []sdk.AccAddress

	// Iterate over messages and their signers.
	for _, msg := range msgs {
		for _, signer := range msg.GetSigners() {
			signers = append(signers, signer)
		}
	}

	return signers
}

func GetCommonAuthenticationData(
	tx sdk.Tx,
) (signers []sdk.AccAddress, signatures []signing.SignatureV2, err error) {
	// Attempt to cast the provided transaction to an authsigning.Tx.
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return nil, nil,
			errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// Retrieve signatures from the transaction.
	signatures, err = sigTx.GetSignaturesV2()
	if err != nil {
		return nil, nil, err
	}

	// Retrieve messages from the transaction.
	msgs := sigTx.GetMsgs()
	signers = GetSigners(msgs)

	return signers, signatures, nil
}

func GenerateAuthenticationData(
	ctx sdk.Context,
	ak *keeper.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,
	account sdk.AccAddress,
	msg sdk.Msg,
	tx sdk.Tx,
	msgIndex int,
	simulate bool,
	replayProtection ReplayProtection,
) (iface.AuthenticationRequest, error) {
	txSigners, txSignatures, err := GetCommonAuthenticationData(tx)
	if err != nil {
		return iface.AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get signs and signatures")
	}

	// Retrieve and build the signer data struct
	baseAccount := ak.GetAccount(ctx, account)
	genesis := ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	if !genesis {
		accNum = baseAccount.GetAccountNumber()
	}
	var sequence uint64
	if baseAccount != nil {
		sequence = baseAccount.GetSequence()
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      sequence,
	}

	// This can also be extracted
	signBytes, err := sigModeHandler.GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, tx)
	if err != nil {
		return iface.AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get signBytes")
	}

	timeoutTx, ok := tx.(sdk.TxWithTimeoutHeight)
	if !ok {
		return iface.AuthenticationRequest{}, errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithTimeoutHeight")
	}
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return iface.AuthenticationRequest{}, errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithMemo")
	}

	txMsgs := tx.GetMsgs()
	msgs := make([]iface.LocalAny, len(txMsgs))
	for i, txMsg := range txMsgs {
		encodedMsg, err := types.NewAnyWithValue(txMsg)
		if err != nil {
			return iface.AuthenticationRequest{}, errorsmod.Wrap(err, "failed to encode msg")
		}
		msgs[i] = iface.LocalAny{
			TypeURL: encodedMsg.TypeUrl,
			Value:   encodedMsg.Value,
		}
	}

	txData := iface.ExplicitTxData{
		ChainID:         chainID,
		AccountNumber:   accNum,
		AccountSequence: sequence,
		TimeoutHeight:   timeoutTx.GetTimeoutHeight(),
		Msgs:            msgs,
		Memo:            memoTx.GetMemo(),
	}

	// Only supporting one signer per message.
	signer := msg.GetSigners()[0]
	var signatures [][]byte
	var msgSignature []byte
	for i, signature := range txSignatures {
		err := replayProtection(&txData, &signature)
		if err != nil {
			return iface.AuthenticationRequest{}, err
		}

		single, ok := signature.Data.(*signing.SingleSignatureData)
		if !ok {
			return iface.AuthenticationRequest{},
				errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast signature to SingleSignatureData")
		}

		signatures = append(signatures, single.Signature)

		if txSigners[i].Equals(signer) {
			msgSignature = single.Signature
		}
	}

	authRequest := iface.AuthenticationRequest{
		Account:   account,
		Msg:       txData.Msgs[msgIndex],
		MsgIndex:  uint64(msgIndex),
		Signature: msgSignature,
		TxData:    txData,
		SignModeTxData: iface.SignModeData{ // TODO: Add other sign modes. Specifically textual when it becomes available
			Direct: signBytes,
		},
		SignatureData: iface.SimplifiedSignatureData{
			Signers:    txSigners,
			Signatures: signatures,
		},
		Simulate:            simulate,
		AuthenticatorParams: nil,
	}

	return authRequest, nil
}
