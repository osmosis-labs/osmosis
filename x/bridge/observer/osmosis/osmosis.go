package osmosis

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync/atomic"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

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
	txConfig           cosmosclient.TxConfig
}

// NewChainClient returns new instance of `Osmosis`
func NewChainClient(
	logger log.Logger,
	osmoClient *Client,
	cometRpc *rpchttp.HTTP,
	txConfig cosmosclient.TxConfig,
) *ChainClient {
	return &ChainClient{
		logger:             logger.With("module", ModuleName),
		osmoClient:         osmoClient,
		cometRpc:           cometRpc,
		stopChan:           make(chan struct{}),
		outboundChan:       make(chan observer.Transfer),
		lastObservedHeight: atomic.Uint64{},
		txConfig:           txConfig,
	}
}

func (c *ChainClient) RpcSend(ctx context.Context, tx []byte) (*coretypes.ResultBroadcastTx, error) {
	res, err := c.cometRpc.BroadcastTxSync(ctx, tx)
	return res, err
}

// Start subscribes to the `NewBlock` events and starts listening to `EventOutboundTransfer` events
func (c *ChainClient) Start(ctx context.Context) error {
	err := c.cometRpc.Start()
	if err != nil {
		return errorsmod.Wrapf(ErrRpcClient, err.Error())
	}

	txs, err := c.cometRpc.Subscribe(ctx, ModuleName, comettypes.EventQueryNewBlock.String())
	if err != nil {
		return errorsmod.Wrapf(ErrRpcClient, err.Error())
	}

	go c.observeEvents(ctx, txs)

	return nil
}

// Stop stops listening to events and closes Osmosis client
func (c *ChainClient) Stop(ctx context.Context) error {
	close(c.stopChan)
	c.osmoClient.Close()
	if err := c.cometRpc.UnsubscribeAll(ctx, ModuleName); err != nil {
		return errorsmod.Wrapf(err, "Failed to unsubscribe from RPC client")
	}
	return c.cometRpc.Stop()
}

// ListenOutboundTransfer returns receive-only channel with `OutboundTransfer` items
func (c *ChainClient) ListenOutboundTransfer() <-chan observer.Transfer {
	return c.outboundChan
}

// SignalInboundTransfer sends `InboundTransfer` to Osmosis
func (c *ChainClient) SignalInboundTransfer(ctx context.Context, in observer.Transfer) error {
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
	bytes, err := c.osmoClient.SignTx(ctx, msg, fees, OsmoGasLimit)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to sign tx for inbound transfer %s", in.Id)
	}
	_, err = c.osmoClient.BroadcastTx(ctx, bytes)
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"Failed to broadcast tx to Osmosis for inbound transfer %s",
			in.Id,
		)
	}
	return nil
}

func (c *ChainClient) observeEvents(ctx context.Context, txs <-chan coretypes.ResultEvent) {
	defer close(c.outboundChan)

	for {
		select {
		case <-c.stopChan:
			return
		case event := <-txs:
			newBlock, ok := event.Data.(comettypes.EventDataNewBlock)
			if !ok {
				continue
			}

			c.lastObservedHeight.Store(math.Max(
				c.lastObservedHeight.Load(),
				uint64(newBlock.Block.Height),
			))

			c.processNewBlockTxs(ctx, uint64(newBlock.Block.Height), newBlock.Block.Txs)
		}
	}
}

func (c *ChainClient) processNewBlockTxs(ctx context.Context, height uint64, txs comettypes.Txs) {
	for _, tx := range txs {
		txHash := hex.EncodeToString(tx.Hash())
		decoded, err := c.txConfig.TxDecoder()(tx)
		if err != nil {
			c.logger.Error(fmt.Sprintf(
				"Failed to decode Tx %s in block %d",
				txHash,
				height,
			))
			continue
		}

		res, err := c.cometRpc.CheckTx(ctx, tx)
		if err != nil {
			c.logger.Error(fmt.Sprintf(
				"Failed to get result for Tx %s in block %d",
				txHash,
				height,
			))
			continue
		}
		if res.IsErr() {
			continue
		}

		for _, msg := range decoded.GetMsgs() {
			outbound, ok := msg.(*bridgetypes.MsgOutboundTransfer)
			if !ok {
				continue
			}

			out := outboundTransferFromMsg(
				height,
				txHash,
				outbound,
			)

			select {
			case c.outboundChan <- out:
			case <-c.stopChan:
				c.logger.Info(
					"Exiting early, msg %s skipped in Tx %s, block %d",
					msg.String(),
					txHash,
					height,
				)
				return
			}
		}
	}
}

func outboundTransferFromMsg(
	height uint64,
	hash string,
	msg *bridgetypes.MsgOutboundTransfer,
) observer.Transfer {
	return observer.Transfer{
		SrcChain: observer.ChainIdOsmosis,
		DstChain: observer.ChainId(msg.AssetId.SourceChain),
		Id:       hash,
		Height:   height,
		Sender:   msg.Sender,
		To:       msg.DestAddr,
		Asset:    msg.AssetId.Denom,
		Amount:   math.Uint(msg.Amount),
	}
}

// Returns current height of the chain
func (c *ChainClient) Height() (uint64, error) {
	return c.lastObservedHeight.Load(), nil
}

// Returns number of required tx confirmations
func (c *ChainClient) ConfirmationsRequired() (uint64, error) {
	// Query bridge module
	return 0, nil
}
