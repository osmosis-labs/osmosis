package bitcoin

import (
	"errors"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/cometbft/cometbft/libs/log"
)

type TxIn struct {
	Id          string
	Sender      string
	Destination string
	Amount      math.Uint
}

type RpcConfig struct {
	Host     string
	Endpoint string
	User     string
	Pass     string
}

type Observer struct {
	logger         log.Logger
	vaultAddr      string
	btcRpc         *rpcclient.Client
	rawTxChan      chan *btcjson.TxRawResult
	globalTxInChan chan TxIn
}

func clientConnected() {
	fmt.Println("Client connected")
}

func blockConnected(height int32, header *wire.BlockHeader, txs []*btcutil.Tx) {
	fmt.Println("Block connected: ", header)
	fmt.Println(txs)
}

func NewObserver(logger log.Logger, cfg RpcConfig, vaultAddr string) (Observer, error) {
	rawTxChan := make(chan *btcjson.TxRawResult)
	txAcceptedVerbose := func(txRaw *btcjson.TxRawResult) {
		rawTxChan <- txRaw
	}

	handlers := rpcclient.NotificationHandlers{
		OnClientConnected:        clientConnected,
		OnFilteredBlockConnected: blockConnected,
		OnTxAcceptedVerbose:      txAcceptedVerbose,
	}
	btcRpc, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:                cfg.Host,
		Endpoint:            cfg.Endpoint,
		User:                cfg.User,
		Pass:                cfg.Pass,
		DisableTLS:          true,
		HTTPPostMode:        false,
		DisableConnectOnNew: true,
		Params:              chaincfg.TestNet3Params.Name,
	}, &handlers)
	if err != nil {
		return Observer{}, errorsmod.Wrapf(err, "Failed to create RPC client")
	}

	return Observer{
		logger:         logger,
		vaultAddr:      vaultAddr,
		btcRpc:         btcRpc,
		rawTxChan:      rawTxChan,
		globalTxInChan: make(chan TxIn),
	}, nil
}

func (o *Observer) Start() error {
	err := o.btcRpc.Connect(1)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to connect with RPC client")
	}

	err = o.btcRpc.NotifyBlocks()
	if err != nil {
		return err
	}
	fmt.Println("Notify block registered")

	err = o.btcRpc.NotifyNewTransactions(true)
	if err != nil {
		return err
	}
	fmt.Println("Notify Tx registered")

	go o.processTransactions()

	return nil
}

func (o *Observer) Stop() error {
	o.btcRpc.Shutdown()
	o.btcRpc.WaitForShutdown()

	return nil
}

func (o *Observer) processTransactions() {
	for tx := range o.rawTxChan {
		txHash, err := chainhash.NewHashFromStr(tx.Vin[0].Txid)
		if err != nil {
			o.logger.Error(fmt.Sprintf("Failed to get Tx hash from Tx id %s", err.Error()))
			continue
		}
		vinTx, err := o.btcRpc.GetRawTransactionVerbose(txHash)
		if err != nil {
			o.logger.Error(fmt.Sprintf("Failed to get raw Tx %s", err.Error()))
			continue
		}
		vout := vinTx.Vout[tx.Vin[0].Vout]
		if len(vout.ScriptPubKey.Addresses) == 0 {
			o.logger.Error("Failed to extract address from vout")
			continue
		}
		sender := vout.ScriptPubKey.Addresses[0]

		output, err := o.getOutput(sender, tx)
		if err != nil {
			o.logger.Error("Failed to get output from Tx", err)
			continue
		}

		amount, err := btcutil.NewAmount(output.Value)
		if err != nil {
			o.logger.Error("Failed to parse float value", err)
			continue
		}

		o.globalTxInChan <- TxIn{
			Id:          tx.Txid,
			Sender:      sender,
			Destination: output.ScriptPubKey.Addresses[0],
			Amount:      math.NewUint(uint64(amount.ToUnit(btcutil.AmountSatoshi))),
		}
	}
}

func (o *Observer) getOutput(sender string, tx *btcjson.TxRawResult) (btcjson.Vout, error) {
	for _, vout := range tx.Vout {
		if strings.EqualFold(vout.ScriptPubKey.Type, "nulldata") {
			continue
		}
		if vout.Value <= 0 {
			continue
		}
		if len(vout.ScriptPubKey.Addresses) != 1 {
			continue
		}
		if vout.ScriptPubKey.Addresses[0] != sender {
			return vout, nil
		}
	}
	return btcjson.Vout{}, errors.New("Failed to get output from Tx")
}
