package keeper

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/pubsub/query"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
)

type Observer struct {
	tmRpc         *rpchttp.HTTP
	observeEvents map[string]struct{}
	stopChan      chan struct{}
	eventsOutChan chan<- abcitypes.Event
}

func NewObserver(rpcUrl string, eventsOutChan chan<- abcitypes.Event) (Observer, error) {
	rpc, err := rpchttp.New(rpcUrl, "/websocket")
	if err != nil {
		return Observer{}, err
	}

	return Observer{
		tmRpc:         rpc,
		observeEvents: make(map[string]struct{}),
		stopChan:      make(chan struct{}),
		eventsOutChan: eventsOutChan,
	}, nil
}

func (o *Observer) Start(queryStr string, observeEvents []string) error {
	err := o.tmRpc.Start()
	if err != nil {
		return err
	}

	query, err := query.New(queryStr)
	if err != nil {
		return err
	}

	txs, err := o.tmRpc.Subscribe(context.Background(), "observer", query.String())
	if err != nil {
		return err
	}
	for _, e := range observeEvents {
		o.observeEvents[e] = struct{}{}
	}

	go o.processEvents(txs)

	return nil
}

func (o *Observer) Stop() error {
	if err := o.tmRpc.UnsubscribeAll(context.Background(), "observer"); err != nil {
		return err
	}
	return o.tmRpc.Stop()
}

func (o *Observer) processEvents(txs <-chan coretypes.ResultEvent) {
	for {
		select {
		case <-o.stopChan:
			return
		case event := <-txs:
			if newBlock, ok := event.Data.(comettypes.EventDataNewBlock); ok {
				results, err := o.tmRpc.BlockResults(context.Background(), &newBlock.Block.Height)
				if err != nil {
					continue
				}

				for _, r := range results.TxsResults {
					if r.IsErr() {
						continue
					}
					for _, e := range r.Events {
						if _, ok := o.observeEvents[e.Type]; !ok {
							continue
						}
						o.eventsOutChan <- e
					}
				}
			}
		}
	}
}
