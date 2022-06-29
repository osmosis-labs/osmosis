<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [osmosis/epochs/genesis.proto](#osmosis/epochs/genesis.proto)
    - [EpochInfo](#osmosis.epochs.v1beta1.EpochInfo)
    - [GenesisState](#osmosis.epochs.v1beta1.GenesisState)
  
- [osmosis/epochs/query.proto](#osmosis/epochs/query.proto)
    - [QueryCurrentEpochRequest](#osmosis.epochs.v1beta1.QueryCurrentEpochRequest)
    - [QueryCurrentEpochResponse](#osmosis.epochs.v1beta1.QueryCurrentEpochResponse)
    - [QueryEpochsInfoRequest](#osmosis.epochs.v1beta1.QueryEpochsInfoRequest)
    - [QueryEpochsInfoResponse](#osmosis.epochs.v1beta1.QueryEpochsInfoResponse)
  
    - [Query](#osmosis.epochs.v1beta1.Query)
  
- [osmosis/gamm/v1beta1/genesis.proto](#osmosis/gamm/v1beta1/genesis.proto)
    - [GenesisState](#osmosis.gamm.v1beta1.GenesisState)
    - [Params](#osmosis.gamm.v1beta1.Params)
  
- [osmosis/gamm/v1beta1/tx.proto](#osmosis/gamm/v1beta1/tx.proto)
    - [MsgExitPool](#osmosis.gamm.v1beta1.MsgExitPool)
    - [MsgExitPoolResponse](#osmosis.gamm.v1beta1.MsgExitPoolResponse)
    - [MsgExitSwapExternAmountOut](#osmosis.gamm.v1beta1.MsgExitSwapExternAmountOut)
    - [MsgExitSwapExternAmountOutResponse](#osmosis.gamm.v1beta1.MsgExitSwapExternAmountOutResponse)
    - [MsgExitSwapShareAmountIn](#osmosis.gamm.v1beta1.MsgExitSwapShareAmountIn)
    - [MsgExitSwapShareAmountInResponse](#osmosis.gamm.v1beta1.MsgExitSwapShareAmountInResponse)
    - [MsgJoinPool](#osmosis.gamm.v1beta1.MsgJoinPool)
    - [MsgJoinPoolResponse](#osmosis.gamm.v1beta1.MsgJoinPoolResponse)
    - [MsgJoinSwapExternAmountIn](#osmosis.gamm.v1beta1.MsgJoinSwapExternAmountIn)
    - [MsgJoinSwapExternAmountInResponse](#osmosis.gamm.v1beta1.MsgJoinSwapExternAmountInResponse)
    - [MsgJoinSwapShareAmountOut](#osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOut)
    - [MsgJoinSwapShareAmountOutResponse](#osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOutResponse)
    - [MsgSwapExactAmountIn](#osmosis.gamm.v1beta1.MsgSwapExactAmountIn)
    - [MsgSwapExactAmountInResponse](#osmosis.gamm.v1beta1.MsgSwapExactAmountInResponse)
    - [MsgSwapExactAmountOut](#osmosis.gamm.v1beta1.MsgSwapExactAmountOut)
    - [MsgSwapExactAmountOutResponse](#osmosis.gamm.v1beta1.MsgSwapExactAmountOutResponse)
    - [SwapAmountInRoute](#osmosis.gamm.v1beta1.SwapAmountInRoute)
    - [SwapAmountOutRoute](#osmosis.gamm.v1beta1.SwapAmountOutRoute)
  
    - [Msg](#osmosis.gamm.v1beta1.Msg)
  
- [osmosis/gamm/v1beta1/query.proto](#osmosis/gamm/v1beta1/query.proto)
    - [QueryNumPoolsRequest](#osmosis.gamm.v1beta1.QueryNumPoolsRequest)
    - [QueryNumPoolsResponse](#osmosis.gamm.v1beta1.QueryNumPoolsResponse)
    - [QueryPoolParamsRequest](#osmosis.gamm.v1beta1.QueryPoolParamsRequest)
    - [QueryPoolParamsResponse](#osmosis.gamm.v1beta1.QueryPoolParamsResponse)
    - [QueryPoolRequest](#osmosis.gamm.v1beta1.QueryPoolRequest)
    - [QueryPoolResponse](#osmosis.gamm.v1beta1.QueryPoolResponse)
    - [QueryPoolsRequest](#osmosis.gamm.v1beta1.QueryPoolsRequest)
    - [QueryPoolsResponse](#osmosis.gamm.v1beta1.QueryPoolsResponse)
    - [QuerySpotPriceRequest](#osmosis.gamm.v1beta1.QuerySpotPriceRequest)
    - [QuerySpotPriceResponse](#osmosis.gamm.v1beta1.QuerySpotPriceResponse)
    - [QuerySwapExactAmountInRequest](#osmosis.gamm.v1beta1.QuerySwapExactAmountInRequest)
    - [QuerySwapExactAmountInResponse](#osmosis.gamm.v1beta1.QuerySwapExactAmountInResponse)
    - [QuerySwapExactAmountOutRequest](#osmosis.gamm.v1beta1.QuerySwapExactAmountOutRequest)
    - [QuerySwapExactAmountOutResponse](#osmosis.gamm.v1beta1.QuerySwapExactAmountOutResponse)
    - [QueryTotalLiquidityRequest](#osmosis.gamm.v1beta1.QueryTotalLiquidityRequest)
    - [QueryTotalLiquidityResponse](#osmosis.gamm.v1beta1.QueryTotalLiquidityResponse)
    - [QueryTotalPoolLiquidityRequest](#osmosis.gamm.v1beta1.QueryTotalPoolLiquidityRequest)
    - [QueryTotalPoolLiquidityResponse](#osmosis.gamm.v1beta1.QueryTotalPoolLiquidityResponse)
    - [QueryTotalSharesRequest](#osmosis.gamm.v1beta1.QueryTotalSharesRequest)
    - [QueryTotalSharesResponse](#osmosis.gamm.v1beta1.QueryTotalSharesResponse)
  
    - [Query](#osmosis.gamm.v1beta1.Query)
  
- [osmosis/lockup/lock.proto](#osmosis/lockup/lock.proto)
    - [PeriodLock](#osmosis.lockup.PeriodLock)
    - [QueryCondition](#osmosis.lockup.QueryCondition)
    - [SyntheticLock](#osmosis.lockup.SyntheticLock)
  
    - [LockQueryType](#osmosis.lockup.LockQueryType)
  
- [osmosis/incentives/gauge.proto](#osmosis/incentives/gauge.proto)
    - [Gauge](#osmosis.incentives.Gauge)
    - [LockableDurationsInfo](#osmosis.incentives.LockableDurationsInfo)
  
- [osmosis/incentives/params.proto](#osmosis/incentives/params.proto)
    - [Params](#osmosis.incentives.Params)
  
- [osmosis/incentives/genesis.proto](#osmosis/incentives/genesis.proto)
    - [GenesisState](#osmosis.incentives.GenesisState)
  
- [osmosis/incentives/query.proto](#osmosis/incentives/query.proto)
    - [ActiveGaugesPerDenomRequest](#osmosis.incentives.ActiveGaugesPerDenomRequest)
    - [ActiveGaugesPerDenomResponse](#osmosis.incentives.ActiveGaugesPerDenomResponse)
    - [ActiveGaugesRequest](#osmosis.incentives.ActiveGaugesRequest)
    - [ActiveGaugesResponse](#osmosis.incentives.ActiveGaugesResponse)
    - [GaugeByIDRequest](#osmosis.incentives.GaugeByIDRequest)
    - [GaugeByIDResponse](#osmosis.incentives.GaugeByIDResponse)
    - [GaugesRequest](#osmosis.incentives.GaugesRequest)
    - [GaugesResponse](#osmosis.incentives.GaugesResponse)
    - [ModuleDistributedCoinsRequest](#osmosis.incentives.ModuleDistributedCoinsRequest)
    - [ModuleDistributedCoinsResponse](#osmosis.incentives.ModuleDistributedCoinsResponse)
    - [ModuleToDistributeCoinsRequest](#osmosis.incentives.ModuleToDistributeCoinsRequest)
    - [ModuleToDistributeCoinsResponse](#osmosis.incentives.ModuleToDistributeCoinsResponse)
    - [QueryLockableDurationsRequest](#osmosis.incentives.QueryLockableDurationsRequest)
    - [QueryLockableDurationsResponse](#osmosis.incentives.QueryLockableDurationsResponse)
    - [RewardsEstRequest](#osmosis.incentives.RewardsEstRequest)
    - [RewardsEstResponse](#osmosis.incentives.RewardsEstResponse)
    - [UpcomingGaugesPerDenomRequest](#osmosis.incentives.UpcomingGaugesPerDenomRequest)
    - [UpcomingGaugesPerDenomResponse](#osmosis.incentives.UpcomingGaugesPerDenomResponse)
    - [UpcomingGaugesRequest](#osmosis.incentives.UpcomingGaugesRequest)
    - [UpcomingGaugesResponse](#osmosis.incentives.UpcomingGaugesResponse)
  
    - [Query](#osmosis.incentives.Query)
  
- [osmosis/incentives/tx.proto](#osmosis/incentives/tx.proto)
    - [MsgAddToGauge](#osmosis.incentives.MsgAddToGauge)
    - [MsgAddToGaugeResponse](#osmosis.incentives.MsgAddToGaugeResponse)
    - [MsgCreateGauge](#osmosis.incentives.MsgCreateGauge)
    - [MsgCreateGaugeResponse](#osmosis.incentives.MsgCreateGaugeResponse)
  
    - [Msg](#osmosis.incentives.Msg)
  
- [osmosis/lockup/genesis.proto](#osmosis/lockup/genesis.proto)
    - [GenesisState](#osmosis.lockup.GenesisState)
  
- [osmosis/lockup/query.proto](#osmosis/lockup/query.proto)
    - [AccountLockedCoinsRequest](#osmosis.lockup.AccountLockedCoinsRequest)
    - [AccountLockedCoinsResponse](#osmosis.lockup.AccountLockedCoinsResponse)
    - [AccountLockedDurationRequest](#osmosis.lockup.AccountLockedDurationRequest)
    - [AccountLockedDurationResponse](#osmosis.lockup.AccountLockedDurationResponse)
    - [AccountLockedLongerDurationDenomRequest](#osmosis.lockup.AccountLockedLongerDurationDenomRequest)
    - [AccountLockedLongerDurationDenomResponse](#osmosis.lockup.AccountLockedLongerDurationDenomResponse)
    - [AccountLockedLongerDurationNotUnlockingOnlyRequest](#osmosis.lockup.AccountLockedLongerDurationNotUnlockingOnlyRequest)
    - [AccountLockedLongerDurationNotUnlockingOnlyResponse](#osmosis.lockup.AccountLockedLongerDurationNotUnlockingOnlyResponse)
    - [AccountLockedLongerDurationRequest](#osmosis.lockup.AccountLockedLongerDurationRequest)
    - [AccountLockedLongerDurationResponse](#osmosis.lockup.AccountLockedLongerDurationResponse)
    - [AccountLockedPastTimeDenomRequest](#osmosis.lockup.AccountLockedPastTimeDenomRequest)
    - [AccountLockedPastTimeDenomResponse](#osmosis.lockup.AccountLockedPastTimeDenomResponse)
    - [AccountLockedPastTimeNotUnlockingOnlyRequest](#osmosis.lockup.AccountLockedPastTimeNotUnlockingOnlyRequest)
    - [AccountLockedPastTimeNotUnlockingOnlyResponse](#osmosis.lockup.AccountLockedPastTimeNotUnlockingOnlyResponse)
    - [AccountLockedPastTimeRequest](#osmosis.lockup.AccountLockedPastTimeRequest)
    - [AccountLockedPastTimeResponse](#osmosis.lockup.AccountLockedPastTimeResponse)
    - [AccountUnlockableCoinsRequest](#osmosis.lockup.AccountUnlockableCoinsRequest)
    - [AccountUnlockableCoinsResponse](#osmosis.lockup.AccountUnlockableCoinsResponse)
    - [AccountUnlockedBeforeTimeRequest](#osmosis.lockup.AccountUnlockedBeforeTimeRequest)
    - [AccountUnlockedBeforeTimeResponse](#osmosis.lockup.AccountUnlockedBeforeTimeResponse)
    - [AccountUnlockingCoinsRequest](#osmosis.lockup.AccountUnlockingCoinsRequest)
    - [AccountUnlockingCoinsResponse](#osmosis.lockup.AccountUnlockingCoinsResponse)
    - [LockedDenomRequest](#osmosis.lockup.LockedDenomRequest)
    - [LockedDenomResponse](#osmosis.lockup.LockedDenomResponse)
    - [LockedRequest](#osmosis.lockup.LockedRequest)
    - [LockedResponse](#osmosis.lockup.LockedResponse)
    - [ModuleBalanceRequest](#osmosis.lockup.ModuleBalanceRequest)
    - [ModuleBalanceResponse](#osmosis.lockup.ModuleBalanceResponse)
    - [ModuleLockedAmountRequest](#osmosis.lockup.ModuleLockedAmountRequest)
    - [ModuleLockedAmountResponse](#osmosis.lockup.ModuleLockedAmountResponse)
    - [SyntheticLockupsByLockupIDRequest](#osmosis.lockup.SyntheticLockupsByLockupIDRequest)
    - [SyntheticLockupsByLockupIDResponse](#osmosis.lockup.SyntheticLockupsByLockupIDResponse)
  
    - [Query](#osmosis.lockup.Query)
  
- [osmosis/lockup/tx.proto](#osmosis/lockup/tx.proto)
    - [MsgBeginUnlocking](#osmosis.lockup.MsgBeginUnlocking)
    - [MsgBeginUnlockingAll](#osmosis.lockup.MsgBeginUnlockingAll)
    - [MsgBeginUnlockingAllResponse](#osmosis.lockup.MsgBeginUnlockingAllResponse)
    - [MsgBeginUnlockingResponse](#osmosis.lockup.MsgBeginUnlockingResponse)
    - [MsgExtendLockup](#osmosis.lockup.MsgExtendLockup)
    - [MsgExtendLockupResponse](#osmosis.lockup.MsgExtendLockupResponse)
    - [MsgLockTokens](#osmosis.lockup.MsgLockTokens)
    - [MsgLockTokensResponse](#osmosis.lockup.MsgLockTokensResponse)
  
    - [Msg](#osmosis.lockup.Msg)
  
- [osmosis/mint/v1beta1/mint.proto](#osmosis/mint/v1beta1/mint.proto)
    - [DistributionProportions](#osmosis.mint.v1beta1.DistributionProportions)
    - [Minter](#osmosis.mint.v1beta1.Minter)
    - [Params](#osmosis.mint.v1beta1.Params)
    - [WeightedAddress](#osmosis.mint.v1beta1.WeightedAddress)
  
- [osmosis/mint/v1beta1/genesis.proto](#osmosis/mint/v1beta1/genesis.proto)
    - [GenesisState](#osmosis.mint.v1beta1.GenesisState)
  
- [osmosis/mint/v1beta1/query.proto](#osmosis/mint/v1beta1/query.proto)
    - [QueryEpochProvisionsRequest](#osmosis.mint.v1beta1.QueryEpochProvisionsRequest)
    - [QueryEpochProvisionsResponse](#osmosis.mint.v1beta1.QueryEpochProvisionsResponse)
    - [QueryParamsRequest](#osmosis.mint.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#osmosis.mint.v1beta1.QueryParamsResponse)
  
    - [Query](#osmosis.mint.v1beta1.Query)
  
- [osmosis/pool-incentives/v1beta1/incentives.proto](#osmosis/pool-incentives/v1beta1/incentives.proto)
    - [DistrInfo](#osmosis.poolincentives.v1beta1.DistrInfo)
    - [DistrRecord](#osmosis.poolincentives.v1beta1.DistrRecord)
    - [LockableDurationsInfo](#osmosis.poolincentives.v1beta1.LockableDurationsInfo)
    - [Params](#osmosis.poolincentives.v1beta1.Params)
  
- [osmosis/pool-incentives/v1beta1/genesis.proto](#osmosis/pool-incentives/v1beta1/genesis.proto)
    - [GenesisState](#osmosis.poolincentives.v1beta1.GenesisState)
  
- [osmosis/pool-incentives/v1beta1/gov.proto](#osmosis/pool-incentives/v1beta1/gov.proto)
    - [ReplacePoolIncentivesProposal](#osmosis.poolincentives.v1beta1.ReplacePoolIncentivesProposal)
    - [UpdatePoolIncentivesProposal](#osmosis.poolincentives.v1beta1.UpdatePoolIncentivesProposal)
  
- [osmosis/pool-incentives/v1beta1/query.proto](#osmosis/pool-incentives/v1beta1/query.proto)
    - [IncentivizedPool](#osmosis.poolincentives.v1beta1.IncentivizedPool)
    - [QueryDistrInfoRequest](#osmosis.poolincentives.v1beta1.QueryDistrInfoRequest)
    - [QueryDistrInfoResponse](#osmosis.poolincentives.v1beta1.QueryDistrInfoResponse)
    - [QueryExternalIncentiveGaugesRequest](#osmosis.poolincentives.v1beta1.QueryExternalIncentiveGaugesRequest)
    - [QueryExternalIncentiveGaugesResponse](#osmosis.poolincentives.v1beta1.QueryExternalIncentiveGaugesResponse)
    - [QueryGaugeIdsRequest](#osmosis.poolincentives.v1beta1.QueryGaugeIdsRequest)
    - [QueryGaugeIdsResponse](#osmosis.poolincentives.v1beta1.QueryGaugeIdsResponse)
    - [QueryGaugeIdsResponse.GaugeIdWithDuration](#osmosis.poolincentives.v1beta1.QueryGaugeIdsResponse.GaugeIdWithDuration)
    - [QueryIncentivizedPoolsRequest](#osmosis.poolincentives.v1beta1.QueryIncentivizedPoolsRequest)
    - [QueryIncentivizedPoolsResponse](#osmosis.poolincentives.v1beta1.QueryIncentivizedPoolsResponse)
    - [QueryLockableDurationsRequest](#osmosis.poolincentives.v1beta1.QueryLockableDurationsRequest)
    - [QueryLockableDurationsResponse](#osmosis.poolincentives.v1beta1.QueryLockableDurationsResponse)
    - [QueryParamsRequest](#osmosis.poolincentives.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#osmosis.poolincentives.v1beta1.QueryParamsResponse)
  
    - [Query](#osmosis.poolincentives.v1beta1.Query)
  
- [osmosis/store/v1beta1/tree.proto](#osmosis/store/v1beta1/tree.proto)
    - [Child](#osmosis.store.v1beta1.Child)
    - [Leaf](#osmosis.store.v1beta1.Leaf)
    - [Node](#osmosis.store.v1beta1.Node)
  
- [osmosis/superfluid/superfluid.proto](#osmosis/superfluid/superfluid.proto)
    - [LockIdIntermediaryAccountConnection](#osmosis.superfluid.LockIdIntermediaryAccountConnection)
    - [OsmoEquivalentMultiplierRecord](#osmosis.superfluid.OsmoEquivalentMultiplierRecord)
    - [SuperfluidAsset](#osmosis.superfluid.SuperfluidAsset)
    - [SuperfluidDelegationRecord](#osmosis.superfluid.SuperfluidDelegationRecord)
    - [SuperfluidIntermediaryAccount](#osmosis.superfluid.SuperfluidIntermediaryAccount)
    - [UnpoolWhitelistedPools](#osmosis.superfluid.UnpoolWhitelistedPools)
  
    - [SuperfluidAssetType](#osmosis.superfluid.SuperfluidAssetType)
  
- [osmosis/superfluid/params.proto](#osmosis/superfluid/params.proto)
    - [Params](#osmosis.superfluid.Params)
  
- [osmosis/superfluid/genesis.proto](#osmosis/superfluid/genesis.proto)
    - [GenesisState](#osmosis.superfluid.GenesisState)
  
- [osmosis/superfluid/gov.proto](#osmosis/superfluid/gov.proto)
    - [RemoveSuperfluidAssetsProposal](#osmosis.superfluid.v1beta1.RemoveSuperfluidAssetsProposal)
    - [SetSuperfluidAssetsProposal](#osmosis.superfluid.v1beta1.SetSuperfluidAssetsProposal)
  
- [osmosis/superfluid/query.proto](#osmosis/superfluid/query.proto)
    - [AllAssetsRequest](#osmosis.superfluid.AllAssetsRequest)
    - [AllAssetsResponse](#osmosis.superfluid.AllAssetsResponse)
    - [AllIntermediaryAccountsRequest](#osmosis.superfluid.AllIntermediaryAccountsRequest)
    - [AllIntermediaryAccountsResponse](#osmosis.superfluid.AllIntermediaryAccountsResponse)
    - [AssetMultiplierRequest](#osmosis.superfluid.AssetMultiplierRequest)
    - [AssetMultiplierResponse](#osmosis.superfluid.AssetMultiplierResponse)
    - [AssetTypeRequest](#osmosis.superfluid.AssetTypeRequest)
    - [AssetTypeResponse](#osmosis.superfluid.AssetTypeResponse)
    - [ConnectedIntermediaryAccountRequest](#osmosis.superfluid.ConnectedIntermediaryAccountRequest)
    - [ConnectedIntermediaryAccountResponse](#osmosis.superfluid.ConnectedIntermediaryAccountResponse)
    - [EstimateSuperfluidDelegatedAmountByValidatorDenomRequest](#osmosis.superfluid.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest)
    - [EstimateSuperfluidDelegatedAmountByValidatorDenomResponse](#osmosis.superfluid.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse)
    - [QueryParamsRequest](#osmosis.superfluid.QueryParamsRequest)
    - [QueryParamsResponse](#osmosis.superfluid.QueryParamsResponse)
    - [SuperfluidDelegationAmountRequest](#osmosis.superfluid.SuperfluidDelegationAmountRequest)
    - [SuperfluidDelegationAmountResponse](#osmosis.superfluid.SuperfluidDelegationAmountResponse)
    - [SuperfluidDelegationsByDelegatorRequest](#osmosis.superfluid.SuperfluidDelegationsByDelegatorRequest)
    - [SuperfluidDelegationsByDelegatorResponse](#osmosis.superfluid.SuperfluidDelegationsByDelegatorResponse)
    - [SuperfluidDelegationsByValidatorDenomRequest](#osmosis.superfluid.SuperfluidDelegationsByValidatorDenomRequest)
    - [SuperfluidDelegationsByValidatorDenomResponse](#osmosis.superfluid.SuperfluidDelegationsByValidatorDenomResponse)
    - [SuperfluidIntermediaryAccountInfo](#osmosis.superfluid.SuperfluidIntermediaryAccountInfo)
    - [SuperfluidUndelegationsByDelegatorRequest](#osmosis.superfluid.SuperfluidUndelegationsByDelegatorRequest)
    - [SuperfluidUndelegationsByDelegatorResponse](#osmosis.superfluid.SuperfluidUndelegationsByDelegatorResponse)
    - [TotalSuperfluidDelegationsRequest](#osmosis.superfluid.TotalSuperfluidDelegationsRequest)
    - [TotalSuperfluidDelegationsResponse](#osmosis.superfluid.TotalSuperfluidDelegationsResponse)
  
    - [Query](#osmosis.superfluid.Query)
  
- [osmosis/superfluid/tx.proto](#osmosis/superfluid/tx.proto)
    - [MsgLockAndSuperfluidDelegate](#osmosis.superfluid.MsgLockAndSuperfluidDelegate)
    - [MsgLockAndSuperfluidDelegateResponse](#osmosis.superfluid.MsgLockAndSuperfluidDelegateResponse)
    - [MsgSuperfluidDelegate](#osmosis.superfluid.MsgSuperfluidDelegate)
    - [MsgSuperfluidDelegateResponse](#osmosis.superfluid.MsgSuperfluidDelegateResponse)
    - [MsgSuperfluidUnbondLock](#osmosis.superfluid.MsgSuperfluidUnbondLock)
    - [MsgSuperfluidUnbondLockResponse](#osmosis.superfluid.MsgSuperfluidUnbondLockResponse)
    - [MsgSuperfluidUndelegate](#osmosis.superfluid.MsgSuperfluidUndelegate)
    - [MsgSuperfluidUndelegateResponse](#osmosis.superfluid.MsgSuperfluidUndelegateResponse)
    - [MsgUnPoolWhitelistedPool](#osmosis.superfluid.MsgUnPoolWhitelistedPool)
    - [MsgUnPoolWhitelistedPoolResponse](#osmosis.superfluid.MsgUnPoolWhitelistedPoolResponse)
  
    - [Msg](#osmosis.superfluid.Msg)
  
- [osmosis/tokenfactory/v1beta1/authorityMetadata.proto](#osmosis/tokenfactory/v1beta1/authorityMetadata.proto)
    - [DenomAuthorityMetadata](#osmosis.tokenfactory.v1beta1.DenomAuthorityMetadata)
  
- [osmosis/tokenfactory/v1beta1/params.proto](#osmosis/tokenfactory/v1beta1/params.proto)
    - [Params](#osmosis.tokenfactory.v1beta1.Params)
  
- [osmosis/tokenfactory/v1beta1/genesis.proto](#osmosis/tokenfactory/v1beta1/genesis.proto)
    - [GenesisDenom](#osmosis.tokenfactory.v1beta1.GenesisDenom)
    - [GenesisState](#osmosis.tokenfactory.v1beta1.GenesisState)
  
- [osmosis/tokenfactory/v1beta1/query.proto](#osmosis/tokenfactory/v1beta1/query.proto)
    - [QueryDenomAuthorityMetadataRequest](#osmosis.tokenfactory.v1beta1.QueryDenomAuthorityMetadataRequest)
    - [QueryDenomAuthorityMetadataResponse](#osmosis.tokenfactory.v1beta1.QueryDenomAuthorityMetadataResponse)
    - [QueryDenomsFromCreatorRequest](#osmosis.tokenfactory.v1beta1.QueryDenomsFromCreatorRequest)
    - [QueryDenomsFromCreatorResponse](#osmosis.tokenfactory.v1beta1.QueryDenomsFromCreatorResponse)
    - [QueryParamsRequest](#osmosis.tokenfactory.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#osmosis.tokenfactory.v1beta1.QueryParamsResponse)
  
    - [Query](#osmosis.tokenfactory.v1beta1.Query)
  
- [osmosis/tokenfactory/v1beta1/tx.proto](#osmosis/tokenfactory/v1beta1/tx.proto)
    - [MsgBurn](#osmosis.tokenfactory.v1beta1.MsgBurn)
    - [MsgBurnResponse](#osmosis.tokenfactory.v1beta1.MsgBurnResponse)
    - [MsgChangeAdmin](#osmosis.tokenfactory.v1beta1.MsgChangeAdmin)
    - [MsgChangeAdminResponse](#osmosis.tokenfactory.v1beta1.MsgChangeAdminResponse)
    - [MsgCreateDenom](#osmosis.tokenfactory.v1beta1.MsgCreateDenom)
    - [MsgCreateDenomResponse](#osmosis.tokenfactory.v1beta1.MsgCreateDenomResponse)
    - [MsgMint](#osmosis.tokenfactory.v1beta1.MsgMint)
    - [MsgMintResponse](#osmosis.tokenfactory.v1beta1.MsgMintResponse)
  
    - [Msg](#osmosis.tokenfactory.v1beta1.Msg)
  
- [osmosis/txfees/v1beta1/feetoken.proto](#osmosis/txfees/v1beta1/feetoken.proto)
    - [FeeToken](#osmosis.txfees.v1beta1.FeeToken)
  
- [osmosis/txfees/v1beta1/genesis.proto](#osmosis/txfees/v1beta1/genesis.proto)
    - [GenesisState](#osmosis.txfees.v1beta1.GenesisState)
  
- [osmosis/txfees/v1beta1/gov.proto](#osmosis/txfees/v1beta1/gov.proto)
    - [UpdateFeeTokenProposal](#osmosis.txfees.v1beta1.UpdateFeeTokenProposal)
  
- [osmosis/txfees/v1beta1/query.proto](#osmosis/txfees/v1beta1/query.proto)
    - [QueryBaseDenomRequest](#osmosis.txfees.v1beta1.QueryBaseDenomRequest)
    - [QueryBaseDenomResponse](#osmosis.txfees.v1beta1.QueryBaseDenomResponse)
    - [QueryDenomPoolIdRequest](#osmosis.txfees.v1beta1.QueryDenomPoolIdRequest)
    - [QueryDenomPoolIdResponse](#osmosis.txfees.v1beta1.QueryDenomPoolIdResponse)
    - [QueryDenomSpotPriceRequest](#osmosis.txfees.v1beta1.QueryDenomSpotPriceRequest)
    - [QueryDenomSpotPriceResponse](#osmosis.txfees.v1beta1.QueryDenomSpotPriceResponse)
    - [QueryFeeTokensRequest](#osmosis.txfees.v1beta1.QueryFeeTokensRequest)
    - [QueryFeeTokensResponse](#osmosis.txfees.v1beta1.QueryFeeTokensResponse)
  
    - [Query](#osmosis.txfees.v1beta1.Query)
  
- [Scalar Value Types](#scalar-value-types)



<a name="osmosis/epochs/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/epochs/genesis.proto



<a name="osmosis.epochs.v1beta1.EpochInfo"></a>

### EpochInfo
EpochInfo is a struct that describes the data going into
a timer defined by the x/epochs module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `identifier` | [string](#string) |  | identifier is a unique reference to this particular timer. |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | start_time is the time at which the timer first ever ticks. If start_time is in the future, the epoch will not begin until the start time. |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  | duration is the time in between epoch ticks. In order for intended behavior to be met, duration should be greater than the chains expected block time. Duration must be non-zero. |
| `current_epoch` | [int64](#int64) |  | current_epoch is the current epoch number, or in other words, how many times has the timer 'ticked'. The first tick (current_epoch=1) is defined as the first block whose blocktime is greater than the EpochInfo start_time. |
| `current_epoch_start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | current_epoch_start_time describes the start time of the current timer interval. The interval is (current_epoch_start_time, current_epoch_start_time + duration] When the timer ticks, this is set to current_epoch_start_time = last_epoch_start_time + duration only one timer tick for a given identifier can occur per block.

NOTE! The current_epoch_start_time may diverge significantly from the wall-clock time the epoch began at. Wall-clock time of epoch start may be >> current_epoch_start_time. Suppose current_epoch_start_time = 10, duration = 5. Suppose the chain goes offline at t=14, and comes back online at t=30, and produces blocks at every successive time. (t=31, 32, etc.) * The t=30 block will start the epoch for (10, 15] * The t=31 block will start the epoch for (15, 20] * The t=32 block will start the epoch for (20, 25] * The t=33 block will start the epoch for (25, 30] * The t=34 block will start the epoch for (30, 35] * The **t=36** block will start the epoch for (35, 40] |
| `epoch_counting_started` | [bool](#bool) |  | epoch_counting_started is a boolean, that indicates whether this epoch timer has began yet. |
| `current_epoch_start_height` | [int64](#int64) |  | current_epoch_start_height is the block height at which the current epoch started. (The block height at which the timer last ticked) |






<a name="osmosis.epochs.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the epochs module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epochs` | [EpochInfo](#osmosis.epochs.v1beta1.EpochInfo) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/epochs/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/epochs/query.proto



<a name="osmosis.epochs.v1beta1.QueryCurrentEpochRequest"></a>

### QueryCurrentEpochRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `identifier` | [string](#string) |  |  |






<a name="osmosis.epochs.v1beta1.QueryCurrentEpochResponse"></a>

### QueryCurrentEpochResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `current_epoch` | [int64](#int64) |  |  |






<a name="osmosis.epochs.v1beta1.QueryEpochsInfoRequest"></a>

### QueryEpochsInfoRequest







<a name="osmosis.epochs.v1beta1.QueryEpochsInfoResponse"></a>

### QueryEpochsInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epochs` | [EpochInfo](#osmosis.epochs.v1beta1.EpochInfo) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.epochs.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `EpochInfos` | [QueryEpochsInfoRequest](#osmosis.epochs.v1beta1.QueryEpochsInfoRequest) | [QueryEpochsInfoResponse](#osmosis.epochs.v1beta1.QueryEpochsInfoResponse) | EpochInfos provide running epochInfos | GET|/osmosis/epochs/v1beta1/epochs|
| `CurrentEpoch` | [QueryCurrentEpochRequest](#osmosis.epochs.v1beta1.QueryCurrentEpochRequest) | [QueryCurrentEpochResponse](#osmosis.epochs.v1beta1.QueryCurrentEpochResponse) | CurrentEpoch provide current epoch of specified identifier | GET|/osmosis/epochs/v1beta1/current_epoch|

 <!-- end services -->



<a name="osmosis/gamm/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/gamm/v1beta1/genesis.proto



<a name="osmosis.gamm.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the gamm module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pools` | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |
| `next_pool_number` | [uint64](#uint64) |  |  |
| `params` | [Params](#osmosis.gamm.v1beta1.Params) |  |  |






<a name="osmosis.gamm.v1beta1.Params"></a>

### Params
Params holds parameters for the incentives module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/gamm/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/gamm/v1beta1/tx.proto



<a name="osmosis.gamm.v1beta1.MsgExitPool"></a>

### MsgExitPool
===================== MsgExitPool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `shareInAmount` | [string](#string) |  |  |
| `tokenOutMins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.gamm.v1beta1.MsgExitPoolResponse"></a>

### MsgExitPoolResponse







<a name="osmosis.gamm.v1beta1.MsgExitSwapExternAmountOut"></a>

### MsgExitSwapExternAmountOut
===================== MsgExitSwapExternAmountOut


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `tokenOut` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `shareInMaxAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgExitSwapExternAmountOutResponse"></a>

### MsgExitSwapExternAmountOutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `shareInAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgExitSwapShareAmountIn"></a>

### MsgExitSwapShareAmountIn
===================== MsgExitSwapShareAmountIn


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `tokenOutDenom` | [string](#string) |  |  |
| `shareInAmount` | [string](#string) |  |  |
| `tokenOutMinAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgExitSwapShareAmountInResponse"></a>

### MsgExitSwapShareAmountInResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenOutAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgJoinPool"></a>

### MsgJoinPool
===================== MsgJoinPool
This is really MsgJoinPoolNoSwap


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `shareOutAmount` | [string](#string) |  |  |
| `tokenInMaxs` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.gamm.v1beta1.MsgJoinPoolResponse"></a>

### MsgJoinPoolResponse







<a name="osmosis.gamm.v1beta1.MsgJoinSwapExternAmountIn"></a>

### MsgJoinSwapExternAmountIn
===================== MsgJoinSwapExternAmountIn
TODO: Rename to MsgJoinSwapExactAmountIn


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `tokenIn` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `shareOutMinAmount` | [string](#string) |  | repeated cosmos.base.v1beta1.Coin tokensIn = 5 [ (gogoproto.moretags) = "yaml:\"tokens_in\"", (gogoproto.nullable) = false ]; |






<a name="osmosis.gamm.v1beta1.MsgJoinSwapExternAmountInResponse"></a>

### MsgJoinSwapExternAmountInResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `shareOutAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOut"></a>

### MsgJoinSwapShareAmountOut
===================== MsgJoinSwapShareAmountOut


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `tokenInDenom` | [string](#string) |  |  |
| `shareOutAmount` | [string](#string) |  |  |
| `tokenInMaxAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOutResponse"></a>

### MsgJoinSwapShareAmountOutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenInAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgSwapExactAmountIn"></a>

### MsgSwapExactAmountIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `routes` | [SwapAmountInRoute](#osmosis.gamm.v1beta1.SwapAmountInRoute) | repeated |  |
| `tokenIn` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `tokenOutMinAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgSwapExactAmountInResponse"></a>

### MsgSwapExactAmountInResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenOutAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.MsgSwapExactAmountOut"></a>

### MsgSwapExactAmountOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `routes` | [SwapAmountOutRoute](#osmosis.gamm.v1beta1.SwapAmountOutRoute) | repeated |  |
| `tokenInMaxAmount` | [string](#string) |  |  |
| `tokenOut` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="osmosis.gamm.v1beta1.MsgSwapExactAmountOutResponse"></a>

### MsgSwapExactAmountOutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenInAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.SwapAmountInRoute"></a>

### SwapAmountInRoute
===================== MsgSwapExactAmountIn


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolId` | [uint64](#uint64) |  |  |
| `tokenOutDenom` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.SwapAmountOutRoute"></a>

### SwapAmountOutRoute
===================== MsgSwapExactAmountOut


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolId` | [uint64](#uint64) |  |  |
| `tokenInDenom` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.gamm.v1beta1.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `JoinPool` | [MsgJoinPool](#osmosis.gamm.v1beta1.MsgJoinPool) | [MsgJoinPoolResponse](#osmosis.gamm.v1beta1.MsgJoinPoolResponse) |  | |
| `ExitPool` | [MsgExitPool](#osmosis.gamm.v1beta1.MsgExitPool) | [MsgExitPoolResponse](#osmosis.gamm.v1beta1.MsgExitPoolResponse) |  | |
| `SwapExactAmountIn` | [MsgSwapExactAmountIn](#osmosis.gamm.v1beta1.MsgSwapExactAmountIn) | [MsgSwapExactAmountInResponse](#osmosis.gamm.v1beta1.MsgSwapExactAmountInResponse) |  | |
| `SwapExactAmountOut` | [MsgSwapExactAmountOut](#osmosis.gamm.v1beta1.MsgSwapExactAmountOut) | [MsgSwapExactAmountOutResponse](#osmosis.gamm.v1beta1.MsgSwapExactAmountOutResponse) |  | |
| `JoinSwapExternAmountIn` | [MsgJoinSwapExternAmountIn](#osmosis.gamm.v1beta1.MsgJoinSwapExternAmountIn) | [MsgJoinSwapExternAmountInResponse](#osmosis.gamm.v1beta1.MsgJoinSwapExternAmountInResponse) |  | |
| `JoinSwapShareAmountOut` | [MsgJoinSwapShareAmountOut](#osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOut) | [MsgJoinSwapShareAmountOutResponse](#osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOutResponse) |  | |
| `ExitSwapExternAmountOut` | [MsgExitSwapExternAmountOut](#osmosis.gamm.v1beta1.MsgExitSwapExternAmountOut) | [MsgExitSwapExternAmountOutResponse](#osmosis.gamm.v1beta1.MsgExitSwapExternAmountOutResponse) |  | |
| `ExitSwapShareAmountIn` | [MsgExitSwapShareAmountIn](#osmosis.gamm.v1beta1.MsgExitSwapShareAmountIn) | [MsgExitSwapShareAmountInResponse](#osmosis.gamm.v1beta1.MsgExitSwapShareAmountInResponse) |  | |

 <!-- end services -->



<a name="osmosis/gamm/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/gamm/v1beta1/query.proto



<a name="osmosis.gamm.v1beta1.QueryNumPoolsRequest"></a>

### QueryNumPoolsRequest
=============================== NumPools






<a name="osmosis.gamm.v1beta1.QueryNumPoolsResponse"></a>

### QueryNumPoolsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `numPools` | [uint64](#uint64) |  |  |






<a name="osmosis.gamm.v1beta1.QueryPoolParamsRequest"></a>

### QueryPoolParamsRequest
=============================== PoolParams


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolId` | [uint64](#uint64) |  |  |






<a name="osmosis.gamm.v1beta1.QueryPoolParamsResponse"></a>

### QueryPoolParamsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="osmosis.gamm.v1beta1.QueryPoolRequest"></a>

### QueryPoolRequest
=============================== Pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolId` | [uint64](#uint64) |  |  |






<a name="osmosis.gamm.v1beta1.QueryPoolResponse"></a>

### QueryPoolResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="osmosis.gamm.v1beta1.QueryPoolsRequest"></a>

### QueryPoolsRequest
=============================== Pools


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="osmosis.gamm.v1beta1.QueryPoolsResponse"></a>

### QueryPoolsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pools` | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="osmosis.gamm.v1beta1.QuerySpotPriceRequest"></a>

### QuerySpotPriceRequest
QuerySpotPriceRequest defines the gRPC request structure for a SpotPrice
query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolId` | [uint64](#uint64) |  |  |
| `base_asset_denom` | [string](#string) |  |  |
| `quote_asset_denom` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.QuerySpotPriceResponse"></a>

### QuerySpotPriceResponse
QuerySpotPriceResponse defines the gRPC response structure for a SpotPrice
query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `spotPrice` | [string](#string) |  | String of the Dec. Ex) 10.203uatom |






<a name="osmosis.gamm.v1beta1.QuerySwapExactAmountInRequest"></a>

### QuerySwapExactAmountInRequest
=============================== EstimateSwapExactAmountIn


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `tokenIn` | [string](#string) |  |  |
| `routes` | [SwapAmountInRoute](#osmosis.gamm.v1beta1.SwapAmountInRoute) | repeated |  |






<a name="osmosis.gamm.v1beta1.QuerySwapExactAmountInResponse"></a>

### QuerySwapExactAmountInResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenOutAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.QuerySwapExactAmountOutRequest"></a>

### QuerySwapExactAmountOutRequest
=============================== EstimateSwapExactAmountOut


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `poolId` | [uint64](#uint64) |  |  |
| `routes` | [SwapAmountOutRoute](#osmosis.gamm.v1beta1.SwapAmountOutRoute) | repeated |  |
| `tokenOut` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.QuerySwapExactAmountOutResponse"></a>

### QuerySwapExactAmountOutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenInAmount` | [string](#string) |  |  |






<a name="osmosis.gamm.v1beta1.QueryTotalLiquidityRequest"></a>

### QueryTotalLiquidityRequest







<a name="osmosis.gamm.v1beta1.QueryTotalLiquidityResponse"></a>

### QueryTotalLiquidityResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `liquidity` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.gamm.v1beta1.QueryTotalPoolLiquidityRequest"></a>

### QueryTotalPoolLiquidityRequest
=============================== PoolLiquidity


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolId` | [uint64](#uint64) |  |  |






<a name="osmosis.gamm.v1beta1.QueryTotalPoolLiquidityResponse"></a>

### QueryTotalPoolLiquidityResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `liquidity` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.gamm.v1beta1.QueryTotalSharesRequest"></a>

### QueryTotalSharesRequest
=============================== TotalShares


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolId` | [uint64](#uint64) |  |  |






<a name="osmosis.gamm.v1beta1.QueryTotalSharesResponse"></a>

### QueryTotalSharesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `totalShares` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.gamm.v1beta1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Pools` | [QueryPoolsRequest](#osmosis.gamm.v1beta1.QueryPoolsRequest) | [QueryPoolsResponse](#osmosis.gamm.v1beta1.QueryPoolsResponse) |  | GET|/osmosis/gamm/v1beta1/pools|
| `NumPools` | [QueryNumPoolsRequest](#osmosis.gamm.v1beta1.QueryNumPoolsRequest) | [QueryNumPoolsResponse](#osmosis.gamm.v1beta1.QueryNumPoolsResponse) |  | GET|/osmosis/gamm/v1beta1/num_pools|
| `TotalLiquidity` | [QueryTotalLiquidityRequest](#osmosis.gamm.v1beta1.QueryTotalLiquidityRequest) | [QueryTotalLiquidityResponse](#osmosis.gamm.v1beta1.QueryTotalLiquidityResponse) |  | GET|/osmosis/gamm/v1beta1/total_liquidity|
| `Pool` | [QueryPoolRequest](#osmosis.gamm.v1beta1.QueryPoolRequest) | [QueryPoolResponse](#osmosis.gamm.v1beta1.QueryPoolResponse) | Per Pool gRPC Endpoints | GET|/osmosis/gamm/v1beta1/pools/{poolId}|
| `PoolParams` | [QueryPoolParamsRequest](#osmosis.gamm.v1beta1.QueryPoolParamsRequest) | [QueryPoolParamsResponse](#osmosis.gamm.v1beta1.QueryPoolParamsResponse) |  | GET|/osmosis/gamm/v1beta1/pools/{poolId}/params|
| `TotalPoolLiquidity` | [QueryTotalPoolLiquidityRequest](#osmosis.gamm.v1beta1.QueryTotalPoolLiquidityRequest) | [QueryTotalPoolLiquidityResponse](#osmosis.gamm.v1beta1.QueryTotalPoolLiquidityResponse) |  | GET|/osmosis/gamm/v1beta1/pools/{poolId}/total_pool_liquidity|
| `TotalShares` | [QueryTotalSharesRequest](#osmosis.gamm.v1beta1.QueryTotalSharesRequest) | [QueryTotalSharesResponse](#osmosis.gamm.v1beta1.QueryTotalSharesResponse) |  | GET|/osmosis/gamm/v1beta1/pools/{poolId}/total_shares|
| `SpotPrice` | [QuerySpotPriceRequest](#osmosis.gamm.v1beta1.QuerySpotPriceRequest) | [QuerySpotPriceResponse](#osmosis.gamm.v1beta1.QuerySpotPriceResponse) | SpotPrice defines a gRPC query handler that returns the spot price given a base denomination and a quote denomination. | GET|/osmosis/gamm/v1beta1/pools/{poolId}/prices|
| `EstimateSwapExactAmountIn` | [QuerySwapExactAmountInRequest](#osmosis.gamm.v1beta1.QuerySwapExactAmountInRequest) | [QuerySwapExactAmountInResponse](#osmosis.gamm.v1beta1.QuerySwapExactAmountInResponse) | Estimate the swap. | GET|/osmosis/gamm/v1beta1/{poolId}/estimate/swap_exact_amount_in|
| `EstimateSwapExactAmountOut` | [QuerySwapExactAmountOutRequest](#osmosis.gamm.v1beta1.QuerySwapExactAmountOutRequest) | [QuerySwapExactAmountOutResponse](#osmosis.gamm.v1beta1.QuerySwapExactAmountOutResponse) |  | GET|/osmosis/gamm/v1beta1/{poolId}/estimate/swap_exact_amount_out|

 <!-- end services -->



<a name="osmosis/lockup/lock.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/lockup/lock.proto



<a name="osmosis.lockup.PeriodLock"></a>

### PeriodLock
PeriodLock is a single unit of lock by period. It's a record of locked coin
at a specific time. It stores owner, duration, unlock time and the amount of
coins locked.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ID` | [uint64](#uint64) |  |  |
| `owner` | [string](#string) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `end_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.lockup.QueryCondition"></a>

### QueryCondition



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lock_query_type` | [LockQueryType](#osmosis.lockup.LockQueryType) |  | type of lock query, ByLockDuration | ByLockTime |
| `denom` | [string](#string) |  | What token denomination are we looking for lockups of |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  | valid when query condition is ByDuration |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | valid when query condition is ByTime |






<a name="osmosis.lockup.SyntheticLock"></a>

### SyntheticLock
SyntheticLock is a single unit of synthetic lockup
TODO: Change this to have
* underlying_lock_id
* synthetic_coin
* end_time
* duration
* owner
We then index synthetic locks by the denom, just like we do with normal
locks. Ideally we even get an interface, so we can re-use that same logic.
I currently have no idea how reward distribution is supposed to be working...
EVENTUALLY
we make a "constrained_coin" field, which is what the current "coins" field
is. Constrained coin field can be a #post-v7 feature, since we aren't
allowing partial unlocks of synthetic lockups.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `underlying_lock_id` | [uint64](#uint64) |  | underlying native lockup id for this synthetic lockup |
| `synth_denom` | [string](#string) |  |  |
| `end_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | used for unbonding synthetic lockups, for active synthetic lockups, this value is set to uninitialized value |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |





 <!-- end messages -->


<a name="osmosis.lockup.LockQueryType"></a>

### LockQueryType


| Name | Number | Description |
| ---- | ------ | ----------- |
| ByDuration | 0 | Queries for locks that are longer than a certain duration |
| ByTime | 1 | Queries for lockups that started before a specific time |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/incentives/gauge.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/incentives/gauge.proto



<a name="osmosis.incentives.Gauge"></a>

### Gauge



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | unique ID of a Gauge |
| `is_perpetual` | [bool](#bool) |  | flag to show if it's perpetual or multi-epoch distribution incentives by third party |
| `distribute_to` | [osmosis.lockup.QueryCondition](#osmosis.lockup.QueryCondition) |  | Rewards are distributed to lockups that are are returned by at least one of these queries |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | total amount of Coins that has been in the gauge. can distribute multiple coins |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | distribution start time |
| `num_epochs_paid_over` | [uint64](#uint64) |  | number of epochs distribution will be done |
| `filled_epochs` | [uint64](#uint64) |  | number of epochs distributed already |
| `distributed_coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | already distributed coins |






<a name="osmosis.incentives.LockableDurationsInfo"></a>

### LockableDurationsInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lockable_durations` | [google.protobuf.Duration](#google.protobuf.Duration) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/incentives/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/incentives/params.proto



<a name="osmosis.incentives.Params"></a>

### Params
Params holds parameters for the incentives module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `distr_epoch_identifier` | [string](#string) |  | distribution epoch identifier |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/incentives/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/incentives/genesis.proto



<a name="osmosis.incentives.GenesisState"></a>

### GenesisState
GenesisState defines the incentives module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.incentives.Params) |  | params defines all the parameters of the module |
| `gauges` | [Gauge](#osmosis.incentives.Gauge) | repeated |  |
| `lockable_durations` | [google.protobuf.Duration](#google.protobuf.Duration) | repeated |  |
| `last_gauge_id` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/incentives/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/incentives/query.proto



<a name="osmosis.incentives.ActiveGaugesPerDenomRequest"></a>

### ActiveGaugesPerDenomRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an pagination for the request. |






<a name="osmosis.incentives.ActiveGaugesPerDenomResponse"></a>

### ActiveGaugesPerDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [Gauge](#osmosis.incentives.Gauge) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an pagination for the response. |






<a name="osmosis.incentives.ActiveGaugesRequest"></a>

### ActiveGaugesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an pagination for the request. |






<a name="osmosis.incentives.ActiveGaugesResponse"></a>

### ActiveGaugesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [Gauge](#osmosis.incentives.Gauge) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an pagination for the response. |






<a name="osmosis.incentives.GaugeByIDRequest"></a>

### GaugeByIDRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |






<a name="osmosis.incentives.GaugeByIDResponse"></a>

### GaugeByIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gauge` | [Gauge](#osmosis.incentives.Gauge) |  |  |






<a name="osmosis.incentives.GaugesRequest"></a>

### GaugesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an pagination for the request. |






<a name="osmosis.incentives.GaugesResponse"></a>

### GaugesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [Gauge](#osmosis.incentives.Gauge) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an pagination for the response. |






<a name="osmosis.incentives.ModuleDistributedCoinsRequest"></a>

### ModuleDistributedCoinsRequest







<a name="osmosis.incentives.ModuleDistributedCoinsResponse"></a>

### ModuleDistributedCoinsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.incentives.ModuleToDistributeCoinsRequest"></a>

### ModuleToDistributeCoinsRequest







<a name="osmosis.incentives.ModuleToDistributeCoinsResponse"></a>

### ModuleToDistributeCoinsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.incentives.QueryLockableDurationsRequest"></a>

### QueryLockableDurationsRequest







<a name="osmosis.incentives.QueryLockableDurationsResponse"></a>

### QueryLockableDurationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lockable_durations` | [google.protobuf.Duration](#google.protobuf.Duration) | repeated |  |






<a name="osmosis.incentives.RewardsEstRequest"></a>

### RewardsEstRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `lock_ids` | [uint64](#uint64) | repeated |  |
| `end_epoch` | [int64](#int64) |  |  |






<a name="osmosis.incentives.RewardsEstResponse"></a>

### RewardsEstResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.incentives.UpcomingGaugesPerDenomRequest"></a>

### UpcomingGaugesPerDenomRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="osmosis.incentives.UpcomingGaugesPerDenomResponse"></a>

### UpcomingGaugesPerDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `upcoming_gauges` | [Gauge](#osmosis.incentives.Gauge) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="osmosis.incentives.UpcomingGaugesRequest"></a>

### UpcomingGaugesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an pagination for the request. |






<a name="osmosis.incentives.UpcomingGaugesResponse"></a>

### UpcomingGaugesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [Gauge](#osmosis.incentives.Gauge) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an pagination for the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.incentives.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ModuleToDistributeCoins` | [ModuleToDistributeCoinsRequest](#osmosis.incentives.ModuleToDistributeCoinsRequest) | [ModuleToDistributeCoinsResponse](#osmosis.incentives.ModuleToDistributeCoinsResponse) | returns coins that is going to be distributed | GET|/osmosis/incentives/v1beta1/module_to_distribute_coins|
| `ModuleDistributedCoins` | [ModuleDistributedCoinsRequest](#osmosis.incentives.ModuleDistributedCoinsRequest) | [ModuleDistributedCoinsResponse](#osmosis.incentives.ModuleDistributedCoinsResponse) | returns coins that are distributed by module so far | GET|/osmosis/incentives/v1beta1/module_distributed_coins|
| `GaugeByID` | [GaugeByIDRequest](#osmosis.incentives.GaugeByIDRequest) | [GaugeByIDResponse](#osmosis.incentives.GaugeByIDResponse) | returns Gauge by id | GET|/osmosis/incentives/v1beta1/gauge_by_id/{id}|
| `Gauges` | [GaugesRequest](#osmosis.incentives.GaugesRequest) | [GaugesResponse](#osmosis.incentives.GaugesResponse) | returns gauges both upcoming and active | GET|/osmosis/incentives/v1beta1/gauges|
| `ActiveGauges` | [ActiveGaugesRequest](#osmosis.incentives.ActiveGaugesRequest) | [ActiveGaugesResponse](#osmosis.incentives.ActiveGaugesResponse) | returns active gauges | GET|/osmosis/incentives/v1beta1/active_gauges|
| `ActiveGaugesPerDenom` | [ActiveGaugesPerDenomRequest](#osmosis.incentives.ActiveGaugesPerDenomRequest) | [ActiveGaugesPerDenomResponse](#osmosis.incentives.ActiveGaugesPerDenomResponse) | returns active gauges per denom | GET|/osmosis/incentives/v1beta1/active_gauges_per_denom|
| `UpcomingGauges` | [UpcomingGaugesRequest](#osmosis.incentives.UpcomingGaugesRequest) | [UpcomingGaugesResponse](#osmosis.incentives.UpcomingGaugesResponse) | returns scheduled gauges | GET|/osmosis/incentives/v1beta1/upcoming_gauges|
| `UpcomingGaugesPerDenom` | [UpcomingGaugesPerDenomRequest](#osmosis.incentives.UpcomingGaugesPerDenomRequest) | [UpcomingGaugesPerDenomResponse](#osmosis.incentives.UpcomingGaugesPerDenomResponse) | returns scheduled gauges per denom | GET|/osmosis/incentives/v1beta1/upcoming_gauges_per_denom|
| `RewardsEst` | [RewardsEstRequest](#osmosis.incentives.RewardsEstRequest) | [RewardsEstResponse](#osmosis.incentives.RewardsEstResponse) | RewardsEst returns an estimate of the rewards at a future specific time. The querier either provides an address or a set of locks for which they want to find the associated rewards. | GET|/osmosis/incentives/v1beta1/rewards_est/{owner}|
| `LockableDurations` | [QueryLockableDurationsRequest](#osmosis.incentives.QueryLockableDurationsRequest) | [QueryLockableDurationsResponse](#osmosis.incentives.QueryLockableDurationsResponse) | returns lockable durations that are valid to give incentives | GET|/osmosis/incentives/v1beta1/lockable_durations|

 <!-- end services -->



<a name="osmosis/incentives/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/incentives/tx.proto



<a name="osmosis.incentives.MsgAddToGauge"></a>

### MsgAddToGauge



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `gauge_id` | [uint64](#uint64) |  |  |
| `rewards` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.incentives.MsgAddToGaugeResponse"></a>

### MsgAddToGaugeResponse







<a name="osmosis.incentives.MsgCreateGauge"></a>

### MsgCreateGauge



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `is_perpetual` | [bool](#bool) |  | flag to show if it's perpetual or multi-epoch distribution incentives by third party |
| `owner` | [string](#string) |  |  |
| `distribute_to` | [osmosis.lockup.QueryCondition](#osmosis.lockup.QueryCondition) |  | distribute condition of a lock which meet one of these conditions |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | can distribute multiple coins |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | distribution start time |
| `num_epochs_paid_over` | [uint64](#uint64) |  | number of epochs distribution will be done |






<a name="osmosis.incentives.MsgCreateGaugeResponse"></a>

### MsgCreateGaugeResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.incentives.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateGauge` | [MsgCreateGauge](#osmosis.incentives.MsgCreateGauge) | [MsgCreateGaugeResponse](#osmosis.incentives.MsgCreateGaugeResponse) |  | |
| `AddToGauge` | [MsgAddToGauge](#osmosis.incentives.MsgAddToGauge) | [MsgAddToGaugeResponse](#osmosis.incentives.MsgAddToGaugeResponse) |  | |

 <!-- end services -->



<a name="osmosis/lockup/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/lockup/genesis.proto



<a name="osmosis.lockup.GenesisState"></a>

### GenesisState
GenesisState defines the lockup module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `last_lock_id` | [uint64](#uint64) |  |  |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |
| `synthetic_locks` | [SyntheticLock](#osmosis.lockup.SyntheticLock) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/lockup/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/lockup/query.proto



<a name="osmosis.lockup.AccountLockedCoinsRequest"></a>

### AccountLockedCoinsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |






<a name="osmosis.lockup.AccountLockedCoinsResponse"></a>

### AccountLockedCoinsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.lockup.AccountLockedDurationRequest"></a>

### AccountLockedDurationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="osmosis.lockup.AccountLockedDurationResponse"></a>

### AccountLockedDurationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountLockedLongerDurationDenomRequest"></a>

### AccountLockedLongerDurationDenomRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `denom` | [string](#string) |  |  |






<a name="osmosis.lockup.AccountLockedLongerDurationDenomResponse"></a>

### AccountLockedLongerDurationDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountLockedLongerDurationNotUnlockingOnlyRequest"></a>

### AccountLockedLongerDurationNotUnlockingOnlyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="osmosis.lockup.AccountLockedLongerDurationNotUnlockingOnlyResponse"></a>

### AccountLockedLongerDurationNotUnlockingOnlyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountLockedLongerDurationRequest"></a>

### AccountLockedLongerDurationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="osmosis.lockup.AccountLockedLongerDurationResponse"></a>

### AccountLockedLongerDurationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountLockedPastTimeDenomRequest"></a>

### AccountLockedPastTimeDenomRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `denom` | [string](#string) |  |  |






<a name="osmosis.lockup.AccountLockedPastTimeDenomResponse"></a>

### AccountLockedPastTimeDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountLockedPastTimeNotUnlockingOnlyRequest"></a>

### AccountLockedPastTimeNotUnlockingOnlyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="osmosis.lockup.AccountLockedPastTimeNotUnlockingOnlyResponse"></a>

### AccountLockedPastTimeNotUnlockingOnlyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountLockedPastTimeRequest"></a>

### AccountLockedPastTimeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="osmosis.lockup.AccountLockedPastTimeResponse"></a>

### AccountLockedPastTimeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountUnlockableCoinsRequest"></a>

### AccountUnlockableCoinsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |






<a name="osmosis.lockup.AccountUnlockableCoinsResponse"></a>

### AccountUnlockableCoinsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.lockup.AccountUnlockedBeforeTimeRequest"></a>

### AccountUnlockedBeforeTimeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `timestamp` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="osmosis.lockup.AccountUnlockedBeforeTimeResponse"></a>

### AccountUnlockedBeforeTimeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.AccountUnlockingCoinsRequest"></a>

### AccountUnlockingCoinsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |






<a name="osmosis.lockup.AccountUnlockingCoinsResponse"></a>

### AccountUnlockingCoinsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.lockup.LockedDenomRequest"></a>

### LockedDenomRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="osmosis.lockup.LockedDenomResponse"></a>

### LockedDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |






<a name="osmosis.lockup.LockedRequest"></a>

### LockedRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lock_id` | [uint64](#uint64) |  |  |






<a name="osmosis.lockup.LockedResponse"></a>

### LockedResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lock` | [PeriodLock](#osmosis.lockup.PeriodLock) |  |  |






<a name="osmosis.lockup.ModuleBalanceRequest"></a>

### ModuleBalanceRequest







<a name="osmosis.lockup.ModuleBalanceResponse"></a>

### ModuleBalanceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.lockup.ModuleLockedAmountRequest"></a>

### ModuleLockedAmountRequest







<a name="osmosis.lockup.ModuleLockedAmountResponse"></a>

### ModuleLockedAmountResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.lockup.SyntheticLockupsByLockupIDRequest"></a>

### SyntheticLockupsByLockupIDRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lock_id` | [uint64](#uint64) |  |  |






<a name="osmosis.lockup.SyntheticLockupsByLockupIDResponse"></a>

### SyntheticLockupsByLockupIDResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `synthetic_locks` | [SyntheticLock](#osmosis.lockup.SyntheticLock) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.lockup.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ModuleBalance` | [ModuleBalanceRequest](#osmosis.lockup.ModuleBalanceRequest) | [ModuleBalanceResponse](#osmosis.lockup.ModuleBalanceResponse) | Return full balance of the module | GET|/osmosis/lockup/v1beta1/module_balance|
| `ModuleLockedAmount` | [ModuleLockedAmountRequest](#osmosis.lockup.ModuleLockedAmountRequest) | [ModuleLockedAmountResponse](#osmosis.lockup.ModuleLockedAmountResponse) | Return locked balance of the module | GET|/osmosis/lockup/v1beta1/module_locked_amount|
| `AccountUnlockableCoins` | [AccountUnlockableCoinsRequest](#osmosis.lockup.AccountUnlockableCoinsRequest) | [AccountUnlockableCoinsResponse](#osmosis.lockup.AccountUnlockableCoinsResponse) | Returns unlockable coins which are not withdrawn yet | GET|/osmosis/lockup/v1beta1/account_unlockable_coins/{owner}|
| `AccountUnlockingCoins` | [AccountUnlockingCoinsRequest](#osmosis.lockup.AccountUnlockingCoinsRequest) | [AccountUnlockingCoinsResponse](#osmosis.lockup.AccountUnlockingCoinsResponse) | Returns unlocking coins | GET|/osmosis/lockup/v1beta1/account_unlocking_coins/{owner}|
| `AccountLockedCoins` | [AccountLockedCoinsRequest](#osmosis.lockup.AccountLockedCoinsRequest) | [AccountLockedCoinsResponse](#osmosis.lockup.AccountLockedCoinsResponse) | Return a locked coins that can't be withdrawn | GET|/osmosis/lockup/v1beta1/account_locked_coins/{owner}|
| `AccountLockedPastTime` | [AccountLockedPastTimeRequest](#osmosis.lockup.AccountLockedPastTimeRequest) | [AccountLockedPastTimeResponse](#osmosis.lockup.AccountLockedPastTimeResponse) | Returns locked records of an account with unlock time beyond timestamp | GET|/osmosis/lockup/v1beta1/account_locked_pasttime/{owner}|
| `AccountLockedPastTimeNotUnlockingOnly` | [AccountLockedPastTimeNotUnlockingOnlyRequest](#osmosis.lockup.AccountLockedPastTimeNotUnlockingOnlyRequest) | [AccountLockedPastTimeNotUnlockingOnlyResponse](#osmosis.lockup.AccountLockedPastTimeNotUnlockingOnlyResponse) | Returns locked records of an account with unlock time beyond timestamp excluding tokens started unlocking | GET|/osmosis/lockup/v1beta1/account_locked_pasttime_not_unlocking_only/{owner}|
| `AccountUnlockedBeforeTime` | [AccountUnlockedBeforeTimeRequest](#osmosis.lockup.AccountUnlockedBeforeTimeRequest) | [AccountUnlockedBeforeTimeResponse](#osmosis.lockup.AccountUnlockedBeforeTimeResponse) | Returns unlocked records with unlock time before timestamp | GET|/osmosis/lockup/v1beta1/account_unlocked_before_time/{owner}|
| `AccountLockedPastTimeDenom` | [AccountLockedPastTimeDenomRequest](#osmosis.lockup.AccountLockedPastTimeDenomRequest) | [AccountLockedPastTimeDenomResponse](#osmosis.lockup.AccountLockedPastTimeDenomResponse) | Returns lock records by address, timestamp, denom | GET|/osmosis/lockup/v1beta1/account_locked_pasttime_denom/{owner}|
| `LockedDenom` | [LockedDenomRequest](#osmosis.lockup.LockedDenomRequest) | [LockedDenomResponse](#osmosis.lockup.LockedDenomResponse) | Returns total locked per denom with longer past given time | GET|/osmosis/lockup/v1beta1/locked_denom|
| `LockedByID` | [LockedRequest](#osmosis.lockup.LockedRequest) | [LockedResponse](#osmosis.lockup.LockedResponse) | Returns lock record by id | GET|/osmosis/lockup/v1beta1/locked_by_id/{lock_id}|
| `SyntheticLockupsByLockupID` | [SyntheticLockupsByLockupIDRequest](#osmosis.lockup.SyntheticLockupsByLockupIDRequest) | [SyntheticLockupsByLockupIDResponse](#osmosis.lockup.SyntheticLockupsByLockupIDResponse) | Returns synthetic lockups by native lockup id | GET|/osmosis/lockup/v1beta1/synthetic_lockups_by_lock_id/{lock_id}|
| `AccountLockedLongerDuration` | [AccountLockedLongerDurationRequest](#osmosis.lockup.AccountLockedLongerDurationRequest) | [AccountLockedLongerDurationResponse](#osmosis.lockup.AccountLockedLongerDurationResponse) | Returns account locked records with longer duration | GET|/osmosis/lockup/v1beta1/account_locked_longer_duration/{owner}|
| `AccountLockedDuration` | [AccountLockedDurationRequest](#osmosis.lockup.AccountLockedDurationRequest) | [AccountLockedDurationResponse](#osmosis.lockup.AccountLockedDurationResponse) | Returns account locked records with a specific duration | GET|/osmosis/lockup/v1beta1/account_locked_duration/{owner}|
| `AccountLockedLongerDurationNotUnlockingOnly` | [AccountLockedLongerDurationNotUnlockingOnlyRequest](#osmosis.lockup.AccountLockedLongerDurationNotUnlockingOnlyRequest) | [AccountLockedLongerDurationNotUnlockingOnlyResponse](#osmosis.lockup.AccountLockedLongerDurationNotUnlockingOnlyResponse) | Returns account locked records with longer duration excluding tokens started unlocking | GET|/osmosis/lockup/v1beta1/account_locked_longer_duration_not_unlocking_only/{owner}|
| `AccountLockedLongerDurationDenom` | [AccountLockedLongerDurationDenomRequest](#osmosis.lockup.AccountLockedLongerDurationDenomRequest) | [AccountLockedLongerDurationDenomResponse](#osmosis.lockup.AccountLockedLongerDurationDenomResponse) | Returns account's locked records for a denom with longer duration | GET|/osmosis/lockup/v1beta1/account_locked_longer_duration_denom/{owner}|

 <!-- end services -->



<a name="osmosis/lockup/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/lockup/tx.proto



<a name="osmosis.lockup.MsgBeginUnlocking"></a>

### MsgBeginUnlocking



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `ID` | [uint64](#uint64) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Amount of unlocking coins. Unlock all if not set. |






<a name="osmosis.lockup.MsgBeginUnlockingAll"></a>

### MsgBeginUnlockingAll



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |






<a name="osmosis.lockup.MsgBeginUnlockingAllResponse"></a>

### MsgBeginUnlockingAllResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `unlocks` | [PeriodLock](#osmosis.lockup.PeriodLock) | repeated |  |






<a name="osmosis.lockup.MsgBeginUnlockingResponse"></a>

### MsgBeginUnlockingResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `success` | [bool](#bool) |  |  |






<a name="osmosis.lockup.MsgExtendLockup"></a>

### MsgExtendLockup
MsgExtendLockup extends the existing lockup's duration.
The new duration is longer than the original.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `ID` | [uint64](#uint64) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  | duration to be set. fails if lower than the current duration, or is unlocking |






<a name="osmosis.lockup.MsgExtendLockupResponse"></a>

### MsgExtendLockupResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `success` | [bool](#bool) |  |  |






<a name="osmosis.lockup.MsgLockTokens"></a>

### MsgLockTokens



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.lockup.MsgLockTokensResponse"></a>

### MsgLockTokensResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ID` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.lockup.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `LockTokens` | [MsgLockTokens](#osmosis.lockup.MsgLockTokens) | [MsgLockTokensResponse](#osmosis.lockup.MsgLockTokensResponse) | LockTokens lock tokens | |
| `BeginUnlockingAll` | [MsgBeginUnlockingAll](#osmosis.lockup.MsgBeginUnlockingAll) | [MsgBeginUnlockingAllResponse](#osmosis.lockup.MsgBeginUnlockingAllResponse) | BeginUnlockingAll begin unlocking all tokens | |
| `BeginUnlocking` | [MsgBeginUnlocking](#osmosis.lockup.MsgBeginUnlocking) | [MsgBeginUnlockingResponse](#osmosis.lockup.MsgBeginUnlockingResponse) | MsgBeginUnlocking begins unlocking tokens by lock ID | |
| `ExtendLockup` | [MsgExtendLockup](#osmosis.lockup.MsgExtendLockup) | [MsgExtendLockupResponse](#osmosis.lockup.MsgExtendLockupResponse) | MsgEditLockup edits the existing lockups by lock ID | |

 <!-- end services -->



<a name="osmosis/mint/v1beta1/mint.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/mint/v1beta1/mint.proto



<a name="osmosis.mint.v1beta1.DistributionProportions"></a>

### DistributionProportions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `staking` | [string](#string) |  | staking defines the proportion of the minted minted_denom that is to be allocated as staking rewards. |
| `pool_incentives` | [string](#string) |  | pool_incentives defines the proportion of the minted minted_denom that is to be allocated as pool incentives. |
| `developer_rewards` | [string](#string) |  | developer_rewards defines the proportion of the minted minted_denom that is to be allocated to developer rewards address. |
| `community_pool` | [string](#string) |  | community_pool defines the proportion of the minted minted_denom that is to be allocated to the community pool. |






<a name="osmosis.mint.v1beta1.Minter"></a>

### Minter
Minter represents the minting state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epoch_provisions` | [string](#string) |  | current epoch provisions |






<a name="osmosis.mint.v1beta1.Params"></a>

### Params
Params holds parameters for the mint module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mint_denom` | [string](#string) |  | type of coin to mint |
| `genesis_epoch_provisions` | [string](#string) |  | epoch provisions from the first epoch |
| `epoch_identifier` | [string](#string) |  | mint epoch identifier |
| `reduction_period_in_epochs` | [int64](#int64) |  | number of epochs take to reduce rewards |
| `reduction_factor` | [string](#string) |  | reduction multiplier to execute on each period |
| `distribution_proportions` | [DistributionProportions](#osmosis.mint.v1beta1.DistributionProportions) |  | distribution_proportions defines the proportion of the minted denom |
| `weighted_developer_rewards_receivers` | [WeightedAddress](#osmosis.mint.v1beta1.WeightedAddress) | repeated | address to receive developer rewards |
| `minting_rewards_distribution_start_epoch` | [int64](#int64) |  | start epoch to distribute minting rewards |






<a name="osmosis.mint.v1beta1.WeightedAddress"></a>

### WeightedAddress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `weight` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/mint/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/mint/v1beta1/genesis.proto



<a name="osmosis.mint.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the mint module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minter` | [Minter](#osmosis.mint.v1beta1.Minter) |  | minter is a space for holding current rewards information. |
| `params` | [Params](#osmosis.mint.v1beta1.Params) |  | params defines all the paramaters of the module. |
| `halven_started_epoch` | [int64](#int64) |  | current halven period start epoch |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/mint/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/mint/v1beta1/query.proto



<a name="osmosis.mint.v1beta1.QueryEpochProvisionsRequest"></a>

### QueryEpochProvisionsRequest
QueryEpochProvisionsRequest is the request type for the
Query/EpochProvisions RPC method.






<a name="osmosis.mint.v1beta1.QueryEpochProvisionsResponse"></a>

### QueryEpochProvisionsResponse
QueryEpochProvisionsResponse is the response type for the
Query/EpochProvisions RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epoch_provisions` | [bytes](#bytes) |  | epoch_provisions is the current minting per epoch provisions value. |






<a name="osmosis.mint.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="osmosis.mint.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.mint.v1beta1.Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.mint.v1beta1.Query"></a>

### Query
Query provides defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#osmosis.mint.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#osmosis.mint.v1beta1.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/osmosis/mint/v1beta1/params|
| `EpochProvisions` | [QueryEpochProvisionsRequest](#osmosis.mint.v1beta1.QueryEpochProvisionsRequest) | [QueryEpochProvisionsResponse](#osmosis.mint.v1beta1.QueryEpochProvisionsResponse) | EpochProvisions current minting epoch provisions value. | GET|/osmosis/mint/v1beta1/epoch_provisions|

 <!-- end services -->



<a name="osmosis/pool-incentives/v1beta1/incentives.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/pool-incentives/v1beta1/incentives.proto



<a name="osmosis.poolincentives.v1beta1.DistrInfo"></a>

### DistrInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `total_weight` | [string](#string) |  |  |
| `records` | [DistrRecord](#osmosis.poolincentives.v1beta1.DistrRecord) | repeated |  |






<a name="osmosis.poolincentives.v1beta1.DistrRecord"></a>

### DistrRecord



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gauge_id` | [uint64](#uint64) |  |  |
| `weight` | [string](#string) |  |  |






<a name="osmosis.poolincentives.v1beta1.LockableDurationsInfo"></a>

### LockableDurationsInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lockable_durations` | [google.protobuf.Duration](#google.protobuf.Duration) | repeated |  |






<a name="osmosis.poolincentives.v1beta1.Params"></a>

### Params



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minted_denom` | [string](#string) |  | minted_denom is the denomination of the coin expected to be minted by the minting module. Pool-incentives module doesn’t actually mint the coin itself, but rather manages the distribution of coins that matches the defined minted_denom. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/pool-incentives/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/pool-incentives/v1beta1/genesis.proto



<a name="osmosis.poolincentives.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the pool incentives module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.poolincentives.v1beta1.Params) |  | params defines all the paramaters of the module. |
| `lockable_durations` | [google.protobuf.Duration](#google.protobuf.Duration) | repeated |  |
| `distr_info` | [DistrInfo](#osmosis.poolincentives.v1beta1.DistrInfo) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/pool-incentives/v1beta1/gov.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/pool-incentives/v1beta1/gov.proto



<a name="osmosis.poolincentives.v1beta1.ReplacePoolIncentivesProposal"></a>

### ReplacePoolIncentivesProposal
ReplacePoolIncentivesProposal is a gov Content type for updating the pool
incentives. If a ReplacePoolIncentivesProposal passes, the proposal’s records
override the existing DistrRecords set in the module. Each record has a
specified gauge id and weight, and the incentives are distributed to each
gauge according to weight/total_weight. The incentives are put in the fee
pool and it is allocated to gauges and community pool by the DistrRecords
configuration. Note that gaugeId=0 represents the community pool.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `records` | [DistrRecord](#osmosis.poolincentives.v1beta1.DistrRecord) | repeated |  |






<a name="osmosis.poolincentives.v1beta1.UpdatePoolIncentivesProposal"></a>

### UpdatePoolIncentivesProposal
For example: if the existing DistrRecords were:
[(Gauge 0, 5), (Gauge 1, 6), (Gauge 2, 6)]
An UpdatePoolIncentivesProposal includes
[(Gauge 1, 0), (Gauge 2, 4), (Gauge 3, 10)]
This would delete Gauge 1, Edit Gauge 2, and Add Gauge 3
The result DistrRecords in state would be:
[(Gauge 0, 5), (Gauge 2, 4), (Gauge 3, 10)]


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `records` | [DistrRecord](#osmosis.poolincentives.v1beta1.DistrRecord) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/pool-incentives/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/pool-incentives/v1beta1/query.proto



<a name="osmosis.poolincentives.v1beta1.IncentivizedPool"></a>

### IncentivizedPool



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool_id` | [uint64](#uint64) |  |  |
| `lockable_duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `gauge_id` | [uint64](#uint64) |  |  |






<a name="osmosis.poolincentives.v1beta1.QueryDistrInfoRequest"></a>

### QueryDistrInfoRequest







<a name="osmosis.poolincentives.v1beta1.QueryDistrInfoResponse"></a>

### QueryDistrInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `distr_info` | [DistrInfo](#osmosis.poolincentives.v1beta1.DistrInfo) |  |  |






<a name="osmosis.poolincentives.v1beta1.QueryExternalIncentiveGaugesRequest"></a>

### QueryExternalIncentiveGaugesRequest







<a name="osmosis.poolincentives.v1beta1.QueryExternalIncentiveGaugesResponse"></a>

### QueryExternalIncentiveGaugesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [osmosis.incentives.Gauge](#osmosis.incentives.Gauge) | repeated |  |






<a name="osmosis.poolincentives.v1beta1.QueryGaugeIdsRequest"></a>

### QueryGaugeIdsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool_id` | [uint64](#uint64) |  |  |






<a name="osmosis.poolincentives.v1beta1.QueryGaugeIdsResponse"></a>

### QueryGaugeIdsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gauge_ids_with_duration` | [QueryGaugeIdsResponse.GaugeIdWithDuration](#osmosis.poolincentives.v1beta1.QueryGaugeIdsResponse.GaugeIdWithDuration) | repeated |  |






<a name="osmosis.poolincentives.v1beta1.QueryGaugeIdsResponse.GaugeIdWithDuration"></a>

### QueryGaugeIdsResponse.GaugeIdWithDuration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gauge_id` | [uint64](#uint64) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="osmosis.poolincentives.v1beta1.QueryIncentivizedPoolsRequest"></a>

### QueryIncentivizedPoolsRequest







<a name="osmosis.poolincentives.v1beta1.QueryIncentivizedPoolsResponse"></a>

### QueryIncentivizedPoolsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `incentivized_pools` | [IncentivizedPool](#osmosis.poolincentives.v1beta1.IncentivizedPool) | repeated |  |






<a name="osmosis.poolincentives.v1beta1.QueryLockableDurationsRequest"></a>

### QueryLockableDurationsRequest







<a name="osmosis.poolincentives.v1beta1.QueryLockableDurationsResponse"></a>

### QueryLockableDurationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lockable_durations` | [google.protobuf.Duration](#google.protobuf.Duration) | repeated |  |






<a name="osmosis.poolincentives.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest







<a name="osmosis.poolincentives.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.poolincentives.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.poolincentives.v1beta1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GaugeIds` | [QueryGaugeIdsRequest](#osmosis.poolincentives.v1beta1.QueryGaugeIdsRequest) | [QueryGaugeIdsResponse](#osmosis.poolincentives.v1beta1.QueryGaugeIdsResponse) | GaugeIds takes the pool id and returns the matching gauge ids and durations | GET|/osmosis/pool-incentives/v1beta1/gauge-ids/{pool_id}|
| `DistrInfo` | [QueryDistrInfoRequest](#osmosis.poolincentives.v1beta1.QueryDistrInfoRequest) | [QueryDistrInfoResponse](#osmosis.poolincentives.v1beta1.QueryDistrInfoResponse) |  | GET|/osmosis/pool-incentives/v1beta1/distr_info|
| `Params` | [QueryParamsRequest](#osmosis.poolincentives.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#osmosis.poolincentives.v1beta1.QueryParamsResponse) |  | GET|/osmosis/pool-incentives/v1beta1/params|
| `LockableDurations` | [QueryLockableDurationsRequest](#osmosis.poolincentives.v1beta1.QueryLockableDurationsRequest) | [QueryLockableDurationsResponse](#osmosis.poolincentives.v1beta1.QueryLockableDurationsResponse) |  | GET|/osmosis/pool-incentives/v1beta1/lockable_durations|
| `IncentivizedPools` | [QueryIncentivizedPoolsRequest](#osmosis.poolincentives.v1beta1.QueryIncentivizedPoolsRequest) | [QueryIncentivizedPoolsResponse](#osmosis.poolincentives.v1beta1.QueryIncentivizedPoolsResponse) |  | GET|/osmosis/pool-incentives/v1beta1/incentivized_pools|
| `ExternalIncentiveGauges` | [QueryExternalIncentiveGaugesRequest](#osmosis.poolincentives.v1beta1.QueryExternalIncentiveGaugesRequest) | [QueryExternalIncentiveGaugesResponse](#osmosis.poolincentives.v1beta1.QueryExternalIncentiveGaugesResponse) |  | GET|/osmosis/pool-incentives/v1beta1/external_incentive_gauges|

 <!-- end services -->



<a name="osmosis/store/v1beta1/tree.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/store/v1beta1/tree.proto



<a name="osmosis.store.v1beta1.Child"></a>

### Child



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `index` | [bytes](#bytes) |  |  |
| `accumulation` | [string](#string) |  |  |






<a name="osmosis.store.v1beta1.Leaf"></a>

### Leaf



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `leaf` | [Child](#osmosis.store.v1beta1.Child) |  |  |






<a name="osmosis.store.v1beta1.Node"></a>

### Node



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `children` | [Child](#osmosis.store.v1beta1.Child) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/superfluid/superfluid.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/superfluid/superfluid.proto



<a name="osmosis.superfluid.LockIdIntermediaryAccountConnection"></a>

### LockIdIntermediaryAccountConnection



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lock_id` | [uint64](#uint64) |  |  |
| `intermediary_account` | [string](#string) |  |  |






<a name="osmosis.superfluid.OsmoEquivalentMultiplierRecord"></a>

### OsmoEquivalentMultiplierRecord
The Osmo-Equivalent-Multiplier Record for epoch N refers to the osmo worth we
treat an LP share as having, for all of epoch N. Eventually this is intended
to be set as the Time-weighted-average-osmo-backing for the entire duration
of epoch N-1. (Thereby locking whats in use for epoch N as based on the prior
epochs rewards) However for now, this is not the TWAP but instead the spot
price at the boundary.  For different types of assets in the future, it could
change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epoch_number` | [int64](#int64) |  |  |
| `denom` | [string](#string) |  | superfluid asset denom, can be LP token or native token |
| `multiplier` | [string](#string) |  |  |






<a name="osmosis.superfluid.SuperfluidAsset"></a>

### SuperfluidAsset
SuperfluidAsset stores the pair of superfluid asset type and denom pair


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `asset_type` | [SuperfluidAssetType](#osmosis.superfluid.SuperfluidAssetType) |  |  |






<a name="osmosis.superfluid.SuperfluidDelegationRecord"></a>

### SuperfluidDelegationRecord
SuperfluidDelegationRecord takes the role of intermediary between LP token
and OSMO tokens for superfluid staking


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegator_address` | [string](#string) |  |  |
| `validator_address` | [string](#string) |  |  |
| `delegation_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `equivalent_staked_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="osmosis.superfluid.SuperfluidIntermediaryAccount"></a>

### SuperfluidIntermediaryAccount
SuperfluidIntermediaryAccount takes the role of intermediary between LP token
and OSMO tokens for superfluid staking


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `val_addr` | [string](#string) |  |  |
| `gauge_id` | [uint64](#uint64) |  | perpetual gauge for rewards distribution |






<a name="osmosis.superfluid.UnpoolWhitelistedPools"></a>

### UnpoolWhitelistedPools



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ids` | [uint64](#uint64) | repeated |  |





 <!-- end messages -->


<a name="osmosis.superfluid.SuperfluidAssetType"></a>

### SuperfluidAssetType


| Name | Number | Description |
| ---- | ------ | ----------- |
| SuperfluidAssetTypeNative | 0 |  |
| SuperfluidAssetTypeLPShare | 1 | SuperfluidAssetTypeLendingShare = 2; // for now not exist |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/superfluid/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/superfluid/params.proto



<a name="osmosis.superfluid.Params"></a>

### Params
Params holds parameters for the superfluid module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_risk_factor` | [string](#string) |  | the risk_factor is to be cut on OSMO equivalent value of lp tokens for superfluid staking, default: 5% |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/superfluid/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/superfluid/genesis.proto



<a name="osmosis.superfluid.GenesisState"></a>

### GenesisState
GenesisState defines the module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.superfluid.Params) |  |  |
| `superfluid_assets` | [SuperfluidAsset](#osmosis.superfluid.SuperfluidAsset) | repeated |  |
| `osmo_equivalent_multipliers` | [OsmoEquivalentMultiplierRecord](#osmosis.superfluid.OsmoEquivalentMultiplierRecord) | repeated |  |
| `intermediary_accounts` | [SuperfluidIntermediaryAccount](#osmosis.superfluid.SuperfluidIntermediaryAccount) | repeated |  |
| `intemediary_account_connections` | [LockIdIntermediaryAccountConnection](#osmosis.superfluid.LockIdIntermediaryAccountConnection) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/superfluid/gov.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/superfluid/gov.proto



<a name="osmosis.superfluid.v1beta1.RemoveSuperfluidAssetsProposal"></a>

### RemoveSuperfluidAssetsProposal
RemoveSuperfluidAssetsProposal is a gov Content type to remove the superfluid
assets by denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `superfluid_asset_denoms` | [string](#string) | repeated |  |






<a name="osmosis.superfluid.v1beta1.SetSuperfluidAssetsProposal"></a>

### SetSuperfluidAssetsProposal
SetSuperfluidAssetsProposal is a gov Content type to update the superfluid
assets


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `assets` | [osmosis.superfluid.SuperfluidAsset](#osmosis.superfluid.SuperfluidAsset) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/superfluid/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/superfluid/query.proto



<a name="osmosis.superfluid.AllAssetsRequest"></a>

### AllAssetsRequest







<a name="osmosis.superfluid.AllAssetsResponse"></a>

### AllAssetsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [SuperfluidAsset](#osmosis.superfluid.SuperfluidAsset) | repeated |  |






<a name="osmosis.superfluid.AllIntermediaryAccountsRequest"></a>

### AllIntermediaryAccountsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="osmosis.superfluid.AllIntermediaryAccountsResponse"></a>

### AllIntermediaryAccountsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [SuperfluidIntermediaryAccountInfo](#osmosis.superfluid.SuperfluidIntermediaryAccountInfo) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="osmosis.superfluid.AssetMultiplierRequest"></a>

### AssetMultiplierRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="osmosis.superfluid.AssetMultiplierResponse"></a>

### AssetMultiplierResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `osmo_equivalent_multiplier` | [OsmoEquivalentMultiplierRecord](#osmosis.superfluid.OsmoEquivalentMultiplierRecord) |  |  |






<a name="osmosis.superfluid.AssetTypeRequest"></a>

### AssetTypeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="osmosis.superfluid.AssetTypeResponse"></a>

### AssetTypeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_type` | [SuperfluidAssetType](#osmosis.superfluid.SuperfluidAssetType) |  |  |






<a name="osmosis.superfluid.ConnectedIntermediaryAccountRequest"></a>

### ConnectedIntermediaryAccountRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `lock_id` | [uint64](#uint64) |  |  |






<a name="osmosis.superfluid.ConnectedIntermediaryAccountResponse"></a>

### ConnectedIntermediaryAccountResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [SuperfluidIntermediaryAccountInfo](#osmosis.superfluid.SuperfluidIntermediaryAccountInfo) |  |  |






<a name="osmosis.superfluid.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest"></a>

### EstimateSuperfluidDelegatedAmountByValidatorDenomRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator_address` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |






<a name="osmosis.superfluid.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse"></a>

### EstimateSuperfluidDelegatedAmountByValidatorDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `total_delegated_coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.superfluid.QueryParamsRequest"></a>

### QueryParamsRequest







<a name="osmosis.superfluid.QueryParamsResponse"></a>

### QueryParamsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.superfluid.Params) |  | params defines the parameters of the module. |






<a name="osmosis.superfluid.SuperfluidDelegationAmountRequest"></a>

### SuperfluidDelegationAmountRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegator_address` | [string](#string) |  |  |
| `validator_address` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |






<a name="osmosis.superfluid.SuperfluidDelegationAmountResponse"></a>

### SuperfluidDelegationAmountResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="osmosis.superfluid.SuperfluidDelegationsByDelegatorRequest"></a>

### SuperfluidDelegationsByDelegatorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegator_address` | [string](#string) |  |  |






<a name="osmosis.superfluid.SuperfluidDelegationsByDelegatorResponse"></a>

### SuperfluidDelegationsByDelegatorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `superfluid_delegation_records` | [SuperfluidDelegationRecord](#osmosis.superfluid.SuperfluidDelegationRecord) | repeated |  |
| `total_delegated_coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `total_equivalent_staked_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="osmosis.superfluid.SuperfluidDelegationsByValidatorDenomRequest"></a>

### SuperfluidDelegationsByValidatorDenomRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator_address` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |






<a name="osmosis.superfluid.SuperfluidDelegationsByValidatorDenomResponse"></a>

### SuperfluidDelegationsByValidatorDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `superfluid_delegation_records` | [SuperfluidDelegationRecord](#osmosis.superfluid.SuperfluidDelegationRecord) | repeated |  |






<a name="osmosis.superfluid.SuperfluidIntermediaryAccountInfo"></a>

### SuperfluidIntermediaryAccountInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `val_addr` | [string](#string) |  |  |
| `gauge_id` | [uint64](#uint64) |  |  |
| `address` | [string](#string) |  |  |






<a name="osmosis.superfluid.SuperfluidUndelegationsByDelegatorRequest"></a>

### SuperfluidUndelegationsByDelegatorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegator_address` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |






<a name="osmosis.superfluid.SuperfluidUndelegationsByDelegatorResponse"></a>

### SuperfluidUndelegationsByDelegatorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `superfluid_delegation_records` | [SuperfluidDelegationRecord](#osmosis.superfluid.SuperfluidDelegationRecord) | repeated |  |
| `total_undelegated_coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `synthetic_locks` | [osmosis.lockup.SyntheticLock](#osmosis.lockup.SyntheticLock) | repeated |  |






<a name="osmosis.superfluid.TotalSuperfluidDelegationsRequest"></a>

### TotalSuperfluidDelegationsRequest







<a name="osmosis.superfluid.TotalSuperfluidDelegationsResponse"></a>

### TotalSuperfluidDelegationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `totalDelegations` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.superfluid.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#osmosis.superfluid.QueryParamsRequest) | [QueryParamsResponse](#osmosis.superfluid.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/osmosis/superfluid/v1beta1/params|
| `AssetType` | [AssetTypeRequest](#osmosis.superfluid.AssetTypeRequest) | [AssetTypeResponse](#osmosis.superfluid.AssetTypeResponse) | Returns superfluid asset type | GET|/osmosis/superfluid/v1beta1/asset_type|
| `AllAssets` | [AllAssetsRequest](#osmosis.superfluid.AllAssetsRequest) | [AllAssetsResponse](#osmosis.superfluid.AllAssetsResponse) | Returns all superfluid asset types | GET|/osmosis/superfluid/v1beta1/all_assets|
| `AssetMultiplier` | [AssetMultiplierRequest](#osmosis.superfluid.AssetMultiplierRequest) | [AssetMultiplierResponse](#osmosis.superfluid.AssetMultiplierResponse) | Returns superfluid asset Multiplier | GET|/osmosis/superfluid/v1beta1/asset_multiplier|
| `AllIntermediaryAccounts` | [AllIntermediaryAccountsRequest](#osmosis.superfluid.AllIntermediaryAccountsRequest) | [AllIntermediaryAccountsResponse](#osmosis.superfluid.AllIntermediaryAccountsResponse) | Returns all superfluid intermediary account | GET|/osmosis/superfluid/v1beta1/all_intermediary_accounts|
| `ConnectedIntermediaryAccount` | [ConnectedIntermediaryAccountRequest](#osmosis.superfluid.ConnectedIntermediaryAccountRequest) | [ConnectedIntermediaryAccountResponse](#osmosis.superfluid.ConnectedIntermediaryAccountResponse) | Returns intermediary account connected to a superfluid staked lock by id | GET|/osmosis/superfluid/v1beta1/connected_intermediary_account/{lock_id}|
| `TotalSuperfluidDelegations` | [TotalSuperfluidDelegationsRequest](#osmosis.superfluid.TotalSuperfluidDelegationsRequest) | [TotalSuperfluidDelegationsResponse](#osmosis.superfluid.TotalSuperfluidDelegationsResponse) | Returns the total amount of osmo superfluidly staked response denominated in uosmo | GET|/osmosis/superfluid/v1beta1/all_superfluid_delegations|
| `SuperfluidDelegationAmount` | [SuperfluidDelegationAmountRequest](#osmosis.superfluid.SuperfluidDelegationAmountRequest) | [SuperfluidDelegationAmountResponse](#osmosis.superfluid.SuperfluidDelegationAmountResponse) | Returns the coins superfluid delegated for a delegator, validator, denom triplet | GET|/osmosis/superfluid/v1beta1/superfluid_delegation_amount|
| `SuperfluidDelegationsByDelegator` | [SuperfluidDelegationsByDelegatorRequest](#osmosis.superfluid.SuperfluidDelegationsByDelegatorRequest) | [SuperfluidDelegationsByDelegatorResponse](#osmosis.superfluid.SuperfluidDelegationsByDelegatorResponse) | Returns all the superfluid poistions for a specific delegator | GET|/osmosis/superfluid/v1beta1/superfluid_delegations/{delegator_address}|
| `SuperfluidUndelegationsByDelegator` | [SuperfluidUndelegationsByDelegatorRequest](#osmosis.superfluid.SuperfluidUndelegationsByDelegatorRequest) | [SuperfluidUndelegationsByDelegatorResponse](#osmosis.superfluid.SuperfluidUndelegationsByDelegatorResponse) |  | GET|/osmosis/superfluid/v1beta1/superfluid_undelegations_by_delegator/{delegator_address}|
| `SuperfluidDelegationsByValidatorDenom` | [SuperfluidDelegationsByValidatorDenomRequest](#osmosis.superfluid.SuperfluidDelegationsByValidatorDenomRequest) | [SuperfluidDelegationsByValidatorDenomResponse](#osmosis.superfluid.SuperfluidDelegationsByValidatorDenomResponse) | Returns all the superfluid positions of a specific denom delegated to one validator | GET|/osmosis/superfluid/v1beta1/superfluid_delegations_by_validator_denom|
| `EstimateSuperfluidDelegatedAmountByValidatorDenom` | [EstimateSuperfluidDelegatedAmountByValidatorDenomRequest](#osmosis.superfluid.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest) | [EstimateSuperfluidDelegatedAmountByValidatorDenomResponse](#osmosis.superfluid.EstimateSuperfluidDelegatedAmountByValidatorDenomResponse) | Returns the amount of a specific denom delegated to a specific validator This is labeled an estimate, because the way it calculates the amount can lead rounding errors from the true delegated amount | GET|/osmosis/superfluid/v1beta1/estimate_superfluid_delegation_amount_by_validator_denom|

 <!-- end services -->



<a name="osmosis/superfluid/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/superfluid/tx.proto



<a name="osmosis.superfluid.MsgLockAndSuperfluidDelegate"></a>

### MsgLockAndSuperfluidDelegate
MsgLockAndSuperfluidDelegate locks coins with the unbonding period duration,
and then does a superfluid lock from the newly created lockup, to the
specified validator addr.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `val_addr` | [string](#string) |  |  |






<a name="osmosis.superfluid.MsgLockAndSuperfluidDelegateResponse"></a>

### MsgLockAndSuperfluidDelegateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ID` | [uint64](#uint64) |  |  |






<a name="osmosis.superfluid.MsgSuperfluidDelegate"></a>

### MsgSuperfluidDelegate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `lock_id` | [uint64](#uint64) |  |  |
| `val_addr` | [string](#string) |  |  |






<a name="osmosis.superfluid.MsgSuperfluidDelegateResponse"></a>

### MsgSuperfluidDelegateResponse







<a name="osmosis.superfluid.MsgSuperfluidUnbondLock"></a>

### MsgSuperfluidUnbondLock



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `lock_id` | [uint64](#uint64) |  |  |






<a name="osmosis.superfluid.MsgSuperfluidUnbondLockResponse"></a>

### MsgSuperfluidUnbondLockResponse







<a name="osmosis.superfluid.MsgSuperfluidUndelegate"></a>

### MsgSuperfluidUndelegate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `lock_id` | [uint64](#uint64) |  |  |






<a name="osmosis.superfluid.MsgSuperfluidUndelegateResponse"></a>

### MsgSuperfluidUndelegateResponse







<a name="osmosis.superfluid.MsgUnPoolWhitelistedPool"></a>

### MsgUnPoolWhitelistedPool
MsgUnPoolWhitelistedPool Unpools every lock the sender has, that is
associated with pool pool_id. If pool_id is not approved for unpooling by
governance, this is a no-op. Unpooling takes the locked gamm shares, and runs
"ExitPool" on it, to get the constituent tokens. e.g. z gamm/pool/1 tokens
ExitPools into constituent tokens x uatom, y uosmo. Then it creates a new
lock for every constituent token, with the duration associated with the lock.
If the lock was unbonding, the new lockup durations should be the time left
until unbond completion.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `pool_id` | [uint64](#uint64) |  |  |






<a name="osmosis.superfluid.MsgUnPoolWhitelistedPoolResponse"></a>

### MsgUnPoolWhitelistedPoolResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exitedLockIds` | [uint64](#uint64) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.superfluid.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `SuperfluidDelegate` | [MsgSuperfluidDelegate](#osmosis.superfluid.MsgSuperfluidDelegate) | [MsgSuperfluidDelegateResponse](#osmosis.superfluid.MsgSuperfluidDelegateResponse) | Execute superfluid delegation for a lockup | |
| `SuperfluidUndelegate` | [MsgSuperfluidUndelegate](#osmosis.superfluid.MsgSuperfluidUndelegate) | [MsgSuperfluidUndelegateResponse](#osmosis.superfluid.MsgSuperfluidUndelegateResponse) | Execute superfluid undelegation for a lockup

Execute superfluid redelegation for a lockup rpc SuperfluidRedelegate(MsgSuperfluidRedelegate) returns (MsgSuperfluidRedelegateResponse); | |
| `SuperfluidUnbondLock` | [MsgSuperfluidUnbondLock](#osmosis.superfluid.MsgSuperfluidUnbondLock) | [MsgSuperfluidUnbondLockResponse](#osmosis.superfluid.MsgSuperfluidUnbondLockResponse) | For a given lock that is being superfluidly undelegated, also unbond the underlying lock. | |
| `LockAndSuperfluidDelegate` | [MsgLockAndSuperfluidDelegate](#osmosis.superfluid.MsgLockAndSuperfluidDelegate) | [MsgLockAndSuperfluidDelegateResponse](#osmosis.superfluid.MsgLockAndSuperfluidDelegateResponse) | Execute lockup lock and superfluid delegation in a single msg | |
| `UnPoolWhitelistedPool` | [MsgUnPoolWhitelistedPool](#osmosis.superfluid.MsgUnPoolWhitelistedPool) | [MsgUnPoolWhitelistedPoolResponse](#osmosis.superfluid.MsgUnPoolWhitelistedPoolResponse) |  | |

 <!-- end services -->



<a name="osmosis/tokenfactory/v1beta1/authorityMetadata.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/tokenfactory/v1beta1/authorityMetadata.proto



<a name="osmosis.tokenfactory.v1beta1.DenomAuthorityMetadata"></a>

### DenomAuthorityMetadata
DenomAuthorityMetadata specifies metadata for addresses that have specific
capabilities over a token factory denom. Right now there is only one Admin
permission, but is planned to be extended to the future.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `Admin` | [string](#string) |  | Can be empty for no admin, or a valid osmosis address |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/tokenfactory/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/tokenfactory/v1beta1/params.proto



<a name="osmosis.tokenfactory.v1beta1.Params"></a>

### Params
Params holds parameters for the tokenfactory module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/tokenfactory/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/tokenfactory/v1beta1/genesis.proto



<a name="osmosis.tokenfactory.v1beta1.GenesisDenom"></a>

### GenesisDenom



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `authority_metadata` | [DenomAuthorityMetadata](#osmosis.tokenfactory.v1beta1.DenomAuthorityMetadata) |  |  |






<a name="osmosis.tokenfactory.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the tokenfactory module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.tokenfactory.v1beta1.Params) |  | params defines the paramaters of the module. |
| `factory_denoms` | [GenesisDenom](#osmosis.tokenfactory.v1beta1.GenesisDenom) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/tokenfactory/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/tokenfactory/v1beta1/query.proto



<a name="osmosis.tokenfactory.v1beta1.QueryDenomAuthorityMetadataRequest"></a>

### QueryDenomAuthorityMetadataRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="osmosis.tokenfactory.v1beta1.QueryDenomAuthorityMetadataResponse"></a>

### QueryDenomAuthorityMetadataResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority_metadata` | [DenomAuthorityMetadata](#osmosis.tokenfactory.v1beta1.DenomAuthorityMetadata) |  |  |






<a name="osmosis.tokenfactory.v1beta1.QueryDenomsFromCreatorRequest"></a>

### QueryDenomsFromCreatorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator` | [string](#string) |  |  |






<a name="osmosis.tokenfactory.v1beta1.QueryDenomsFromCreatorResponse"></a>

### QueryDenomsFromCreatorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denoms` | [string](#string) | repeated |  |






<a name="osmosis.tokenfactory.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="osmosis.tokenfactory.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.tokenfactory.v1beta1.Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.tokenfactory.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#osmosis.tokenfactory.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#osmosis.tokenfactory.v1beta1.QueryParamsResponse) | Params returns the total set of minting parameters. | GET|/osmosis/tokenfactory/v1beta1/params|
| `DenomAuthorityMetadata` | [QueryDenomAuthorityMetadataRequest](#osmosis.tokenfactory.v1beta1.QueryDenomAuthorityMetadataRequest) | [QueryDenomAuthorityMetadataResponse](#osmosis.tokenfactory.v1beta1.QueryDenomAuthorityMetadataResponse) |  | GET|/osmosis/tokenfactory/v1beta1/denoms/{denom}/authority_metadata|
| `DenomsFromCreator` | [QueryDenomsFromCreatorRequest](#osmosis.tokenfactory.v1beta1.QueryDenomsFromCreatorRequest) | [QueryDenomsFromCreatorResponse](#osmosis.tokenfactory.v1beta1.QueryDenomsFromCreatorResponse) |  | GET|/osmosis/tokenfactory/v1beta1/denoms_from_creator/{creator}|

 <!-- end services -->



<a name="osmosis/tokenfactory/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/tokenfactory/v1beta1/tx.proto



<a name="osmosis.tokenfactory.v1beta1.MsgBurn"></a>

### MsgBurn
MsgBurn is the sdk.Msg type for allowing an admin account to burn
a token.  For now, we only support burning from the sender account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="osmosis.tokenfactory.v1beta1.MsgBurnResponse"></a>

### MsgBurnResponse







<a name="osmosis.tokenfactory.v1beta1.MsgChangeAdmin"></a>

### MsgChangeAdmin
MsgChangeAdmin is the sdk.Msg type for allowing an admin account to reassign
adminship of a denom to a new account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `newAdmin` | [string](#string) |  |  |






<a name="osmosis.tokenfactory.v1beta1.MsgChangeAdminResponse"></a>

### MsgChangeAdminResponse







<a name="osmosis.tokenfactory.v1beta1.MsgCreateDenom"></a>

### MsgCreateDenom
MsgCreateDenom is the sdk.Msg type for allowing an account to create
a new denom. It requires a sender address and a subdenomination.
The (sender_address, sub_denomination) pair must be unique and cannot be
re-used. The resulting denom created is `factory/{creator
address}/{subdenom}`. The resultant denom's admin is originally set to be the
creator, but this can be changed later. The token denom does not indicate the
current admin.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `subdenom` | [string](#string) |  | subdenom can be up to 44 "alphanumeric" characters long. |






<a name="osmosis.tokenfactory.v1beta1.MsgCreateDenomResponse"></a>

### MsgCreateDenomResponse
MsgCreateDenomResponse is the return value of MsgCreateDenom
It returns the full string of the newly created denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `new_token_denom` | [string](#string) |  |  |






<a name="osmosis.tokenfactory.v1beta1.MsgMint"></a>

### MsgMint
MsgMint is the sdk.Msg type for allowing an admin account to mint
more of a token.  For now, we only support minting to the sender account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="osmosis.tokenfactory.v1beta1.MsgMintResponse"></a>

### MsgMintResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.tokenfactory.v1beta1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateDenom` | [MsgCreateDenom](#osmosis.tokenfactory.v1beta1.MsgCreateDenom) | [MsgCreateDenomResponse](#osmosis.tokenfactory.v1beta1.MsgCreateDenomResponse) |  | |
| `Mint` | [MsgMint](#osmosis.tokenfactory.v1beta1.MsgMint) | [MsgMintResponse](#osmosis.tokenfactory.v1beta1.MsgMintResponse) |  | |
| `Burn` | [MsgBurn](#osmosis.tokenfactory.v1beta1.MsgBurn) | [MsgBurnResponse](#osmosis.tokenfactory.v1beta1.MsgBurnResponse) |  | |
| `ChangeAdmin` | [MsgChangeAdmin](#osmosis.tokenfactory.v1beta1.MsgChangeAdmin) | [MsgChangeAdminResponse](#osmosis.tokenfactory.v1beta1.MsgChangeAdminResponse) | ForceTransfer is deactivated for now because we need to think through edge cases rpc ForceTransfer(MsgForceTransfer) returns (MsgForceTransferResponse); | |

 <!-- end services -->



<a name="osmosis/txfees/v1beta1/feetoken.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/txfees/v1beta1/feetoken.proto



<a name="osmosis.txfees.v1beta1.FeeToken"></a>

### FeeToken
FeeToken is a struct that specifies a coin denom, and pool ID pair.
This marks the token as eligible for use as a tx fee asset in Osmosis.
Its price in osmo is derived through looking at the provided pool ID.
The pool ID must have osmo as one of its assets.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `poolID` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/txfees/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/txfees/v1beta1/genesis.proto



<a name="osmosis.txfees.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the txfees module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `basedenom` | [string](#string) |  |  |
| `feetokens` | [FeeToken](#osmosis.txfees.v1beta1.FeeToken) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/txfees/v1beta1/gov.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/txfees/v1beta1/gov.proto



<a name="osmosis.txfees.v1beta1.UpdateFeeTokenProposal"></a>

### UpdateFeeTokenProposal
UpdateFeeTokenProposal is a gov Content type for adding a new whitelisted fee
token. It must specify a denom along with gamm pool ID to use as a spot price
calculator. It can be used to add a new denom to the whitelist It can also be
used to update the Pool to associate with the denom. If Pool ID is set to 0,
it will remove the denom from the whitelisted set.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `feetoken` | [FeeToken](#osmosis.txfees.v1beta1.FeeToken) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/txfees/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/txfees/v1beta1/query.proto



<a name="osmosis.txfees.v1beta1.QueryBaseDenomRequest"></a>

### QueryBaseDenomRequest







<a name="osmosis.txfees.v1beta1.QueryBaseDenomResponse"></a>

### QueryBaseDenomResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_denom` | [string](#string) |  |  |






<a name="osmosis.txfees.v1beta1.QueryDenomPoolIdRequest"></a>

### QueryDenomPoolIdRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="osmosis.txfees.v1beta1.QueryDenomPoolIdResponse"></a>

### QueryDenomPoolIdResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolID` | [uint64](#uint64) |  |  |






<a name="osmosis.txfees.v1beta1.QueryDenomSpotPriceRequest"></a>

### QueryDenomSpotPriceRequest
QueryDenomSpotPriceRequest defines grpc request structure for querying spot
price for the specified tx fee denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="osmosis.txfees.v1beta1.QueryDenomSpotPriceResponse"></a>

### QueryDenomSpotPriceResponse
QueryDenomSpotPriceRequest defines grpc response structure for querying spot
price for the specified tx fee denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `poolID` | [uint64](#uint64) |  |  |
| `spot_price` | [string](#string) |  |  |






<a name="osmosis.txfees.v1beta1.QueryFeeTokensRequest"></a>

### QueryFeeTokensRequest







<a name="osmosis.txfees.v1beta1.QueryFeeTokensResponse"></a>

### QueryFeeTokensResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fee_tokens` | [FeeToken](#osmosis.txfees.v1beta1.FeeToken) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.txfees.v1beta1.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `FeeTokens` | [QueryFeeTokensRequest](#osmosis.txfees.v1beta1.QueryFeeTokensRequest) | [QueryFeeTokensResponse](#osmosis.txfees.v1beta1.QueryFeeTokensResponse) | FeeTokens returns a list of all the whitelisted fee tokens and their corresponding pools It does not include the BaseDenom, which has its own query endpoint | GET|/osmosis/txfees/v1beta1/fee_tokens|
| `DenomSpotPrice` | [QueryDenomSpotPriceRequest](#osmosis.txfees.v1beta1.QueryDenomSpotPriceRequest) | [QueryDenomSpotPriceResponse](#osmosis.txfees.v1beta1.QueryDenomSpotPriceResponse) |  | GET|/osmosis/txfees/v1beta1/spot_price_by_denom|
| `DenomPoolId` | [QueryDenomPoolIdRequest](#osmosis.txfees.v1beta1.QueryDenomPoolIdRequest) | [QueryDenomPoolIdResponse](#osmosis.txfees.v1beta1.QueryDenomPoolIdResponse) |  | GET|/osmosis/txfees/v1beta1/denom_pool_id/{denom}|
| `BaseDenom` | [QueryBaseDenomRequest](#osmosis.txfees.v1beta1.QueryBaseDenomRequest) | [QueryBaseDenomResponse](#osmosis.txfees.v1beta1.QueryBaseDenomResponse) |  | GET|/osmosis/txfees/v1beta1/base_denom|

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

