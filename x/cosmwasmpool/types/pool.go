package types

import (
	"github.com/gogo/protobuf/proto"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// CosmWasmExtension
type CosmWasmExtension interface {
	poolmanagertypes.PoolI

	GetCodeId() uint64

	GetInstantiateMsg() []byte

	GetContractAddress() string

	SetContractAddress(contractAddress string)

	GetStoreModel() proto.Message

	SetWasmKeeper(wasmKeeper WasmKeeper)
}
