package poolstransformer

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/sqsdomain"

	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"

	appparams "github.com/osmosis-labs/osmosis/v25/app/params"
	"github.com/osmosis-labs/osmosis/v25/x/concentrated-liquidity/client/queryproto"
	concentratedtypes "github.com/osmosis-labs/osmosis/v25/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

// poolTransformer is a transformer for pools.
// It instruments each pool with chain native balances and
// OSMO based TVL.
// NOTE:
// - TVL is calculated using spot price. TODO: use TWAP (https://app.clickup.com/t/86a182835)
// - If error in TVL calculation, TVL is set to the value that could be computed and the pool struct
// has a flag to indicate that there was an error in TVL calculation.
type poolTransformer struct {
	gammKeeper         domain.PoolKeeper
	concentratedKeeper domain.ConcentratedKeeper
	cosmWasmKeeper     domain.CosmWasmPoolKeeper
	wasmKeeper         domain.WasmKeeper
	bankKeeper         domain.BankKeeper
	protorevKeeper     domain.ProtorevKeeper
	poolManagerKeeper  domain.PoolManagerKeeper
	assetListGetter    domain.AssetListGetter
}

// denomRoutingInfo encapsulates the routing information for a pool.
// It has a pool ID of the pool that is paired with OSMO.
// It has a spot price from that pool with OSMO as the base asset.
type denomRoutingInfo struct {
	PoolID uint64
	Price  osmomath.BigDec
}

const (
	UOSMO          = appparams.BaseCoinUnit
	uosmoPrecision = 6

	noTokenPrecisionErrorFmtStr   = "error getting token precision %s"
	spotPriceErrorFmtStr          = "error calculating spot price for denom %s, %s"
	spotPricePrecisionErrorFmtStr = "error calculating spot price from route overwrites due to precision for denom %s"
	multiHopSpotPriceErrorFmtStr  = "error calculating spot price via multihop swap, %s"

	// placeholder value to disable route updates at the end of every block.
	// nolint: unused
	routeIngestDisablePlaceholder = 0

	// https://app.osmosis.zone/pool/1263
	usdcPool    = 1263
	usdcDenom   = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
	stATOMDenom = "ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901"
	atomDenom   = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	usdtDenom   = "ibc/2108F2D81CBE328F371AD0CEF56691B18A86E08C3651504E42487D9EE92DDE9C"
	oneOSMO     = 1_000_000
)

var (
	uosmoPrecisionBigDec = osmomath.NewBigDec(uosmoPrecision)
	oneOsmoInt           = osmomath.NewInt(oneOSMO)
	oneOsmoBigDec        = osmomath.NewBigDec(oneOSMO)

	oneOsmoCoin = sdk.NewCoin(appparams.BaseCoinUnit, oneOsmoInt)
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
}

var _ domain.PoolsTransformer = &poolTransformer{}

// NewPoolTransformer returns a new pool ingester.
func NewPoolTransformer(assetListGetter domain.AssetListGetter, keepers domain.SQSIngestKeepers) domain.PoolsTransformer {
	return &poolTransformer{
		gammKeeper:         keepers.GammKeeper,
		concentratedKeeper: keepers.ConcentratedKeeper,
		cosmWasmKeeper:     keepers.CosmWasmPoolKeeper,
		wasmKeeper:         keepers.WasmKeeper,
		bankKeeper:         keepers.BankKeeper,
		protorevKeeper:     keepers.ProtorevKeeper,
		poolManagerKeeper:  keepers.PoolManagerKeeper,
		assetListGetter:    assetListGetter,
	}
}

