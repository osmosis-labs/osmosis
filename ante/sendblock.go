package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cast"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

type SendBlockOptions struct {
	PermittedOnlySendTo map[string]string
}

// NewSendBlockOptions returns options for sending new blocks.
func NewSendBlockOptions(appOpts servertypes.AppOptions) SendBlockOptions {
	return SendBlockOptions{
		PermittedOnlySendTo: parsePermittedOnlySendTo(appOpts),
	}
}

// parsePermittedOnlySendTo parses the mapping PermittedOnlySendTo.
func parsePermittedOnlySendTo(opts servertypes.AppOptions) map[string]string {
	valueInterface := opts.Get("permitted-only-send-to")
	if valueInterface == nil {
		return make(map[string]string)
	}
	return cast.ToStringMapString(valueInterface) // equal with viper.GetStringMapString
}

type SendBlockDecorator struct {
	Options SendBlockOptions
}

// NewSendBlockDecorator are a part of auth module AnteDecorators that are recursively chained together into a single AntiHandler.
func NewSendBlockDecorator(options SendBlockOptions) *SendBlockDecorator {
	return &SendBlockDecorator{
		Options: options, // TODO: hydrate from configuration
	}
}

// AnteHandle is used for performing basic validity checks on a transaction such that it can be thrown out of the mempool.
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
	if len(decorator.Options.PermittedOnlySendTo) == 0 {
		return nil
	}
	for _, msg := range msgs {
		signers := msg.GetSigners()
		for _, signer := range signers {
			if permittedTo, ok := decorator.Options.PermittedOnlySendTo[signer.String()]; ok {
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
