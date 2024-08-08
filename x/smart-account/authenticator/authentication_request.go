package authenticator

import (
	"fmt"

	txsigning "cosmossdk.io/x/tx/signing"

	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

//
// These structs define the data structure for authentication, used with AuthenticationRequest struct.
//

// SignModeData represents the signing modes with direct bytes and textual representation.
type SignModeData struct {
	Direct  []byte `json:"sign_mode_direct"`
	Textual string `json:"sign_mode_textual"`
}

// LocalAny holds a message with its type URL and byte value. This is necessary because the type Any fails
// to serialize and deserialize properly in nested contexts.
type LocalAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

// SimplifiedSignatureData contains lists of signers and their corresponding signatures.
type SimplifiedSignatureData struct {
	Signers    []sdk.AccAddress `json:"signers"`
	Signatures [][]byte         `json:"signatures"`
}

// ExplicitTxData encapsulates key transaction data like chain ID, account info, and messages.
type ExplicitTxData struct {
	ChainID         string     `json:"chain_id"`
	AccountNumber   uint64     `json:"account_number"`
	AccountSequence uint64     `json:"sequence"`
	TimeoutHeight   uint64     `json:"timeout_height"`
	Msgs            []LocalAny `json:"msgs"`
	Memo            string     `json:"memo"`
}

// GetSignerAndSignatures gets an array of signer and an array of signatures from the transaction
// checks they're the same length and returns both.
//
// A signer can only have one signature, so if it appears in multiple messages, the signatures must be
// the same, and it will only be returned once by this function. This is to mimic the way the classic
// sdk authentication works, and we will probably want to change this in the future
func GetSignerAndSignatures(tx sdk.Tx) (signers []sdk.AccAddress, signatures []signing.SignatureV2, err error) {
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
	signerBytes, err := sigTx.GetSigners()
	if err != nil {
		return nil, nil, err
	}

	for _, signer := range signerBytes {
		signers = append(signers, sdk.AccAddress(signer))
	}

	// check that signer length and signature length are the same
	if len(signatures) != len(signers) {
		return nil, nil,
			errorsmod.Wrap(sdkerrors.ErrTxDecode, fmt.Sprintf("invalid number of signer;  expected: %d, got %d", len(signers), len(signatures)))
	}

	return signers, signatures, nil
}

// getSignerData returns the signer data for a given account. This is part of the data that needs to be signed.
func getSignerData(ctx sdk.Context, ak authante.AccountKeeper, account sdk.AccAddress) authsigning.SignerData {
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

	return authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      sequence,
	}
}

// extractExplicitTxData makes the transaction data concrete for the authentication request. This is necessary to
// pass the parsed data to the cosmwasm authenticator.
func extractExplicitTxData(tx sdk.Tx, signerData authsigning.SignerData) (ExplicitTxData, error) {
	timeoutTx, ok := tx.(sdk.TxWithTimeoutHeight)
	if !ok {
		return ExplicitTxData{}, errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithTimeoutHeight")
	}
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return ExplicitTxData{}, errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithMemo")
	}

	// Encode messages as Anys and manually convert them to a struct we can serialize to json for cosmwasm.
	txMsgs := tx.GetMsgs()
	msgs := make([]LocalAny, len(txMsgs))
	for i, txMsg := range txMsgs {
		encodedMsg, err := types.NewAnyWithValue(txMsg)
		if err != nil {
			return ExplicitTxData{}, errorsmod.Wrap(err, "failed to encode msg")
		}
		msgs[i] = LocalAny{
			TypeURL: encodedMsg.TypeUrl,
			Value:   encodedMsg.Value,
		}
	}

	return ExplicitTxData{
		ChainID:         signerData.ChainID,
		AccountNumber:   signerData.AccountNumber,
		AccountSequence: signerData.Sequence,
		TimeoutHeight:   timeoutTx.GetTimeoutHeight(),
		Msgs:            msgs,
		Memo:            memoTx.GetMemo(),
	}, nil
}

