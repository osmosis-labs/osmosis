package wasm

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"

	epochstypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v10/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v10/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v10/x/pool-incentives/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v10/x/superfluid/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/types"
)

func StargateQuerier(queryRouter *baseapp.GRPCQueryRouter, codec codec.Codec) func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
		// reqBinding, whitelisted := StargateLayerRequestBindings.Load(request.Path)
		// if !whitelisted {
		// 	return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", request.Path)}
		// }
		_, whitelisted := StargateLayerRequestBindings.Load(request.Path)
		if !whitelisted {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", request.Path)}
		}

		route := queryRouter.Route(request.Path)
		if route == nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("No route to query '%s'", request.Path)}
		}

		req := abci.RequestQuery{
			Data: request.Data,
			Path: request.Path,
		}
		res, err := route(ctx, req)
		if err != nil {
			return nil, err
		}

		resBinding, whitelisted := StargateLayerResponseBindings.Load(request.Path)
		if !whitelisted {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", request.Path)}
		}

		bz, err := NormalizeReponsesAndJsonfy(resBinding, res.Value, codec)
		if err != nil {
			return nil, err
		}
		return bz, nil
	}
}

func NormalizeRequestsAndUnjsonfy(binding interface{}, bz []byte, codec codec.Codec) ([]byte, error) {
	// all values are proto message
	message, ok := binding.(proto.Message)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}

	err := codec.UnmarshalJSON(bz, message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	bz, err = proto.Marshal(message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	return bz, nil
}

func NormalizeReponsesAndJsonfy(binding interface{}, bz []byte, codec codec.Codec) ([]byte, error) {
	// all values are proto message
	message, ok := binding.(proto.Message)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}

	// unmarshal binary into stargate response data structure
	err := proto.Unmarshal(bz, message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	// build new deterministic response
	_, err = proto.Marshal(message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	// clear proto message
	message.Reset()

	err = proto.Unmarshal(bz, message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	// jsonfy
	bz, err = codec.MarshalJSON(message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	return bz, nil
}

var StargateLayerRequestBindings sync.Map
var StargateLayerResponseBindings sync.Map

func init() {
	StargateLayerRequestBindings.Store("/osmosis.epochs.v1beta1.Query/EpochInfos", &epochstypes.QueryEpochsInfoRequest{})
	StargateLayerResponseBindings.Store("/osmosis.epochs.v1beta1.Query/EpochInfos", &epochstypes.QueryEpochsInfoResponse{})

	StargateLayerRequestBindings.Store("/osmosis.epochs.v1beta1.Query/CurrentEpoch", &epochstypes.QueryCurrentEpochRequest{})
	StargateLayerResponseBindings.Store("/osmosis.epochs.v1beta1.Query/CurrentEpoch", &epochstypes.QueryCurrentEpochResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/Pools", &gammtypes.QueryPoolsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/Pools", &gammtypes.QueryPoolsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/NumPools", &gammtypes.QueryNumPoolsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/NumPools", &gammtypes.QueryNumPoolsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/PoolParams", &gammtypes.QueryPoolParamsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/PoolParams", &gammtypes.QueryPoolParamsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", &gammtypes.QueryTotalPoolLiquidityRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", &gammtypes.QueryTotalPoolLiquidityResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/TotalShares", &gammtypes.QueryTotalSharesRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/TotalShares", &gammtypes.QueryTotalSharesResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/SpotPrice", &gammtypes.QuerySpotPriceRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/SpotPrice", &gammtypes.QuerySpotPriceResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn", &gammtypes.QuerySwapExactAmountInRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn", &gammtypes.QuerySwapExactAmountInResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", &gammtypes.QuerySwapExactAmountOutRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", &gammtypes.QuerySwapExactAmountOutResponse{})

	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", &gammtypes.QuerySwapExactAmountOutRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", &gammtypes.QuerySwapExactAmountOutResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/ModuleToDistributeCoins", &incentivestypes.ModuleToDistributeCoinsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/ModuleToDistributeCoins", &incentivestypes.ModuleToDistributeCoinsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/ModuleDistributedCoins", &incentivestypes.ModuleDistributedCoinsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/ModuleDistributedCoins", &incentivestypes.ModuleDistributedCoinsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/GaugeByID", &incentivestypes.GaugeByIDRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/GaugeByID", &incentivestypes.GaugeByIDResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/Gauges", &incentivestypes.GaugesRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/Gauges", &incentivestypes.GaugesResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/ActiveGauges", &incentivestypes.ActiveGaugesRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/ActiveGauges", &incentivestypes.ActiveGaugesResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/ActiveGaugesPerDenom", &incentivestypes.ActiveGaugesPerDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/ActiveGaugesPerDenom", &incentivestypes.ActiveGaugesPerDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/UpcomingGauges", &incentivestypes.UpcomingGaugesRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/UpcomingGauges", &incentivestypes.UpcomingGaugesResponse{})

	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/UpcomingGaugesPerDenom", &incentivestypes.UpcomingGaugesPerDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/UpcomingGaugesPerDenom", &incentivestypes.UpcomingGaugesPerDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/ModuleBalance", &lockuptypes.ModuleBalanceRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/ModuleBalance", &lockuptypes.ModuleBalanceResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/ModuleLockedAmount", &lockuptypes.ModuleLockedAmountRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/ModuleLockedAmount", &lockuptypes.ModuleLockedAmountResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountUnlockableCoins", &lockuptypes.AccountUnlockableCoinsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountUnlockableCoins", &lockuptypes.AccountUnlockableCoinsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountUnlockingCoins", &lockuptypes.AccountUnlockingCoinsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountUnlockingCoins", &lockuptypes.AccountUnlockingCoinsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedCoins", &lockuptypes.AccountLockedCoinsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedCoins", &lockuptypes.AccountLockedCoinsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedPastTime", &lockuptypes.AccountLockedPastTimeRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedPastTime", &lockuptypes.AccountLockedPastTimeResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedPastTimeNotUnlockingOnly", &lockuptypes.AccountLockedPastTimeNotUnlockingOnlyRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedPastTimeNotUnlockingOnly", &lockuptypes.AccountLockedPastTimeNotUnlockingOnlyResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountUnlockedBeforeTime", &lockuptypes.AccountUnlockedBeforeTimeRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountUnlockedBeforeTime", &lockuptypes.AccountUnlockedBeforeTimeResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedPastTimeDenom", &lockuptypes.AccountLockedPastTimeDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedPastTimeDenom", &lockuptypes.AccountLockedPastTimeDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/LockedDenom", &lockuptypes.LockedDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/LockedDenom", &lockuptypes.LockedDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/LockedByID", &lockuptypes.LockedRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/LockedByID", &lockuptypes.LockedResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/SyntheticLockupsByLockupID", &lockuptypes.SyntheticLockupsByLockupIDRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/SyntheticLockupsByLockupID", &lockuptypes.SyntheticLockupsByLockupIDResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDuration", &lockuptypes.AccountLockedLongerDurationRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDuration", &lockuptypes.AccountLockedLongerDurationResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedDuration", &lockuptypes.AccountLockedDurationRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedDuration", &lockuptypes.AccountLockedDurationResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDurationNotUnlockingOnly", &lockuptypes.AccountLockedLongerDurationNotUnlockingOnlyRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDurationNotUnlockingOnly", &lockuptypes.AccountLockedLongerDurationNotUnlockingOnlyResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDurationDenom", &lockuptypes.AccountLockedLongerDurationDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDurationDenom", &lockuptypes.AccountLockedLongerDurationDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDurationDenom", &lockuptypes.AccountLockedLongerDurationDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.lockup.v1beta1.Query/AccountLockedLongerDurationDenom", &lockuptypes.AccountLockedLongerDurationDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.mint.v1beta1.Query/Params", &minttypes.QueryParamsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.mint.v1beta1.Query/Params", &minttypes.QueryParamsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.mint.v1beta1.Query/EpochProvisions", &minttypes.QueryEpochProvisionsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.mint.v1beta1.Query/EpochProvisions", &minttypes.QueryEpochProvisionsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.pool-incentives.v1beta1.Query/GaugeIds", &poolincentivestypes.QueryGaugeIdsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.pool-incentives.v1beta1.Query/GaugeIds", &poolincentivestypes.QueryGaugeIdsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.pool-incentives.v1beta1.Query/DistrInfo", &poolincentivestypes.QueryDistrInfoRequest{})
	StargateLayerResponseBindings.Store("/osmosis.pool-incentives.v1beta1.Query/DistrInfo", &poolincentivestypes.QueryDistrInfoResponse{})

	StargateLayerRequestBindings.Store("/osmosis.pool-incentives.v1beta1.Query/Params", &poolincentivestypes.QueryParamsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.pool-incentives.v1beta1.Query/Params", &poolincentivestypes.QueryParamsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.pool-incentives.v1beta1.Query/LockableDurations", &poolincentivestypes.QueryLockableDurationsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.pool-incentives.v1beta1.Query/LockableDurations", &poolincentivestypes.QueryLockableDurationsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.pool-incentives.v1beta1.Query/IncentivizedPools", &poolincentivestypes.QueryIncentivizedPoolsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.pool-incentives.v1beta1.Query/IncentivizedPools", &poolincentivestypes.QueryIncentivizedPoolsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.pool-incentives.v1beta1.Query/ExternalIncentiveGauges", &poolincentivestypes.QueryExternalIncentiveGaugesRequest{})
	StargateLayerResponseBindings.Store("/osmosis.pool-incentives.v1beta1.Query/ExternalIncentiveGauges", &poolincentivestypes.QueryExternalIncentiveGaugesResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/Params", &superfluidtypes.QueryParamsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/Params", &superfluidtypes.QueryParamsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/AssetType", &superfluidtypes.AssetTypeRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/AssetType", &superfluidtypes.AssetTypeResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/AllAssets", &superfluidtypes.AllAssetsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/AllAssets", &superfluidtypes.AllAssetsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/AssetMultiplier", &superfluidtypes.AssetMultiplierRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/AssetMultiplier", &superfluidtypes.AssetMultiplierResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/AllAssets", &superfluidtypes.AllAssetsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/AllAssets", &superfluidtypes.AllAssetsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/AllIntermediaryAccounts", &superfluidtypes.AllIntermediaryAccountsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/AllIntermediaryAccounts", &superfluidtypes.AllIntermediaryAccountsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/ConnectedIntermediaryAccount", &superfluidtypes.ConnectedIntermediaryAccountRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/ConnectedIntermediaryAccount", &superfluidtypes.ConnectedIntermediaryAccountResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/TotalSuperfluidDelegations", &superfluidtypes.TotalSuperfluidDelegationsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/TotalSuperfluidDelegations", &superfluidtypes.TotalSuperfluidDelegationsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidDelegationAmount", &superfluidtypes.SuperfluidDelegationAmountRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidDelegationAmount", &superfluidtypes.SuperfluidDelegationAmountResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidDelegationsByDelegator", &superfluidtypes.SuperfluidDelegationsByDelegatorRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidDelegationsByDelegator", &superfluidtypes.SuperfluidDelegationsByDelegatorResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidUndelegationsByDelegator", &superfluidtypes.SuperfluidUndelegationsByDelegatorRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidUndelegationsByDelegator", &superfluidtypes.SuperfluidUndelegationsByDelegatorResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidDelegationsByValidatorDenom", &superfluidtypes.SuperfluidDelegationsByValidatorDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/SuperfluidDelegationsByValidatorDenom", &superfluidtypes.SuperfluidDelegationsByValidatorDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.superfluid.v1beta1.Query/EstimateSuperfluidDelegatedAmountByValidatorDenom", &superfluidtypes.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{})
	StargateLayerResponseBindings.Store("/osmosis.superfluid.v1beta1.Query/EstimateSuperfluidDelegatedAmountByValidatorDenom", &superfluidtypes.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse{})

	StargateLayerRequestBindings.Store("/osmosis.tokenfactory.v1beta1.Query/Params", &tokenfactorytypes.QueryParamsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.tokenfactory.v1beta1.Query/Params", &tokenfactorytypes.QueryParamsResponse{})

	StargateLayerRequestBindings.Store("/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata", &tokenfactorytypes.QueryDenomAuthorityMetadataRequest{})
	StargateLayerResponseBindings.Store("/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata", &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{})

	StargateLayerRequestBindings.Store("/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator", &tokenfactorytypes.QueryDenomsFromCreatorRequest{})
	StargateLayerResponseBindings.Store("/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator", &tokenfactorytypes.QueryDenomsFromCreatorResponse{})

}
