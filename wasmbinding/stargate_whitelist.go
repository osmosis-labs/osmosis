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

	epochtypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	gammv2types "github.com/osmosis-labs/osmosis/v12/x/gamm/v2types"
	incentivestypes "github.com/osmosis-labs/osmosis/v12/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v12/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v12/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v12/x/pool-incentives/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v12/x/superfluid/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v12/x/tokenfactory/types"
	twapquerytypes "github.com/osmosis-labs/osmosis/v12/x/twap/client/queryproto"
	twapv2querytypes "github.com/osmosis-labs/osmosis/v12/x/twap/client/v2queryproto"
	txfeestypes "github.com/osmosis-labs/osmosis/v12/x/txfees/types"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

func init() {
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
	setWhitelistedQuery("/osmosis.gamm.v1beta1.Query/SpotPrice", &gammtypes.QuerySpotPriceResponse{})
	setWhitelistedQuery("/osmosis.gamm.v2.Query/SpotPrice", &gammv2types.QuerySpotPriceResponse{})

	// incentives
	setWhitelistedQuery("/osmosis.incentives.Query/ModuleToDistributeCoins", &incentivestypes.ModuleToDistributeCoinsResponse{})
	setWhitelistedQuery("/osmosis.incentives.Query/LockableDurations", &incentivestypes.QueryLockableDurationsResponse{})

	// lockup
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleBalance", &lockuptypes.ModuleBalanceResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/ModuleLockedAmount", &lockuptypes.ModuleLockedAmountResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockableCoins", &lockuptypes.AccountUnlockableCoinsResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/AccountUnlockingCoins", &lockuptypes.AccountUnlockingCoinsResponse{})
	setWhitelistedQuery("/osmosis.lockup.Query/LockedDenom", &lockuptypes.LockedDenomResponse{})

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

	// txfees
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/FeeTokens", &txfeestypes.QueryFeeTokensResponse{})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/DenomSpotPrice", &txfeestypes.QueryDenomSpotPriceResponse{})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/DenomPoolId", &txfeestypes.QueryDenomPoolIdResponse{})
	setWhitelistedQuery("/osmosis.txfees.v1beta1.Query/BaseDenom", &txfeestypes.QueryBaseDenomResponse{})

	// tokenfactory
	setWhitelistedQuery("/osmosis.tokenfactory.v1beta1.Query/params", &tokenfactorytypes.QueryParamsResponse{})
	setWhitelistedQuery("/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata", &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{})
	// Does not include denoms_from_creator, TBD if this is the index we want contracts to use instead of admin

	// twap
	setWhitelistedQuery("/osmosis.twap.v2.Query/ArithmeticTwap", &twapv2querytypes.ArithmeticTwapResponse{})
	setWhitelistedQuery("/osmosis.twap.v2.Query/ArithmeticTwapToNow", &twapv2querytypes.ArithmeticTwapToNowResponse{})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/Params", &twapquerytypes.ParamsResponse{})
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
