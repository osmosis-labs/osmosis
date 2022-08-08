package wasmbinding

import (
	"sync"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

// StargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var StargateWhitelist sync.Map

func init() {
	// TODO determine safety

	// auth
	StargateWhitelist.Store("/cosmos.auth.v1beta1.Query/Account", &authtypes.QueryAccountResponse{})
	StargateWhitelist.Store("/cosmos.auth.v1beta1.Query/Params", &authtypes.QueryParamsResponse{})
	// StargateWhitelist.Store("cosmos.auth.v1beta1.Query/Accounts")

	//authz
	// StargateWhitelist.Store("cosmos.authz.v1beta1.Query/Grants")

	// bank
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/Balance", &banktypes.QueryBalanceResponse{})
	// StargateWhitelist.Store("cosmos.bank.v1beta1.Query/AllBalances")
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/DenomMetadata", &banktypes.QueryDenomsMetadataResponse{})
	// StargateWhitelist.Store("cosmos.bank.v1beta1.Query/DenomsMetadatas")
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/Params", &banktypes.QueryParamsResponse{})
	// StargateWhitelist.Store("cosmos.bank.v1beta1.Query/TotalSupply")
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/SupplyOf", &banktypes.QuerySupplyOfResponse{})

	// distribution
	// StargateWhitelist.Store("cosmos.distribution.v1beta1.Query/CommunityPool")
	StargateWhitelist.Store("/cosmos.distribution.v1beta1.Query/Params", &distributiontypes.QueryParamsResponse{})
	// StargateWhitelist.Store("cosmos.distribution.v1beta1.Query/DelegationRewards")
	// StargateWhitelist.Store("cosmos.distribution.v1beta1.Query/DelegationTotalRewards")
	// StargateWhitelist.Store("cosmos.distribution.v1beta1.Query/DelegatorValidators")
	// StargateWhitelist.Store("cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress")
	StargateWhitelist.Store("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", &distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	StargateWhitelist.Store("/cosmos.distribution.v1beta1.Query/ValidatorCommission", &distributiontypes.QueryValidatorCommissionResponse{})
	// StargateWhitelist.Store("cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards")
	// StargateWhitelist.Store("cosmos.distribution.v1beta1.Query/ValidatorSlashes")

	// evidence
	// StargateWhitelist.Store("cosmos.evidence.v1beta1.Query/AllEvidence")
	StargateWhitelist.Store("cosmos.evidence.v1beta1.Query/Evidence", &evidencetypes.QueryEvidenceResponse{})

	// feegrant
	// StargateWhitelist.Store("cosmos.feegrant.v1beta1.Query/Allowance")
	// StargateWhitelist.Store("cosmos.feegrant.v1beta1.Query/Allowances")

	// gov
	StargateWhitelist.Store("cosmos.gov.v1beta1.Query/Deposit", &govtypes.QueryDepositResponse{})
	// StargateWhitelist.Store("cosmos.gov.v1beta1.Query/Deposits")
	StargateWhitelist.Store("cosmos.gov.v1beta1.Query/Params", &govtypes.QueryParamsResponse{})
	// StargateWhitelist.Store("cosmos.gov.v1beta1.Query/Proposal")
	// StargateWhitelist.Store("cosmos.gov.v1beta1.Query/Proposals")
	// StargateWhitelist.Store("cosmos.gov.v1beta1.Query/TallyResult")
	StargateWhitelist.Store("cosmos.gov.v1beta1.Query/Vote", &govtypes.QueryVoteResponse{})
	// StargateWhitelist.Store("cosmos.gov.v1beta1.Query/Votes")

	// params (don't add for now because this module will be removed)
	// StargateWhitelist.Store("cosmos.params.v1beta1.Query/Params", params)

	// slashing
	StargateWhitelist.Store("cosmos.slashing.v1beta1.Query/Params", &slashingtypes.QueryParamsResponse{})
	StargateWhitelist.Store("cosmos.slashing.v1beta1.Query/SigningInfo", &slashingtypes.QuerySigningInfoResponse{})
	// StargateWhitelist.Store("cosmos.slashing.v1beta1.Query/SigningInfos")

	// staking
	StargateWhitelist.Store("cosmos.staking.v1beta1.Query/Delegation", &stakingtypes.QueryDelegationResponse{})
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/DelegatorDelegations")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/DelegatorValidator")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/DelegatorValidators")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/HistoricalInfo")
	StargateWhitelist.Store("cosmos.staking.v1beta1.Query/Params", &stakingtypes.QueryParamsResponse{})
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/Pool")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/Redelegations")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/UnbondingDelegation")
	StargateWhitelist.Store("cosmos.staking.v1beta1.Query/Validator", &stakingtypes.QueryValidatorResponse{})
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/ValidatorDelegations")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations")
	// StargateWhitelist.Store("cosmos.staking.v1beta1.Query/Validators")

	// upgrade
	// StargateWhitelist.Store("cosmos.upgrade.v1beta1.Query/AppliedPlan")
	// StargateWhitelist.Store("cosmos.upgrade.v1beta1.Query/CurrentPlan")
	// StargateWhitelist.Store("cosmos.upgrade.v1beta1.Query/ModuleVersions")
	// StargateWhitelist.Store("cosmos.upgrade.v1beta1.Query/UpgradedConsensusState")

	// wasm
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/AllContractState")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/Code")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/Codes")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/ContractHistory")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/ContractInfo")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/ContractsByCode")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/PinnedCodes")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/RawContractState")
	// StargateWhitelist.Store("cosmwasm.wasm.v1.Query/SmartContractState")

	// ibc transfer
	// StargateWhitelist.Store("ibc.applications.transfer.v1.Query/DenomHash")
	// StargateWhitelist.Store("ibc.applications.transfer.v1.Query/DenomTrace")
	// StargateWhitelist.Store("ibc.applications.transfer.v1.Query/DenomTraces")
	// StargateWhitelist.Store("ibc.applications.transfer.v1.Query/Params")

	// mint
	// TODO

	// gamm
	// StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/Pools", )
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/NumPools", &gammtypes.QueryNumPoolsResponse{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/TotalLiquidity", &gammtypes.QueryTotalLiquidityResponse{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolResponse{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/PoolParams", &gammtypes.QueryPoolParamsResponse{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", &gammtypes.QueryTotalPoolLiquidityRequest{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/TotalShares", &gammtypes.QueryTotalSharesResponse{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/SpotPrice", &gammtypes.QuerySpotPriceResponse{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn", &gammtypes.QuerySwapExactAmountInResponse{})
	StargateWhitelist.Store("osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", &gammtypes.QuerySwapExactAmountOutResponse{})

	// epoch
	StargateWhitelist.Store("/osmosis.epochs.v1beta1.Query/EpochInfos", &epochtypes.QueryCurrentEpochResponse{})
}
