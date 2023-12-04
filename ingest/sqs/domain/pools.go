package domain

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	clqueryproto "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

// PoolI represents a generalized Pool interface.
type PoolI interface {
	// GetId returns the ID of the pool.
	GetId() uint64
	// GetType returns the type of the pool (Balancer, Stableswap, Concentrated, etc.)
	GetType() poolmanagertypes.PoolType

	GetTotalValueLockedUOSMO() osmomath.Int

	GetPoolDenoms() []string

	GetUnderlyingPool() poolmanagertypes.PoolI

	GetSQSPoolModel() SQSPool

	// GetTickModel returns the tick model for the pool
	// If this is a concentrated pool. Errors otherwise
	// Also errors if this is a concentrated pool but
	// the tick model is not set
	GetTickModel() (*TickModel, error)

	// SetTickModel sets the tick model for the pool
	// If this is not a concentrated pool, errors
	SetTickModel(*TickModel) error

	// Validate validates the pool
	// Returns nil if the pool is valid
	// Returns error if the pool is invalid
	Validate(minUOSMOTVL osmomath.Int) error
}

type SQSPool struct {
	TotalValueLockedUSDC  osmomath.Int `json:"total_value_locked_uosmo"`
	TotalValueLockedError string       `json:"total_value_locked_error,omitempty"`
	// Only CL and Cosmwasm pools need balances appended
	Balances     sdk.Coins    `json:"balances"`
	PoolDenoms   []string     `json:"pool_denoms"`
	SpreadFactor osmomath.Dec `json:"spread_factor"`
}

type LiquidityDepthsWithRange = clqueryproto.LiquidityDepthWithRange

type TickModel struct {
	Ticks            []LiquidityDepthsWithRange `json:"ticks,omitempty"`
	CurrentTickIndex int64                      `json:"current_tick_index,omitempty"`
	HasNoLiquidity   bool                       `json:"has_no_liquidity,omitempty"`
}

type PoolWrapper struct {
	ChainModel poolmanagertypes.PoolI `json:"underlying_pool"`
	SQSModel   SQSPool                `json:"sqs_model"`
	TickModel  *TickModel             `json:"tick_model,omitempty"`
}

var _ PoolI = &PoolWrapper{}

func NewPool(model poolmanagertypes.PoolI, spreadFactor osmomath.Dec, balances sdk.Coins) PoolI {
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

// GetTotalValueLockedUOSMO implements PoolI.
func (p *PoolWrapper) GetTotalValueLockedUOSMO() osmomath.Int {
	return p.SQSModel.TotalValueLockedUSDC
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

// SetTickModel implements PoolI.
func (p *PoolWrapper) SetTickModel(tickModel *TickModel) error {
	if p.GetType() != poolmanagertypes.Concentrated {
		return fmt.Errorf("pool (%d) is not a concentrated pool, type (%d)", p.GetId(), p.GetType())
	}

	p.TickModel = tickModel

	return nil
}

func (p *PoolWrapper) Validate(minUOSMOTVL osmomath.Int) error {
	sqsModel := p.GetSQSPoolModel()
	poolDenoms := p.GetPoolDenoms()

	if len(poolDenoms) < 2 {
		return fmt.Errorf("pool (%d) has fewer than 2 denoms (%d)", p.GetId(), len(poolDenoms))
	}

	// Note that balances are allowed to be zero because zero coins are filtered out.

	// Validate TVL
	if sqsModel.TotalValueLockedUSDC.LT(minUOSMOTVL) {
		return fmt.Errorf("pool (%d) has less than minimum tvl, pool tvl (%s), minimum tvl (%s)", p.GetId(), sqsModel.TotalValueLockedUSDC, minUOSMOTVL)
	}

	return nil
}
