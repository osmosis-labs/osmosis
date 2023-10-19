package domain

import (
	"context"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// PoolI represents a generalized Pool interface.
type PoolI interface {
	// GetId returns the ID of the pool.
	GetId() uint64
	// GetType returns the type of the pool (Balancer, Stableswap, Concentrated, etc.)
	GetType() poolmanagertypes.PoolType

	GetLiquidity() string

	GetSpreadFactor() string

	GetDenoms() []string

	GetWeights() []string
}

type Pool struct {
	Id           uint64   `json:"id"`
	Type         int      `json:"type"`
	SpreadFactor string   `json:"spread_factor"`
	Denoms       []string `json:"pool_denoms"`
	Balances     string   `json:"balances"`
	Liquidity    string   `json:"liquidity,omitempty"`
	Weights      []string `json:"weights,omitempty"`
}

var _ PoolI = &Pool{}

// GetId implements PoolI.
func (p *Pool) GetId() uint64 {
	return p.Id
}

// GetType implements PoolI.
func (p *Pool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.PoolType(p.Type)
}

// GetDenoms implements PoolI.
func (p *Pool) GetDenoms() []string {
	return p.Denoms
}

// GetLiquidity implements PoolI.
func (p *Pool) GetLiquidity() string {
	return p.Liquidity
}

// GetSpreadFactor implements PoolI.
func (p *Pool) GetSpreadFactor() string {
	return p.SpreadFactor
}

// GetWeights implements PoolI.
func (p *Pool) GetWeights() []string {
	return p.Weights
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
