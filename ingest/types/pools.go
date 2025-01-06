package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sort"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v28/ingest/types/cosmwasmpool"
	"github.com/osmosis-labs/osmosis/v28/ingest/types/passthroughdomain"

	clqueryproto "github.com/osmosis-labs/osmosis/v28/x/concentrated-liquidity/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v28/x/poolmanager/types"
)

// PoolI represents a generalized Pool interface.
type PoolI interface {
	// GetId returns the ID of the pool.
	GetId() uint64
	// GetType returns the type of the pool (Balancer, Stableswap, Concentrated, etc.)
	GetType() poolmanagertypes.PoolType

	GetPoolLiquidityCap() osmomath.Int

	GetPoolDenoms() []string

	GetUnderlyingPool() poolmanagertypes.PoolI

	GetSQSPoolModel() SQSPool

	// GetTickModel returns the tick model for the pool
	// if this is a concentrated pool. Errors otherwise
	// Also errors if this is a concentrated pool but
	// the tick model is not set
	GetTickModel() (*TickModel, error)
}

type LiquidityDepthsWithRange = clqueryproto.LiquidityDepthWithRange

type TickModel struct {
	Ticks            []LiquidityDepthsWithRange `json:"ticks,omitempty"`
	CurrentTickIndex int64                      `json:"current_tick_index,omitempty"`
	HasNoLiquidity   bool                       `json:"has_no_liquidity,omitempty"`
}

type SQSPool struct {
	PoolLiquidityCap      osmomath.Int `json:"pool_liquidity_cap"`
	PoolLiquidityCapError string       `json:"pool_liquidity_error,omitempty"`
	// Only CL and Cosmwasm pools need balances appended
	Balances     sdk.Coins    `json:"balances"`
	PoolDenoms   []string     `json:"pool_denoms"`
	SpreadFactor osmomath.Dec `json:"spread_factor"`

	// Only CosmWasm pools need CosmWasmPoolModel appended
	CosmWasmPoolModel *cosmwasmpool.CosmWasmPoolModel `json:"cosmwasm_pool_model,omitempty"`
}

type PoolWrapper struct {
	ChainModel poolmanagertypes.PoolI                   `json:"underlying_pool"`
	SQSModel   SQSPool                                  `json:"sqs_model"`
	APRData    passthroughdomain.PoolAPRDataStatusWrap  `json:"apr_data,omitempty"`
	FeesData   passthroughdomain.PoolFeesDataStatusWrap `json:"fees_data,omitempty"`
	TickModel  *TickModel                               `json:"tick_model,omitempty"`
}

var _ PoolI = &PoolWrapper{}

func NewPool(model poolmanagertypes.PoolI, spreadFactor osmomath.Dec, balances sdk.Coins) *PoolWrapper {
	return &PoolWrapper{
		ChainModel: model,
		SQSModel: SQSPool{
			SpreadFactor: spreadFactor,
			Balances:     balances,
		},
	}
}

// GetId implements PoolI.
func (p *PoolWrapper) GetId() uint64 {
	return p.ChainModel.GetId()
}

// GetType implements PoolI.
func (p *PoolWrapper) GetType() poolmanagertypes.PoolType {
	return p.ChainModel.GetType()
}

// GetPoolLiquidityCap implements PoolI.
func (p *PoolWrapper) GetPoolLiquidityCap() osmomath.Int {
	return p.SQSModel.PoolLiquidityCap
}

// GetPoolDenoms implements PoolI.
func (p *PoolWrapper) GetPoolDenoms() []string {
	// sort pool denoms
	sort.Strings(p.SQSModel.PoolDenoms)
	return p.SQSModel.PoolDenoms
}

// GetUnderlyingPool implements PoolI.
func (p *PoolWrapper) GetUnderlyingPool() poolmanagertypes.PoolI {
	return p.ChainModel
}

// GetSQSPoolModel implements PoolI.
func (p *PoolWrapper) GetSQSPoolModel() SQSPool {
	return p.SQSModel
}

// GetTickModel implements PoolI.
func (p *PoolWrapper) GetTickModel() (*TickModel, error) {
	if p.GetType() != poolmanagertypes.Concentrated {
		return nil, fmt.Errorf("pool (%d) is not a concentrated pool, type (%d)", p.GetId(), p.GetType())
	}

	if p.TickModel == nil {
		return nil, ConcentratedPoolNoTickModelError{PoolId: p.GetId()}
	}

	return p.TickModel, nil
}
