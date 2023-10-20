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

	GetTotalValueLockedUSDC() osmomath.Int

	GetPoolDenoms() []string
}

type Pool struct {
	UnderlyingPool       poolmanagertypes.PoolI `json:"pool"`
	TotalValueLockedUSDC osmomath.Int           `json:"total_value_locked_usdc"`
	// Only CL and Cosmwasm pools need balances appended
	Balances sdk.Coins `json:"balances,omitempty"`
}

var _ PoolI = &Pool{}

// GetId implements PoolI.
func (p *Pool) GetId() uint64 {
	return p.UnderlyingPool.GetId()
}

// GetType implements PoolI.
func (p *Pool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.PoolType(p.UnderlyingPool.GetType())
}

// GetTotalValueLockedUSDC implements PoolI.
func (p *Pool) GetTotalValueLockedUSDC() osmomath.Int {
	return p.TotalValueLockedUSDC
}

// GetPoolDenoms implements PoolI.
func (p *Pool) GetPoolDenoms() []string {
	denoms := make([]string, 0, len(p.Balances))
	for _, balance := range p.Balances {
		denoms = append(denoms, balance.Denom)
	}
	return denoms
}

// PoolsRepository represent the pool's repository contract
type PoolsRepository interface {
	// GetAllConcentrated returns concentrated pools sorted by ID.
	GetAllConcentrated(context.Context) ([]PoolI, error)
	// GetAllCFMM returns CFMM pools sorted by ID.
	GetAllCFMM(context.Context) ([]PoolI, error)
	// GetAllCosmWasm returns CosmWasm pools sorted by ID.
	GetAllCosmWasm(context.Context) ([]PoolI, error)

	// StoreConcentrated stores concentrated pools.
	// Returns error if any occurs when interacting with repository.
	StoreConcentrated(context.Context, []PoolI) error

	// StoreCFMM stores CFMM pools.
	// Returns error if any occurs when interacting with repository.
	StoreCFMM(context.Context, []PoolI) error

	// StoreCosmWasm stores CosmWasm pools.
	// Returns error if any occurs when interacting with repository.
	StoreCosmWasm(context.Context, []PoolI) error
}

// PoolsUsecase represent the pool's usecases
type PoolsUsecase interface {
	GetAllPools(ctx context.Context) ([]PoolI, error)
}
