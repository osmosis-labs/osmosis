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
	Height      int64
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
	vaultAddr          string
	currentHeight      int64
	btcRpc             *rpcclient.Client
	globalTxInChan     chan TxIn
	stopChan           chan struct{}
	observeSleepPeriod time.Duration
}

// NewObserver returns new instance of `Observer` with BTC RPC client
func NewObserver(
	logger log.Logger,
	cfg RpcConfig,
	vaultAddr string,
	initialHeight int64,
	observeSleepPeriod time.Duration,
) (Observer, error) {
	err := cfg.Validate()
	if err != nil {
		return Observer{}, errorsmod.Wrapf(err, "Invalid RPC configuration")
	}
	if len(vaultAddr) == 0 {
		return Observer{}, fmt.Errorf("Empty vaultAddr")
	}

	btcRpc, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         cfg.Host,
		HTTPPostMode: true,
		DisableTLS:   cfg.DisableTls,
		User:         cfg.User,
		Pass:         cfg.Pass,
		Params:       chaincfg.TestNet3Params.Name,
	}, nil)
	if err != nil {
		return Observer{}, errorsmod.Wrapf(err, "Failed to create RPC client")
	}

	return Observer{
		logger:             logger.With("module", ModuleNameObserver),
		vaultAddr:          vaultAddr,
		currentHeight:      initialHeight,
		btcRpc:             btcRpc,
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

// FetchBlock processes block transactions at the given `height`
func (o *Observer) fetchBlock(height int64) error {
	hash, err := o.btcRpc.GetBlockHash(height)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to get block hash")
	}

	blockVerbose, err := o.btcRpc.GetBlockVerboseTx(hash)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to get verbose block")
	}

	for _, tx := range blockVerbose.Tx {
		txIn, err := o.processTx(blockVerbose.Height, &tx)
		if err != nil {
			o.logger.Error(fmt.Sprintf("Failed to process Tx %s: %s", tx.Txid, err.Error()))
			continue
		}
		o.globalTxInChan <- txIn
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
				o.logger.Error("Failed to fetch block", err)
				time.Sleep(o.observeSleepPeriod)
				continue
			}
			o.currentHeight = o.currentHeight + 1
		}
	}
}

func (o *Observer) processTx(height int64, tx *btcjson.TxRawResult) (TxIn, error) {
	sender, err := o.getSender(tx)
	if err != nil {
		return TxIn{}, errorsmod.Wrapf(err, "Failed to get Tx sender")
	}

	output, err := o.getVout(sender, tx)
	if err != nil {
		o.logger.Error("Failed to get output from Tx", err)
		fmt.Println("Failed output")
		return TxIn{}, err
	}

	dest, err := o.getDestination(output)
	if err != nil {
		return TxIn{}, errorsmod.Wrapf(err, "Failed to get destination address")
	}
	if dest != o.vaultAddr {
		return TxIn{}, fmt.Errorf("Invalid destination address")
	}

	amount, err := o.getAmount(output)
	if err != nil {
		return TxIn{}, errorsmod.Wrapf(err, "Failed to get amount")
	}

	memo, err := o.getMemo(tx)
	if err != nil {
		return TxIn{}, errorsmod.Wrapf(err, "Failed to get memo")
	}

	return TxIn{
		Id:          tx.Txid,
		Height:      height,
		Sender:      sender,
		Destination: dest,
		Amount:      amount,
		Memo:        memo,
	}, err
}

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

func (o *Observer) getVout(sender string, tx *btcjson.TxRawResult) (btcjson.Vout, error) {
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
			return vout, nil
		}
	}
	return btcjson.Vout{}, fmt.Errorf("Failed to get Vout")
}

func (o *Observer) getDestination(vout btcjson.Vout) (string, error) {
	addresses, err := o.getAddressesFromScriptPubKey(vout.ScriptPubKey)
	if err != nil {
		return "", errorsmod.Wrapf(err, "Failed to get destination address")
	}
	if len(addresses) == 0 {
		return "", errorsmod.Wrapf(err, "Destination address not found")
	}
	return addresses[0], nil
}

func (o *Observer) getAmount(vout btcjson.Vout) (math.Uint, error) {
	amount, err := btcutil.NewAmount(vout.Value)
	if err != nil {
		return math.Uint{}, errorsmod.Wrapf(err, "Failed to parse float value")
	}
	return math.NewUint(uint64(amount.ToUnit(btcutil.AmountSatoshi))), nil
}

func (o *Observer) getMemo(tx *btcjson.TxRawResult) (string, error) {
	for _, vout := range tx.Vout {
		if !strings.EqualFold(vout.ScriptPubKey.Type, "nulldata") {
			continue
		}
		fields := strings.Fields(vout.ScriptPubKey.Asm)
		if len(fields) == 2 {
			decoded, err := hex.DecodeString(fields[1])
			if err != nil {
				fmt.Println("Failed to decode field", fields[1])
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
