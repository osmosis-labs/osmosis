package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

/*
GetSignersAndSignatures retrieves the signers and their respective signatures for either a specific
message (identified by its index) or for all messages in the provided list. The function returns lists
of account addresses and their corresponding signatures.

Parameters:
  - msgs: A list of messages, each potentially having multiple signers.
  - suppliedSignatures: A list of signatures corresponding to the signers of the messages. Each signer
    has exactly one signature, and they appear in the order of the signers in the messages list.
  - feePayer: A string representing the account address of the fee payer. The fee payer is an optional
    additional signer whose signature might be present in the suppliedSignatures list.
  - msgIndex: An integer indicating the index of a specific message for which signers and signatures
    are to be retrieved. If set to -1, the function returns signers and signatures for all messages.

Assumptions:
  - Each signer in the messages list has a unique signature in the suppliedSignatures list. The order
    of signatures matches the order of appearance of the signers.
  - If the fee payer is provided and has not been seen among the signers in the messages, its signature
    is assumed to be the last signature on the list (after the signatures of the other signers from the
    messages).
  - The function assumes that any address conversion from string will succeed for addresses already
    present in the signerToSignature map, as they have been successfully converted before.

Returns:
  - A list of account addresses representing the signers.
  - A list of signatures corresponding to the returned signers.
  - An error, if any occurs during the processing (e.g., invalid fee payer address).

The primary use case for this function is to validate transactions by matching signers with their
signatures. It ensures that all required signers for a specific message or for all messages have
provided valid signatures.
*/
func GetSignersAndSignatures(
	msgs []sdk.Msg,
	suppliedSignatures []signing.SignatureV2,
	feePayer string,
	// we use the message index to get signers and signatures for
	// a specific message, with all messages.
	msgIndex int,
) ([]sdk.AccAddress, []signing.SignatureV2, error) {
	// Map to associate each signer with its signature.
	signerToSignature := make(map[string]signing.SignatureV2)
	sigIndex := 0
	specificMsg := msgIndex != -1
	var resultSigners []sdk.AccAddress

	// Iterate over messages and their signers.
	for i, msg := range msgs {
		for _, signer := range msg.GetSigners() {
			signerStr := signer.String()
			if _, exists := signerToSignature[signerStr]; !exists {
				// sanity check for runtime error: index out of range with
				// the sigIndex can be more that the supplied signatures
				if sigIndex >= len(suppliedSignatures) {
					return nil, nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not enough signatures provided")
				}
				signerToSignature[signerStr] = suppliedSignatures[sigIndex]
				sigIndex++
			} else {
				// ensure that number of signers and signatures are the same
				// this also ensures that each message has a signer
				// NOTE: This is the same behavior that currently exists in the node
				return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signers")
			}

			// If dealing with a specific message, capture its signers.
			if specificMsg && i == msgIndex {
				resultSigners = append(resultSigners, signer)
			}
		}
	}

	// If no specific message is given, get all signers from the map.
	if !specificMsg {
		for signer := range signerToSignature {
			addr, err := sdk.AccAddressFromBech32(signer)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid signer address")
			}
			resultSigners = append(resultSigners, addr)
		}
	}

	// Handle the feePayer.
	if feePayer != "" {
		if _, exists := signerToSignature[feePayer]; !exists {
			feePayerAddr, err := sdk.AccAddressFromBech32(feePayer)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid fee payer address")
			}
			resultSigners = append(resultSigners, feePayerAddr)
			// sanity check for runtime error: index out of range with
			// the sigIndex can be more that the supplied signatures
			if sigIndex >= len(suppliedSignatures) {
				return nil, nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not enough signatures provided for fee payer")
			}
			signerToSignature[feePayer] = suppliedSignatures[sigIndex]
		}
		// TODO: Consider always returning the fee payer separately
	}

	// Construct the result signatures list based on the result signers.
	var resultSignatures []signing.SignatureV2
	for _, signer := range resultSigners {
		resultSignatures = append(resultSignatures, signerToSignature[signer.String()])
	}

	return resultSigners, resultSignatures, nil
}

// GetCommonAuthenticationData retrieves common authentication data from a transaction for Cosmos SDK.
// It extracts signers, signatures, and other necessary information from the provided transaction.
// It is used in both the PassKeyAuthenticator and the SignatureVerificationAuthenticator
//
// Parameters:
// - ctx: The context of the current operation.
// - tx: The transaction to extract authentication data from.
// - messageIndex: The index of the message within the transaction.
// - simulate: A boolean indicating whether to simulate the transaction.
//
// Returns:
// - signers: A list of account addresses that signed the transaction.
// - signatures: A list of signature objects.
// - sigTx: The transaction with signature information.
// - err: An error if any issues are encountered during the extraction.
func GetCommonAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int8,
	simulate bool,
) (signers []sdk.AccAddress, signatures []signing.SignatureV2, sigTx authsigning.Tx, err error) {
	// Attempt to cast the provided transaction to an authsigning.Tx.
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return nil, nil, nil,
			sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// Retrieve signatures from the transaction.
	signatures, err = sigTx.GetSignaturesV2()
	if err != nil {
		return nil, nil, nil, err
	}

	// Retrieve messages from the transaction.
	msgs := sigTx.GetMsgs()

	// Ensure the transaction is of type sdk.FeeTx.
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, nil, nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Parse signers and signatures from the transaction.
	signers, signatures, err = GetSignersAndSignatures(
		msgs,
		signatures,
		feeTx.FeePayer().String(),
		int(messageIndex),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	// Return the extracted data.
	return signers, signatures, sigTx, nil
}
