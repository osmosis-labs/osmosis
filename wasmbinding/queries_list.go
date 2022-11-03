package wasmbinding

import (
	"github.com/cosmos/cosmos-sdk/codec"
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
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v12/x/tokenfactory/types"
	twapquerytypes "github.com/osmosis-labs/osmosis/v12/x/twap/client/queryproto"
	txfeestypes "github.com/osmosis-labs/osmosis/v12/x/txfees/types"
)

var (
	QueriesList = []codec.ProtoMarshaler{
		&authtypes.QueryAccountResponse{},
		&authtypes.QueryParamsResponse{},
		&banktypes.QueryBalanceResponse{},
		&banktypes.QueryDenomsMetadataResponse{},
		&banktypes.QueryParamsResponse{},
		&banktypes.QuerySupplyOfResponse{},
		&distributiontypes.QueryParamsResponse{},
		&distributiontypes.QueryDelegatorWithdrawAddressResponse{},
		&distributiontypes.QueryValidatorCommissionResponse{},
		&govtypes.QueryDepositResponse{},
		&govtypes.QueryParamsResponse{},
		&govtypes.QueryVoteResponse{},
		&slashingtypes.QueryParamsResponse{},
		&slashingtypes.QuerySigningInfoResponse{},
		&stakingtypes.QueryDelegationResponse{},
		&stakingtypes.QueryParamsResponse{},
		&stakingtypes.QueryValidatorResponse{},
		&epochtypes.QueryEpochsInfoResponse{},
		&epochtypes.QueryCurrentEpochResponse{},
		&gammtypes.QueryNumPoolsResponse{},
		&gammtypes.QueryTotalLiquidityResponse{},
		&gammtypes.QueryPoolResponse{},
		&gammtypes.QueryPoolParamsResponse{},
		&gammtypes.QueryTotalPoolLiquidityResponse{},
		&gammtypes.QueryTotalSharesResponse{},
		&gammtypes.QuerySpotPriceResponse{},
		&incentivestypes.ModuleToDistributeCoinsResponse{},
		&incentivestypes.QueryLockableDurationsResponse{},
		&lockuptypes.ModuleBalanceResponse{},
		&lockuptypes.ModuleLockedAmountResponse{},
		&lockuptypes.AccountUnlockableCoinsResponse{},
		&lockuptypes.AccountUnlockingCoinsResponse{},
		&lockuptypes.LockedDenomResponse{},
		&minttypes.QueryEpochProvisionsResponse{},
		&minttypes.QueryParamsResponse{},
		&poolincentivestypes.QueryGaugeIdsResponse{},
		&superfluidtypes.QueryParamsResponse{},
		&superfluidtypes.AssetTypeResponse{},
		&superfluidtypes.AllAssetsResponse{},
		&superfluidtypes.AssetMultiplierResponse{},
		&txfeestypes.QueryFeeTokensResponse{},
		&txfeestypes.QueryDenomSpotPriceResponse{},
		&txfeestypes.QueryDenomPoolIdResponse{},
		&txfeestypes.QueryBaseDenomResponse{},
		&tokenfactorytypes.QueryParamsResponse{},
		&tokenfactorytypes.QueryDenomAuthorityMetadataResponse{},
		&twapquerytypes.ArithmeticTwapResponse{},
		&twapquerytypes.ArithmeticTwapToNowResponse{},
		&twapquerytypes.ParamsResponse{},
	}
)