// extractSignatures returns the signature data for each signature in the transaction and the one for the current signer.
//
// This function also checks for replay attacks. The replay protection needs to be able to match the signature to the
// corresponding signer, which involves iterating over the signatures. To avoid iterating over the signatures twice,
// we do replay protection here instead of in a separate replay protection function.
//
// Only SingleSignatureData is supported. Multisigs can be implemented by using partitioned compound authenticators
func extractSignatures(txSigners []sdk.AccAddress, txSignatures []signing.SignatureV2, txData ExplicitTxData, account sdk.AccAddress, replayProtection ReplayProtection) (signatures [][]byte, msgSignature []byte, err error) {
	for i, signature := range txSignatures {
		single, ok := signature.Data.(*signing.SingleSignatureData)
		if !ok {
			return nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast signature to SingleSignatureData")
		}

		signatures = append(signatures, single.Signature)

		if txSigners[i].Equals(account) {
			err = replayProtection(&txData, &signature)
			if err != nil {
				return nil, nil, err
			}
			msgSignature = single.Signature
		}
	}
	return signatures, msgSignature, nil
}

// GenerateAuthenticationRequest creates an AuthenticationRequest for the transaction.
func GenerateAuthenticationRequest(
	ctx sdk.Context,
	cdc codec.Codec,
	ak authante.AccountKeeper,
	sigModeHandler *txsigning.HandlerMap,
	account sdk.AccAddress,
	feePayer sdk.AccAddress,
	feeGranter sdk.AccAddress,
	fee sdk.Coins,
	msg sdk.Msg,
	tx sdk.Tx,
	msgIndex int,
	simulate bool,
	replayProtection ReplayProtection,
) (AuthenticationRequest, error) {
	// Only supporting one signer per message. This will be enforced in sdk v0.50
	signers, _, err := cdc.GetMsgV1Signers(msg)
	if err != nil {
		return AuthenticationRequest{}, err
	}
	signer := sdk.AccAddress(signers[0])
	if !signer.Equals(account) {
		return AuthenticationRequest{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid signer")
	}

	// Get the signers and signatures from the transaction. A signer can only have one signature, so if it
	// appears in multiple messages, the signatures must be the same, and it will only be returned once by
	// this function. This is to mimic the way the classic sdk authentication works, and we will probably want
	// to change this in the future
	txSigners, txSignatures, err := GetSignerAndSignatures(tx)
	if err != nil {
		return AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get signers and signatures")
	}

	// Get the signer data for the account. This is needed in the SignDoc
	signerData := getSignerData(ctx, ak, account)

	// Get the concrete transaction data to be passed to the authenticators
	txData, err := extractExplicitTxData(tx, signerData)
	if err != nil {
		return AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get explicit tx data")
	}

	// Get the signatures for the transaction and execute replay protection
	signatures, msgSignature, err := extractSignatures(txSigners, txSignatures, txData, account, replayProtection)
	if err != nil {
		return AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get signatures")
	}

	// Build the authentication request
	authRequest := AuthenticationRequest{
		Account:    account,
		FeePayer:   feePayer,
		FeeGranter: feeGranter,
		Fee:        fee,
		Msg:        txData.Msgs[msgIndex],
		MsgIndex:   uint64(msgIndex),
		Signature:  msgSignature,
		TxData:     txData,
		SignModeTxData: SignModeData{
			Direct: []byte("signBytes"),
		},
		SignatureData: SimplifiedSignatureData{
			Signers:    txSigners,
			Signatures: signatures,
		},
		Simulate:            simulate,
		AuthenticatorParams: nil,
	}

	// We do not generate the sign bytes if simulate is true
	if simulate {
		return authRequest, nil
	}

	// Get the sign bytes for the transaction
	signBytes, err := authsigning.GetSignBytesAdapter(ctx, sigModeHandler, signing.SignMode_SIGN_MODE_DIRECT, signerData, tx)
	if err != nil {
		return AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get signBytes")
	}

	// TODO: Add other sign modes. Specifically json when it becomes available
	authRequest.SignModeTxData = SignModeData{
		Direct: signBytes,
	}

	return authRequest, nil
}
