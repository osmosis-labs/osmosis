package poolstransformer

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ingesttypes "github.com/osmosis-labs/osmosis/v28/ingest/types"
	sqscosmwasmpool "github.com/osmosis-labs/osmosis/v28/ingest/types/cosmwasmpool"

	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v28/x/cosmwasmpool/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"

	appparams "github.com/osmosis-labs/osmosis/v28/app/params"
	"github.com/osmosis-labs/osmosis/v28/x/concentrated-liquidity/client/queryproto"
	concentratedtypes "github.com/osmosis-labs/osmosis/v28/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v28/x/poolmanager/types"
)

// poolTransformer is a transformer for pools.
// It instruments each pool with chain native balances and
// OSMO based TVL.
// NOTE:
// - TVL is calculated using spot price. TODO: use TWAP (https://app.clickup.com/t/86a182835)
// - If error in TVL calculation, TVL is set to the value that could be computed and the pool struct
// has a flag to indicate that there was an error in TVL calculation.
type poolTransformer struct {
	gammKeeper         commondomain.PoolKeeper
	concentratedKeeper commondomain.ConcentratedKeeper
	cosmWasmKeeper     commondomain.CosmWasmPoolKeeper
	wasmKeeper         commondomain.WasmKeeper
	bankKeeper         commondomain.BankKeeper
	protorevKeeper     commondomain.ProtorevKeeper
	poolManagerKeeper  commondomain.PoolManagerKeeper

	// Pool ID that is used for converting between USDC and UOSMO.
	defaultUSDCUOSMOPoolID uint64
}

const (
	UOSMO         = appparams.BaseCoinUnit
	usdcPrecision = 6

	spotPriceErrorFmtStr          = "error calculating spot price for denom %s, %s"
	spotPricePrecisionErrorFmtStr = "error calculating spot price from route overwrites due to precision for denom %s"
	multiHopSpotPriceErrorFmtStr  = "error calculating spot price via multihop swap, %s"
	// Empty string placeholder for no pool liquidity capitalization error
	noPoolLiquidityCapError = ""

	// placeholder value to disable route updates at the end of every block.
	// nolint: unused
	routeIngestDisablePlaceholder = 0

	usdcDenom       = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
	stATOMDenom     = "ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901"
	atomDenom       = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	usdtDenom       = "ibc/2108F2D81CBE328F371AD0CEF56691B18A86E08C3651504E42487D9EE92DDE9C"
	oneOSMO         = 1_000_000
	contractInfoKey = "contract_info"
)

var (
	oneOsmoInt    = osmomath.NewInt(oneOSMO)
	oneOsmoBigDec = osmomath.NewBigDec(oneOSMO)

	oneOsmoCoin                = sdk.NewCoin(appparams.BaseCoinUnit, oneOsmoInt)
	usdcPrecisionScalingFactor = osmomath.NewBigDec(10).PowerIntegerMut(usdcPrecision)
)

// These are the routes that we use for pricing certain tokens against OSMO
// for determining TVL.
var uosmoRoutesFromDenom map[string][]poolmanagertypes.SwapAmountOutRoute = map[string][]poolmanagertypes.SwapAmountOutRoute{
	// stATOM
	stATOMDenom: {
		{
			PoolId: 1283,

			// stAtom
			TokenInDenom: stATOMDenom,
		},
		{
			PoolId: 1265,
			// ATOM
			TokenInDenom: atomDenom,
		},
	},
}

// For stablecoin denoms that we do not have routes set for calculating TVL, we simply assume
// that they have the same spot price as USDC. This is sufficient for correctness of naive ranking.
var stablesOverwrite map[string]struct{} = map[string]struct{}{
	// Tether USD (Wormhole)
	usdtDenom: {},
	usdcDenom: {},
}

var _ domain.PoolsTransformer = &poolTransformer{}

