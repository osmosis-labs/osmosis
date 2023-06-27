package wasmbinding

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	gammv2types "github.com/osmosis-labs/osmosis/v16/x/gamm/v2types"

	concentratedliquidityquery "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/client/queryproto"
	downtimequerytypes "github.com/osmosis-labs/osmosis/v16/x/downtime-detector/client/queryproto"
	gammtypes "github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v16/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v16/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v16/x/pool-incentives/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v16/x/poolmanager/client/queryproto"
	superfluidtypes "github.com/osmosis-labs/osmosis/v16/x/superfluid/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v16/x/tokenfactory/types"
	twapquerytypes "github.com/osmosis-labs/osmosis/v16/x/twap/client/queryproto"
	txfeestypes "github.com/osmosis-labs/osmosis/v16/x/txfees/types"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// A map to store the factory functions
var stargateWhitelist = make(map[string]func() codec.ProtoMarshaler)

// Mutex to make the map access thread-safe
var mutex = &sync.RWMutex{}

// Note: When adding a migration here, we should also add it to the Async ICQ params in the upgrade.
// In the future we may want to find a better way to keep these in sync

//nolint:staticcheck
func init() {
	// ibc queries
	//
	// transfer
	setWhitelistedQuery("/ibc.applications.transfer.v1.Query/DenomTrace", func() codec.ProtoMarshaler {
		return &ibctransfertypes.QueryDenomTraceResponse{}
	})

	// cosmos-sdk queries
	//
	// auth
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Account", func() codec.ProtoMarshaler {
		return &authtypes.QueryAccountResponse{}
	})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &authtypes.QueryParamsResponse{}
	})

	// bank
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Balance", func() codec.ProtoMarshaler {
		return &banktypes.QueryBalanceResponse{}
	})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomMetadata", func() codec.ProtoMarshaler {
		return &banktypes.QueryDenomsMetadataResponse{}
	})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &banktypes.QueryParamsResponse{}
	})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SupplyOf", func() codec.ProtoMarshaler {
		return &banktypes.QuerySupplyOfResponse{}
	})

	// distribution
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &distributiontypes.QueryParamsResponse{}
	})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", func() codec.ProtoMarshaler {
		return &distributiontypes.QueryDelegatorWithdrawAddressResponse{}
	})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorCommission", func() codec.ProtoMarshaler {
		return &distributiontypes.QueryValidatorCommissionResponse{}
	})

	// gov
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Deposit", func() codec.ProtoMarshaler {
		return &govtypes.QueryDepositResponse{}
	})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &govtypes.QueryParamsResponse{}
	})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Vote", func() codec.ProtoMarshaler {
		return &govtypes.QueryVoteResponse{}
	})

	// slashing
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &slashingtypes.QueryParamsResponse{}
	})
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/SigningInfo", func() codec.ProtoMarshaler {
		return &slashingtypes.QuerySigningInfoResponse{}
	})

	// staking
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Delegation", func() codec.ProtoMarshaler {
		return &stakingtypes.QueryDelegationResponse{}
	})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &stakingtypes.QueryParamsResponse{}
	})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Validator", func() codec.ProtoMarshaler {
		return &stakingtypes.QueryValidatorResponse{}
	})

	// osmosis queries

	// epochs
	setWhitelistedQuery("/osmosis.epochs.v1beta1.Query/EpochInfos", func() codec.ProtoMarshaler {
		return &epochtypes.QueryEpochsInfoResponse{}
	})
	setWhitelistedQuery("/osmosis.epochs.v1beta1.Query/CurrentEpoch", func() codec.ProtoMarshaler {
		return &epochtypes.QueryCurrentEpochResponse{}
	})

	// gamm
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/NumPools", func() codec.ProtoMarshaler {
		return &gammtypes.QueryNumPoolsResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalLiquidity", func() codec.ProtoMarshaler {
		return &gammtypes.QueryTotalLiquidityResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/Pool", func() codec.ProtoMarshaler {
		return &gammtypes.QueryPoolResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/PoolParams", func() codec.ProtoMarshaler {
		return &gammtypes.QueryPoolParamsResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", func() codec.ProtoMarshaler {
		return &gammtypes.QueryTotalPoolLiquidityResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalShares", func() codec.ProtoMarshaler {
		return &gammtypes.QueryTotalSharesResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcJoinPoolShares", func() codec.ProtoMarshaler {
		return &gammtypes.QueryCalcJoinPoolSharesResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcExitPoolCoinsFromShares", func() codec.ProtoMarshaler {
		return &gammtypes.QueryCalcExitPoolCoinsFromSharesResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcJoinPoolNoSwapShares", func() codec.ProtoMarshaler {
		return &gammtypes.QueryCalcJoinPoolNoSwapSharesResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/PoolType", func() codec.ProtoMarshaler {
		return &gammtypes.QueryPoolTypeResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v2.Query/SpotPrice", func() codec.ProtoMarshaler {
		return &gammv2types.QuerySpotPriceResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn", func() codec.ProtoMarshaler {
		return &gammtypes.QuerySwapExactAmountInResponse{}
	})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", func() codec.ProtoMarshaler {
		return &gammtypes.QuerySwapExactAmountOutResponse{}
	})

	// incentives
	setWhitelistedQuery("/osmosis.incentives.Query/ModuleToDistributeCoins", func() codec.ProtoMarshaler {
		return &incentivestypes.ModuleToDistributeCoinsResponse{}
	})
	setWhitelistedQuery("/osmosis.incentives.Query/LockableDurations", func() codec.ProtoMarshaler {
		return &incentivestypes.QueryLockableDurationsResponse{}
	})

	// lockup
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleBalance", func() codec.ProtoMarshaler {
		return &lockuptypes.ModuleBalanceResponse{}
	})
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleLockedAmount", func() codec.ProtoMarshaler {
		return &lockuptypes.ModuleLockedAmountResponse{}
	})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockableCoins", func() codec.ProtoMarshaler {
		return &lockuptypes.AccountUnlockableCoinsResponse{}
	})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockingCoins", func() codec.ProtoMarshaler {
		return &lockuptypes.AccountUnlockingCoinsResponse{}
	})
	setWhitelistedQuery("/osmosis.lockup.Query/LockedDenom", func() codec.ProtoMarshaler {
		return &lockuptypes.LockedDenomResponse{}
	})
	setWhitelistedQuery("/osmosis.lockup.Query/LockedByID", func() codec.ProtoMarshaler {
		return &lockuptypes.LockedResponse{}
	})
	setWhitelistedQuery("/osmosis.lockup.Query/NextLockID", func() codec.ProtoMarshaler {
		return &lockuptypes.NextLockIDResponse{}
	})
	setWhitelistedQuery("/osmosis.lockup.Query/LockRewardReceiver", func() codec.ProtoMarshaler {
		return &lockuptypes.LockRewardReceiverResponse{}
	})

	// mint
	setWhitelistedQuery("/osmosis.mint.v1beta1.Query/EpochProvisions", func() codec.ProtoMarshaler {
		return &minttypes.QueryEpochProvisionsResponse{}
	})
	setWhitelistedQuery("/osmosis.mint.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &minttypes.QueryParamsResponse{}
	})

	// pool-incentives
	setWhitelistedQuery("/osmosis.poolincentives.v1beta1.Query/GaugeIds", func() codec.ProtoMarshaler {
		return &poolincentivestypes.QueryGaugeIdsResponse{}
	})

	// superfluid
	setWhitelistedQuery("/osmosis.superfluid.Query/Params", func() codec.ProtoMarshaler {
		return &superfluidtypes.QueryParamsResponse{}
	})
	setWhitelistedQuery("/osmosis.superfluid.Query/AssetType", func() codec.ProtoMarshaler {
		return &superfluidtypes.AssetTypeResponse{}
	})
	setWhitelistedQuery("/osmosis.superfluid.Query/AllAssets", func() codec.ProtoMarshaler {
		return &superfluidtypes.AllAssetsResponse{}
	})
	setWhitelistedQuery("/osmosis.superfluid.Query/AssetMultiplier", func() codec.ProtoMarshaler {
		return &superfluidtypes.AssetMultiplierResponse{}
	})

	// poolmanager
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/NumPools", func() codec.ProtoMarshaler {
		return &poolmanagerqueryproto.NumPoolsResponse{}
	})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn", func() codec.ProtoMarshaler {
		return &poolmanagerqueryproto.EstimateSwapExactAmountInResponse{}
	})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut", func() codec.ProtoMarshaler {
		return &poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{}
	})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountIn", func() codec.ProtoMarshaler {
		return &poolmanagerqueryproto.EstimateSwapExactAmountInResponse{}
	})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountOut", func() codec.ProtoMarshaler {
		return &poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{}
	})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/Pool", func() codec.ProtoMarshaler {
		return &poolmanagerqueryproto.PoolResponse{}
	})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/SpotPrice", func() codec.ProtoMarshaler {
		return &poolmanagerqueryproto.SpotPriceResponse{}
	})

	// txfees
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/FeeTokens", func() codec.ProtoMarshaler {
		return &txfeestypes.QueryFeeTokensResponse{}
	})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/DenomSpotPrice", func() codec.ProtoMarshaler {
		return &txfeestypes.QueryDenomSpotPriceResponse{}
	})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/DenomPoolId", func() codec.ProtoMarshaler {
		return &txfeestypes.QueryDenomPoolIdResponse{}
	})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/BaseDenom", func() codec.ProtoMarshaler {
		return &txfeestypes.QueryBaseDenomResponse{}
	})

	// tokenfactory
	setWhitelistedQuery("/osmosis.tokenfactory.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &tokenfactorytypes.QueryParamsResponse{}
	})
	setWhitelistedQuery("/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata", func() codec.ProtoMarshaler {
		return &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{}
	})
	// Does not include denoms_from_creator, TBD if this is the index we want contracts to use instead of admin

	// twap
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/ArithmeticTwap", func() codec.ProtoMarshaler {
		return &twapquerytypes.ArithmeticTwapResponse{}
	})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow", func() codec.ProtoMarshaler {
		return &twapquerytypes.ArithmeticTwapToNowResponse{}
	})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/GeometricTwap", func() codec.ProtoMarshaler {
		return &twapquerytypes.GeometricTwapResponse{}
	})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/GeometricTwapToNow", func() codec.ProtoMarshaler {
		return &twapquerytypes.GeometricTwapToNowResponse{}
	})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &twapquerytypes.ParamsResponse{}
	})

	// downtime-detector
	setWhitelistedQuery("/osmosis.downtimedetector.v1beta1.Query/RecoveredSinceDowntimeOfLength", func() codec.ProtoMarshaler {
		return &downtimequerytypes.RecoveredSinceDowntimeOfLengthResponse{}
	})

	// concentrated-liquidity
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/Pools", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.PoolsResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/UserPositions", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.UserPositionsResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/LiquidityPerTickRange", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.LiquidityPerTickRangeResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/LiquidityNetInDirection", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.LiquidityNetInDirectionResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/ClaimableSpreadRewards", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.ClaimableSpreadRewardsResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/ClaimableIncentives", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.ClaimableIncentivesResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/PositionById", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.PositionByIdResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/Params", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.ParamsResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/PoolAccumulatorRewards", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.PoolAccumulatorRewardsResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/IncentiveRecords", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.IncentiveRecordsResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/TickAccumulatorTrackers", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.TickAccumulatorTrackersResponse{}
	})
	setWhitelistedQuery("/osmosis.concentratedliquidity.v1beta1.Query/CFMMPoolIdLinkFromConcentratedPoolId", func() codec.ProtoMarshaler {
		return &concentratedliquidityquery.CFMMPoolIdLinkFromConcentratedPoolIdResponse{}
	})
}

// GetWhitelistedQuery returns the whitelisted query at the provided path.
// If the query does not exist, or it was setup wrong by the chain, this returns an error.
func GetWhitelistedQuery(queryPath string) (codec.ProtoMarshaler, error) {
	mutex.RLock()
	factoryFunc, isWhitelisted := stargateWhitelist[queryPath]
	mutex.RUnlock()
	if !isWhitelisted {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}

	return factoryFunc(), nil
}

func setWhitelistedQuery(queryPath string, factoryFunc func() codec.ProtoMarshaler) {
	mutex.Lock()
	stargateWhitelist[queryPath] = factoryFunc
	mutex.Unlock()
}

func GetStargateWhitelistedPaths() (keys []string) {
	mutex.RLock()
	defer mutex.RUnlock()

	// Iterate over the map and collect the keys
	for key := range stargateWhitelist {
		keys = append(keys, key)
	}
	return keys
}
