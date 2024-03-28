package bitcoin

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cometbft/cometbft/libs/log"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
)

var (
	ModuleName = "bitcoin-chain"
)

type Bitcoin struct {
	logger             log.Logger
	btcRpc             *rpcclient.Client
	vaultAddr          string
	stopChan           chan struct{}
	outboundChan       chan observer.OutboundTransfer
	observeSleepPeriod time.Duration
	lastObservedHeight uint64
}

// NewBitcoin returns new instance of `Bitcoin`
func NewBitcoin(
	logger log.Logger,
	btcRpc *rpcclient.Client,
	vaultAddr string,
	observeSleepPeriod time.Duration,
	lastObservedHeight uint64,
) (*Bitcoin, error) {
	if len(vaultAddr) == 0 {
		return nil, errorsmod.Wrapf(ErrInvalidCfg, "Invalid BTC vault address")
	}

	return &Bitcoin{
		logger:             logger.With("module", ModuleName),
		btcRpc:             btcRpc,
		vaultAddr:          vaultAddr,
		stopChan:           make(chan struct{}),
		outboundChan:       make(chan observer.OutboundTransfer),
		observeSleepPeriod: observeSleepPeriod,
		lastObservedHeight: lastObservedHeight,
	}, nil
}

// Start starts observing Bitcoin blocks for outbound transfers
func (b *Bitcoin) Start(context.Context) error {
	go b.observeBlocks()

	return nil
}

// Stop stops observing Bitcoin blocks and shutdowns RPC client
func (b *Bitcoin) Stop(context.Context) error {
	close(b.stopChan)
	b.btcRpc.Shutdown()
	b.btcRpc.WaitForShutdown()
	return nil
}

// ListenOutboundTransfer returns receive-only channel with outbound transfer items
func (b *Bitcoin) ListenOutboundTransfer() <-chan observer.OutboundTransfer {
	return b.outboundChan
}

// SignalInboundTransfer sends `InboundTransfer` to Bitcoin
func (b *Bitcoin) SignalInboundTransfer(ctx context.Context, in observer.InboundTransfer) error {
	return fmt.Errorf("Not implemented")
}

// Returns current height of the Bitcoin chain
func (b *Bitcoin) Height() (uint64, error) {
	height, err := b.btcRpc.GetBlockCount()
	if err != nil {
		return 0, errorsmod.Wrapf(ErrRpcClient, "Failed to get current height %s", err.Error())
	}
	return uint64(height), nil
}

// Returns number of required tx confirmations
func (b *Bitcoin) ConfirmationsRequired() (uint64, error) {
	// Query bridge module
	return 0, nil
}

func (b *Bitcoin) observeBlocks() {
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
					b.logger.Error(fmt.Sprintf("Failed to fetch block %d: %s", b.lastObservedHeight+1, err.Error()))
				}
				time.Sleep(b.observeSleepPeriod)
				continue
			}
		}
	}
}

func (b *Bitcoin) fetchNewBlock() error {
	nextHeight := b.lastObservedHeight + 1
	hash, err := b.btcRpc.GetBlockHash(int64(nextHeight))
	if err != nil {
		if rpcErr, ok := err.(*btcjson.RPCError); ok && rpcErr.Code == btcjson.ErrRPCInvalidParameter {
			return ErrBlockUnavailable
		}
		return errorsmod.Wrapf(err, "Failed to get block hash")
	}

	blockVerbose, err := b.btcRpc.GetBlockVerboseTx(hash)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to get verbose block")
	}

	for _, tx := range blockVerbose.Tx {
		txIn, isRelevant, err := b.processTx(uint64(blockVerbose.Height), &tx)
		if isRelevant && err != nil {
			b.logger.Error(fmt.Sprintf("Failed to process Tx %s: %s", tx.Txid, err.Error()))
		}
		if isRelevant && err == nil {
			select {
			case b.outboundChan <- txIn:
			case <-b.stopChan:
				b.logger.Info("Observer exiting early, tx skipped: ", tx.Txid)
				return nil
			}
		}
	}
	b.lastObservedHeight += 1
	return nil
}

