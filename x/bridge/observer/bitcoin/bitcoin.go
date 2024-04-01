package bitcoin

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
	bridgetypes "github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

var ModuleName = "bitcoin-chain"

type ChainClient struct {
	logger             log.Logger
	btcRpc             *rpcclient.Client
	vaultAddr          string
	stopChan           chan struct{}
	outboundChan       chan observer.Transfer
	observeSleepPeriod time.Duration
	lastObservedHeight atomic.Uint64
	chainParams        chaincfg.Params
}

// NewChainClient returns new instance of `ChainClient`
func NewChainClient(
	logger log.Logger,
	btcRpc *rpcclient.Client,
	vaultAddr string,
	observeSleepPeriod time.Duration,
	lastObservedHeight uint64,
	chainParams chaincfg.Params,
) (*ChainClient, error) {
	_, err := btcutil.DecodeAddress(vaultAddr, &chainParams)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidCfg, "Invalid BTC vault address")
	}

	c := &ChainClient{
		logger:             logger.With("module", ModuleName),
		btcRpc:             btcRpc,
		vaultAddr:          vaultAddr,
		stopChan:           make(chan struct{}),
		outboundChan:       make(chan observer.Transfer),
		observeSleepPeriod: observeSleepPeriod,
		lastObservedHeight: atomic.Uint64{},
		chainParams:        chainParams,
	}
	c.lastObservedHeight.Store(lastObservedHeight)

	return c, nil
}

// Start starts observing Bitcoin blocks for outbound transfers
func (b *ChainClient) Start(context.Context) error {
	go b.observeBlocks()

	b.logger.Info("Started Bitcoin chain client")
	return nil
}

// Stop stops observing Bitcoin blocks and shutdowns RPC client
func (b *ChainClient) Stop(ctx context.Context) error {
	close(b.stopChan)
	b.btcRpc.Shutdown()

	const shutdownTimeout = 3 * time.Second
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()
	stop := make(chan struct{})

	go func() {
		b.btcRpc.WaitForShutdown()
	}()

	select {
	case <-ctx.Done():
		b.logger.Error("Bitcoin chain client: shutdown timeout exceeded")
	case <-stop:
	}

	b.logger.Info("Stopped Bitcoin chain client")
	return nil
}

// ListenOutboundTransfer returns receive-only channel with outbound transfer items
func (b *ChainClient) ListenOutboundTransfer() <-chan observer.Transfer {
	return b.outboundChan
}

// SignalInboundTransfer sends `InboundTransfer` to Bitcoin
func (b *ChainClient) SignalInboundTransfer(context.Context, observer.Transfer) error {
	return fmt.Errorf("Not implemented")
}

// Height returns current height of the Bitcoin chain
func (b *ChainClient) Height(context.Context) (uint64, error) {
	height, err := b.btcRpc.GetBlockCount()
	if err != nil {
		return 0, errorsmod.Wrapf(ErrRpcClient, "Failed to get current height %s", err.Error())
	}
	return uint64(height), nil
}

// ConfirmationsRequired returns number of required tx confirmations
func (b *ChainClient) ConfirmationsRequired(context.Context, bridgetypes.AssetID) (uint64, error) {
	return 0, fmt.Errorf("not supported for the BTC chain")
}

// observeBlocks main loop for fetching new Bitcoin blocks
func (b *ChainClient) observeBlocks() {
	defer close(b.outboundChan)

	for {
		select {
		case <-b.stopChan:
			return
		default:
			err := b.fetchNewBlock()
			if err != nil {
				// Do not log error if block with this height doesn't exist yet
				if !errors.Is(err, ErrBlockUnavailable) {
					b.logger.Error(fmt.Sprintf(
						"Failed to fetch block %d: %s",
						b.lastObservedHeight.Load()+1,
						err.Error(),
					))
				}
				time.Sleep(b.observeSleepPeriod)
				continue
			}
		}
	}
}