// NewPoolTransformer returns a new pool ingester.
func NewPoolTransformer(keepers commondomain.PoolExtractorKeepers, defaultUSDCUOSMOPoolID uint64) domain.PoolsTransformer {
	return &poolTransformer{
		gammKeeper:         keepers.GammKeeper,
		concentratedKeeper: keepers.ConcentratedKeeper,
		cosmWasmKeeper:     keepers.CosmWasmPoolKeeper,
		wasmKeeper:         keepers.WasmKeeper,
		bankKeeper:         keepers.BankKeeper,
		protorevKeeper:     keepers.ProtorevKeeper,
		poolManagerKeeper:  keepers.PoolManagerKeeper,

		defaultUSDCUOSMOPoolID: defaultUSDCUOSMOPoolID,
	}
}

// processPoolState processes the pool state. an
func (pi *poolTransformer) Transform(ctx sdk.Context, blockPools commondomain.BlockPools) ([]ingesttypes.PoolI, ingesttypes.TakerFeeMap, error) {
	// Create a map from denom to its price.
	priceInfoMap := make(map[string]osmomath.BigDec)

	denomPairToTakerFeeMap := make(map[ingesttypes.DenomPair]osmomath.Dec, 0)

	// Get all pools
	cfmmPools := blockPools.CFMMPools
	concentratedPools := blockPools.ConcentratedPools
	cosmWasmPools := blockPools.CosmWasmPools

	allPoolsParsed := make([]ingesttypes.PoolI, 0, len(cfmmPools)+len(concentratedPools)+len(cosmWasmPools))

	// Parse CFMM pool to the standard SQS types.
	for _, pool := range cfmmPools {
		// Parse CFMM pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, priceInfoMap, denomPairToTakerFeeMap)
		if err != nil {
			// Silently skip pools on error to avoid breaking ingest of all other pools.
			continue
		}

		allPoolsParsed = append(allPoolsParsed, pool)
	}

	for _, pool := range concentratedPools {
		// Parse concentrated pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, priceInfoMap, denomPairToTakerFeeMap)
		if err != nil {
			// Silently skip pools on error to avoid breaking ingest of all other pools.
			continue
		}

		allPoolsParsed = append(allPoolsParsed, pool)
	}

	for _, pool := range cosmWasmPools {
		// Parse cosmwasm pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, priceInfoMap, denomPairToTakerFeeMap)
		if err != nil {
			// Silently skip pools on error to avoid breaking ingest of all other pools.
			continue
		}

		allPoolsParsed = append(allPoolsParsed, pool)
	}

	ctx.Logger().Info("finish extracting pools", "height", ctx.BlockHeight(), "num_cfmm", len(cfmmPools), "num_concentrated", len(concentratedPools), "num_cosmwasm", len(cosmWasmPools))

	return allPoolsParsed, denomPairToTakerFeeMap, nil
}

