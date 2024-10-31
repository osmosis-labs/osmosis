package types

import (
	context "context"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/cosmwasmpool keeper.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// PoolManagerKeeper defines the interface needed to be fulfilled for
// the poolmanager keeper.
type PoolManagerKeeper interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
	GetNextPoolId(ctx sdk.Context) uint64
}

// ContractKeeper defines the interface needed to be fulfilled for
// the contract keeper.
type ContractKeeper interface {
	Instantiate(
		ctx sdk.Context,
		codeID uint64,
		creator, admin sdk.AccAddress,
		initMsg []byte,
		label string,
		deposit sdk.Coins,
	) (sdk.AccAddress, []byte, error)

	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)

	Execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) ([]byte, error)

	Create(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, instantiateAccess *wasmtypes.AccessConfig) (codeID uint64, checksum []byte, err error)

	Migrate(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, newCodeID uint64, msg []byte) ([]byte, error)
}

// ContractKeeper defines the interface needed to be fulfilled for
// the WasmKeeper.
type WasmKeeper interface {
	QuerySmart(ctx context.Context, contractAddress sdk.AccAddress, queryMsg []byte) ([]byte, error)
	QueryGasLimit() storetypes.Gas

	GetContractInfo(ctx context.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
}
