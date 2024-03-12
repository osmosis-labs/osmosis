package observer

import (
	"context"
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/libs/pubsub/query"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type Observer struct {
	tmRpc      *rpchttp.HTTP
	cancelChan chan struct{}
	wg         *sync.WaitGroup
}

func NewObesrver(rpcUrl string) (Observer, error) {
	rpc, err := rpchttp.New(rpcUrl, "/websocket")
	if err != nil {
		return Observer{}, err
	}

	return Observer{
		tmRpc:      rpc,
		cancelChan: make(chan struct{}),
		wg:         &sync.WaitGroup{},
	}, nil
}

func (o *Observer) Start(queryStr string) error {
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

	o.wg.Add(1)
	go o.observeEvents(txs)

	return nil
}

func (o *Observer) Stop() error {
	close(o.cancelChan)
	o.wg.Wait()
	return o.tmRpc.Stop()
}

func (o *Observer) observeEvents(txs <-chan coretypes.ResultEvent) {
	defer o.wg.Done()
	for {
		select {
		case <-o.cancelChan:
			return
		case event := <-txs:
			e, ok := event.Data.(types.EventOutboundTransfer)
			if ok {
				fmt.Println("Got OutboundTransfer: ", e)
			} else {
				fmt.Println("Got unknown event", event)
			}
		}
	}
}