// convertPool converts a pool to the standard SQS pool type.
// It instruments the pool with chain native balances and OSMO based TVL.
// If error occurs in TVL estimation, it is silently skipped and the error flag
// set to true in the pool model.
// Note:
// - TVL is calculated using spot price. TODO: use TWAP (https://app.clickup.com/t/86a182835)
// - TVL does not account for token precision.
// (https://app.clickup.com/t/86a18287v)
func (pi *poolTransformer) convertPool(
	ctx sdk.Context,
	pool poolmanagertypes.PoolI,
	denomPriceInfoMap map[string]osmomath.BigDec,
	denomPairToTakerFeeMap ingesttypes.TakerFeeMap,
) (sqsPool ingesttypes.PoolI, err error) {
	defer func() {
		r := recover()
		if r != nil {
			telemetry.IncrCounter(1, "sqs_ingest_convert_pool_panic", "pool_"+strconv.FormatUint(pool.GetId(), 10))

			err = fmt.Errorf("sqs ingest pool (%d) conversion panicked: %v", pool.GetId(), r)
			ctx.Logger().Error(err.Error())
		}
	}()

	balances := pi.bankKeeper.GetAllBalances(ctx, pool.GetAddress())

	// Convert pool denoms to map for faster lookup.
	poolDenomsMap := getPoolDenomsMap(pool.GetPoolDenoms(ctx))

	spreadFactor := pool.GetSpreadFactor(ctx)

	// Get pool denoms. Although these can be inferred from balances, this is safer.
	// If we used balances, for pools with no liquidity, we would not be able to get the denoms.
	denoms, err := pi.poolManagerKeeper.RouteGetPoolDenoms(ctx, pool.GetId())
	if err != nil {
		return nil, err
	}

	var cosmWasmPoolModel *sqscosmwasmpool.CosmWasmPoolModel
	if pool.GetType() == poolmanagertypes.CosmWasm {
		poolId := pool.GetId()
		poolAddress := pool.GetAddress()

		cwPool, ok := pool.(cosmwasmpooltypes.CosmWasmExtension)
		if !ok {
			return nil, fmt.Errorf("pool (%d) with type (%d) is not a CosmWasmExtension", poolId, pool.GetType())
		}

		balances = cwPool.GetTotalPoolLiquidity(ctx)

		// Sort balances for consistency with `sdk.Coins` assumptions.
		// For example, finding a denom by name using binary search requires coins to be sorted
		balances.Sort()

		// This must never happen, but if it does, and there is no checks, the query will fail silently.
		// We make sure to return an error here.
		if pi.wasmKeeper == nil {
			return nil, fmt.Errorf("pool (%d) with type (%d) requires `poolTransformer` to have `wasmKeeper` but got `nil`", poolId, pool.GetType())
		}

		initedCosmWasmPoolModel := pi.initCosmWasmPoolModel(ctx, pool)
		cosmWasmPoolModel = &initedCosmWasmPoolModel

		// special transformation based on different cw pool
		if cosmWasmPoolModel.IsAlloyTransmuter() {
			err = pi.updateAlloyTransmuterInfo(ctx, poolId, poolAddress, cosmWasmPoolModel, &denoms)
			if err != nil {
				return nil, err
			}
		} else if cosmWasmPoolModel.IsOrderbook() {
			err = pi.updateOrderbookInfo(ctx, poolId, poolAddress, cosmWasmPoolModel)
			if err != nil {
				return nil, err
			}
		}
	}

	// Note that this must follow the call to GetPoolDenoms() and GetSpreadFactor.
	// Otherwise, the CosmWasmPool model panics.
	pool = pool.AsSerializablePool()

	// filtered balances consisting only of the pool denom balances.
	filteredBalances := filterBalances(balances, poolDenomsMap)

	// Compute pool liquidity capitalization in UOSMO.
	poolLiquidityCapUOSMO, poolLiquidityCapErrorStr := pi.computeUOSMOPoolLiquidityCap(ctx, filteredBalances, denomPriceInfoMap)

	// Convert pool liquidity capitalization from UOSMO to USDC.
	poolLiquidityCapUSDC, poolLiquidityCapUSDCErrorStr := pi.computeUSDCPoolLiquidityCapFromUOSMO(ctx, poolLiquidityCapUOSMO)

	// Join error strings for pool liquidity cap.
	poolLiquidityCapErrorStr = strings.Join([]string{poolLiquidityCapErrorStr, poolLiquidityCapUSDCErrorStr}, " ")

	// Sort denoms for consistent ordering.
	sort.Strings(denoms)

	// Mutates denomPairToTakerFeeMap with the taker fee for every uniquer denom pair in the denoms list.
	err = retrieveTakerFeeToMapIfNotExists(ctx, denoms, denomPairToTakerFeeMap, pi.poolManagerKeeper)
	if err != nil {
		return nil, err
	}

	// Get the tick model for concentrated pools
	var tickModel *ingesttypes.TickModel

	// For CL pools, get the tick data
	if pool.GetType() == poolmanagertypes.Concentrated {
		tickData, currentTickIndex, err := pi.concentratedKeeper.GetTickLiquidityForFullRange(ctx, pool.GetId())
		// If there is no error, we set the tick model
		if err == nil {
			tickModel = &ingesttypes.TickModel{
				Ticks:            tickData,
				CurrentTickIndex: currentTickIndex,
			}
			// If there is no liquidity, we set the tick model to nil and update no liquidity flag
		} else if errors.Is(err, concentratedtypes.RanOutOfTicksForPoolError{PoolId: pool.GetId()}) {
			tickModel = &ingesttypes.TickModel{
				Ticks:            []queryproto.LiquidityDepthWithRange{},
				CurrentTickIndex: -1,
				HasNoLiquidity:   true,
			}

			// On any other error, we return the error
		} else {
			return nil, err
		}
	}

	return &ingesttypes.PoolWrapper{
		ChainModel: pool,
		SQSModel: ingesttypes.SQSPool{
			PoolLiquidityCap:      poolLiquidityCapUSDC,
			PoolLiquidityCapError: poolLiquidityCapErrorStr,
			Balances:              filteredBalances,
			PoolDenoms:            denoms,
			SpreadFactor:          spreadFactor,
			CosmWasmPoolModel:     cosmWasmPoolModel,
		},
		TickModel: tickModel,
	}, nil
}