// processPoolState processes the pool state. an
func (pi *poolTransformer) Transform(ctx sdk.Context, blockPools domain.BlockPools) ([]sqsdomain.PoolI, sqsdomain.TakerFeeMap, error) {
	goCtx := sdk.WrapSDKContext(ctx)

	// TODO: can be cached
	tokenPrecisionMap, err := pi.assetListGetter.GetDenomPrecisions(goCtx)
	if err != nil {
		return nil, nil, err
	}

	// Create a map from denom to routable pool ID.
	denomToRoutablePoolIDMap := make(map[string]denomRoutingInfo)

	denomPairToTakerFeeMap := make(map[sqsdomain.DenomPair]osmomath.Dec, 0)

	// Get all pools
	cfmmPools := blockPools.CFMMPools
	concentratedPools := blockPools.ConcentratedPools
	cosmWasmPools := blockPools.CosmWasmPools

	allPoolsParsed := make([]sqsdomain.PoolI, 0, len(cfmmPools)+len(concentratedPools)+len(cosmWasmPools))

	// Parse CFMM pool to the standard SQS types.
	for _, pool := range cfmmPools {
		// Parse CFMM pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, denomToRoutablePoolIDMap, denomPairToTakerFeeMap, tokenPrecisionMap)
		if err != nil {
			// Silently skip pools on error to avoid breaking ingest of all other pools.
			continue
		}

		allPoolsParsed = append(allPoolsParsed, pool)
	}

	for _, pool := range concentratedPools {
		// Parse concentrated pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, denomToRoutablePoolIDMap, denomPairToTakerFeeMap, tokenPrecisionMap)
		if err != nil {
			// Silently skip pools on error to avoid breaking ingest of all other pools.
			continue
		}

		allPoolsParsed = append(allPoolsParsed, pool)
	}

	for _, pool := range cosmWasmPools {
		// Parse cosmwasm pool to the standard SQS types.
		pool, err := pi.convertPool(ctx, pool, denomToRoutablePoolIDMap, denomPairToTakerFeeMap, tokenPrecisionMap)
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
	denomToRoutingInfoMap map[string]denomRoutingInfo,
	denomPairToTakerFeeMap sqsdomain.TakerFeeMap,
	tokenPrecisionMap map[string]int,
) (sqsPool sqsdomain.PoolI, err error) {
	defer func() {
		r := recover()
		if r != nil {
			telemetry.IncrCounter(1, "sqs_ingest_convert_pool_panic", "pool_"+strconv.FormatUint(pool.GetId(), 10))

			err = fmt.Errorf("sqs ingest pool (%d) conversion panicked: %v", pool.GetId(), r)
			ctx.Logger().Error(err.Error())
		}
	}()

	balances := pi.bankKeeper.GetAllBalances(ctx, pool.GetAddress())
	poolDenoms := pool.GetPoolDenoms(ctx)
	poolDenomsMap := map[string]struct{}{}

	// Convert pool denoms to map
	for _, poolDenom := range poolDenoms {
		poolDenomsMap[poolDenom] = struct{}{}
	}

	spreadFactor := pool.GetSpreadFactor(ctx)

	// Get pool denoms. Although these can be inferred from balances, this is safer.
	// If we used balances, for pools with no liquidity, we would not be able to get the denoms.
	denoms, err := pi.poolManagerKeeper.RouteGetPoolDenoms(ctx, pool.GetId())
	if err != nil {
		return nil, err
	}

	var cosmWasmPoolModel *sqsdomain.CosmWasmPoolModel
	if pool.GetType() == poolmanagertypes.CosmWasm {
		cwPool, ok := pool.(cosmwasmpooltypes.CosmWasmExtension)
		if !ok {
			return nil, fmt.Errorf("pool (%d) with type (%d) is not a CosmWasmExtension", pool.GetId(), pool.GetType())
		}

		balances = cwPool.GetTotalPoolLiquidity(ctx)

		// This must never happen, but if it does, and there is no checks, the query will fail silently
		// so we panic here to make sure we catch this error
		if pi.wasmKeeper == nil {
			panic("wasmKeeper is nil")
		}

		bz := pi.wasmKeeper.QueryRaw(ctx, cwPool.GetAddress(), []byte("contract_info"))
		if len(bz) == 0 {
			// only log since cw pool contracts are not required to conform cw2
			ctx.Logger().Info(
				"contract_info not found for CosmWasm pool",
				"pool_id", pool.GetId(),
				"contract_address", pool.GetAddress(),
			)
		} else {
			var contractInfo *sqsdomain.ContractInfo
			cosmWasmPoolModel = &sqsdomain.CosmWasmPoolModel{}

			if err := json.Unmarshal(bz, &contractInfo); err != nil {
				// only log since cw pool contracts are not required to conform cw2
				ctx.Logger().Info(
					"CosmWasm pool does not conform cw2",
					"pool_id", pool.GetId(),
					"contract_address", pool.GetAddress(),
					"contract_info", string(bz),
				)
			} else {
				cosmWasmPoolModel.ContractInfo = *contractInfo

				// special transformation based on different cw pool
				if cosmWasmPoolModel.IsAlloyTransmuter() {
					err = pi.updateAlloyTrasmuterInfo(ctx, pool, cwPool, cosmWasmPoolModel, &denoms)
				}

				if err != nil {
					return nil, err
				}
			}
		}
	}

	osmoPoolTVL := osmomath.ZeroInt()

	// Note that this must follow the call to GetPoolDenoms() and GetSpreadFactor.
	// Otherwise, the CosmWasmPool model panics.
	pool = pool.AsSerializablePool()

	// filtered balances consisting only of the pool denom balances.
	filteredBalances := sdk.NewCoins()

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

		// update filtered balances only with pool tokens
		filteredBalances = filteredBalances.Add(balance)

		if balance.Denom == UOSMO {
			osmoPoolTVL = osmoPoolTVL.Add(balance.Amount)
			continue
		}

		// Spot price with uosmo as base asset.
		var uosmoBaseAssetSpotPrice osmomath.BigDec

		// Check if routable poolID already exists for the denom
		routingInfo, ok := denomToRoutingInfoMap[balance.Denom]
		if !ok {
			basePrecison, ok := tokenPrecisionMap[balance.Denom]
			if !ok {
				errorInTVLStr = fmt.Sprintf(noTokenPrecisionErrorFmtStr, balance.Denom)
				ctx.Logger().Debug(errorInTVLStr)
				continue
			}

			// Attempt to get a single-hop pool from on-chain routes.
			poolForDenomPair, err := pi.protorevKeeper.GetPoolForDenomPair(ctx, UOSMO, balance.Denom)
			if err == nil {
				// If on-chain route is present, calculate spot price with uosmo.
				uosmoBaseAssetSpotPrice, err = pi.poolManagerKeeper.RouteCalculateSpotPrice(ctx, poolForDenomPair, balance.Denom, UOSMO)
				if err != nil {
					errorInTVLStr = fmt.Sprintf(spotPriceErrorFmtStr, balance.Denom, err)
					ctx.Logger().Debug(errorInTVLStr)
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
						errorInTVLStr = fmt.Sprintf(multiHopSpotPriceErrorFmtStr, err)
						continue
					}

					denomBigDecAmtIn := osmomath.BigDecFromSDKInt(denomAmtIn)
					if denomBigDecAmtIn.IsZero() {
						ctx.Logger().Info("error inverting price from route overwrite", "denom", balance.Denom)
						errorInTVLStr = "error inverting price from route overwrite"
						continue
					}

					uosmoBaseAssetSpotPrice = oneOsmoBigDec.QuoMut(denomBigDecAmtIn)
				} else if isStableCoin {
					// We (very) naively assume that stablecoin has the same price as USDC for TVL ranking of pools in the router.
					uosmoBaseAssetSpotPrice, err = pi.poolManagerKeeper.RouteCalculateSpotPrice(ctx, usdcPool, usdcDenom, UOSMO)
					if err != nil {
						errorInTVLStr = fmt.Sprintf(spotPriceErrorFmtStr, balance.Denom, err)
						ctx.Logger().Debug(errorInTVLStr)
						continue
					}

					// Set base preecision to USDC
					basePrecison, ok = tokenPrecisionMap[usdcDenom]
					if !ok {
						errorInTVLStr = "no precision for denom " + usdcDenom
						continue
					}
				} else {
					// If there is no method to compute TVL for this denom, attach error and silently skip it.
					errorInTVLStr = err.Error()
					ctx.Logger().Debug("no overwrite present", "denom", balance.Denom)
					continue
				}
			}

			// Scale on-chain spot price to the correct token precision.
			precisionMultiplier := uosmoPrecisionBigDec.Quo(osmomath.NewBigDec(int64(basePrecison)))

			uosmoBaseAssetSpotPrice = uosmoBaseAssetSpotPrice.Mul(precisionMultiplier)

			if uosmoBaseAssetSpotPrice.IsZero() {
				errorInTVLStr = "failed to calculate spot price due to it becoming zero from truncations " + balance.Denom
				continue
			}

			routingInfo = denomRoutingInfo{
				PoolID: poolForDenomPair,
				Price:  uosmoBaseAssetSpotPrice,
			}
		}

		tvlAddition := osmomath.BigDecFromSDKInt(balance.Amount).QuoMut(routingInfo.Price).Dec().TruncateInt()
		osmoPoolTVL = osmoPoolTVL.Add(tvlAddition)
	}

	// Sort denoms for consistent ordering.
	sort.Strings(denoms)

	// Mutates denomPairToTakerFeeMap with the taker fee for every uniquer denom pair in the denoms list.
	err = retrieveTakerFeeToMapIfNotExists(ctx, denoms, denomPairToTakerFeeMap, pi.poolManagerKeeper)
	if err != nil {
		return nil, err
	}

	// Get the tick model for concentrated pools
	var tickModel *sqsdomain.TickModel

	// For CL pools, get the tick data
	if pool.GetType() == poolmanagertypes.Concentrated {
		tickData, currentTickIndex, err := pi.concentratedKeeper.GetTickLiquidityForFullRange(ctx, pool.GetId())
		// If there is no error, we set the tick model
		if err == nil {
			tickModel = &sqsdomain.TickModel{
				Ticks:            tickData,
				CurrentTickIndex: currentTickIndex,
			}
			// If there is no liquidity, we set the tick model to nil and update no liquidity flag
		} else if errors.Is(err, concentratedtypes.RanOutOfTicksForPoolError{PoolId: pool.GetId()}) {
			tickModel = &sqsdomain.TickModel{
				Ticks:            []queryproto.LiquidityDepthWithRange{},
				CurrentTickIndex: -1,
				HasNoLiquidity:   true,
			}

			// On any other error, we return the error
		} else {
			return nil, err
		}
	}

	return &sqsdomain.PoolWrapper{
		ChainModel: pool,
		SQSModel: sqsdomain.SQSPool{
			TotalValueLockedUSDC:  osmoPoolTVL,
			TotalValueLockedError: errorInTVLStr,
			Balances:              filteredBalances,
			PoolDenoms:            denoms,
			SpreadFactor:          spreadFactor,
			CosmWasmPoolModel:     cosmWasmPoolModel,
		},
		TickModel: tickModel,
	}, nil
}

