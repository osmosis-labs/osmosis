package main

import "C"

import (
	// std
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	// helpers
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	// tendermint
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	// cosmos sdk
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// wasmd
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	// cosmwasm-testing
	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types/result"
	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types/testenv"
)

var (
	envCounter  uint64 = 0
	envRegister        = sync.Map{}
	mu          sync.Mutex
)

//export InitTestEnv
func InitTestEnv() uint64 {
	// Allow testing unoptimized contract
	wasmtypes.MaxWasmSize = 1024 * 1024 * 1024 * 1024 * 1024
	// Temp fix for concurrency issue
	mu.Lock()
	defer mu.Unlock()

	env := new(testenv.TestEnv)
	env.App = testenv.SetupOsmosisApp()

	env.Ctx = env.App.BaseApp.NewContext(false, tmproto.Header{Height: 0, ChainID: "osmosis-1", Time: time.Now().UTC()})

	env.BeginNewBlock(false)

	reqEndBlock := abci.RequestEndBlock{Height: env.Ctx.BlockHeight()}
	env.App.EndBlock(reqEndBlock)
	env.App.Commit()

	envCounter += 1
	id := envCounter

	envRegister.Store(id, *env)

	return id
}

//export InitAccount
func InitAccount(envId uint64, coinsJson string) *C.char {
	env := loadEnv(envId)
	var coins sdk.Coins

	if err := json.Unmarshal([]byte(coinsJson), &coins); err != nil {
		panic(err)
	}

	priv := secp256k1.GenPrivKey()
	accAddr := sdk.AccAddress(priv.PubKey().Address())

	err := simapp.FundAccount(env.App.BankKeeper, env.Ctx, accAddr, coins)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to fund account"))
	}

	base64Priv := base64.StdEncoding.EncodeToString(priv.Bytes())

	envRegister.Store(envId, env)

	return C.CString(base64Priv)
}

//export BeginBlock
func BeginBlock(envId uint64) {
	env := loadEnv(envId)
	env.BeginNewBlock(false)
	envRegister.Store(envId, env)
}

//export EndBlock
func EndBlock(envId uint64) {
	env := loadEnv(envId)
	reqEndBlock := abci.RequestEndBlock{Height: env.Ctx.BlockHeight()}
	env.App.EndBlock(reqEndBlock)
	env.App.Commit()
	envRegister.Store(envId, env)
}

//export Execute
func Execute(envId uint64, base64ReqDeliverTx string) *C.char {
	env := loadEnv(envId)
	// Temp fix for concurrency issue
	mu.Lock()
	defer mu.Unlock()

	reqDeliverTxBytes, err := base64.StdEncoding.DecodeString(base64ReqDeliverTx)
	if err != nil {
		panic(err)
	}

	reqDeliverTx := abci.RequestDeliverTx{}
	err = proto.Unmarshal(reqDeliverTxBytes, &reqDeliverTx)
	if err != nil {
		return encodeErrToResultBytes(result.ExecuteError, err)
	}

	resDeliverTx := env.App.DeliverTx(reqDeliverTx)
	bz, err := proto.Marshal(&resDeliverTx)

	if err != nil {
		panic(err)
	}

	envRegister.Store(envId, env)

	return encodeBytesResultBytes(bz)
}

//export Query
func Query(envId uint64, path, base64QueryMsgBytes string) *C.char {
	env := loadEnv(envId)
	queryMsgBytes, err := base64.StdEncoding.DecodeString(base64QueryMsgBytes)
	if err != nil {
		panic(err)
	}

	req := abci.RequestQuery{}
	req.Data = queryMsgBytes

	route := env.App.GRPCQueryRouter().Route(path)
	if route == nil {
		err := errors.New("No route found for `" + path + "`")
		return encodeErrToResultBytes(result.QueryError, err)
	}
	res, err := route(env.Ctx, req)

	if err != nil {
		return encodeErrToResultBytes(result.QueryError, err)
	}

	return encodeBytesResultBytes(res.Value)
}

//export AccountSequence
func AccountSequence(envId uint64, bech32Address string) uint64 {
	env := loadEnv(envId)

	addr, err := sdk.AccAddressFromBech32(bech32Address)

	if err != nil {
		panic(err)
	}

	seq, err := env.App.AppKeepers.AccountKeeper.GetSequence(env.Ctx, addr)

	if err != nil {
		panic(err)
	}

	return seq
}

//export AccountNumber
func AccountNumber(envId uint64, bech32Address string) uint64 {
	env := loadEnv(envId)

	addr, err := sdk.AccAddressFromBech32(bech32Address)

	if err != nil {
		panic(err)
	}

	acc := env.App.AppKeepers.AccountKeeper.GetAccount(env.Ctx, addr)
	return acc.GetAccountNumber()
}

//export Simulate
func Simulate(envId uint64, base64TxBytes string) *C.char { // => base64GasInfo
	env := loadEnv(envId)
	// Temp fix for concurrency issue
	mu.Lock()
	defer mu.Unlock()

	txBytes, err := base64.StdEncoding.DecodeString(base64TxBytes)
	if err != nil {
		panic(err)
	}

	gasInfo, _, err := env.App.Simulate(txBytes)

	if err != nil {
		return encodeErrToResultBytes(result.ExecuteError, err)
	}

	bz, err := proto.Marshal(&gasInfo)
	if err != nil {
		panic(err)
	}

	return encodeBytesResultBytes(bz)
}

// ========= utils =========

func loadEnv(envId uint64) testenv.TestEnv {
	item, ok := envRegister.Load(envId)
	env := testenv.TestEnv(item.(testenv.TestEnv))
	if !ok {
		panic(fmt.Sprintf("env not found: %d", envId))
	}
	return env
}

func encodeErrToResultBytes(code byte, err error) *C.char {
	return C.CString(result.EncodeResultFromError(code, err))
}

func encodeBytesResultBytes(bytes []byte) *C.char {
	return C.CString(result.EncodeResultFromOk(bytes))
}

// must define main for ffi build
func main() {}
