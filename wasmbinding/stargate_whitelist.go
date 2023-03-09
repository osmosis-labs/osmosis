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

	downtimequerytypes "github.com/osmosis-labs/osmosis/v15/x/downtime-detector/client/queryproto"
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	gammv2types "github.com/osmosis-labs/osmosis/v15/x/gamm/v2types"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v15/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
	superfluidtypes "github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v15/x/tokenfactory/types"
	twapquerytypes "github.com/osmosis-labs/osmosis/v15/x/twap/client/queryproto"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

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

	// bank
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Balance", &banktypes.QueryBalanceResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomMetadata", &banktypes.QueryDenomsMetadataResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Params", &banktypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SupplyOf", &banktypes.QuerySupplyOfResponse{})

	// distribution
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/Params", &distributiontypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", &distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorCommission", &distributiontypes.QueryValidatorCommissionResponse{})

	// gov
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Deposit", &govtypes.QueryDepositResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Params", &govtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Vote", &govtypes.QueryVoteResponse{})

	// slashing
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/Params", &slashingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/SigningInfo", &slashingtypes.QuerySigningInfoResponse{})

	// staking
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Delegation", &stakingtypes.QueryDelegationResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Params", &stakingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Validator", &stakingtypes.QueryValidatorResponse{})

	// osmosis queries

	// epochs
	setWhitelistedQuery("/osmosis.epochs.v1beta1.Query/EpochInfos", &epochtypes.QueryEpochsInfoResponse{})
	setWhitelistedQuery("/osmosis.epochs.v1beta1.Query/CurrentEpoch", &epochtypes.QueryCurrentEpochResponse{})

	// gamm
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/NumPools", &gammtypes.QueryNumPoolsResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalLiquidity", &gammtypes.QueryTotalLiquidityResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/PoolParams", &gammtypes.QueryPoolParamsResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", &gammtypes.QueryTotalPoolLiquidityResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/TotalShares", &gammtypes.QueryTotalSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcJoinPoolShares", &gammtypes.QueryCalcJoinPoolSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcExitPoolCoinsFromShares", &gammtypes.QueryCalcExitPoolCoinsFromSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/CalcJoinPoolNoSwapShares", &gammtypes.QueryCalcJoinPoolNoSwapSharesResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/PoolType", &gammtypes.QueryPoolTypeResponse{})
	setWhitelistedQuery("/osmosis.gamm.v2.Query/SpotPrice", &gammv2types.QuerySpotPriceResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn", &gammtypes.QuerySwapExactAmountInResponse{})
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", &gammtypes.QuerySwapExactAmountOutResponse{})

	// incentives
	setWhitelistedQuery("/osmosis.incentives.Query/ModuleToDistributeCoins", &incentivestypes.ModuleToDistributeCoinsResponse{})
	setWhitelistedQuery("/osmosis.incentives.Query/LockableDurations", &incentivestypes.QueryLockableDurationsResponse{})

	// lockup
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleBalance", &lockuptypes.ModuleBalanceResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleLockedAmount", &lockuptypes.ModuleLockedAmountResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockableCoins", &lockuptypes.AccountUnlockableCoinsResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockingCoins", &lockuptypes.AccountUnlockingCoinsResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/LockedDenom", &lockuptypes.LockedDenomResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/LockedByID", &lockuptypes.LockedResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/NextLockID", &lockuptypes.NextLockIDResponse{})

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

	// poolmanager
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/NumPools", &poolmanagerqueryproto.NumPoolsResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn", &poolmanagerqueryproto.EstimateSwapExactAmountInResponse{})
	setWhitelistedQuery("/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut", &poolmanagerqueryproto.EstimateSwapExactAmountOutRequest{})

	// txfees
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/FeeTokens", &txfeestypes.QueryFeeTokensResponse{})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/DenomSpotPrice", &txfeestypes.QueryDenomSpotPriceResponse{})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/DenomPoolId", &txfeestypes.QueryDenomPoolIdResponse{})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/BaseDenom", &txfeestypes.QueryBaseDenomResponse{})

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
}

// GetWhitelistedQuery returns the whitelisted query at the provided path.
// If the query does not exist, or it was setup wrong by the chain, this returns an error.
func GetWhitelistedQuery(queryPath string) (codec.ProtoMarshaler, error) {
	protoResponseAny, isWhitelisted := stargateWhitelist.Load(queryPath)
	if !isWhitelisted {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}
	protoResponseType, ok := protoResponseAny.(codec.ProtoMarshaler)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}
	return protoResponseType, nil
}

func setWhitelistedQuery(queryPath string, protoType codec.ProtoMarshaler) {
	stargateWhitelist.Store(queryPath, protoType)
}

func GetStargateWhitelistedPaths() (keys []string) {
	// Iterate over the map and collect the keys
	stargateWhitelist.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			panic("key is not a string")
		}
		keys = append(keys, keyStr)
		return true
	})

	return keys
}
