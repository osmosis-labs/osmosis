package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Verify all signatures for a tx and return an error if any are invalid. Note,
// the AuthenticatorDecorator will not check signatures on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type AuthenticatorDecorator struct {
	ak              authante.AccountKeeper
	signModeHandler authsigning.SignModeHandler
}

func NewAuthenticatorDecorator(
	ak authante.AccountKeeper,
	signModeHandler authsigning.SignModeHandler,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		ak:              ak,
		signModeHandler: signModeHandler,
		//TODO: Add authenticator struct here for later user
	}
}

// AnteHandle is the authenticator decorator ante handler
// this is used to validate multiple signatures
func (ad AuthenticatorDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	ad.ValidateSignature(ctx, tx, simulate)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// ValidateSignature validates the txn signature using Secp256k1 keys
func (ad AuthenticatorDecorator) ValidateSignature(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
) (newCtx sdk.Context, err error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}

	signerAddrs := sigTx.GetSigners()

	// check that signer length and signature length are the same
	if len(sigs) != len(signerAddrs) {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"invalid number of signer;  expected: %d, got %d",
			len(signerAddrs),
			len(sigs))
	}

	for i, sig := range sigs {
		acc, err := authante.GetSignerAcc(ctx, ad.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}

		// retrieve pubkey
		pubKey := acc.GetPubKey()
		if !simulate && pubKey == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		// Check account sequence number.
		if sig.Sequence != acc.GetSequence() {
			return ctx, sdkerrors.Wrapf(
				sdkerrors.ErrWrongSequence,
				"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
			)
		}

		// retrieve signer data
		genesis := ctx.IsGenesis() || ctx.BlockHeight() == 0
		chainID := ctx.ChainID()
		var accNum uint64
		if !genesis {
			accNum = acc.GetAccountNumber()
		}
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNum,
			Sequence:      acc.GetSequence(),
		}

		// no need to verify signatures on recheck tx
		if !simulate && !ctx.IsReCheckTx() {
			err := authsigning.VerifySignature(pubKey, signerData, sig.Data, ad.signModeHandler, tx)
			if err != nil {
				var errMsg string
				if authante.OnlyLegacyAminoSigners(sig.Data) {
					// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
					// and therefore communicate sequence number as a potential cause of error.
					errMsg = fmt.Sprintf(
						"signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)",
						accNum,
						acc.GetSequence(),
						chainID,
					)
				} else {
					errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s)",
						accNum,
						chainID,
					)
				}
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, errMsg)

			}
		}
	}
	return ctx, nil
}
