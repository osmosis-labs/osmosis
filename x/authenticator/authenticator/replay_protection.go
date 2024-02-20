package authenticator

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// make replay protection into an interface. SequenceMatch is a default implementation
type ReplayProtection func(txData *iface.ExplicitTxData, signature *signing.SignatureV2) error

func SequenceMatch(txData *iface.ExplicitTxData, signature *signing.SignatureV2) error {
	if signature.Sequence != txData.Sequence {
		return errorsmod.Wrap(sdkerrors.ErrInvalidSequence, fmt.Sprintf("account sequence mismatch, expected %d, got %d", txData.Sequence, signature.Sequence))
	}
	return nil
}

func NoReplayProtection(txData *iface.ExplicitTxData, signature *signing.SignatureV2) error {
	return nil
}
