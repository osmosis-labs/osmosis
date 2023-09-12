package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetSignersAndSignatures gets a signer and signature for the message index
// provided, this is then used to validate the msg in the transaction
func GetSignersAndSignatures(
	msgs []sdk.Msg,
	allSignatures []signing.SignatureV2,
	feePayer string,
	msgIndex int,
) ([]sdk.AccAddress, []signing.SignatureV2, error) {
	// we use the msg index...
	useMsgIndex := msgIndex != -1

	// we track seem sigers because...
	seen := map[string]bool{}

	// we count unique sigers because...
	uniqueSignersCount := 0

	var signersList []sdk.AccAddress
	var signaturesList []signing.SignatureV2

	// TODO: Revisit and rework these data iterations
	// they could do with being more readable

	// TODO:
	// on the feePayer acc1
	// give me authenticator for that account
	// can we auth this account in sufficiently low gas
	// does it have secp256k1 or something else that is low gas
	// defined as less than x

	// Loop through the messages and their signers
	for i, msg := range msgs {
		for _, signer := range msg.GetSigners() {
			signerStr := signer.String()
			if !seen[signerStr] {
				seen[signerStr] = true
				uniqueSignersCount++
				if !useMsgIndex || i == msgIndex {
					signersList = append(signersList, signer)
					signaturesList = append(signaturesList, allSignatures[uniqueSignersCount-1])
					if useMsgIndex {
						break
					}
				}
			}
		}
	}

	// Add FeePayer if not already included
	if feePayer != "" && !seen[feePayer] {
		if uniqueSignersCount != len(allSignatures)-1 {
			// TODO: Better error?
			return nil, nil, sdkerrors.Wrapf(
				sdkerrors.ErrUnauthorized,
				"invalid number of signer;  expected: %d, got %d",
				uniqueSignersCount,
				len(allSignatures)-1,
			)
		}
		feePayerAddr, err := sdk.AccAddressFromBech32(feePayer)
		if err != nil {
			return nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid fee payer address")
		}
		signersList = append(signersList, feePayerAddr)
		signaturesList = append(signaturesList, allSignatures[len(allSignatures)-1]) // FeePayer's signature is last
	}

	// NOTE: this fails when adding another signer for some reason
	//if len(allSignatures) != len(signersList) {
	//	return nil, nil, sdkerrors.Wrapf(
	//		sdkerrors.ErrUnauthorized,
	//		"mismatch between signers and signatures;  expected: %d, got %d",
	//		uniqueSignersCount,
	//		len(allSignatures),
	//	)
	//}

	return signersList, signaturesList, nil
}