// getPoolDenomsMap converts pool denoms to a map for faster lookup.
func getPoolDenomsMap(poolDenoms []string) map[string]struct{} {
	poolDenomsMap := make(map[string]struct{}, len(poolDenoms))

	// Convert pool denoms to map
	for _, poolDenom := range poolDenoms {
		poolDenomsMap[poolDenom] = struct{}{}
	}
	return poolDenomsMap
}

// filterBalances filters out balances that do not belong to the pool.
// Note that there are edge cases where gamm shares or some random
// garbage tokens are in the balance that do not belong to the pool.
// A mainnet example is pool ID 2 with the following extra denoms:
// ibc/65BCD5909ED3D9E6223529017BC828ECBECCBE3F63D444EC44CE7412EF8C82D6
// ibc/778F0504E33BBB66D0950FE12E29BA81C258ED0A10CCEF9CB0096BA9E22C5D61
// As a result, we skilently skip them
func filterBalances(originalBalances sdk.Coins, poolDenomsMap map[string]struct{}) sdk.Coins {
	// filtered balances consisting only of the pool denom balances.
	filteredBalances := sdk.Coins{}

	for _, balance := range originalBalances {
		_, exists := poolDenomsMap[balance.Denom]
		if !exists {
			continue
		}

		filteredBalances = append(filteredBalances, balance)
	}

	return filteredBalances
}

