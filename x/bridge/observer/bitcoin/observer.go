package bitcoin

import (
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
)

const ModuleNameObserver = "btc-observer"

type TxIn struct {
	Id          string
	Height      uint64
	Sender      string
	Destination string
	Amount      math.Uint
	Memo        string
}

type RpcConfig struct {
	Host       string
	DisableTls bool
	User       string
	Pass       string
}

func (o RpcConfig) Validate() error {
	if len(o.Host) == 0 {
		return fmt.Errorf("Invalid `Host` value")
	} else if len(o.User) == 0 {
		return fmt.Errorf("Invalid `User` value")
	} else if len(o.Pass) == 0 {
		return fmt.Errorf("Invalid `Pass` value")
	} else {
		return nil
	}
}

type Observer struct {
	logger             log.Logger
	btcRpc             *rpcclient.Client
	vaultAddr          string
	currentHeight      uint64
	globalTxInChan     chan TxIn
	stopChan           chan struct{}
	observeSleepPeriod time.Duration
}

// NewObserver returns new instance of `Observer` with BTC RPC client
func NewObserver(
	logger log.Logger,
	btcRpc *rpcclient.Client,
	vaultAddr string,
	initialHeight uint64,
	observeSleepPeriod time.Duration,
) (Observer, error) {
	if len(vaultAddr) == 0 {
		return Observer{}, errorsmod.Wrapf(ErrInvalidCfg, "Invalid vaultAddr")
	}

	return Observer{
		logger:             logger.With("module", ModuleNameObserver),
		btcRpc:             btcRpc,
		vaultAddr:          vaultAddr,
		currentHeight:      initialHeight,
		globalTxInChan:     make(chan TxIn),
		stopChan:           make(chan struct{}),
		observeSleepPeriod: observeSleepPeriod,
	}, nil
}

// Start starts observing BTC blocks for incoming Txs to the given address
func (o *Observer) Start() {
	go o.observeBlocks()
}

// Stop stops observation loop and RPC client
func (o *Observer) Stop() {
	close(o.stopChan)
	o.btcRpc.Shutdown()
	o.btcRpc.WaitForShutdown()
}

// TxIns returns receive-only part of observed Txs channel
func (o *Observer) TxIns() <-chan TxIn {
	return o.globalTxInChan
}

func (o *Observer) CurrentHeight() uint64 {
	return o.currentHeight
}

// FetchBlock processes block transactions at the given `height`
func (o *Observer) fetchBlock(height uint64) error {
	hash, err := o.btcRpc.GetBlockHash(int64(height))
	if err != nil {
		if rpcErr, ok := err.(*btcjson.RPCError); ok && rpcErr.Code == btcjson.ErrRPCInvalidParameter {
			return ErrBlockUnavailable
		}
		return errorsmod.Wrapf(err, "Failed to get block hash")
	}

	blockVerbose, err := o.btcRpc.GetBlockVerboseTx(hash)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to get verbose block")
	}

	for _, tx := range blockVerbose.Tx {
		txIn, isRelevant, err := o.processTx(uint64(blockVerbose.Height), &tx)
		if isRelevant && err != nil {
			o.logger.Error(fmt.Sprintf("Failed to process Tx %s: %s", tx.Txid, err.Error()))
		}
		if isRelevant && err == nil {
			select {
			case o.globalTxInChan <- txIn:
			case <-o.stopChan:
				o.logger.Info("Observer exiting early, tx skipped: ", tx.Txid)
				return nil
			}
		}
	}
	return nil
}

func (o *Observer) observeBlocks() {
	defer close(o.globalTxInChan)

	for {
		select {
		case <-o.stopChan:
			return
		default:
			err := o.fetchBlock(o.currentHeight)
			if err != nil {
				// Do not log error if block with this height doesn't exist yet
				if !errors.Is(err, ErrBlockUnavailable) {
					o.logger.Error(fmt.Sprintf("Failed to fetch block %d: %s", o.currentHeight, err.Error()))
				}
				time.Sleep(o.observeSleepPeriod)
				continue
			}
			o.currentHeight = o.currentHeight + 1
		}
	}
}

func (o *Observer) processTx(height uint64, tx *btcjson.TxRawResult) (TxIn, bool, error) {
	sender, err := o.getSender(tx)
	if err != nil {
		return TxIn{}, false, errorsmod.Wrapf(err, "Failed to get Tx sender")
	}

	dest, amount, err := o.getOutput(sender, tx)
	if err != nil {
		return TxIn{}, false, errorsmod.Wrapf(err, "Failed to get Tx output")
	}
	isRelevant := dest == o.vaultAddr

	memo, err := o.getMemo(tx)
	if err != nil {
		return TxIn{}, isRelevant, errorsmod.Wrapf(err, "Failed to get Tx memo")
	}

	return TxIn{
		Id:          tx.Txid,
		Height:      height,
		Sender:      sender,
		Destination: dest,
		Amount:      amount,
		Memo:        memo,
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
func (o *Observer) getSender(tx *btcjson.TxRawResult) (string, error) {
	if len(tx.Vin) == 0 {
		return "", fmt.Errorf("Vin is empty for Tx %s", tx.Txid)
	}

	txHash, err := chainhash.NewHashFromStr(tx.Vin[0].Txid)
	if err != nil {
		return "", errorsmod.Wrapf(err, "Failed to get Vin Tx hash for Tx %s", tx.Txid)
	}

	vinTx, err := o.btcRpc.GetRawTransactionVerbose(txHash)
	if err != nil {
		return "", errorsmod.Wrapf(err, "Failed to get Vin Tx with hash %s", txHash)
	}

	vout := vinTx.Vout[tx.Vin[0].Vout]
	addresses, err := o.getAddressesFromScriptPubKey(vout.ScriptPubKey)
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
func (o *Observer) getOutput(sender string, tx *btcjson.TxRawResult) (string, math.Uint, error) {
	for _, vout := range tx.Vout {
		if strings.EqualFold(vout.ScriptPubKey.Type, "nulldata") {
			continue
		}
		if vout.Value <= 0 {
			continue
		}

		addresses, err := o.getAddressesFromScriptPubKey(vout.ScriptPubKey)
		if err != nil || len(addresses) != 1 {
			continue
		}

		if addresses[0] != sender {
			amount, err := o.getAmount(vout)
			if err != nil {
				continue
			}
			return addresses[0], amount, nil
		}
	}
	return "", math.Uint{}, fmt.Errorf("Failed to get Vout")
}

// getAmount retrieves amount of tokens sent
func (o *Observer) getAmount(vout btcjson.Vout) (math.Uint, error) {
	amount, err := btcutil.NewAmount(vout.Value)
	if err != nil {
		return math.Uint{}, errorsmod.Wrapf(err, "Failed to parse float value")
	}
	return math.NewUint(uint64(amount.ToUnit(btcutil.AmountSatoshi))), nil
}

// getMemo retrieves data behind `OP_RETURN` Vout
func (o *Observer) getMemo(tx *btcjson.TxRawResult) (string, error) {
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
				o.logger.Error("Failed to decode OP_RETURN field", err)
				continue
			}
			return string(decoded), nil
		}
	}
	return "", fmt.Errorf("Memo not found")
}

func (o *Observer) getAddressesFromScriptPubKey(key btcjson.ScriptPubKeyResult) ([]string, error) {
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
