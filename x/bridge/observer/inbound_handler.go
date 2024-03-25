package observer

import (
	"context"
	"sync"
	"time"

	"cosmossdk.io/math"
	btcclient "github.com/btcsuite/btcd/rpcclient"
	"github.com/cometbft/cometbft/libs/log"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/observer/bitcoin"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/observer/osmosis"
	bridgetypes "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

var (
	SourceChain      = "bitcoin"
	SourceChainDenom = "btc"
	OsmoGasLimit     = uint64(200000)
	OsmoFeeAmount    = sdktypes.NewIntFromUint64(1000)
	OsmoFeeDenom     = "uosmo"
)

type InboundHandler struct {
	btcObserver           bitcoin.Observer
	btcClient             *btcclient.Client
	osmoClient            *osmosis.Client
	txQueue               []bitcoin.TxIn
	confirmationsRequired uint64
	connTimeout           time.Duration
	lock                  sync.Mutex
	stopChan              chan struct{}
}

func NewInboundHandler(
	logger log.Logger,
	btcObserver bitcoin.Observer,
	btcClient *btcclient.Client,
	osmoClient *osmosis.Client,
	confirmationsRequired uint64,
	connTimeout time.Duration,
) (InboundHandler, error) {
	return InboundHandler{
		btcObserver:           btcObserver,
		btcClient:             btcClient,
		osmoClient:            osmoClient,
		txQueue:               []bitcoin.TxIn{},
		confirmationsRequired: confirmationsRequired,
		connTimeout:           connTimeout,
		lock:                  sync.Mutex{},
		stopChan:              make(chan struct{}),
	}, nil
}

func (h *InboundHandler) Start() {
	h.btcObserver.Start()
	go h.collectTxIns()
	go h.processTxQueue()
}

func (h *InboundHandler) Stop() {
	h.btcObserver.Stop()
	close(h.stopChan)
}

func (h *InboundHandler) collectTxIns() {
	for {
		select {
		case <-h.stopChan:
			return
		case txIn := <-h.btcObserver.TxIns():
			h.lock.Lock()
			h.txQueue = append(h.txQueue, txIn)
			h.lock.Unlock()
		}
	}
}

func (h *InboundHandler) processTxQueue() {
	for {
		select {
		case <-h.stopChan:
			h.sendTxs()
			return
		case <-time.After(time.Minute):
			h.sendTxs()
		}
	}
}

func (h *InboundHandler) sendTxs() {
	h.lock.Lock()
	defer h.lock.Unlock()

	btcHeight := h.btcObserver.CurrentHeight()
	newTxQueue := []bitcoin.TxIn{}
	for _, tx := range h.txQueue {
		if btcHeight-tx.Height < h.confirmationsRequired {
			newTxQueue = append(newTxQueue, tx)
		} else {
			err := h.signAndBroadcastTx(tx)
			if err != nil {
				newTxQueue = append(newTxQueue, tx)
			}
		}
	}
	h.txQueue = newTxQueue
}

func (h *InboundHandler) signAndBroadcastTx(tx bitcoin.TxIn) error {
	msg := bridgetypes.NewMsgInboundTransfer(
		tx.Id,
		tx.Sender,
		tx.Destination,
		bridgetypes.AssetID{
			SourceChain: SourceChain,
			Denom:       SourceChainDenom,
		},
		math.Int(tx.Amount),
	)
	ctx, cancel := context.WithTimeout(context.Background(), h.connTimeout)
	defer cancel()
	fees := sdktypes.NewCoins(sdktypes.NewCoin(OsmoFeeDenom, OsmoFeeAmount))
	bytes, err := h.osmoClient.SignTx(ctx, msg, fees, OsmoGasLimit)
	if err != nil {
		return err
	}
	_, err = h.osmoClient.BroadcastTx(ctx, bytes)
	if err != nil {
		return err
	}
	return nil
}
