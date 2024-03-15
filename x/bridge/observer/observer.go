package observer

import (
	"context"
	"errors"

	errorsmod "cosmossdk.io/errors"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/pubsub/query"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
)

const ModuleNameObserver = "observer"

type Observer struct {
	logger        log.Logger
	cometRpc      *rpchttp.HTTP
	stopChan      chan struct{}
	eventsOutChan chan abcitypes.Event
}

// NewObserver returns new instance of `Observer` with RPC client created
func NewObserver(logger log.Logger, rpcUrl string) (Observer, error) {
	if len(rpcUrl) == 0 {
		return Observer{}, errors.New("RPC URL can't be empty")
	}

	rpc, err := rpchttp.New(rpcUrl, "/websocket")
	if err != nil {
		return Observer{}, errorsmod.Wrapf(err, "Failed to create RPC client")
	}

	return Observer{
		logger:        logger.With("module", ModuleNameObserver),
		cometRpc:      rpc,
		stopChan:      make(chan struct{}),
		eventsOutChan: make(chan abcitypes.Event),
	}, nil
}

// Start starts the RPC client, subscribes to events for provided query and starts listening to the events
func (o *Observer) Start(ctx context.Context, queryStr string, observeEvents []string) error {
	err := o.cometRpc.Start()
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to start RPC client")
	}

	query, err := query.New(queryStr)
	if err != nil {
		return errorsmod.Wrapf(err, "Invalid query")
	}

	txs, err := o.cometRpc.Subscribe(ctx, ModuleNameObserver, query.String())
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to subscribe to RPC client")
	}

	o.logger.Info("Observer starting listening for events at RPC", o.cometRpc.Remote())
	go o.processEvents(ctx, txs, observeEvents)

	return nil
}

// Stop stops listening to events, unsubscribes from the RPC client and stops the RPC channel
func (o *Observer) Stop(ctx context.Context) error {
	close(o.stopChan)
	if err := o.cometRpc.UnsubscribeAll(ctx, ModuleNameObserver); err != nil {
		return errorsmod.Wrapf(err, "Failed to unsubscribe from RPC client")
	}
	return o.cometRpc.Stop()
}

// Events returns receive-only part of the observed events
func (o *Observer) Events() <-chan abcitypes.Event {
	return o.eventsOutChan
}

func (o *Observer) processEvents(ctx context.Context, txs <-chan coretypes.ResultEvent, observeEvents []string) {
	defer close(o.eventsOutChan)
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
				results, err := o.cometRpc.BlockResults(ctx, &newBlock.Block.Height)
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
						select {
						case o.eventsOutChan <- e:
						case <-o.stopChan:
							o.logger.Info("Observer exiting early, event skipped: ", e)
							return
						}
					}
				}
			}
		}
	}
}
