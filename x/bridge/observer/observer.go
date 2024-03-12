package keeper

import (
	"context"

	"github.com/cometbft/cometbft/libs/pubsub/query"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
)

type Observer struct {
	tmRpc      *rpchttp.HTTP
	eventsChan <-chan coretypes.ResultEvent
}

func NewObesrver(rpcUrl string) (Observer, error) {
	rpc, err := rpchttp.New(rpcUrl, "/websocket")
	if err != nil {
		return Observer{}, err
	}

	return Observer{
		tmRpc: rpc,
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
	o.eventsChan = txs

	return nil
}

func (o *Observer) Stop() error {
	return o.tmRpc.Stop()
}

func (o *Observer) GetEvents() <-chan coretypes.ResultEvent {
	return o.eventsChan
}
