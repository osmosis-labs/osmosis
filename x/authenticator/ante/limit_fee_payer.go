package ante

import (
	"bytes"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// LimitFeePayerDecorator enforces that the tx fee payer has not been set manually
// to an account different to the signer of the first message. This is a requirement
// for the authenticator module.
// The only user of a manually set fee payer is with fee grants, which are not
// available on osmosis
type LimitFeePayerDecorator struct{}

// AnteHandle performs an AnteHandler check that returns an error if the tx has
// fee payer set that is not the first signer of the first message
func (mfd LimitFeePayerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// The fee payer by default is the first signer of the transaction
	feePayer := feeTx.FeePayer()

	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must contain at least one message")
	}
	signers := msgs[0].GetSigners()
	if len(signers) == 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx message must contain at least one signer")
	}

	if !bytes.Equal(feePayer, signers[0]) {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "fee payer must be the first signer")
	}

	return next(ctx, tx, simulate)
}
