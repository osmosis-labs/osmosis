package domain

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
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
}

type SQSPool struct {
	TotalValueLockedUSDC      osmomath.Int `json:"total_value_locked_usdc"`
	IsErrorInTotalValueLocked bool         `json:"is_error_in_total_value_locked"`
	// Only CL and Cosmwasm pools need balances appended
	Balances sdk.Coins `json:"balances,string,omitempty"`
}

type PoolWrapper struct {
	ChainModel poolmanagertypes.PoolI `json:"underlying_pool"`
	SQSModel   SQSPool                `json:"sqs_model"`
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
	denoms := make([]string, 0, len(p.SQSModel.Balances))
	for _, balance := range p.SQSModel.Balances {
		denoms = append(denoms, balance.Denom)
	}
	return denoms
}

// GetUnderlyingPool implements PoolI.
func (p *PoolWrapper) GetUnderlyingPool() poolmanagertypes.PoolI {
	return p.ChainModel
}

// GetSQSPoolModel implements PoolI.
func (p *PoolWrapper) GetSQSPoolModel() SQSPool {
	return p.SQSModel
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
	StorePools(ctx context.Context, cfmmPools []PoolI, concentratedPools []PoolI, cosmwasmPools []PoolI) error
}

// PoolsUsecase represent the pool's usecases
type PoolsUsecase interface {
	GetAllPools(ctx context.Context) ([]PoolI, error)
}

// RouterUsecase represent the router's usecases
type RouterUsecase interface {
	// GetOptimalQuote returns the optimal quote for the given tokenIn and tokenOutDenom.
	GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (Quote, error)
	// GetBestSingleRouteQuote returns the best single route quote for the given tokenIn and tokenOutDenom.
	GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (Quote, error)
}