func (pi *poolTransformer) updateAlloyTrasmuterInfo(
	ctx sdk.Context,
	pool poolmanagertypes.PoolI,
	cwPool cosmwasmpooltypes.CosmWasmExtension,
	cosmWasmPoolModel *sqsdomain.CosmWasmPoolModel,
	denoms *[]string,
) error {
	bz, err := pi.wasmKeeper.QuerySmart(ctx, cwPool.GetAddress(), []byte(`{"list_asset_configs":{}}`))
	if err != nil {
		return fmt.Errorf(
			"error querying list_asset_configs for pool (%d) contrat_address (%s): %w",
			pool.GetId(), pool.GetAddress(), err,
		)
	}
	var assetConfigsResponse struct {
		AssetConfigs []sqsdomain.TransmuterAssetConfig `json:"asset_configs"`
	}

	if err := json.Unmarshal(bz, &assetConfigsResponse); err != nil {
		return fmt.Errorf(
			"error unmarshalling asset_configs for pool (%d) contrat_address (%s): %w",
			pool.GetId(), pool.GetAddress(), err,
		)
	}

	bz, err = pi.wasmKeeper.QuerySmart(ctx, cwPool.GetAddress(), []byte(`{"get_share_denom":{}}`))
	if err != nil {
		return fmt.Errorf(
			"error querying get_share_denom for pool (%d) contrat_address (%s): %w",
			pool.GetId(), pool.GetAddress(), err,
		)
	}

	var getShareDenomResponse struct {
		ShareDenom string `json:"share_denom"`
	}

	if err := json.Unmarshal(bz, &getShareDenomResponse); err != nil {
		return fmt.Errorf(
			"error unmarshalling share_denom for pool (%d) contrat_address (%s): %w",
			pool.GetId(), pool.GetAddress(), err,
		)
	}

	// append alloyed denom to denoms
	*denoms = append(*denoms, getShareDenomResponse.ShareDenom)

	cosmWasmPoolModel.Data.AlloyTransmuter = &sqsdomain.AlloyTransmuterData{
		AlloyedDenom: getShareDenomResponse.ShareDenom,
		AssetConfigs: assetConfigsResponse.AssetConfigs,
	}

	return nil
}
