package model

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
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
	return sdk.MustAccAddressFromBech32(p.ContractAddress)
}

func (p CosmWasmPool) GetId() uint64 {
	panic("CosmWasmPool.GetId not implemented")
}

func (p CosmWasmPool) GetSpreadFactor(ctx sdk.Context) osmomath.Dec {
	panic("CosmWasmPool.GetSpreadFactor not implemented")
}

func (p CosmWasmPool) GetExitFee(ctx sdk.Context) osmomath.Dec {
	panic("CosmWasmPool.GetExitFee not implemented")
}

func (p CosmWasmPool) IsActive(ctx sdk.Context) bool {
	panic("CosmWasmPool.IsActive not implemented")
}

func (p CosmWasmPool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (osmomath.Dec, error) {
	panic("CosmWasmPool.SpotPrice not implemented")
}

func (p CosmWasmPool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.CosmWasm
}

func (p CosmWasmPool) GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins {
	panic("CosmWasmPool.GetTotalPoolLiquidity not implemented")
}

// types.CosmWasmExtension interface implementation

func (p CosmWasmPool) GetCodeId() uint64 {
	return p.CodeId
}

func (p CosmWasmPool) GetInstantiateMsg() []byte {
	return p.InstantiateMsg
}

func (p CosmWasmPool) GetContractAddress() string {
	return p.ContractAddress
}

func (p *CosmWasmPool) SetContractAddress(contractAddress string) {
	p.ContractAddress = contractAddress
}

func (p CosmWasmPool) GetStoreModel() poolmanagertypes.PoolI {
	return &p
}

func (p CosmWasmPool) SetWasmKeeper(wasmKeeper types.WasmKeeper) {
	panic("CosmWasmPool.SetWasmKeeeper not implemented")
}
