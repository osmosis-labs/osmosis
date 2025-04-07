package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	//nolint:staticcheck
	"github.com/golang/protobuf/proto"
	markettypes "github.com/osmosis-labs/osmosis/v27/x/market/types"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/wasmbinding"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	concentratedliquidityquery "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/queryproto"
	downtimequerytypes "github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/queryproto"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	superfluidtypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
	twapquerytypes "github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

// convert requested proto struct into proto marshalled bytes
func DebugProtoMarshalledBytes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proto-marshalled-bytes [query-path] [module] [struct-name] [struct-arguments...]",
		Short: "Convert request proto struct into proto marhsalled bytes ",
		Long: `Convert request proto struct into proto marshalled bytes.
Especially useful when debugging proto marshalled bytes or debugging stargate queries

Example:
	symphonyd debug proto-marshalled-bytes "/cosmos.bank.v1beta1.Query/Balance" bank QueryBalanceRequest "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w" stake
	`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			structArguments := args[3:]
			structI, err := GetStructAndFill(args[0], args[1], args[2], structArguments...)
			if err != nil {
				return err
			}

			// convert back to proto message
			protoMessage, ok := structI.(proto.Message)
			if !ok {
				return fmt.Errorf("error when converting back to proto message")
			}
			bytes, err := proto.Marshal(protoMessage)
			if err != nil {
				return err
			}

			cmd.Println(bytes)
			return nil
		},
	}

	return cmd
}

