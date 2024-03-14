package keeper

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/pubsub/query"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
)

type Observer struct {
	logger        log.Logger
	tmRpc         *rpchttp.HTTP
	stopChan      chan struct{}
	eventsOutChan chan<- abcitypes.Event
}

func NewObserver(logger log.Logger, rpcUrl string, eventsOutChan chan<- abcitypes.Event) (Observer, error) {
	rpc, err := rpchttp.New(rpcUrl, "/websocket")
	if err != nil {
		return Observer{}, err
	}

	return Observer{
		logger:        logger,
		tmRpc:         rpc,
		stopChan:      make(chan struct{}),
		eventsOutChan: eventsOutChan,
	}, nil
}

func (o *Observer) Start(ctx context.Context, queryStr string, observeEvents []string) error {
	err := o.tmRpc.Start()
	if err != nil {
		o.logger.Error("Observer failed to start RPC client", err.Error())
		return err
	}

	query, err := query.New(queryStr)
	if err != nil {
		o.logger.Error("Invalid query expression", err.Error())
		return err
	}

	txs, err := o.tmRpc.Subscribe(ctx, "observer", query.String())
	if err != nil {
		o.logger.Error("Observer failed to subscribe to RPC client", err.Error())
		return err
	}

	o.logger.Info("Observer starting listening for events at RPC", o.tmRpc.Remote())
	go o.processEvents(ctx, txs, observeEvents)

	return nil
}

func (o *Observer) Stop(ctx context.Context) error {
	if err := o.tmRpc.UnsubscribeAll(ctx, "observer"); err != nil {
		o.logger.Error("Observer failed to unsubscribe from RPC client", err.Error())
		return err
	}
	close(o.stopChan)
	return o.tmRpc.Stop()
}

func (o *Observer) processEvents(ctx context.Context, txs <-chan coretypes.ResultEvent, observeEvents []string) {
	events := make(map[string]struct{})
	for _, e := range observeEvents {
		events[e] = struct{}{}
	}

	for {
		select {
		case <-o.stopChan:
			return
		case event := <-txs:
			if newBlock, ok := event.Data.(comettypes.EventDataNewBlock); ok {
				results, err := o.tmRpc.BlockResults(ctx, &newBlock.Block.Height)
				if err != nil {
					o.logger.Error("Observer failed to fetch block results for block", newBlock.Block.Height)
					continue
				}

				for _, r := range results.TxsResults {
					if r.IsErr() {
						continue
					}
					for _, e := range r.Events {
						if _, ok := events[e.Type]; !ok {
							continue
						}
						o.eventsOutChan <- e
					}
				}
			}
		}
	}
}