func (b *Bitcoin) processTx(height uint64, tx *btcjson.TxRawResult) (observer.OutboundTransfer, bool, error) {
	sender, err := b.getSender(tx)
	if err != nil {
		return observer.OutboundTransfer{}, false, errorsmod.Wrapf(err, "Failed to get Tx sender")
	}

	dest, amount, err := b.getOutput(sender, tx)
	if err != nil {
		return observer.OutboundTransfer{}, false, errorsmod.Wrapf(err, "Failed to get Tx output")
	}
	isRelevant := dest == b.vaultAddr

	memo, err := b.getMemo(tx)
	if err != nil {
		return observer.OutboundTransfer{}, isRelevant, errorsmod.Wrapf(err, "Failed to get Tx memo")
	}

	return observer.OutboundTransfer{
		DstChain: observer.ChainId_OSMO,
		Id:       tx.Hash,
		Height:   height,
		Sender:   sender,
		To:       memo,
		Asset:    string(observer.Denom_BITCOIN),
		Amount:   amount,
	}, isRelevant, nil
}

// getSender retrieves sender's address from Tx
// There is no straightforward way to determine sender of the Tx, the flow used in this impl goes like this:
// 1. Get the `txid` and `Vout` index from `Vin[0]` from incoming Tx
//   - if we have multiple `Vin` entries the best thing we can do is to assume
//     that all of the Vins owned by the same person
//
// 2. Get Tx by `txid` to get the transaction the input is originated from
// 3. Get the `Vout` entry from this Tx by its index
// 4. Get the sender address from `Vout`
//   - if `addresses` field is available - get the first address from it
//     (again we assume that all of them are owned by the same person)
//   - otherwise - try to decode address from the script
func (b *Bitcoin) getSender(tx *btcjson.TxRawResult) (string, error) {
	if len(tx.Vin) == 0 {
		return "", fmt.Errorf("Vin is empty for Tx %s", tx.Txid)
	}

	txHash, err := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	if err != nil {
		return "", errorsmod.Wrapf(err, "Failed to get Vin Tx hash for Tx %s", tx.Txid)
	}

	vinTx, err := b.btcRpc.GetRawTransactionVerbose(txHash)
	if err != nil {
		return "", errorsmod.Wrapf(err, "Failed to get Vin Tx with hash %s", txHash)
	}

	vout := vinTx.Vout[tx.Vin[0].Vout]
	addresses, err := b.getAddressesFromScriptPubKey(vout.ScriptPubKey)
	if err != nil {
		return "", errorsmod.Wrapf(err, "Failed to get addresses from tx %s", txHash)
	}
	if len(addresses) == 0 {
		return "", fmt.Errorf("No addresses found in Vout in tx %s", txHash)
	}

	return addresses[0], nil
}

// getOutput retrieves receiver address and amount of tokens from Tx
// We try to find a `Vout` with a single receiver address (our vault)
// to get the Tx receiver and amount of tokens
// We go through all of the `Vout`'s and pick the first one that is not addressed back to the sender
func (b *Bitcoin) getOutput(sender string, tx *btcjson.TxRawResult) (string, math.Uint, error) {
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

		if addresses[0] != sender {
			amount, err := b.getAmount(vout)
			if err != nil {
				continue
			}
			return addresses[0], amount, nil
		}
	}
	return "", math.Uint{}, fmt.Errorf("Failed to get Vout")
}

// getAmount retrieves amount of tokens sent
func (b *Bitcoin) getAmount(vout btcjson.Vout) (math.Uint, error) {
	amount, err := btcutil.NewAmount(vout.Value)
	if err != nil {
		return math.Uint{}, errorsmod.Wrapf(err, "Failed to parse float value")
	}
	return math.NewUint(uint64(amount.ToUnit(btcutil.AmountSatoshi))), nil
}

// getMemo retrieves data behind `OP_RETURN` Vout
func (b *Bitcoin) getMemo(tx *btcjson.TxRawResult) (string, error) {
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
				b.logger.Error("Failed to decode OP_RETURN field", err)
				continue
			}
			return string(decoded), nil
		}
	}
	return "", fmt.Errorf("Memo not found")
}

func (b *Bitcoin) getAddressesFromScriptPubKey(key btcjson.ScriptPubKeyResult) ([]string, error) {
	if len(key.Addresses) > 0 {
		return key.Addresses, nil
	}
	if len(key.Hex) == 0 {
		return nil, errors.New("Empty scriptPubKey hex")
	}
	buf, err := hex.DecodeString(key.Hex)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to decode scriptPubKey hex")
	}
	_, extractedAddresses, _, err := txscript.ExtractPkScriptAddrs(buf, &chaincfg.TestNet3Params)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to extract address from scriptPubKey")
	}
	var addresses []string
	for _, addr := range extractedAddresses {
		addresses = append(addresses, addr.String())
	}
	return addresses, nil
}
