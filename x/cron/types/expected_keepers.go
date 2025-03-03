package types

import (
	wasmvmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ContractOpsKeeper contains mutable operations on a contract.
type ContractOpsKeeper interface {
	// Sudo allows to call privileged entry point of a contract.
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmvmtypes.ContractInfo
}
