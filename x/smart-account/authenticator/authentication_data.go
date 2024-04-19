package authenticator

import (
	"fmt"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	errorsmod "cosmossdk.io/errors"
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

// LocalAny holds a message with its type URL and byte value.
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

// GetSignersAndSignatures gets an array of signer and an array of signatures from the transaction
// checks their the same length and returns both
func GetSignerAndSignatures(
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
	signers = sigTx.GetSigners()

	// check that signer length and signature length are the same
	if len(signatures) != len(signers) {
		return nil, nil,
			errorsmod.Wrap(sdkerrors.ErrTxDecode, fmt.Sprintf("invalid number of signer;  expected: %d, got %d", len(signers), len(signatures)))
	}

	return signers, signatures, nil
}

func GenerateAuthenticationRequest(
	ctx sdk.Context,
	ak authante.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,
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
	txSigners, txSignatures, err := GetSignerAndSignatures(tx)
	if err != nil {
		return AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get signers and signatures")
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
		return AuthenticationRequest{}, errorsmod.Wrap(err, "failed to get signBytes")
	}

	timeoutTx, ok := tx.(sdk.TxWithTimeoutHeight)
	if !ok {
		return AuthenticationRequest{}, errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithTimeoutHeight")
	}
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return AuthenticationRequest{}, errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast tx to TxWithMemo")
	}

	txMsgs := tx.GetMsgs()
	msgs := make([]LocalAny, len(txMsgs))
	for i, txMsg := range txMsgs {
		encodedMsg, err := types.NewAnyWithValue(txMsg)
		if err != nil {
			return AuthenticationRequest{}, errorsmod.Wrap(err, "failed to encode msg")
		}
		msgs[i] = LocalAny{
			TypeURL: encodedMsg.TypeUrl,
			Value:   encodedMsg.Value,
		}
	}

	txData := ExplicitTxData{
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
			return AuthenticationRequest{}, err
		}

		single, ok := signature.Data.(*signing.SingleSignatureData)
		if !ok {
			return AuthenticationRequest{},
				errorsmod.Wrap(sdkerrors.ErrInvalidType, "failed to cast signature to SingleSignatureData")
		}

		signatures = append(signatures, single.Signature)

		if txSigners[i].Equals(signer) {
			msgSignature = single.Signature
		}
	}

	authRequest := AuthenticationRequest{
		Account:    account,
		FeePayer:   feePayer,
		FeeGranter: feeGranter,
		Fee:        fee,
		Msg:        txData.Msgs[msgIndex],
		MsgIndex:   uint64(msgIndex),
		Signature:  msgSignature,
		TxData:     txData,
		SignModeTxData: SignModeData{ // TODO: Add other sign modes. Specifically textual when it becomes available
			Direct: signBytes,
		},
		SignatureData: SimplifiedSignatureData{
			Signers:    txSigners,
			Signatures: signatures,
		},
		Simulate:            simulate,
		AuthenticatorParams: nil,
	}

	return authRequest, nil
}
