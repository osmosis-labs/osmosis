package model

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/cosmwasm/msg"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cosmwasmutils "github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
)

// Pool encapsulates all data and behavior for interacting with a CW pool.
//
// Note: CW Pool has 2 pool models:
// - CosmWasmPool which is a proto-generated store model used for serialization into state.
// - Pool struct that encapsulates the CosmWasmPool and wasmKeeper for calling the contract.
//
// CosmWasmPool implements the poolmanager.PoolI interface but it panics on all methods.
// The reason is that access to wasmKeeper is required to call the contract.
//
// Instead, all interactions and poolmanager.PoolI methods are to be performed
// on the Pool struct. The reason why we cannot have a Pool struct only is
// because it cannot be serialized into state due to having a non-serializable wasmKeeper field.
type Pool struct {
	CosmWasmPool
	WasmKeeper types.WasmKeeper
}

var (
	_ poolmanagertypes.PoolI  = &Pool{}
	_ types.CosmWasmExtension = &Pool{}
)

// NewCosmWasmPool creates a new CosmWasm pool with the specified parameters.
func NewCosmWasmPool(poolId uint64, codeId uint64, instantiateMsg []byte) *Pool {
	pool := Pool{
		CosmWasmPool: CosmWasmPool{
			ContractAddress: "", // N.B. This is to be set in InitializePool()
			PoolId:          poolId,
			CodeId:          codeId,
			InstantiateMsg:  instantiateMsg,
		},
		WasmKeeper: nil, // N.B.: this is set in InitializePool().
	}

	return &pool
}

// poolmanager.PoolI interface implementation

// GetAddress returns the address of the cosmwasm pool
func (p Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.ContractAddress)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", p.GetId()))
	}
	return addr
}

// GetId returns the id of the cosmwasm pool
func (p Pool) GetId() uint64 {
	return p.PoolId
}

// String returns the json marshalled string of the pool
func (p Pool) String() string {
	return p.CosmWasmPool.String()
}

// GetSpreadFactor returns the swap fee of the pool.
func (p Pool) GetSpreadFactor(ctx sdk.Context) osmomath.Dec {
	request := msg.GetSwapFeeQueryMsg{}
	response := cosmwasmutils.MustQuery[msg.GetSwapFeeQueryMsg, msg.GetSwapFeeQueryMsgResponse](ctx, p.WasmKeeper, p.ContractAddress, request)
	return response.SwapFee
}

// IsActive returns true if the pool is active
func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

// SpotPrice returns the spot price of the pool.
func (p Pool) SpotPrice(ctx sdk.Context, quoteAssetDenom string, baseAssetDenom string) (osmomath.BigDec, error) {
	request := msg.SpotPriceQueryMsg{
		SpotPrice: msg.SpotPrice{
			QuoteAssetDenom: quoteAssetDenom,
			BaseAssetDenom:  baseAssetDenom,
		},
	}
	response, err := cosmwasmutils.Query[msg.SpotPriceQueryMsg, msg.SpotPriceQueryMsgResponse](ctx, p.WasmKeeper, p.ContractAddress, request)
	if err != nil {
		return osmomath.BigDec{}, err
	}
	return osmomath.MustNewBigDecFromStr(response.SpotPrice), nil
}

// GetType returns the type of the pool.
func (p Pool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.CosmWasm
}

// GetTotalPoolLiquidity returns the total pool liquidity
func (p Pool) GetTotalPoolLiquidity(ctx sdk.Context) sdk.Coins {
	request := msg.GetTotalPoolLiquidityQueryMsg{}
	response := cosmwasmutils.MustQuery[msg.GetTotalPoolLiquidityQueryMsg, msg.GetTotalPoolLiquidityQueryMsgResponse](ctx, p.WasmKeeper, p.ContractAddress, request)
	return response.TotalPoolLiquidity
}

// types.CosmWasmExtension interface implementation

func (p Pool) GetCodeId() uint64 {
	return p.CodeId
}

func (p Pool) GetInstantiateMsg() []byte {
	return p.InstantiateMsg
}

func (p Pool) GetContractAddress() string {
	return p.ContractAddress
}

func (p *Pool) SetContractAddress(contractAddress string) {
	p.ContractAddress = contractAddress
}

func (p Pool) GetStoreModel() poolmanagertypes.PoolI {
	return &p.CosmWasmPool
}

// Set the wasm keeper.
func (p *Pool) SetWasmKeeper(wasmKeeper types.WasmKeeper) {
	p.WasmKeeper = wasmKeeper
}

// GetPoolDenoms implements types.PoolI.
func (p *Pool) GetPoolDenoms(ctx sdk.Context) []string {
	poolLiquidity := p.GetTotalPoolLiquidity(ctx)
	return poolLiquidity.Denoms()
}

func (p Pool) AsSerializablePool() poolmanagertypes.PoolI {
	return &CosmWasmPool{
		ContractAddress: p.ContractAddress,
		PoolId:          p.PoolId,
		CodeId:          p.CodeId,
		InstantiateMsg:  p.InstantiateMsg,
	}
}

func (p *CosmWasmPool) AsSerializablePool() poolmanagertypes.PoolI {
	return p
}
