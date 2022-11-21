package wasmbinding

import (
	"fmt"
	"reflect"
	"sort"
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

// Design struct:
//  - We're using GRPCQueriesInfo struct to store the query path and corresponding response. Slides sort in alphabetical order.
//    So QueryPaths[i] match with QueryReponses[i]
//  - RegisterService implement QueryServer interface in query.pb files. Purpose to get all query path of module
//  - RegisterQueryReponse: Get all query responses of module.
// We go through all module and get query paths & responses. Then use `queries.list` to filter whitelisted queries.

// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

type GRPCQueriesInfo struct {
	QueryPaths    []string
	QueryReponses []codec.ProtoMarshaler
}

func Contains(sl []codec.ProtoMarshaler, ele codec.ProtoMarshaler) bool {
	for _, value := range sl {
		if reflect.DeepEqual(value, ele) {
			return true
		}
	}
	return false
}

func (g *GRPCQueriesInfo) RegisterQueryReponse(queryServer interface{}) {
	handlers := reflect.TypeOf(queryServer).Elem()
	// adds a top-level query handler based on the gRPC service name
	for i := 0; i < handlers.NumMethod(); i++ {
		qResponse := reflect.New(handlers.Method(i).Type.Out(0).Elem())
		qResponseType, ok := qResponse.Interface().(codec.ProtoMarshaler)
		if !ok {
			panic("can't")
		}
		g.QueryReponses = append(g.QueryReponses, qResponseType)
	}
}

// RegisterService collects all the query path of a module query service
func (g *GRPCQueriesInfo) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	for _, method := range sd.Methods {
		fqName := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		g.QueryPaths = append(g.QueryPaths, fqName)
	}
	sort.Strings(g.QueryPaths)
}

func (g *GRPCQueriesInfo) Whitelist() {
	for id := range g.QueryPaths {
		if Contains(QueriesList, g.QueryReponses[id]) {
			setWhitelistedQuery(g.QueryPaths[id], g.QueryReponses[id])
		}
	}
}

func init() {
	// cosmos-sdk queries
	g := &GRPCQueriesInfo{}
	// auth
	authtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*authtypes.QueryServer)(nil))

	// bank
	banktypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*banktypes.QueryServer)(nil))

	// distribution
	distributiontypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*distributiontypes.QueryServer)(nil))

	// gov
	govtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*govtypes.QueryServer)(nil))

	// slashing
	slashingtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*slashingtypes.QueryServer)(nil))

	// staking
	stakingtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*stakingtypes.QueryServer)(nil))

	// osmosis queries

	// epochs
	epochtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*epochtypes.QueryServer)(nil))

	// gamm
	gammtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*gammtypes.QueryServer)(nil))

	// gamm v2
	gammv2types.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*gammv2types.QueryServer)(nil))

	// incentives
	incentivestypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*incentivestypes.QueryServer)(nil))

	// lockup
	lockuptypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*lockuptypes.QueryServer)(nil))

	// mint
	minttypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*minttypes.QueryServer)(nil))

	// pool-incentives
	poolincentivestypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*poolincentivestypes.QueryServer)(nil))

	// superfluid
	superfluidtypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*superfluidtypes.QueryServer)(nil))

	// txfees
	txfeestypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*txfeestypes.QueryServer)(nil))

	// tokenfactory
	tokenfactorytypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*tokenfactorytypes.QueryServer)(nil))
	// Does not include denoms_from_creator, TBD if this is the index we want contracts to use instead of admin

	// twap
	twapquerytypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*twapquerytypes.QueryServer)(nil))

	// twap v2
	twapv2querytypes.RegisterQueryServer(g, nil)
	g.RegisterQueryReponse((*twapv2querytypes.QueryServer)(nil))

	g.Whitelist()
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
