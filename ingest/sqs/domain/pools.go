package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	clqueryproto "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/client/queryproto"
	concentratedmodel "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/model"
	cosmwasmpoolmodel "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v20/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v20/x/gamm/pool-models/stableswap"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// PoolI represents a generalized Pool interface.
type PoolI interface {
	json.Marshaler
	json.Unmarshaler

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

type serializePool struct {
	ChainModelData []byte                    `json:"chain_model"`
	SQSPoolData    SQSPool                   `json:"sqs_model"`
	TickModelData  *TickModel                `json:"tick_model,omitempty"`
	PoolType       poolmanagertypes.PoolType `json:"pool_type"`
}

// UnmarshalJSON implements PoolI.
func (p *PoolWrapper) UnmarshalJSON(data []byte) error {
	var serializedPool serializePool
	err := json.Unmarshal(data, &serializedPool)
	if err != nil {
		return err
	}

	switch serializedPool.PoolType {
	case poolmanagertypes.Concentrated:
		var concentratedPool concentratedmodel.Pool
		err = json.Unmarshal(serializedPool.ChainModelData, &concentratedPool)
		if err != nil {
			return err
		}
		p.ChainModel = &concentratedPool
	case poolmanagertypes.Balancer:
		var balancerPool balancer.Pool
		err = json.Unmarshal(serializedPool.ChainModelData, &balancerPool)
		if err != nil {
			return err
		}
		p.ChainModel = &balancerPool
	case poolmanagertypes.Stableswap:
		var stableswapPool stableswap.Pool
		err = json.Unmarshal(serializedPool.ChainModelData, &stableswapPool)
		if err != nil {
			return err
		}
		p.ChainModel = &stableswapPool
	case poolmanagertypes.CosmWasm:
		var cosmwasmPool cosmwasmpoolmodel.Pool
		err = json.Unmarshal(serializedPool.ChainModelData, &cosmwasmPool)
		if err != nil {
			return err
		}
		p.ChainModel = &cosmwasmPool
	default:
		return fmt.Errorf("invalid pool type (%d)", serializedPool.PoolType)
	}

	p.SQSModel = serializedPool.SQSPoolData
	p.TickModel = serializedPool.TickModelData

	return nil
}

// MarshalJSON implements PoolI.
func (p *PoolWrapper) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(p.ChainModel)
	if err != nil {
		return nil, err
	}

	var serializedPool serializePool
	serializedPool.ChainModelData = bytes
	serializedPool.PoolType = p.GetType()

	return json.Marshal(serializedPool)
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
	// ClearAllPools atomically clears all pools.
	ClearAllPools(ctx context.Context, tx Tx) error
}

// PoolsUsecase represent the pool's usecases
type PoolsUsecase interface {
	GetAllPools(ctx context.Context) ([]PoolI, error)
}