// computeUOSMOPoolLiquidityCap computes the pool liquidity cap in UOSMO.
// For each denom balance has the following cases:
// 1. The balance is UOSMO. In that case, it is added to the total.
// 2. Routing information is present in priceInfoMap for the denom. In that case, the spot price is used to convert the balance to UOSMO.
// 3. If there is no routing information, we attempt to get a single-hop pool from on-chain routes.
// 4. If there is no on-chain route, we check if there is a route overwrite for the denom.
// 5. If the denom is a stablecoin, we assume that it has the same price as USDC and use USDC routing information.
// 6. If there is no method to compute pool liquidity cap for this denom, we silently skip it and return a non-empty error string.
// The routing information is updated in the cases where it was not present before calling this function.
// Returns the pool liquidity cap in UOSMO and an error string if there was an error in computing the pool liquidity cap.
func (pi *poolTransformer) computeUOSMOPoolLiquidityCap(ctx sdk.Context, balances sdk.Coins, denomPriceMap map[string]osmomath.BigDec) (osmomath.Int, string) {
	poolLiquidityCap := osmomath.ZeroInt()
	var poolLiquidityCapErrorStr string

	for _, balance := range balances {
		if balance.Denom == UOSMO {
			poolLiquidityCap = poolLiquidityCap.Add(balance.Amount)
			continue
		}

		// Check if spot price is already computed for a denom
		// spot price with uosmo as base asset.
		uosmoBaseAssetSpotPrice, ok := denomPriceMap[balance.Denom]
		if !ok {
			// Attempt to get a single-hop pool from on-chain routes.
			poolForDenomPair, err := pi.protorevKeeper.GetPoolForDenomPair(ctx, UOSMO, balance.Denom)
			if err == nil {
				// If on-chain route is present, calculate spot price with uosmo.
				uosmoBaseAssetSpotPrice, err = pi.poolManagerKeeper.RouteCalculateSpotPrice(ctx, poolForDenomPair, balance.Denom, UOSMO)
				if err != nil {
					poolLiquidityCapErrorStr = fmt.Sprintf(spotPriceErrorFmtStr, balance.Denom, err)
					ctx.Logger().Debug(poolLiquidityCapErrorStr)
					continue
				}
			} else {
				ctx.Logger().Debug("error getting OSMO-based pool from Skip route", "denom", balance.Denom, "error", err)

				// Check if there exists a route from current denom to uosmo.
				routes, hasRouteOverwrite := uosmoRoutesFromDenom[balance.Denom]

				// Check if this is a stablecoin
				_, isStableCoin := stablesOverwrite[balance.Denom]

				if hasRouteOverwrite {
					ctx.Logger().Debug("uosmo routes are present", "denom", balance.Denom)

					// Estimate how many tokens in we get for 1 OSMO
					denomAmtIn, err := pi.poolManagerKeeper.MultihopEstimateInGivenExactAmountOut(ctx, routes, oneOsmoCoin)
					if err != nil {
						ctx.Logger().Debug("error computing multihop from route overwrite", "denom", balance.Denom, "error", err)
						poolLiquidityCapErrorStr = fmt.Sprintf(multiHopSpotPriceErrorFmtStr, err)
						continue
					}

					denomBigDecAmtIn := osmomath.BigDecFromSDKInt(denomAmtIn)
					if denomBigDecAmtIn.IsZero() {
						ctx.Logger().Info("error inverting price from route overwrite", "denom", balance.Denom)
						poolLiquidityCapErrorStr = "error inverting price from route overwrite"
						continue
					}

					uosmoBaseAssetSpotPrice = oneOsmoBigDec.QuoMut(denomBigDecAmtIn)
				} else if isStableCoin {
					// We (very) naively assume that stablecoin has the same price as USDC for TVL ranking of pools in the router.
					uosmoBaseAssetSpotPrice, err = pi.poolManagerKeeper.RouteCalculateSpotPrice(ctx, pi.defaultUSDCUOSMOPoolID, usdcDenom, UOSMO)
					if err != nil {
						poolLiquidityCapErrorStr = fmt.Sprintf(spotPriceErrorFmtStr, balance.Denom, err)
						ctx.Logger().Debug(poolLiquidityCapErrorStr)
						continue
					}
				} else {
					// If there is no method to compute pool liquidity cap for this denom, attach error and silently skip it.
					poolLiquidityCapErrorStr = err.Error()
					ctx.Logger().Debug("no overwrite present", "denom", balance.Denom)
					continue
				}
			}

			if uosmoBaseAssetSpotPrice.IsZero() {
				poolLiquidityCapErrorStr = "failed to calculate spot price due to it becoming zero from truncations " + balance.Denom
				continue
			}
		}

		liquidityCapContribution := osmomath.BigDecFromSDKInt(balance.Amount).QuoMut(uosmoBaseAssetSpotPrice).Dec().TruncateInt()
		poolLiquidityCap = poolLiquidityCap.Add(liquidityCapContribution)
	}

	return poolLiquidityCap, poolLiquidityCapErrorStr
}

