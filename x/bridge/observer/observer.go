package observer

import (
	"context"
	"fmt"
	"sync"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
)

const ModuleName = "observer"

type Observer struct {
	logger     log.Logger
	chains     map[ChainId]Chain
	outTxQueue map[ChainId][]OutboundTransfer
	outLock    sync.Mutex
	sendPeriod time.Duration
	stopChan   chan struct{}
}

// NewObserver returns new instance of `Observer`
func NewObserver(logger log.Logger, chains map[ChainId]Chain, sendPeriod time.Duration) Observer {
	return Observer{
		logger:     logger.With("module", ModuleName),
		chains:     chains,
		outTxQueue: make(map[ChainId][]OutboundTransfer),
		outLock:    sync.Mutex{},
		sendPeriod: sendPeriod,
		stopChan:   make(chan struct{}),
	}
}

// Start starts all underlying chains and starts processing transfers
func (o *Observer) Start(ctx context.Context) error {
	for id, c := range o.chains {
		err := c.Start(ctx)
		if err != nil {
			return errorsmod.Wrapf(err, "Failed to start chain %s", id)
		}
	}

	go o.collectOutbound()
	go o.processOutbound(ctx)

	return nil
}

// Stop stops all underlying chains and stops processing transfers
func (o *Observer) Stop() error {
	close(o.stopChan)
	for id, c := range o.chains {
		err := c.Stop()
		if err != nil {
			return errorsmod.Wrapf(err, "Failed to stop chain %s", id)
		}
	}
	return nil
}

func (o *Observer) collectOutbound() {
	aggregate := make(chan struct {
		ChainId
		OutboundTransfer
	})
	for id, chain := range o.chains {
		go func(id ChainId, ch <-chan OutboundTransfer) {
			for t := range ch {
				select {
				case aggregate <- struct {
					ChainId
					OutboundTransfer
				}{id, t}:
				case <-o.stopChan:
					return
				}
			}
		}(id, chain.ListenOutboundTransfer())
	}

	for {
		select {
		case <-o.stopChan:
			return
		case out := <-aggregate:
			dstChain := o.chains[out.OutboundTransfer.DstChain]
			if dstChain == nil {
				o.logger.Error(fmt.Sprintf("Unknown destination chain %s in outbound transfer %s", out.OutboundTransfer.DstChain, out.Id))
			}
			o.outLock.Lock()
			o.outTxQueue[out.ChainId] = append(o.outTxQueue[out.ChainId], out.OutboundTransfer)
			o.outLock.Unlock()
		}
	}
}

func (o *Observer) processOutbound(ctx context.Context) {
	for {
		select {
		case <-o.stopChan:
			o.sendOutbound(ctx)
			return
		case <-time.After(o.sendPeriod):
			o.sendOutbound(ctx)
		}
	}
}

func (o *Observer) sendOutbound(ctx context.Context) {
	o.outLock.Lock()
	defer o.outLock.Unlock()

	for srcId, queue := range o.outTxQueue {
		srcChain := o.chains[srcId]
		height, err := srcChain.Height()
		if err != nil {
			o.logger.Error(fmt.Sprintf("Failed to get height for %s: %s", srcId, err.Error()))
			continue
		}
		confirmationsRequired, err := srcChain.ConfirmationsRequired()
		if err != nil {
			o.logger.Error(fmt.Sprintf("Failed to get confirmations required for %s: %s", srcId, err.Error()))
		}
		newQueue := []OutboundTransfer{}
		for _, out := range queue {
			if height-out.Height < confirmationsRequired {
				newQueue = append(newQueue, out)
			} else {
				in := InboundTransfer{
					SrcChain: srcId,
					Id:       out.Id,
					Height:   out.Height,
					Sender:   out.Sender,
					To:       out.To,
					Asset:    out.Asset,
					Amount:   out.Amount,
				}
				err = o.chains[out.DstChain].SignalInboundTransfer(ctx, in)
				if err != nil {
					newQueue = append(newQueue, out)
				}
			}
		}
		o.outTxQueue[srcId] = newQueue
	}
}
