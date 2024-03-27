package osmosis

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	cosmosproto "github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/observer"
	bridgetypes "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

var (
	ModuleName    = "osmosis-chain"
	OsmoGasLimit  = uint64(200000)
	OsmoFeeAmount = sdktypes.NewIntFromUint64(1000)
	OsmoFeeDenom  = "uosmo"
)

type Osmosis struct {
	logger       log.Logger
	osmoClient   *Client
	cometRpc     *rpchttp.HTTP
	chains       map[observer.ChainId]observer.Chain
	stopChan     chan struct{}
	outboundChan chan observer.OutboundTransfer
}

// NewOsmosis returns new instance of `Osmosis`
func NewOsmosis(
	logger log.Logger,
	osmoClient *Client,
	cometRpc *rpchttp.HTTP,
	chains map[observer.ChainId]observer.Chain,
) *Osmosis {
	return &Osmosis{
		logger:       logger.With("module", ModuleName),
		osmoClient:   osmoClient,
		cometRpc:     cometRpc,
		chains:       chains,
		stopChan:     make(chan struct{}),
		outboundChan: make(chan observer.OutboundTransfer),
	}
}

// Start subscribes to the `NewBlock` events and starts listening to `EventOutboundTransfer` events
func (o *Osmosis) Start(ctx context.Context) error {
	err := o.cometRpc.Start()
	if err != nil {
		return errorsmod.Wrapf(ErrRpcClient, err.Error())
	}

	txs, err := o.cometRpc.Subscribe(ctx, ModuleName, comettypes.EventQueryNewBlock.String())
	if err != nil {
		return errorsmod.Wrapf(ErrRpcClient, err.Error())
	}

	go o.observeEvents(ctx, txs)

	return nil
}

// Stop stops listening to events and closes Osmosis client
func (o *Osmosis) Stop(ctx context.Context) error {
	close(o.stopChan)
	o.osmoClient.Close()
	if err := o.cometRpc.UnsubscribeAll(ctx, ModuleName); err != nil {
		return errorsmod.Wrapf(err, "Failed to unsubscribe from RPC client")
	}
	return o.cometRpc.Stop()
}

// ListenOutboundTransfer returns receive-only channel with `OutboundTransfer` items
func (o *Osmosis) ListenOutboundTransfer() <-chan observer.OutboundTransfer {
	return o.outboundChan
}

// SignalInboundTransfer sends `InboundTransfer` to Osmosis
func (o *Osmosis) SignalInboundTransfer(ctx context.Context, in observer.InboundTransfer) error {
	msg := bridgetypes.NewMsgInboundTransfer(
		in.Id,
		in.Sender,
		in.To,
		bridgetypes.AssetID{
			SourceChain: string(in.SrcChain),
			Denom:       in.Asset,
		},
		math.Int(in.Amount),
	)
	fees := sdktypes.NewCoins(sdktypes.NewCoin(OsmoFeeDenom, OsmoFeeAmount))
	bytes, err := o.osmoClient.SignTx(ctx, msg, fees, OsmoGasLimit)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to sign tx for inbound transfer %s", in.Id)
	}
	_, err = o.osmoClient.BroadcastTx(ctx, bytes)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to broadcast tx to Osmosis for inbound transfer %s", in.Id)
	}
	return nil
}

func (o *Osmosis) observeEvents(
	ctx context.Context,
	txs <-chan coretypes.ResultEvent,
) {
	defer close(o.outboundChan)

	eventToObserve := cosmosproto.MessageName(&bridgetypes.EventOutboundTransfer{})
	for {
		select {
		case <-o.stopChan:
			return
		case event := <-txs:
			newBlock, ok := event.Data.(comettypes.EventDataNewBlock)
			if !ok {
				continue
			}

			results, err := o.cometRpc.TxSearch(
				ctx,
				fmt.Sprintf("tx.height=%d", newBlock.Block.Height),
				false,
				nil,
				nil,
				"",
			)
			if err != nil {
				o.logger.Error("Failed to fetch Txs at height ", newBlock.Block.Height)
				continue
			}

			for _, r := range results.Txs {
				if r.TxResult.IsErr() {
					continue
				}
				for _, e := range r.TxResult.Events {
					if e.Type != eventToObserve {
						continue
					}

					out, err := outboundTransferFromEvent(uint64(r.Height), r.Hash.String(), e)
					if err != nil {
						continue
					}
					select {
					case o.outboundChan <- out:
					case <-o.stopChan:
						o.logger.Info("Exiting early, event %s skipped in tx %s", e.String(), r.Hash.String())
						return
					}
				}
			}
		}
	}
}

func outboundTransferFromEvent(height uint64, hash string, e abci.Event) (observer.OutboundTransfer, error) {
	mes, err := sdktypes.ParseTypedEvent(e)
	if err != nil {
		return observer.OutboundTransfer{}, errorsmod.Wrapf(err, "Failed to parse typed event")
	}
	ev, ok := mes.(*bridgetypes.EventOutboundTransfer)
	if !ok {
		return observer.OutboundTransfer{}, fmt.Errorf("Failed to parse EventOutboundTransfer from event")
	}

	return observer.OutboundTransfer{
		DstChain: observer.ChainId_BITCOIN,
		Id:       hash,
		Height:   height,
		Sender:   ev.Sender,
		To:       ev.DestAddr,
		Asset:    ev.AssetId.Denom,
		Amount:   math.Uint(ev.Amount),
	}, nil
}