// computeUSDCPoolLiquidityCapFromUOSMO computes the pool liquidity cap in USDC from UOSMO.
// If the pool liquidity cap in UOSMO is zero, it returns zero.
// Otherwise, it calculates the spot price from UOSMO to USDC and converts the pool liquidity cap to USDC.
// Returns the pool liquidity cap in USDC and an error string if there was an error in computing the pool liquidity cap.
// If there was an error in computing the spot price, the error string is set to the error message and zero is returned.
func (pi *poolTransformer) computeUSDCPoolLiquidityCapFromUOSMO(ctx sdk.Context, poolLiquidityCapUOSMO osmomath.Int) (osmomath.Int, string) {
	if !poolLiquidityCapUOSMO.IsZero() {
		usdcQuotePrice, err := pi.poolManagerKeeper.RouteCalculateSpotPrice(ctx, pi.defaultUSDCUOSMOPoolID, usdcDenom, UOSMO)
		if err != nil {
			// Note: should never happen in practice.
			poolLiquidityCapErrorStr := fmt.Sprintf(spotPriceErrorFmtStr, usdcDenom, err)
			ctx.Logger().Debug(poolLiquidityCapErrorStr)
			return osmomath.ZeroInt(), poolLiquidityCapErrorStr
		} else {
			poolLiquidityCapUSDCScaled := osmomath.BigDecFromSDKInt(poolLiquidityCapUOSMO).MulMut(usdcQuotePrice)

			// Apply exponent
			// If truncation occurs, the real value is insignificant and we can ignore it.
			poolLiquidityCapUSDC := poolLiquidityCapUSDCScaled.QuoMut(usdcPrecisionScalingFactor)

			// Note, we round up so that pools that have non-zero liquidity get propagated to the router
			// and reflect this context in the pool liquidity cap filtering. Otherwise, pools with zero liquidity get filtered out at the ingest level
			// completely, breaking our edge case tests for supporting low liquidity routes.
			return poolLiquidityCapUSDC.Dec().Ceil().TruncateInt(), noPoolLiquidityCapError
		}
	}

	return poolLiquidityCapUOSMO, noPoolLiquidityCapError
}

// queryContractInfo queries the cw2 contract info from the given contract address.
func (pi *poolTransformer) queryContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) (sqscosmwasmpool.ContractInfo, error) {
	bz := pi.wasmKeeper.QueryRaw(ctx, contractAddress, []byte(contractInfoKey))
	if len(bz) == 0 {
		return sqscosmwasmpool.ContractInfo{}, fmt.Errorf("contract info not found: %s", contractAddress)
	} else {
		var contractInfo sqscosmwasmpool.ContractInfo
		if err := json.Unmarshal(bz, &contractInfo); err != nil {
			return sqscosmwasmpool.ContractInfo{}, fmt.Errorf("error unmarshalling contract info: %w", err)
		} else {
			return contractInfo, nil
		}
	}
}

// initCosmWasmPoolModel initialize the CosmWasmPoolModel with the contract info of the given pool.
// If the contract info is not found, it logs the error and continues since it's not required for the pool to conform cw2.
func (pi *poolTransformer) initCosmWasmPoolModel(
	ctx sdk.Context,
	pool poolmanagertypes.PoolI,
) sqscosmwasmpool.CosmWasmPoolModel {
	contractInfo, err := pi.queryContractInfo(ctx, pool.GetAddress())
	if err != nil {
		// only log since cw pool contracts are not required to conform cw2
		ctx.Logger().Info(
			"CosmWasm pool does not conform cw2",
			"pool_id", pool.GetId(),
			"contract_address", pool.GetAddress(),
			"err", err.Error(),
		)

		return sqscosmwasmpool.CosmWasmPoolModel{}
	} else {
		// initialize the CosmWasmPoolModel with the contract info
		return sqscosmwasmpool.CosmWasmPoolModel{
			ContractInfo: contractInfo,
		}
	}
}
