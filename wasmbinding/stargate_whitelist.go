package wasmbinding

import (
	"fmt"
	"reflect"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	grpc "google.golang.org/grpc"

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

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

// This is
type GRPCQueriesInfo struct {
	QueryPaths    []string
	QueryReponses []codec.ProtoMarshaler
}

func (g GRPCQueriesInfo) RegisterQueryReponse(queryServer interface{}) {
	handlers := reflect.TypeOf(queryServer).Elem()

	// adds a top-level query handler based on the gRPC service name
	for i := 0; i < handlers.NumMethod(); i++ {
		qResponse := reflect.New(handlers.Method(i).Type.Out(0).Elem())

		fmt.Println(qResponse.CanInterface(), "can interface")
		fmt.Println(qResponse.Interface())
		qResponseType, ok := qResponse.Interface().(codec.ProtoMarshaler)
		if !ok {
			panic("can't")
		}
		g.QueryReponses = append(g.QueryReponses, qResponseType)
	}
}

//
func (g GRPCQueriesInfo) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	for _, method := range sd.Methods {
		fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		g.QueryPaths = append(g.QueryPaths, fqName)
	}
}

func init() {
	// cosmos-sdk queries
	g := GRPCQueriesInfo{}

	// auth
	authtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*authtypes.QueryServer)(nil))

	// bank
	banktypes.RegisterQueryServer(g, nil)

	// distribution
	distributiontypes.RegisterQueryServer(g, nil)

	// gov
	govtypes.RegisterQueryServer(g, nil)

	// slashing
	slashingtypes.RegisterQueryServer(g, nil)

	// staking
	stakingtypes.RegisterQueryServer(g, nil)

	// osmosis queries

	// epochs
	epochtypes.RegisterQueryServer(g, nil)

	// gamm
	gammtypes.RegisterQueryServer(g, nil)

	// incentives
	incentivestypes.RegisterQueryServer(g, nil)

	// lockup
	lockuptypes.RegisterQueryServer(g, nil)

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
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/ArithmeticTwap", &twapquerytypes.ArithmeticTwapResponse{})
	setWhitelistedQuery("/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow", &twapquerytypes.ArithmeticTwapToNowResponse{})
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
