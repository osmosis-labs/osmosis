package wasmbinding

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	gammv2types "github.com/osmosis-labs/osmosis/v27/x/gamm/v2types"

	"github.com/cosmos/gogoproto/proto"

	concentratedliquidityquery "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/queryproto"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client/queryproto"
	downtimequerytypes "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/queryproto"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
	twapquerytypes "github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

// stargateResponsePools keeps whitelist and its deterministic
// response binding for stargate queries.
// CONTRACT: since results of queries go into blocks, queries being added here should always be
// deterministic or can cause non-determinism in the state machine.
//
// The query is multi-threaded so we're using a sync.Pool
// to manage the allocation and de-allocation of newly created
// pb objects.
var stargateResponsePools = make(map[string]*sync.Pool)

// Note: When adding a migration here, we should also add it to the Async ICQ params in the upgrade.
// In the future we may want to find a better way to keep these in sync

//nolint:staticcheck
func init() {
	// ibc queries
	setWhitelistedQuery("/ibc.applications.transfer.v1.Query/DenomTrace", &ibctransfertypes.QueryDenomTraceResponse{})

	// cosmos-sdk queries

	// auth
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Account", &authtypes.QueryAccountResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Params", &authtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/ModuleAccounts", &authtypes.QueryModuleAccountsResponse{})

	// bank
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Balance", &banktypes.QueryBalanceResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomMetadata", &banktypes.QueryDenomMetadataResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomsMetadata", &banktypes.QueryDenomsMetadataResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Params", &banktypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SupplyOf", &banktypes.QuerySupplyOfResponse{})

	// distribution
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/Params", &distributiontypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", &distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorCommission", &distributiontypes.QueryValidatorCommissionResponse{})

	// gov
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Deposit", &govtypesv1.QueryDepositResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Params", &govtypesv1.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Vote", &govtypesv1.QueryVoteResponse{})

	// slashing
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/Params", &slashingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/SigningInfo", &slashingtypes.QuerySigningInfoResponse{})

	// staking
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Delegation", &stakingtypes.QueryDelegationResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Params", &stakingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Validator", &stakingtypes.QueryValidatorResponse{})

	// osmosis queries
	// cosmwasm pool
	setWhitelistedQuery("/osmosis.cosmwasmpool.v1beta1.Query/Pools", &cosmwasmpooltypes.PoolsResponse{})
	setWhitelistedQuery("/osmosis.cosmwasmpool.v1beta1.Query/Params", &cosmwasmpooltypes.ParamsResponse{})
	setWhitelistedQuery("/osmosis.cosmwasmpool.v1beta1.Query/ContractInfoByPoolId", &cosmwasmpooltypes.ContractInfoByPoolIdResponse{})

	// epochs
	setWhitelistedQuery("/osmosis.epochs.v1beta1.Query/EpochInfos", &epochtypes.QueryEpochsInfoResponse{})
	setWhitelistedQuery("/osmosis.epochs.v1beta1.Query/CurrentEpoch", &epochtypes.QueryCurrentEpochResponse{})

	// gamm
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/NumPools", &gammtypes.QueryNumPoolsResponse{}) // ==> use x/poolmanager
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalLiquidity", &gammtypes.QueryTotalLiquidityResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolResponse{}) // ==> use x/poolmanager
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/PoolParams", &gammtypes.QueryPoolParamsResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", &gammtypes.QueryTotalPoolLiquidityResponse{}) // ==> use x/poolmanager
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalShares", &gammtypes.QueryTotalSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcJoinPoolShares", &gammtypes.QueryCalcJoinPoolSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcExitPoolCoinsFromShares", &gammtypes.QueryCalcExitPoolCoinsFromSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcJoinPoolNoSwapShares", &gammtypes.QueryCalcJoinPoolNoSwapSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/PoolType", &gammtypes.QueryPoolTypeResponse{})
	setWhitelistedQuery("/osmosis.gamm.v2.Query/SpotPrice", &gammv2types.QuerySpotPriceResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn", &gammtypes.QuerySwapExactAmountInResponse{})   // ==> use x/poolmanager
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", &gammtypes.QuerySwapExactAmountOutResponse{}) // ==> use x/poolmanager

	// incentives
	setWhitelistedQuery("/osmosis.incentives.Query/ModuleToDistributeCoins", &incentivestypes.ModuleToDistributeCoinsResponse{})
	setWhitelistedQuery("/osmosis.incentives.Query/LockableDurations", &incentivestypes.QueryLockableDurationsResponse{})
	setWhitelistedQuery("/osmosis.incentives.Query/GaugeByID", &incentivestypes.GaugeByIDResponse{})

	// lockup
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleBalance", &lockuptypes.ModuleBalanceResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleLockedAmount", &lockuptypes.ModuleLockedAmountResponse{})
	// Warning: it iterates over every single lock account has, which means this query can have unbounded gas
	setWhitelistedQuery("/osmosis.lockup.Query/AccountLockedCoins", &lockuptypes.AccountLockedCoinsResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockableCoins", &lockuptypes.AccountUnlockableCoinsResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockingCoins", &lockuptypes.AccountUnlockingCoinsResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/LockedDenom", &lockuptypes.LockedDenomResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/LockedByID", &lockuptypes.LockedResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/NextLockID", &lockuptypes.NextLockIDResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/LockRewardReceiver", &lockuptypes.LockRewardReceiverResponse{})

	// mint
	setWhitelistedQuery("/osmosis.mint.v1beta1.Query/EpochProvisions", &minttypes.QueryEpochProvisionsResponse{})
	setWhitelistedQuery("/osmosis.mint.v1beta1.Query/Params", &minttypes.QueryParamsResponse{})

	// pool-incentives
	setWhitelistedQuery("/osmosis.poolincentives.v1beta1.Query/GaugeIds", &poolincentivestypes.QueryGaugeIdsResponse{})

	// superfluid
	setWhitelistedQuery("/osmosis.superfluid.Query/Params", &superfluidtypes.QueryParamsResponse{})
	setWhitelistedQuery("/osmosis.superfluid.Query/AssetType", &superfluidtypes.AssetTypeResponse{})
	setWhitelistedQuery("/osmosis.superfluid.Query/AllAssets", &superfluidtypes.AllAssetsResponse{})
	setWhitelistedQuery("/osmosis.superfluid.Query/AssetMultiplier", &superfluidtypes.AssetMultiplierResponse{})

	// smartaccount
	setWhitelistedQuery("/osmosis.smartaccount.v1beta1.Query/GetAuthenticator", &smartaccounttypes.GetAuthenticatorResponse{})
	setWhitelistedQuery("/osmosis.smartaccount.v1beta1.Query/GetAuthenticators", &smartaccounttypes.GetAuthenticatorsResponse{})

	// poolmanager
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/NumPools", &poolmanagerqueryproto.NumPoolsResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn", &poolmanagerqueryproto.EstimateSwapExactAmountInResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut", &poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountIn", &poolmanagerqueryproto.EstimateSwapExactAmountInResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountOut", &poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/Pool", &poolmanagerqueryproto.PoolResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/SpotPrice", &poolmanagerqueryproto.SpotPriceResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/TotalPoolLiquidity", &poolmanagerqueryproto.TotalPoolLiquidityResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/Params", &poolmanagerqueryproto.ParamsResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/TradingPairTakerFee", &poolmanagerqueryproto.TradingPairTakerFeeResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateTradeBasedOnPriceImpact", &poolmanagerqueryproto.EstimateTradeBasedOnPriceImpactResponse{})

	// txfees
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/FeeTokens", &txfeestypes.QueryFeeTokensResponse{})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/DenomSpotPrice", &txfeestypes.QueryDenomSpotPriceResponse{})

	// tokenfactory
	setWhitelistedQuery("/osmosis.tokenfactory.v1beta1.Query/Params", &tokenfactorytypes.QueryParamsResponse{})
	setWhitelistedQuery("/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata", &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{})
	// Does not include denoms_from_creator, TBD if this is the index we want contracts to use instead of admin

	// twap
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/ArithmeticTwap", &twapquerytypes.ArithmeticTwapResponse{})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow", &twapquerytypes.ArithmeticTwapToNowResponse{})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/GeometricTwap", &twapquerytypes.GeometricTwapResponse{})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/GeometricTwapToNow", &twapquerytypes.GeometricTwapToNowResponse{})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/Params", &twapquerytypes.ParamsResponse{})

	// downtime-detector
	setWhitelistedQuery("/osmosis.downtimedetector.v1beta1.Query/RecoveredSinceDowntimeOfLength", &downtimequerytypes.RecoveredSinceDowntimeOfLengthResponse{})

	// concentrated-liquidity
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/UserPositions", &concentratedliquidityquery.UserPositionsResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/LiquidityPerTickRange", &concentratedliquidityquery.LiquidityPerTickRangeResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/ClaimableSpreadRewards", &concentratedliquidityquery.ClaimableSpreadRewardsResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/ClaimableIncentives", &concentratedliquidityquery.ClaimableIncentivesResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/PositionById", &concentratedliquidityquery.PositionByIdResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/Params", &concentratedliquidityquery.ParamsResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/PoolAccumulatorRewards", &concentratedliquidityquery.PoolAccumulatorRewardsResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/IncentiveRecords", &concentratedliquidityquery.IncentiveRecordsResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/TickAccumulatorTrackers", &concentratedliquidityquery.TickAccumulatorTrackersResponse{})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/CFMMPoolIdLinkFromConcentratedPoolId", &concentratedliquidityquery.CFMMPoolIdLinkFromConcentratedPoolIdResponse{})
}

