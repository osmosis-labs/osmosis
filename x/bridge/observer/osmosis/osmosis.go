package osmosis

import (
	"context"
	"fmt"
	"sync/atomic"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	cosmosproto "github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

var (
	ModuleName    = "osmosis-chain"
	OsmoGasLimit  = uint64(200000)
	OsmoFeeAmount = sdktypes.NewIntFromUint64(1000)
	OsmoFeeDenom  = "uosmo"
)

type ChainClient struct {
	logger             log.Logger
	osmoClient         *Client
	cometRpc           *rpchttp.HTTP
	stopChan           chan struct{}
	outboundChan       chan observer.Transfer
	lastObservedHeight atomic.Uint64
	signerAddr         string // TODO: validate this address (probably using a private key)! otherwise everyone could act on behalf of the validator
}

// NewChainClient returns new instance of `Osmosis`
func NewChainClient(
	logger log.Logger,
	osmoClient *Client,
	cometRpc *rpchttp.HTTP,
	signerAddr string,
) *ChainClient {
	return &ChainClient{
		logger:             logger.With("module", ModuleName),
		osmoClient:         osmoClient,
		cometRpc:           cometRpc,
		stopChan:           make(chan struct{}),
		outboundChan:       make(chan observer.Transfer),
		lastObservedHeight: atomic.Uint64{},
		signerAddr:         signerAddr,
	}
}

// Start subscribes to the `NewBlock` events and starts listening to `EventOutboundTransfer` events
func (o *ChainClient) Start(ctx context.Context) error {
	err := o.cometRpc.Start()
	if err != nil {
		return errorsmod.Wrapf(ErrRpcClient, err.Error())
	}

	txs, err := o.cometRpc.Subscribe(ctx, ModuleName, comettypes.EventQueryNewBlock.String())
	if err != nil {
		return errorsmod.Wrapf(ErrRpcClient, err.Error())
	}

	go o.observeEvents(ctx, txs)

	o.logger.Info("Started Osmosis chain client")
	return nil
}

// Stop stops listening to events and closes Osmosis client
func (o *ChainClient) Stop(ctx context.Context) error {
	close(o.stopChan)
	o.osmoClient.Close()
	if err := o.cometRpc.UnsubscribeAll(ctx, ModuleName); err != nil {
		return errorsmod.Wrapf(err, "Failed to unsubscribe from RPC client")
	}

	err := o.cometRpc.Stop()
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to stop comet RPC")
	}

	o.logger.Info("Stopped Osmosis chain client")
	return nil
}

// ListenOutboundTransfer returns receive-only channel with `OutboundTransfer` items
func (o *ChainClient) ListenOutboundTransfer() <-chan observer.Transfer {
	return o.outboundChan
}

// SignalInboundTransfer sends `InboundTransfer` to Osmosis
func (o *ChainClient) SignalInboundTransfer(ctx context.Context, in observer.Transfer) error {
	msg := bridgetypes.NewMsgInboundTransfer(
		in.Id,
		o.signerAddr, // NB! a current node should be a sender!
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
		return errorsmod.Wrapf(
			err,
			"Failed to broadcast tx to Osmosis for inbound transfer %s",
			in.Id,
		)
	}
	return nil
}

func (o *ChainClient) observeEvents(
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

			o.lastObservedHeight.Store(math.Max(
				o.lastObservedHeight.Load(),
				uint64(newBlock.Block.Height),
			))
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
						o.logger.Info(
							"Exiting early, event %s skipped in tx %s",
							e.String(),
							r.Hash.String(),
						)
						return
					}
				}
			}
		}
	}
}

func outboundTransferFromEvent(height uint64, hash string, e abci.Event) (observer.Transfer, error) {
	mes, err := sdktypes.ParseTypedEvent(e)
	if err != nil {
		return observer.Transfer{}, errorsmod.Wrapf(err, "Failed to parse typed event")
	}
	ev, ok := mes.(*bridgetypes.EventOutboundTransfer)
	if !ok {
		return observer.Transfer{}, fmt.Errorf("Failed to parse EventOutboundTransfer from event")
	}

	return observer.Transfer{
		SrcChain: observer.ChainIdOsmosis,
		DstChain: observer.ChainIdBitcoin,
		Id:       hash,
		Height:   height,
		Sender:   ev.Sender,
		To:       ev.DestAddr,
		Asset:    ev.AssetId.Denom,
		Amount:   math.Uint(ev.Amount),
	}, nil
}

// Height returns current height of the chain
func (o *ChainClient) Height(context.Context) (uint64, error) {
	return o.lastObservedHeight.Load(), nil
}

// ConfirmationsRequired returns number of required tx confirmations
func (o *ChainClient) ConfirmationsRequired(
	ctx context.Context,
	id bridgetypes.AssetID,
) (uint64, error) {
	cr, err := o.osmoClient.ConfirmationsRequired(ctx, id)
	if err != nil {
		return 0, errorsmod.Wrapf(err, "Failed to get confirmations required")
	}
	return cr, nil
}
