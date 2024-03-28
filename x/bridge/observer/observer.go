package observer

import (
	"context"
	"fmt"
	"sync"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"

	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

const ModuleName = "observer"

type TxQueueItem struct {
	Tx                    Transfer
	ConfirmationsRequired uint64
}

type Observer struct {
	logger     log.Logger
	chains     map[ChainId]Client
	outTxQueue map[ChainId][]TxQueueItem
	outLock    sync.Mutex
	sendPeriod time.Duration
	stopChan   chan struct{}
}

// NewObserver returns new instance of `Observer`
func NewObserver(
	logger log.Logger,
	chains map[ChainId]Client,
	sendPeriod time.Duration,
) Observer {
	return Observer{
		logger:     logger.With("module", ModuleName),
		chains:     chains,
		outTxQueue: make(map[ChainId][]TxQueueItem),
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

	go o.collectOutbound(ctx)
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

func (o *Observer) collectOutbound(ctx context.Context) {
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
		o.addTxToQueue(ctx, out)
	}
}

func (o *Observer) addTxToQueue(ctx context.Context, tx Transfer) {
	o.outLock.Lock()
	defer o.outLock.Unlock()

	_, ok := o.chains[tx.DstChain]
	if !ok {
		o.logger.Error(fmt.Sprintf(
			"Unknown destination chain %s in outbound transfer %s",
			tx.DstChain,
			tx.Id,
		))
		return
	}
	cr, err := o.getConfirmationsRequired(ctx, tx)
	if err != nil {
		o.logger.Error(fmt.Sprintf(
			"Failed to get confirmations required for outbound transfer %s",
			tx.Id,
		))
		return
	}
	o.outTxQueue[tx.SrcChain] = append(o.outTxQueue[tx.SrcChain], TxQueueItem{tx, cr})
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
		height, err := srcChain.Height(ctx)
		if err != nil {
			o.logger.Error(fmt.Sprintf("Failed to get height for %s: %s", srcId, err.Error()))
			continue
		}

		var newQueue []TxQueueItem
		for _, out := range queue {
			confirmed, err := o.isTxConfirmed(ctx, height, &out)
			if !confirmed || err != nil {
				if err != nil {
					o.logger.Error(fmt.Sprintf("Failed to confirm tx %s", err.Error()))
				}
				newQueue = append(newQueue, out)
				continue
			}
			err = o.chains[out.Tx.DstChain].SignalInboundTransfer(ctx, out.Tx)
			if err != nil {
				newQueue = append(newQueue, out)
			}
		}
		o.outTxQueue[srcId] = newQueue
	}
}

func (o *Observer) getConfirmationsRequired(ctx context.Context, tx Transfer) (uint64, error) {
	if tx.SrcChain == ChainIdOsmosis {
		return 0, nil
	}

	chain, ok := o.chains[ChainIdOsmosis]
	if !ok {
		return 0, fmt.Errorf("Chain client for %s not found", ChainIdOsmosis)
	}
	cr, err := chain.ConfirmationsRequired(ctx, bridgetypes.AssetID{
		SourceChain: string(tx.SrcChain),
		Denom:       tx.Asset,
	})
	if err != nil {
		return 0, err
	}
	return cr, nil
}

func (o *Observer) isTxConfirmed(
	ctx context.Context,
	curHeight uint64,
	item *TxQueueItem,
) (bool, error) {
	if curHeight < item.ConfirmationsRequired+item.Tx.Height {
		return false, nil
	}
	cr, err := o.getConfirmationsRequired(ctx, item.Tx)
	if err != nil {
		return false, errorsmod.Wrapf(
			err,
			"Failed to get confirmations required for outbound transfer %s",
			item.Tx.Id,
		)
	}
	if curHeight < cr+item.Tx.Height {
		item.ConfirmationsRequired = cr
		return false, nil
	}
	return true, nil
}
