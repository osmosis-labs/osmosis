package model

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var (
	_ poolmanagertypes.PoolI  = &CosmWasmPool{}
	_ types.CosmWasmExtension = &CosmWasmPool{}
)

// String returns the json marshalled string of the pool
func (p CosmWasmPool) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// poolmanager.PoolI interface implementation

func (p CosmWasmPool) GetAddress() sdk.AccAddress {
	panic("CosmWasmPool.GetAddress not implemented")
}

func (p CosmWasmPool) GetId() uint64 {
	panic("CosmWasmPool.GetId not implemented")
}

func (p CosmWasmPool) GetSpreadFactor(ctx sdk.Context) sdk.Dec {
	panic("CosmWasmPool.GetSpreadFactor not implemented")
}

func (p CosmWasmPool) GetExitFee(ctx sdk.Context) sdk.Dec {
	panic("CosmWasmPool.GetExitFee not implemented")
}

func (p CosmWasmPool) IsActive(ctx sdk.Context) bool {
	panic("CosmWasmPool.IsActive not implemented")
}

func (p CosmWasmPool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	panic("CosmWasmPool.SpotPrice not implemented")
}

func (p CosmWasmPool) GetType() poolmanagertypes.PoolType {
	panic("CosmWasmPool.GetType not implemented")
}

func (p CosmWasmPool) GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins {
	panic("CosmWasmPool.GetTotalPoolLiquidity not implemented")
}

// types.CosmWasmExtension interface implementation

func (p CosmWasmPool) GetCodeId() uint64 {
	panic("CosmWasmPool.GetType not implemented")
}

func (p CosmWasmPool) GetInstantiateMsg() []byte {
	panic("CosmWasmPool.GetInstantiateMsg not implemented")
}

func (p CosmWasmPool) GetContractAddress() string {
	panic("CosmWasmPool.GetContractAddress not implemented")
}

func (p CosmWasmPool) SetContractAddress(contractAddress string) {
	panic("CosmWasmPool.SetContractAddress not implemented")
}

func (p CosmWasmPool) GetStoreModel() proto.Message {
	panic("CosmWasmPool.GetStoreModel not implemented")
}

func (p CosmWasmPool) SetWasmKeeper(wasmKeeper types.WasmKeeper) {
	panic("CosmWasmPool.SetWasmKeeeper not implemented")
}
