package redis

import (
	"context"
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/common"

	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	"github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/client/queryproto"
	cltypes "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
	concentratedtypes "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// poolIngester is an ingester for pools.
// It implements ingest.Ingester.
// It reads all pools from the state and writes them to the pools repository.
// As part of that, it instruments each pool with chain native balances and
// OSMO based TVL.
// NOTE:
// - TVL is calculated using spot price. TODO: use TWAP (https://app.clickup.com/t/86a182835)
// - TVL does not account for token precision. TODO: use assetlist for pulling token precision data
// (https://app.clickup.com/t/86a18287v)
// - If error in TVL calculation, TVL is set to the value that could be computed and the pool struct
// has a flag to indicate that there was an error in TVL calculation.
type poolIngester struct {
	poolsRepository    mvc.PoolsRepository
	routerRepository   mvc.RouterRepository
	tokensUseCase      domain.TokensUsecase
	repositoryManager  mvc.TxManager
	gammKeeper         common.PoolKeeper
	concentratedKeeper common.ConcentratedKeeper
	cosmWasmKeeper     common.CosmWasmPoolKeeper
	bankKeeper         common.BankKeeper
	protorevKeeper     common.ProtorevKeeper
	poolManagerKeeper  common.PoolManagerKeeper
	logger             log.Logger

	txDecoder sdk.TxDecoder

	hasFetchedInitialData bool

	routerConfig domain.RouterConfig
}

// denomRoutingInfo encapsulates the routing information for a pool.
// It has a pool ID of the pool that is paired with OSMO.
// It has a spot price from that pool with OSMO as the base asset.
type denomRoutingInfo struct {
	PoolID uint64
	Price  osmomath.BigDec
}

const (
	UOSMO          = "uosmo"
	uosmoPrecision = 6

	noTokenPrecisionErrorFmtStr = "error getting token precision %s"
	spotPriceErrorFmtStr        = "error calculating spot price for denom %s, %s"

	// placeholder value to disable route updates at the end of every block.
	routeIngestDisablePlaceholder = 0
)

var uosmoPrecisionBigDec = osmomath.NewBigDec(uosmoPrecision)

// NewPoolIngester returns a new pool ingester.
func NewPoolIngester(poolsRepository mvc.PoolsRepository, routerRepository mvc.RouterRepository, tokensUseCase domain.TokensUsecase, repositoryManager mvc.TxManager, routerConfig domain.RouterConfig, keepers common.SQSIngestKeepers, txDecoder sdk.TxDecoder) mvc.AtomicIngester {
	return &poolIngester{
		poolsRepository:    poolsRepository,
		routerRepository:   routerRepository,
		tokensUseCase:      tokensUseCase,
		repositoryManager:  repositoryManager,
		gammKeeper:         keepers.GammKeeper,
		concentratedKeeper: keepers.ConcentratedKeeper,
		cosmWasmKeeper:     keepers.CosmWasmPoolKeeper,
		bankKeeper:         keepers.BankKeeper,
		protorevKeeper:     keepers.ProtorevKeeper,
		poolManagerKeeper:  keepers.PoolManagerKeeper,
		routerConfig:       routerConfig,

		hasFetchedInitialData: false,
		txDecoder:             txDecoder,
	}
}

// ProcessBlock implements ingest.Ingester.
func (pi *poolIngester) ProcessBlock(ctx sdk.Context, tx mvc.Tx) error {
	return pi.processPoolState(ctx, tx)
}

var _ mvc.AtomicIngester = &poolIngester{}

// processPoolState processes the pool state. an
func (pi *poolIngester) processPoolState(ctx sdk.Context, tx mvc.Tx) error {
	goCtx := sdk.WrapSDKContext(ctx)

	concentratedPoolIDUpdateMap := make(map[uint64]struct{})

	// TODO: can be cached
	tokenPrecisionMap, err := pi.tokensUseCase.GetDenomPrecisions(goCtx)
	if err != nil {
		return err
	}

	// Create a map from denom to routable pool ID.
	denomToRoutablePoolIDMap := make(map[string]denomRoutingInfo)

	// Get all pools by type.

	// CFMM pools

	cfmmPools, err := pi.gammKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// Concentrated pools

	concentratedPools, err := pi.concentratedKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// CosmWasm pools

	cosmWasmPools, err := pi.cosmWasmKeeper.GetPoolsWithWasmKeeper(ctx)
	if err != nil {
		return err
	}

	denomPairToTakerFeeMap := make(map[domain.DenomPair]osmomath.Dec, 0)

	allPoolsParsed := make([]domain.PoolI, 0, len(cfmmPools)+len(concentratedPools)+len(cosmWasmPools))

	// Parse CFMM pool to the standard SQS types.
	for _, pool := range cfmmPools {
		// Parse CFMM pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, denomToRoutablePoolIDMap, denomPairToTakerFeeMap, tokenPrecisionMap)
		if err != nil {
			return err
		}

		allPoolsParsed = append(allPoolsParsed, pool)
	}

	for _, pool := range concentratedPools {
		// Updating concentrated pools is expensive. As a result, we only update them in the following cases:
		// - Initial fetch
		// - If concentrated pool state is updated
		// - TODO: every epoch.
		_, isConcentratedTickStateUpdated := concentratedPoolIDUpdateMap[pool.GetId()]

		// Update only on the initial fetch or if the tick state is updated
		if !pi.hasFetchedInitialData || isConcentratedTickStateUpdated {
			// Parse concentrated pool to the standard SQS types.
			pool, err := pi.convertPool(ctx, pool, denomToRoutablePoolIDMap, denomPairToTakerFeeMap, tokenPrecisionMap)
			if err != nil {
				return err
			}

			var tickModel *domain.TickModel

			// For CL pools, get the tick data
			tickData, currentTickIndex, err := pi.concentratedKeeper.GetTickLiquidityForFullRange(ctx, pool.GetId())
			// If there is no error, we set the tick model
			if err == nil {
				tickModel = &domain.TickModel{
					Ticks:            tickData,
					CurrentTickIndex: currentTickIndex,
				}
				// If there is no liquidity, we set the tick model to nil and update no liquidity flag
			} else if err != nil && errors.Is(err, concentratedtypes.RanOutOfTicksForPoolError{PoolId: pool.GetId()}) {
				tickModel = &domain.TickModel{
					Ticks:            []queryproto.LiquidityDepthWithRange{},
					CurrentTickIndex: -1,
					HasNoLiquidity:   true,
				}

				// On any other error, we return the error
			} else {
				return err
			}

			// Set tick model
			pool.SetTickModel(tickModel)

			allPoolsParsed = append(allPoolsParsed, pool)
		}
	}

	for _, pool := range cosmWasmPools {
		// Parse cosmwasm pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, denomToRoutablePoolIDMap, denomPairToTakerFeeMap, tokenPrecisionMap)
		if err != nil {
			return err
		}

		allPoolsParsed = append(allPoolsParsed, pool)
	}

	pi.logger.Info("ingesting pools to Redis", zap.Int64("height", ctx.BlockHeight()), zap.Int("num_cfmm", len(cfmmPools)), zap.Int("num_concentrated", len(concentratedPools)), zap.Int("num_cosmwasm", len(cosmWasmPools)))

	err = pi.poolsRepository.StorePools(goCtx, tx, allPoolsParsed)
	if err != nil {
		return err
	}

	// persist taker fees
	err = pi.persistTakerFees(ctx, tx, denomPairToTakerFeeMap)
	if err != nil {
		return err
	}

	// Update routes every RouteUpdateHeightInterval blocks unless RouteUpdateHeightInterval is 0.
	if pi.routerConfig.RouteUpdateHeightInterval > routeIngestDisablePlaceholder && ctx.BlockHeight()%int64(pi.routerConfig.RouteUpdateHeightInterval) == 0 {
		allPools := make([]domain.PoolI, 0, len(allPoolsParsed))

		pi.logger.Debug("getting routes for pools", zap.Int64("height", ctx.BlockHeight()))

		pi.updateRoutes(sdk.WrapSDKContext(ctx), tx, allPools, denomPairToTakerFeeMap)
	}

	pi.hasFetchedInitialData = true

	return nil
}

// updateRoutes updates the routes for all denom pairs in the taker fee map. The taker fee map value is unused.
// It returns a channel that is closed when all routes are updated.
// TODO: test
func (pi *poolIngester) updateRoutes(ctx context.Context, tx mvc.Tx, pools []domain.PoolI, denomPairToTakerFeeMap map[domain.DenomPair]osmomath.Dec) chan domain.DenomPair {
	// Initialize a channel that will be closed when all routes are updated.
	completionChan := make(chan domain.DenomPair, len(denomPairToTakerFeeMap))

	defer func() {
		// Close completion channel before returning.
		close(completionChan)
	}()

	for denomPair := range denomPairToTakerFeeMap {
		denomPair := denomPair
		// router
		router := routerusecase.NewRouter([]uint64{}, pi.routerConfig.MaxPoolsPerRoute, pi.routerConfig.MaxRoutes, pi.routerConfig.MaxSplitRoutes, pi.routerConfig.MaxSplitIterations, pi.routerConfig.MinOSMOLiquidity, pi.logger)
		router = routerusecase.WithSortedPools(router, pools)

		go func(denomPair domain.DenomPair) {
			// TODO: abstract this better

			candidateRoutes, err := router.GetCandidateRoutes(denomPair.Denom0, denomPair.Denom1)
			if err != nil {
				pi.logger.Error("error getting routes", zap.Error(err))
				return
			}

			err = pi.routerRepository.SetRoutesTx(ctx, tx, denomPair.Denom0, denomPair.Denom1, candidateRoutes)
			if err != nil {
				pi.logger.Error("error setting routes", zap.Error(err))
				return
			}

			// In the other direction. This can be optimized later.

			candidateRoutes, err = router.GetCandidateRoutes(denomPair.Denom1, denomPair.Denom0)
			if err != nil {
				pi.logger.Error("error getting routes", zap.Error(err))
				return
			}

			err = pi.routerRepository.SetRoutesTx(ctx, tx, denomPair.Denom1, denomPair.Denom0, candidateRoutes)
			if err != nil {
				pi.logger.Error("error setting routes", zap.Error(err))
				return
			}

			completionChan <- denomPair
		}(denomPair)
	}

	return completionChan
}

// convertPool converts a pool to the standard SQS pool type.
// It instruments the pool with chain native balances and OSMO based TVL.
// If error occurs in TVL estimation, it is silently skipped and the error flag
// set to true in the pool model.
// Note:
// - TVL is calculated using spot price. TODO: use TWAP (https://app.clickup.com/t/86a182835)
// - TVL does not account for token precision. TODO: use assetlist for pulling token precision data
// (https://app.clickup.com/t/86a18287v)
func (pi *poolIngester) convertPool(
	ctx sdk.Context,
	pool poolmanagertypes.PoolI,
	denomToRoutingInfoMap map[string]denomRoutingInfo,
	denomPairToTakerFeeMap domain.TakerFeeMap,
	tokenPrecisionMap map[string]int,
) (domain.PoolI, error) {
	balances := pi.bankKeeper.GetAllBalances(ctx, pool.GetAddress())

	osmoPoolTVL := osmomath.ZeroInt()

	poolDenoms := pool.GetPoolDenoms(ctx)
	poolDenomsMap := map[string]struct{}{}

	// Convert pool denoms to map
	for _, poolDenom := range poolDenoms {
		poolDenomsMap[poolDenom] = struct{}{}
	}

	spreadFactor := pool.GetSpreadFactor(ctx)

	// Note that this must follow the call to GetPoolDenoms() and GetSpreadFactor.
	// Otherwise, the CosmWasmPool model panics.
	pool = pool.AsSerializablePool()

	var errorInTVLStr string
	for _, balance := range balances {
		// Note that there are edge cases where gamm shares or some random
		// garbage tokens are in the balance that do not belong to the pool.
		// A mainnet example is pool ID 2 with the following extra denoms:
		// ibc/65BCD5909ED3D9E6223529017BC828ECBECCBE3F63D444EC44CE7412EF8C82D6
		// ibc/778F0504E33BBB66D0950FE12E29BA81C258ED0A10CCEF9CB0096BA9E22C5D61
		// As a result, we skilently skip them
		// TODO: cover with test
		_, exists := poolDenomsMap[balance.Denom]
		if !exists {
			continue
		}

		if balance.Denom == UOSMO {
			osmoPoolTVL = osmoPoolTVL.Add(balance.Amount)
			continue
		}

		// Check if routable poolID already exists for the denom
		routingInfo, ok := denomToRoutingInfoMap[balance.Denom]
		if !ok {
			poolForDenomPair, err := pi.protorevKeeper.GetPoolForDenomPair(ctx, UOSMO, balance.Denom)
			if err != nil {
				pi.logger.Debug("error getting OSMO-based pool", zap.String("denom", balance.Denom), zap.Error(err))
				errorInTVLStr = err.Error()
				continue
			}

			basePrecison, ok := tokenPrecisionMap[balance.Denom]
			if !ok {
				errorInTVLStr = fmt.Sprintf(noTokenPrecisionErrorFmtStr, balance.Denom)
				pi.logger.Debug(errorInTVLStr)
				continue
			}

			uosmoBaseAssetSpotPrice, err := pi.poolManagerKeeper.RouteCalculateSpotPrice(ctx, poolForDenomPair, balance.Denom, UOSMO)
			if err != nil {
				errorInTVLStr = fmt.Sprintf(spotPriceErrorFmtStr, balance.Denom, err)
				pi.logger.Debug(errorInTVLStr)
				continue
			}

			// Scale on-chain spot price to the correct token precision.
			precisionMultiplier := uosmoPrecisionBigDec.Quo(osmomath.NewBigDec(int64(basePrecison)))

			uosmoBaseAssetSpotPrice = uosmoBaseAssetSpotPrice.Mul(precisionMultiplier)

			routingInfo = denomRoutingInfo{
				PoolID: poolForDenomPair,
				Price:  uosmoBaseAssetSpotPrice,
			}
		}

		tvlAddition := osmomath.BigDecFromSDKInt(balance.Amount).QuoMut(routingInfo.Price).Dec().TruncateInt()
		osmoPoolTVL = osmoPoolTVL.Add(tvlAddition)
	}

	// Get pool denoms. Although these can be inferred from balances, this is safer.
	// If we used balances, for pools with no liquidity, we would not be able to get the denoms.
	denoms, err := pi.poolManagerKeeper.RouteGetPoolDenoms(ctx, pool.GetId())
	if err != nil {
		return nil, err
	}

	// Sort denoms for consistent ordering.
	sort.Strings(denoms)

	// Mutates denomPairToTakerFeeMap with the taker fee for every uniquer denom pair in the denoms list.
	err = retrieveTakerFeeToMapIfNotExists(ctx, denoms, denomPairToTakerFeeMap, pi.poolManagerKeeper)
	if err != nil {
		return nil, err
	}
	return &domain.PoolWrapper{
		ChainModel: pool,
		SQSModel: domain.SQSPool{
			TotalValueLockedUSDC:  osmoPoolTVL,
			TotalValueLockedError: errorInTVLStr,
			Balances:              balances,
			PoolDenoms:            denoms,
			SpreadFactor:          spreadFactor,
		},
	}, nil
}

// persistTakerFees persists all taker fees to the router repository.
func (pi *poolIngester) persistTakerFees(ctx sdk.Context, tx mvc.Tx, takerFeeMap domain.TakerFeeMap) error {
	for denomPair, takerFee := range takerFeeMap {
		err := pi.routerRepository.SetTakerFee(sdk.WrapSDKContext(ctx), tx, denomPair.Denom0, denomPair.Denom1, takerFee)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetLogger implements ingest.AtomicIngester.
func (pi *poolIngester) SetLogger(logger log.Logger) {
	pi.logger = logger
}

func (pi *poolIngester) parseTx(ctx sdk.Context) error {
	sdkTxBytes := ctx.TxBytes()

	// Decode the transaction
	sdkTx, err := pi.txDecoder(sdkTxBytes)
	if err != nil {
		return err
	}

	messages := sdkTx.GetMsgs()

	// map of pool IDs that had concentrated pool updates.
	// These include:
	// - new pool
	// - create position
	// - add to position
	// - withdraw position
	// - claim spread factor rewards
	// - claim incentive rewards
	// - swap
	concentratedPoolUpdates := make(map[uint64]struct{})

	for _, msg := range messages {
		switch msg.(type) {
		case *cltypes.MsgCreatePosition:

			createPositionMsg := msg.(*cltypes.MsgCreatePosition)
			poolID := createPositionMsg.PoolId

			concentratedPoolUpdates[poolID] = struct{}{}
		case *cltypes.MsgAddToPosition:

			addToPositionMsg := msg.(*cltypes.MsgAddToPosition)

			// Get the position
			position, err := pi.concentratedKeeper.GetPosition(ctx, addToPositionMsg.PositionId)
			if err != nil {
				return err
			}

			// Get the pool ID
			poolID := position.PoolId

			concentratedPoolUpdates[poolID] = struct{}{}
		case *cltypes.MsgWithdrawPosition:

			withdrawPositionMsg := msg.(*cltypes.MsgWithdrawPosition)

			// Get the position
			position, err := pi.concentratedKeeper.GetPosition(ctx, withdrawPositionMsg.PositionId)
			if err != nil {
				return err
			}

			// Get the pool ID
			poolID := position.PoolId

			// Updates both pool model and tick model
			concentratedPoolUpdates[poolID] = struct{}{}
		case *cltypes.MsgCollectIncentives:

			collectIncentivesMsg := msg.(*cltypes.MsgCollectIncentives)

			for _, positionID := range collectIncentivesMsg.PositionIds {
				// Get the position
				position, err := pi.concentratedKeeper.GetPosition(ctx, positionID)
				if err != nil {
					return err
				}

				// Get the pool ID
				poolID := position.PoolId

				// Updates only pool model
				concentratedPoolUpdates[poolID] = struct{}{}
			}
		case *cltypes.MsgCollectSpreadRewards:

			collectIncentivesMsg := msg.(*cltypes.MsgCollectSpreadRewards)

			collectIncentivesPoolIDs, err := pi.getPoolIDsFromPositionIDs(ctx, collectIncentivesMsg.PositionIds)
			if err != nil {
				return err
			}

			concentratedPoolUpdates = MergeMaps(concentratedPoolUpdates, collectIncentivesPoolIDs)

		case *poolmanagertypes.MsgSwapExactAmountIn:

			msgSwapExactAmountIn := msg.(*poolmanagertypes.MsgSwapExactAmountIn)

			for _, route := range msgSwapExactAmountIn.Routes {
				// Get the pool ID
				poolID := route.PoolId

				// Updates both
				concentratedPoolUpdates[poolID] = struct{}{}
			}

		case *poolmanagertypes.MsgSwapExactAmountOut:

		case *poolmanagertypes.MsgSplitRouteSwapExactAmountIn:

		case *poolmanagertypes.MsgSplitRouteSwapExactAmountOut:

		default:
			// do nothing
		}
	}

	return nil
}

// returns pool IDs associated with the given position IDs.
func (pi *poolIngester) getPoolIDsFromPositionIDs(ctx sdk.Context, positionIDs []uint64) (map[uint64]struct{}, error) {
	poolIDs := make(map[uint64]struct{})

	for _, positionID := range positionIDs {
		// Get the pool ID
		poolID, err := pi.getPoolIDFromPositionID(ctx, positionID)
		if err != nil {
			return nil, err
		}

		poolIDs[poolID] = struct{}{}
	}

	return poolIDs, nil
}

// returns pool ID associated with the given position ID.
func (pi *poolIngester) getPoolIDFromPositionID(ctx sdk.Context, positionID uint64) (uint64, error) {
	// Get the position
	position, err := pi.concentratedKeeper.GetPosition(ctx, positionID)
	if err != nil {
		return 0, err
	}

	// Get the pool ID
	poolID := position.PoolId

	return poolID, nil
}

// MergeMaps merges two maps and returns a new map containing the merged result.
// TODO: move to osmoutils.
func MergeMaps[K comparable, T any](map1, map2 map[K]T) map[K]T {
	result := make(map[K]T)

	// Copy values from the first map
	for key, value := range map1 {
		result[key] = value
	}

	// Copy values from the second map, overwriting existing keys
	for key, value := range map2 {
		result[key] = value
	}

	return result
}
