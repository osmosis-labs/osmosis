package wasmbinding

import (
	"sync"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v12/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v12/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v12/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v12/x/pool-incentives/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v12/x/superfluid/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v12/x/txfees/types"
)

// StargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var StargateWhitelist sync.Map

func init() {
	// cosmos-sdk queries

	// auth
	StargateWhitelist.Store("/cosmos.auth.v1beta1.Query/Account", &authtypes.QueryAccountResponse{})
	StargateWhitelist.Store("/cosmos.auth.v1beta1.Query/Params", &authtypes.QueryParamsResponse{})

	// bank
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/Balance", &banktypes.QueryBalanceResponse{})
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/DenomMetadata", &banktypes.QueryDenomsMetadataResponse{})
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/Params", &banktypes.QueryParamsResponse{})
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/SupplyOf", &banktypes.QuerySupplyOfResponse{})

	// distribution
	StargateWhitelist.Store("/cosmos.distribution.v1beta1.Query/Params", &distributiontypes.QueryParamsResponse{})
	StargateWhitelist.Store("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", &distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	StargateWhitelist.Store("/cosmos.distribution.v1beta1.Query/ValidatorCommission", &distributiontypes.QueryValidatorCommissionResponse{})

	// gov
	StargateWhitelist.Store("/cosmos.gov.v1beta1.Query/Deposit", &govtypes.QueryDepositResponse{})
	StargateWhitelist.Store("/cosmos.gov.v1beta1.Query/Params", &govtypes.QueryParamsResponse{})
	StargateWhitelist.Store("/cosmos.gov.v1beta1.Query/Vote", &govtypes.QueryVoteResponse{})

	// slashing
	StargateWhitelist.Store("/cosmos.slashing.v1beta1.Query/Params", &slashingtypes.QueryParamsResponse{})
	StargateWhitelist.Store("/cosmos.slashing.v1beta1.Query/SigningInfo", &slashingtypes.QuerySigningInfoResponse{})

	// staking
	StargateWhitelist.Store("/cosmos.staking.v1beta1.Query/Delegation", &stakingtypes.QueryDelegationResponse{})
	StargateWhitelist.Store("/cosmos.staking.v1beta1.Query/Params", &stakingtypes.QueryParamsResponse{})
	StargateWhitelist.Store("/cosmos.staking.v1beta1.Query/Validator", &stakingtypes.QueryValidatorResponse{})

	// osmosis queries

	//epochs
	StargateWhitelist.Store("/osmosis.epochs.v1beta1.Query/EpochInfos", &epochtypes.QueryEpochsInfoResponse{})
	StargateWhitelist.Store("/osmosis.epochs.v1beta1.Query/CurrentEpoch", &epochtypes.QueryCurrentEpochResponse{})

	// gamm
	StargateWhitelist.Store("/osmosis.gamm.v1beta1.Query/NumPools", &gammtypes.QueryNumPoolsResponse{})
	StargateWhitelist.Store("/osmosis.gamm.v1beta1.Query/TotalLiquidity", &gammtypes.QueryTotalLiquidityResponse{})
	StargateWhitelist.Store("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolResponse{})
	StargateWhitelist.Store("/osmosis.gamm.v1beta1.Query/PoolParams", &gammtypes.QueryPoolParamsResponse{})
	StargateWhitelist.Store("/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", &gammtypes.QueryTotalPoolLiquidityResponse{})
	StargateWhitelist.Store("/osmosis.gamm.v1beta1.Query/TotalShares", &gammtypes.QueryTotalSharesResponse{})
	StargateWhitelist.Store("/osmosis.gamm.v1beta1.Query/SpotPrice", &gammtypes.QuerySpotPriceResponse{})

	// incentives
	StargateWhitelist.Store("/osmosis.incentives.Query/ModuleToDistributeCoins", &incentivestypes.ModuleToDistributeCoinsResponse{})
	StargateWhitelist.Store("/osmosis.incentives.Query/ModuleDistributedCoins", &incentivestypes.ModuleDistributedCoinsResponse{})
	StargateWhitelist.Store("/osmosis.incentives.Query/LockableDurations", &incentivestypes.QueryLockableDurationsResponse{})

	// lockup
	StargateWhitelist.Store("/osmosis.lockup.Query/ModuleBalance", &lockuptypes.ModuleBalanceResponse{})
	StargateWhitelist.Store("/osmosis.lockup.Query/ModuleLockedAmount", &lockuptypes.ModuleLockedAmountResponse{})
	StargateWhitelist.Store("/osmosis.lockup.Query/AccountUnlockableCoins", &lockuptypes.AccountUnlockableCoinsResponse{})
	StargateWhitelist.Store("/osmosis.lockup.Query/AccountUnlockingCoins", &lockuptypes.AccountUnlockingCoinsResponse{})
	StargateWhitelist.Store("/osmosis.lockup.Query/LockedDenom", &lockuptypes.LockedDenomResponse{})

	// mint
	StargateWhitelist.Store("/osmosis.mint.v1beta1.Query/EpochProvisions", &minttypes.QueryEpochProvisionsResponse{})
	StargateWhitelist.Store("/osmosis.mint.v1beta1.Query/Params", &minttypes.QueryParamsResponse{})

	// pool-incentives
	StargateWhitelist.Store("/osmosis.poolincentives.v1beta1.Query/GaugeIds", &poolincentivestypes.QueryGaugeIdsResponse{})

	// superfluid
	StargateWhitelist.Store("/osmosis.superfluid.Query/Params", &superfluidtypes.QueryParamsResponse{})
	StargateWhitelist.Store("/osmosis.superfluid.Query/AssetType", &superfluidtypes.AssetTypeResponse{})
	StargateWhitelist.Store("/osmosis.superfluid.Query/AllAssets", &superfluidtypes.AllAssetsResponse{})
	StargateWhitelist.Store("/osmosis.superfluid.Query/AssetMultiplier", &superfluidtypes.AssetMultiplierResponse{})

	// txfees
	StargateWhitelist.Store("/osmosis.txfees.v1beta1.Query/FeeTokens", &txfeestypes.QueryFeeTokensResponse{})
	StargateWhitelist.Store("/osmosis.txfees.v1beta1.Query/DenomSpotPrice", &txfeestypes.QueryDenomSpotPriceResponse{})
	StargateWhitelist.Store("/osmosis.txfees.v1beta1.Query/DenomPoolId", &txfeestypes.QueryDenomPoolIdResponse{})
	StargateWhitelist.Store("/osmosis.txfees.v1beta1.Query/BaseDenom", &txfeestypes.QueryBaseDenomResponse{})
}
