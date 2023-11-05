package domain

import (
	"context"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	clqueryproto "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/client/queryproto"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
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

	// Validate validates the pool
	// Returns nil if the pool is valid
	// Returns error if the pool is invalid
	Validate(minUOSMOTVL osmomath.Int) error
}

type SQSPool struct {
	TotalValueLockedUSDC      osmomath.Int `json:"total_value_locked_uosmo"`
	IsErrorInTotalValueLocked bool         `json:"is_error_in_total_value_locked"`
	// Only CL and Cosmwasm pools need balances appended
	Balances   sdk.Coins `json:"balances,string"`
	PoolDenoms []string  `json:"pool_denoms"`
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

func NewPool(model poolmanagertypes.PoolI) PoolI {
	return &PoolWrapper{
		ChainModel: model,
	}
}

// GetId implements PoolI.
func (p *PoolWrapper) GetId() uint64 {
	return p.ChainModel.GetId()
}

// GetType implements PoolI.
func (p *PoolWrapper) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.PoolType(p.ChainModel.GetType())
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

func (p *PoolWrapper) Validate(minUOSMOTVL osmomath.Int) error {
	poolType := p.GetType()

	sqsModel := p.GetSQSPoolModel()

	if len(sqsModel.PoolDenoms) < 2 {
		return fmt.Errorf("pool (%d) has fewer than 2 denoms (%d)", p.GetId(), len(sqsModel.PoolDenoms))
	}

	// Note that balances are allowed to be zero because zero coins are filtered out.

	// Validate TVL
	if sqsModel.TotalValueLockedUSDC.LT(minUOSMOTVL) {
		return fmt.Errorf("pool (%d) has less than minimum tvl, pool tvl (%s), minimum tvl (%s)", p.GetId(), sqsModel.TotalValueLockedUSDC, minUOSMOTVL)
	}

	// Validate CL pools specifically
	if poolType == poolmanagertypes.Concentrated {
		tickModel, err := p.GetTickModel()

		if err != nil {
			return err
		}

		if tickModel.HasNoLiquidity {
			return fmt.Errorf("concentrated pool (%d) has no liquidity", p.GetId())
		}

		if tickModel.CurrentTickIndex < 0 {
			return fmt.Errorf("concentrated pool (%d) has invalid tick index (%d)", p.GetId(), tickModel.CurrentTickIndex)
		}

		if tickModel.CurrentTickIndex >= int64(len(tickModel.Ticks)) {
			return fmt.Errorf("concentrated pool (%d) has invalid tick index (%d) for ticks length (%d)", p.GetId(), tickModel.CurrentTickIndex, len(tickModel.Ticks))
		}

		if len(tickModel.Ticks) == 0 {
			return fmt.Errorf("concentrated pool (%d) has no ticks", p.GetId())
		}

		return nil
	} else {
		// Validate all pools other than concentrated
		if p.TickModel != nil {
			return fmt.Errorf("pool (%d) has tick model set but is not a concentrated pool, (%d)", p.GetId(), p.GetType())
		}
	}

	return nil
}

// PoolsRepository represent the pool's repository contract
type PoolsRepository interface {
	// GetAllPools atomically reads and returns all on-chain pools sorted by ID.
	GetAllPools(context.Context) ([]PoolI, error)
	// GetAllConcentrated atomically reads and returns concentrated pools sorted by ID.
	GetAllConcentrated(context.Context) ([]PoolI, error)
	// GetAllCFMM atomically reads and  returns CFMM pools sorted by ID.
	GetAllCFMM(context.Context) ([]PoolI, error)
	// GetAllCosmWasm atomically reads and returns CosmWasm pools sorted by ID.
	GetAllCosmWasm(context.Context) ([]PoolI, error)
	// StorePools atomically stores the given pools.
	StorePools(ctx context.Context, tx Tx, cfmmPools []PoolI, concentratedPools []PoolI, cosmwasmPools []PoolI) error
}

// PoolsUsecase represent the pool's usecases
type PoolsUsecase interface {
	GetAllPools(ctx context.Context) ([]PoolI, error)
}
