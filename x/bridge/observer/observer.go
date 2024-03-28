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
	outTxQueue map[ChainId][]Transfer
	outLock    sync.Mutex
	sendPeriod time.Duration
	stopChan   chan struct{}
}

// NewObserver returns new instance of `Observer`
func NewObserver(logger log.Logger, chains map[ChainId]Chain, sendPeriod time.Duration) Observer {
	return Observer{
		logger:     logger.With("module", ModuleName),
		chains:     chains,
		outTxQueue: make(map[ChainId][]Transfer),
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
func (o *Observer) Stop(ctx context.Context) error {
	for id, c := range o.chains {
		err := c.Stop(ctx)
		if err != nil {
			o.logger.Error(fmt.Sprintf("Failed to stop chain %s", id))
			continue
		}
	}
	close(o.stopChan)
	return nil
}

func (o *Observer) collectOutbound() {
	aggregate := make(chan Transfer)
	wg := sync.WaitGroup{}

	for _, chain := range o.chains {
		wg.Add(1)
		ch := chain.ListenOutboundTransfer()
		go func() {
			defer wg.Done()
			for t := range ch {
				aggregate <- t
			}
		}()
	}

	go func() {
		wg.Wait()
		close(aggregate)
	}()

	for out := range aggregate {
		o.outLock.Lock()
		_, ok := o.chains[out.DstChain]
		if !ok {
			o.logger.Error(fmt.Sprintf(
				"Unknown destination chain %s in outbound transfer %s",
				out.DstChain,
				out.Id,
			))
			o.outLock.Unlock()
			continue
		}
		o.outTxQueue[out.SrcChain] = append(o.outTxQueue[out.SrcChain], out)
		o.outLock.Unlock()
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
			o.logger.Error(fmt.Sprintf(
				"Failed to get confirmations required for %s: %s",
				srcId,
				err.Error(),
			))
			continue
		}
		var newQueue []Transfer
		for _, out := range queue {
			if height-out.Height < confirmationsRequired {
				newQueue = append(newQueue, out)
				continue
			}
			err = o.chains[out.DstChain].SignalInboundTransfer(ctx, out)
			if err != nil {
				newQueue = append(newQueue, out)
			}
		}
		o.outTxQueue[srcId] = newQueue
	}
}