// fetchNewBlock fetches new Bitcoin block and extracts relevant transactions.
func (b *ChainClient) fetchNewBlock() error {
	nextHeight := b.lastObservedHeight.Load() + 1

	// TODO: Now we infinitely retry to fetch one block if we get errors on either
	// GetBlockHash or GetBlockVerboseTx. Decide what to do in that case.
	// Skip the block after several retries?

	hash, err := b.btcRpc.GetBlockHash(int64(nextHeight))
	if err != nil {
		var rpcErr *btcjson.RPCError
		if errors.As(err, &rpcErr) && rpcErr.Code == btcjson.ErrRPCInvalidParameter {
			// The block doesn't exist yet, skip it for now
			return ErrBlockUnavailable
		}
		return errorsmod.Wrapf(err, "Failed to get block hash")
	}

	b.logger.With("height", nextHeight, "hash", hash).Info("Fetching bitcoin block")

	blockVerbose, err := b.btcRpc.GetBlockVerboseTx(hash)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to get verbose block")
	}

	for _, tx := range blockVerbose.Tx {
		txIn, isRelevant, err := b.processTx(uint64(blockVerbose.Height), &tx)
		if err != nil {
			b.logger.Error(fmt.Sprintf("Failed to process Tx %s: %s", tx.Txid, err.Error()))
			continue
		}
		if !isRelevant {
			continue
		}

		select {
		case b.outboundChan <- txIn:
		case <-b.stopChan:
			b.logger.Info("Observer exiting early, tx skipped: ", tx.Txid)
			return nil
		}
	}

	b.lastObservedHeight.Add(1)
	return nil
}

// processTx builds `Transfer` from provided Bitcoin transaction
func (b *ChainClient) processTx(height uint64, tx *btcjson.TxRawResult) (observer.Transfer, bool, error) {
	amount, relevant := b.isTxRelevant(tx)
	if !relevant {
		return observer.Transfer{}, false, nil
	}

	memo, contains := getMemo(tx)
	if !contains {
		return observer.Transfer{}, false, fmt.Errorf("failed to get Tx memo")
	}

	return observer.Transfer{
		SrcChain: observer.ChainIdBitcoin,
		DstChain: observer.ChainIdOsmosis,
		Id:       tx.Hash,
		Height:   height,
		Sender:   "", // NB! the sender is set in the osmosis chain client
		To:       memo,
		Asset:    string(observer.DenomBitcoin),
		Amount:   amount,
	}, true, nil
}

// isTxRelevant checks if a tx contains transfers to the BTC vault and calculates the total amount to transfer.
// This method goes through all the Vouts and accumulates their values if Vouts.account[0] == vault.
// Returns true is the tx is not relevant. Otherwise, false and the total amount to transfer.
func (b *ChainClient) isTxRelevant(tx *btcjson.TxRawResult) (math.Uint, bool) {
	var amount = sdk.NewUint(0)
	for _, vout := range tx.Vout {
		if strings.EqualFold(vout.ScriptPubKey.Type, "nulldata") {
			continue
		}
		if vout.Value <= 0 {
			continue
		}

		addresses, err := b.getAddressesFromScriptPubKey(vout.ScriptPubKey)
		if err != nil || len(addresses) != 1 {
			continue
		}

		if slices.Contains(addresses, b.vaultAddr) {
			a, err := getAmount(vout)
			if err != nil {
				continue
			}
			amount = amount.Add(a)
		}
	}
	return amount, !amount.IsZero()
}

// getAmount retrieves amount of tokens sent
func getAmount(vout btcjson.Vout) (math.Uint, error) {
	amount, err := btcutil.NewAmount(vout.Value)
	if err != nil {
		return math.Uint{}, errorsmod.Wrapf(err, "Failed to parse float value")
	}
	return math.NewUint(uint64(amount.ToUnit(btcutil.AmountSatoshi))), nil
}

// getMemo retrieves data behind `OP_RETURN` Vout. Return false if can't get memo.
func getMemo(tx *btcjson.TxRawResult) (string, bool) {
	for _, vout := range tx.Vout {
		if !strings.EqualFold(vout.ScriptPubKey.Type, "nulldata") {
			continue
		}
		fields := strings.Fields(vout.ScriptPubKey.Asm)
		// We should have an array like ["OP_RETURN", "$MEMO_DATA$"]
		if len(fields) == 2 {
			// Decode $MEMO_DATA$ entry from the array
			decoded, err := hex.DecodeString(fields[1])
			if err != nil {
				// TODO: log?
				continue
			}
			// TODO: verify the memo format
			return string(decoded), true
		}
	}
	return "", false
}

// getAddressesFromScriptPubKey extracts addresses from the Bitcoin script
func (b *ChainClient) getAddressesFromScriptPubKey(key btcjson.ScriptPubKeyResult) ([]string, error) {
	if len(key.Hex) == 0 {
		return nil, errors.New("Empty scriptPubKey hex")
	}
	buf, err := hex.DecodeString(key.Hex)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to decode scriptPubKey hex")
	}
	_, extractedAddresses, _, err := txscript.ExtractPkScriptAddrs(buf, &b.chainParams)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to extract address from scriptPubKey")
	}
	var addresses []string
	for _, addr := range extractedAddresses {
		addresses = append(addresses, addr.String())
	}
	return addresses, nil
}
