package observer

import (
	"context"
	"fmt"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/types"
)

type Observer struct {
	tmRpc *rpchttp.HTTP
}

func NewObesrver() Observer {
	rpc, err := rpchttp.New("https://rpc.testnet.osmosis.zone:26657", "/websocket") // local node
	fmt.Println(err)
	if err != nil {
		panic("Tendermint RPC client failed to create")
	}

	return Observer{
		tmRpc: rpc,
	}
}

func (o *Observer) Start() {
	err := o.tmRpc.Start()
	fmt.Println("Start err: ", err)
	if err != nil {
		panic(err)
	}
	defer o.tmRpc.Stop()

	fmt.Println("prectx")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	fmt.Println("ctx: ", ctx)
	defer cancel()

	health, err := o.tmRpc.Health(ctx)
	fmt.Println("health: ", health)
	fmt.Println("err: ", err)

	query := "tm.event='Tx'" // test event
	txs, err := o.tmRpc.Subscribe(ctx, "observer", query)
	if err != nil {
		panic("Tendermint RPC client failed to subscribe")
	}

	go func() {
		for e := range txs {
			fmt.Println("got", e.Data.(types.EventDataTx))
		}
	}()
}