// IsWhitelistedQuery returns if the query is not whitelisted.
func IsWhitelistedQuery(queryPath string) error {
	_, isWhitelisted := stargateResponsePools[queryPath]
	if !isWhitelisted {
		return wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}
	return nil
}

// getWhitelistedQuery returns the whitelisted query at the provided path.
// If the query does not exist, or it was setup wrong by the chain, this returns an error.
// CONTRACT: must call returnStargateResponseToPool in order to avoid pointless allocs.
func getWhitelistedQuery(queryPath string) (proto.Message, error) {
	protoResponseAny, isWhitelisted := stargateResponsePools[queryPath]
	if !isWhitelisted {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}
	protoMarshaler, ok := protoResponseAny.Get().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to assert type to proto.Messager")
	}
	return protoMarshaler, nil
}

type protoTypeG[T any] interface {
	*T
	proto.Message
}

// setWhitelistedQuery sets the whitelisted query at the provided path.
// This method also creates a sync.Pool for the provided protoMarshaler.
// We use generics so we can properly instantiate an object that the
// queryPath expects as a response.
func setWhitelistedQuery[T any, PT protoTypeG[T]](queryPath string, _ PT) {
	stargateResponsePools[queryPath] = &sync.Pool{
		New: func() any {
			return PT(new(T))
		},
	}
}

// returnStargateResponseToPool returns the provided protoMarshaler to the appropriate pool based on it's query path.
func returnStargateResponseToPool(queryPath string, pb proto.Message) {
	stargateResponsePools[queryPath].Put(pb)
}

func GetStargateWhitelistedPaths() (keys []string) {
	// Iterate over the map and collect the keys
	keys = make([]string, 0, len(stargateResponsePools))
	for k := range stargateResponsePools {
		keys = append(keys, k)
	}
	return keys
}