//nolint:staticcheck
func GetStructAndFill(queryPath, module, structName string, structArguments ...string) (interface{}, error) {
	const ParamRequest = "QueryParamsRequest"
	err := wasmbinding.IsWhitelistedQuery(queryPath)
	if err != nil {
		return nil, err
	}

	switch module {
	case "auth":
		switch structName {
		case "QueryAccountRequest":
			v := &authtypes.QueryAccountRequest{}
			v.Address = structArguments[0]
			return v, nil

		case ParamRequest:
			v := &authtypes.QueryParamsRequest{}
			return v, nil
		}
	case "bank":
		switch structName {
		case "QueryBalanceRequest":
			v := &banktypes.QueryBalanceRequest{}
			v.Address = structArguments[0]
			v.Denom = structArguments[1]
			return v, nil
		case "QueryDenomsMetadataRequest":
			v := &banktypes.QueryDenomsMetadataRequest{}
			return v, nil
		case ParamRequest:
			v := &banktypes.QueryParamsRequest{}
			return v, nil
		case "QuerySupplyOfRequest":
			v := &banktypes.QuerySupplyOfRequest{}
			v.Denom = structArguments[0]
			return v, nil
		}
	case "distribution":
		switch structName {
		case ParamRequest:
			v := &distributiontypes.QueryParamsRequest{}
			return v, nil
		case "QueryDelegatorWithdrawAddressRequest":
			v := &distributiontypes.QueryDelegatorWithdrawAddressRequest{}
			return v, nil
		case "QueryValidatorCommissionRequest":
			v := &distributiontypes.QueryValidatorCommissionRequest{}
			v.ValidatorAddress = structArguments[0]
			return v, nil
		}
	case "gov":
		switch structName {
		case "QueryDepositRequest":
			v := &govtypesv1.QueryDepositRequest{}
			proposalId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.ProposalId = proposalId
			v.Depositor = structArguments[1]
			return v, nil
		case ParamRequest:
			v := &govtypesv1.QueryParamsRequest{}
			return v, nil
		case "QueryVoteRequest":
			v := &govtypesv1.QueryVoteRequest{}
			proposalId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.ProposalId = proposalId
			v.Voter = structArguments[1]
			return v, nil
		}
	case "slashing":
		switch structName {
		case ParamRequest:
			v := &slashingtypes.QueryParamsRequest{}
			return v, nil
		case "QuerySigningInfoRequest":
			v := &slashingtypes.QuerySigningInfoRequest{}
			v.ConsAddress = structArguments[0]
			return v, nil
		}
	case "staking":
		switch structName {
		case "QueryDelegationRequest":
			v := &stakingtypes.QueryDelegationRequest{}
			v.DelegatorAddr = structArguments[0]
			v.ValidatorAddr = structArguments[1]
			return v, nil
		case ParamRequest:
			v := &stakingtypes.QueryParamsRequest{}
			return v, nil
		case "QueryValidatorRequest":
			v := &stakingtypes.QueryValidatorRequest{}
			v.ValidatorAddr = structArguments[0]
			return v, nil
		}
	case "epochs":
		switch structName {
		case "QueryEpochsInfoRequest":
			v := &epochtypes.QueryEpochsInfoRequest{}
			return v, nil
		case "QueryCurrentEpochRequest":
			v := &epochtypes.QueryCurrentEpochRequest{}
			return v, nil
		}
	case "gamm":
		switch structName {
		case "QueryNumPoolsRequest":
			v := &gammtypes.QueryNumPoolsRequest{}
			return v, nil
		case "QueryTotalLiquidityRequest":
			v := &gammtypes.QueryTotalLiquidityRequest{}
			return v, nil
		case "QueryPoolRequest":
			v := &gammtypes.QueryPoolRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "QueryPoolParamsRequest":
			v := &gammtypes.QueryPoolParamsRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "QueryTotalPoolLiquidityRequest":
			v := &gammtypes.QueryTotalPoolLiquidityRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "QueryTotalSharesRequest":
			v := &gammtypes.QueryTotalSharesRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "QueryCalcJoinPoolSharesRequest":
			v := &gammtypes.QueryCalcJoinPoolSharesRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			coins, err := sdk.ParseCoinsNormalized(structArguments[1])
			if err != nil {
				return nil, err
			}
			v.TokensIn = coins
			return v, nil
		case "QueryCalcExitPoolCoinsFromSharesRequest":
			v := &gammtypes.QueryCalcExitPoolCoinsFromSharesRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			sdkInt, ok := osmomath.NewIntFromString(structArguments[1])
			if !ok {
				return nil, fmt.Errorf("failed to parse to osmomath.Int")
			}
			v.ShareInAmount = sdkInt
			return v, nil
		case "QueryCalcJoinPoolNoSwapSharesRequest":
			v := &gammtypes.QueryCalcJoinPoolNoSwapSharesRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			coins, err := sdk.ParseCoinsNormalized(structArguments[1])
			if err != nil {
				return nil, err
			}
			v.TokensIn = coins
			return v, nil
		case "QueryPoolTypeRequest":
			v := &gammtypes.QueryPoolTypeRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "QuerySpotPriceRequest":
			v := &gammtypes.QuerySpotPriceRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			v.BaseAssetDenom = structArguments[1]
			v.QuoteAssetDenom = structArguments[2]
			return v, nil
		case "QuerySwapExactAmountInRequest":
			return nil, fmt.Errorf("swap route parsing not supported yet")
		case "QuerySwapExactAmountOutRequest":
			return nil, fmt.Errorf("swap route parsing not supported yet")
		}
	case "incentives":
		switch structName {
		case "ModuleToDistributeCoinsRequest":
			v := &incentivestypes.ModuleToDistributeCoinsRequest{}
			return v, nil
		case "QueryLockableDurationsRequest":
			v := &incentivestypes.QueryLockableDurationsRequest{}
			return v, nil
		}
	case "lockup":
		switch structName {
		case "ModuleBalanceRequest":
		case "ModuleToDistributeCoinsRequest":
			v := &lockuptypes.ModuleBalanceRequest{}
			return v, nil
		case "ModuleLockedAmountRequest":
			v := &lockuptypes.ModuleLockedAmountRequest{}
			return v, nil
		case "AccountUnlockableCoinsRequest":
			v := &lockuptypes.AccountUnlockableCoinsRequest{}
			v.Owner = structArguments[0]
			return v, nil
		case "AccountUnlockingCoinsRequest":
			v := &lockuptypes.AccountUnlockingCoinsRequest{}
			v.Owner = structArguments[0]
			return v, nil
		case "LockedDenomRequest":
			v := &lockuptypes.LockedDenomRequest{}
			v.Denom = structArguments[0]
			duration, err := time.ParseDuration(structArguments[1])
			if err != nil {
				return nil, err
			}
			v.Duration = duration
			return v, nil
		case "LockedRequest":
			v := &lockuptypes.LockedRequest{}
			lockId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.LockId = lockId
			return v, nil
		case "NextLockIDRequest":
			v := &lockuptypes.NextLockIDRequest{}
			return v, nil
		case "LockRewardReceiverRequest":
			v := &lockuptypes.LockRewardReceiverRequest{}
			lockId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.LockId = lockId
			return v, nil
		}
	case "mint":
		switch structName {
		case "QueryEpochProvisionsRequest":
			v := &minttypes.QueryEpochProvisionsRequest{}
			return v, nil
		case ParamRequest:
			v := &minttypes.QueryParamsRequest{}
			return v, nil
		}
	case "pool-incentives":
		switch structName {
		case "QueryGaugeIdsRequest":
			v := &poolincentivestypes.QueryGaugeIdsRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case ParamRequest:
			v := &poolincentivestypes.QueryParamsRequest{}
			return v, nil
		}
	case "superfluid":
		switch structName {
		case ParamRequest:
			v := &superfluidtypes.QueryParamsRequest{}
			return v, nil
		case "AssetTypeRequest":
			v := &superfluidtypes.AssetTypeRequest{}
			v.Denom = structArguments[0]
			return v, nil
		case "AllAssetsRequest":
			v := &superfluidtypes.AllAssetsRequest{}
			return v, nil
		case "AssetMultiplierRequest":
			v := &superfluidtypes.AssetMultiplierRequest{}
			v.Denom = structArguments[0]
			return v, nil
		}
	case "poolmanager":
		switch structName {
		case "NumPoolsRequest":
			v := &poolmanagerqueryproto.NumPoolsRequest{}
			return v, nil
		case "EstimateSwapExactAmountInRequest":
			return nil, fmt.Errorf("swap route parsing not supported yet")
		case "EstimateSwapExactAmountOutRequest":
			return nil, fmt.Errorf("swap route parsing not supported yet")
		case "PoolRequest":
			v := &poolmanagerqueryproto.PoolRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "SpotPriceRequest":
			v := &poolmanagerqueryproto.SpotPriceRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			v.BaseAssetDenom = structArguments[1]
			v.QuoteAssetDenom = structArguments[2]
			return v, nil
		case "EstimateTradeBasedOnPriceImpactRequest":
			return nil, fmt.Errorf("swap route parsing not supported yet")
		}
	case "market":
		switch structName {
		case "SwapRequest":
			v := &markettypes.QuerySwapRequest{}
			v.OfferCoin = structArguments[0]
			v.AskDenom = structArguments[1]
			return v, nil
		}
	case "txfees":
		switch structName {
		case "QueryFeeTokensRequest":
			v := &txfeestypes.QueryFeeTokensRequest{}
			return v, nil
		case "QueryDenomSpotPriceRequest":
			v := &txfeestypes.QueryDenomSpotPriceRequest{}
			return v, nil
		}
	case "twap":
		switch structName {
		case "ArithmeticTwapRequest":
			v := &twapquerytypes.ArithmeticTwapRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			v.BaseAsset = structArguments[1]
			v.QuoteAsset = structArguments[2]
			startTime, err := osmoutils.ParseTimeString(structArguments[3])
			if err != nil {
				return nil, err
			}
			endTime, err := osmoutils.ParseTimeString(structArguments[4])
			if err != nil {
				return nil, err
			}
			v.StartTime = startTime
			v.EndTime = &endTime

			return v, nil
		case "ArithmeticTwapToNowRequest":
			v := &twapquerytypes.ArithmeticTwapToNowRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			v.BaseAsset = structArguments[1]
			v.QuoteAsset = structArguments[2]
			startTime, err := osmoutils.ParseTimeString(structArguments[3])
			if err != nil {
				return nil, err
			}
			v.StartTime = startTime
			return v, nil
		case "GeometricTwapRequest":
			v := &twapquerytypes.GeometricTwapRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			v.BaseAsset = structArguments[1]
			v.QuoteAsset = structArguments[2]
			startTime, err := osmoutils.ParseTimeString(structArguments[3])
			if err != nil {
				return nil, err
			}
			endTime, err := osmoutils.ParseTimeString(structArguments[4])
			if err != nil {
				return nil, err
			}
			v.StartTime = startTime
			v.EndTime = &endTime
			return v, nil
		case "GeometricTwapToNowRequest":
			v := &twapquerytypes.GeometricTwapToNowRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			v.BaseAsset = structArguments[1]
			v.QuoteAsset = structArguments[2]
			startTime, err := osmoutils.ParseTimeString(structArguments[3])
			if err != nil {
				return nil, err
			}
			v.StartTime = startTime
			return v, nil
		case "ParamsRequest":
			v := &twapquerytypes.ParamsRequest{}
			return v, nil
		}
	case "downtime-detector":
		switch structName {
		case "RecoveredSinceDowntimeOfLengthRequest":
			v := &downtimequerytypes.RecoveredSinceDowntimeOfLengthRequest{}
			return v, nil
		}
	case "concentrated-liquidity":
		switch structName {
		case "PoolsRequest":
			v := &concentratedliquidityquery.PoolsRequest{}
			return v, nil
		case "UserPositionsRequest":
			v := &concentratedliquidityquery.UserPositionsRequest{}
			v.Address = structArguments[0]
			poolId, err := strconv.ParseUint(structArguments[1], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "LiquidityPerTickRangeRequest":
			v := &concentratedliquidityquery.LiquidityPerTickRangeRequest{}
			poolId, err := strconv.ParseUint(structArguments[1], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "LiquidityNetInDirectionRequest":
			v := &concentratedliquidityquery.LiquidityNetInDirectionRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			v.TokenIn = structArguments[1]
			startTick, err := strconv.ParseInt(structArguments[2], 10, 64)
			if err != nil {
				return nil, err
			}
			v.StartTick = startTick
			useCurTick, err := strconv.ParseBool(structArguments[3])
			if err != nil {
				return nil, err
			}
			v.UseCurTick = useCurTick
			boundTick, err := strconv.ParseInt(structArguments[4], 10, 64)
			if err != nil {
				return nil, err
			}
			v.BoundTick = boundTick
			useNoBound, err := strconv.ParseBool(structArguments[5])
			if err != nil {
				return nil, err
			}
			v.UseNoBound = useNoBound
			return v, nil
		case "ClaimableSpreadRewardsRequest":
			v := &concentratedliquidityquery.ClaimableSpreadRewardsRequest{}
			positionId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PositionId = positionId
			return v, nil
		case "ClaimableIncentivesRequest":
			v := &concentratedliquidityquery.ClaimableIncentivesRequest{}
			positionId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PositionId = positionId
			return v, nil
		case "PositionByIdRequest":
			v := &concentratedliquidityquery.PositionByIdRequest{}
			positionId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PositionId = positionId
			return v, nil
		case "ParamsRequest":
			v := &concentratedliquidityquery.ParamsRequest{}
			return v, nil
		case "PoolAccumulatorRewardsRequest":
			v := &concentratedliquidityquery.PoolAccumulatorRewardsRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "IncentiveRecordsRequest":
			v := &concentratedliquidityquery.IncentiveRecordsRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			return v, nil
		case "TickAccumulatorTrackersRequest":
			v := &concentratedliquidityquery.TickAccumulatorTrackersRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.PoolId = poolId
			tickIndex, err := strconv.ParseInt(structArguments[1], 10, 64)
			if err != nil {
				return nil, err
			}
			v.TickIndex = tickIndex
			return v, nil
		case "CFMMPoolIdLinkFromConcentratedPoolIdRequest":
			v := &concentratedliquidityquery.CFMMPoolIdLinkFromConcentratedPoolIdRequest{}
			poolId, err := strconv.ParseUint(structArguments[0], 10, 64)
			if err != nil {
				return nil, err
			}
			v.ConcentratedPoolId = poolId
			return v, nil
		}
	}

	return nil, errors.New("unknown module/struct")
}
