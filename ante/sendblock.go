package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type SendBlockDecorator struct {
	// permittedOnlySendTo stores the only permitted send destination address for a blocked address.
	permittedOnlySendTo map[string]string // XXX: temporary. change to proper configuration
}

func NewSendBlockDecorator(permittedOnlySendTo map[string]string) *SendBlockDecorator {
	return &SendBlockDecorator{
		permittedOnlySendTo: permittedOnlySendTo, // TODO: hydrate from configuration
	}
}

func (decorator *SendBlockDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	if ctx.IsCheckTx() && !simulate {
		if err := decorator.CheckIfBlocked(tx.GetMsgs()); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// CheckIfBlocked returns error if following are true:
// 1. decorator.permittedOnlySendTo has msg.GetSigners() has its key, and
// 2-1. msg is not a SendMsg, or
// 2-2. msg is SendMsg and the destination is not decorator.permittedOnlySendTo[msg.Sender]
func (decorator *SendBlockDecorator) CheckIfBlocked(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		signers := msg.GetSigners()
		for _, signer := range signers {
			if permittedTo, ok := decorator.permittedOnlySendTo[signer.String()]; ok {
				sendmsg, ok := msg.(*bank.MsgSend)
				if !ok {
					return fmt.Errorf("signer is not allowed to send transactions: %s", signer)
				}
				if sendmsg.ToAddress != permittedTo {
					return fmt.Errorf("signer is not allowed to send tokens: %s", signer)
				}
			}
		}
	}
	return nil
}
