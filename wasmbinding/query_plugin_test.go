package wasmbinding_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	proto "github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/protobuf/runtime/protoiface"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"github.com/osmosis-labs/osmosis/v12/app"
	epochtypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"

	"github.com/osmosis-labs/osmosis/v12/wasmbinding"
)

type StargateTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.OsmosisApp
}

func (suite *StargateTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
}

func TestStargateTestSuite(t *testing.T) {
	suite.Run(t, new(StargateTestSuite))
}

func (suite *StargateTestSuite) TestStargateQuerier() {
	testCases := []struct {
		name                   string
		testSetup              func()
		path                   string
		requestData            func() []byte
		responseProtoStruct    interface{}
		expectedQuerierError   bool
		expectedUnMarshalError bool
	}{
		{
			name: "happy path",
			path: "/osmosis.epochs.v1beta1.Query/EpochInfos",
			requestData: func() []byte {
				epochrequest := epochtypes.QueryEpochsInfoRequest{}
				bz, err := proto.Marshal(&epochrequest)
				suite.Require().NoError(err)
				return bz
			},
			responseProtoStruct: &epochtypes.QueryEpochsInfoResponse{},
		},
		{
			name: "unregistered path(not whitelisted)",
			path: "/osmosis.epochs.v1beta1.Query/CurrentEpoch",
			requestData: func() []byte {
				currentEpochRequest := epochtypes.QueryCurrentEpochRequest{}
				bz, err := proto.Marshal(&currentEpochRequest)
				suite.Require().NoError(err)
				return bz
			},
			expectedQuerierError: true,
		},
		{
			name: "invalid query router route",
			testSetup: func() {
				wasmbinding.StargateWhitelist.Store("invalid/query/router/route", epochtypes.QueryEpochsInfoRequest{})
			},
			path: "invalid/query/router/route",
			requestData: func() []byte {
				return []byte{}
			},
			expectedQuerierError: true,
		},
		{
			name: "unmatching path and data in request",
			path: "/osmosis.epochs.v1beta1.Query/EpochInfos",
			requestData: func() []byte {
				epochrequest := epochtypes.QueryCurrentEpochRequest{}
				bz, err := proto.Marshal(&epochrequest)
				suite.Require().NoError(err)
				return bz
			},
			responseProtoStruct:    &epochtypes.QueryCurrentEpochResponse{},
			expectedUnMarshalError: true,
		},
		{
			name: "error in unmarshalling response",
			// set up whitelist with wrong data
			testSetup: func() {
				wasmbinding.StargateWhitelist.Store("/osmosis.epochs.v1beta1.Query/EpochInfos", interface{}(nil))
			},
			path: "/osmosis.epochs.v1beta1.Query/EpochInfos",
			requestData: func() []byte {
				return []byte{}
			},
			responseProtoStruct:  &epochtypes.QueryCurrentEpochResponse{},
			expectedQuerierError: true,
		},
		{
			name: "error in grpc querier",
			// set up whitelist with wrong data
			testSetup: func() {
				wasmbinding.StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/AllBalances", banktypes.QueryAllBalancesRequest{})
			},
			path: "/cosmos.bank.v1beta1.Query/AllBalances",
			requestData: func() []byte {
				bankrequest := banktypes.QueryAllBalancesRequest{}
				bz, err := proto.Marshal(&bankrequest)
				suite.Require().NoError(err)
				return bz
			},
			responseProtoStruct:  &banktypes.QueryAllBalancesRequest{},
			expectedQuerierError: true,
		},

		// TODO: errors in wrong query in state machine
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			if tc.testSetup != nil {
				tc.testSetup()
			}

			stargateQuerier := wasmbinding.StargateQuerier(*suite.app.GRPCQueryRouter(), suite.app.AppCodec())
			stargateRequest := &wasmvmtypes.StargateQuery{
				Path: tc.path,
				Data: tc.requestData(),
			}
			stargateResponse, err := stargateQuerier(suite.ctx, stargateRequest)
			if tc.expectedQuerierError {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)

				protoResponse, ok := tc.responseProtoStruct.(proto.Message)
				suite.Require().True(ok)

				// test correctness by unmarshalling json response into proto struct
				err = suite.app.AppCodec().UnmarshalJSON(stargateResponse, protoResponse)
				if tc.expectedUnMarshalError {
					suite.Require().Error(err)
				} else {
					suite.Require().NoError(err)
					suite.Require().NotNil(protoResponse)
				}
			}
		})
	}
}

func (suite *StargateTestSuite) TestConvertProtoToJsonMarshal() {
	testCases := []struct {
		name                  string
		queryPath             string
		protoResponseStruct   proto.Message
		originalResponse      string
		expectedProtoResponse proto.Message
		expectedError         bool
	}{
		{
			name:                "successful conversion from proto response to json marshalled response",
			queryPath:           "/cosmos.bank.v1beta1.Query/AllBalances",
			originalResponse:    "0a090a036261721202333012050a03666f6f",
			protoResponseStruct: &banktypes.QueryAllBalancesResponse{},
			expectedProtoResponse: &banktypes.QueryAllBalancesResponse{
				Balances: sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(30))),
				Pagination: &query.PageResponse{
					NextKey: []byte("foo"),
				},
			},
		},
		{
			name:                "invalid proto response struct",
			queryPath:           "/cosmos.bank.v1beta1.Query/AllBalances",
			originalResponse:    "0a090a036261721202333012050a03666f6f",
			protoResponseStruct: protoiface.MessageV1(nil),
			expectedError:       true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			originalVersionBz, err := hex.DecodeString(tc.originalResponse)
			suite.Require().NoError(err)

			jsonMarshalledResponse, err := wasmbinding.ConvertProtoToJSONMarshal(tc.protoResponseStruct, originalVersionBz, suite.app.AppCodec())
			if tc.expectedError {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)

			// check response by json marshalling proto response into json response manually
			jsonMarshalExpectedResponse, err := suite.app.AppCodec().MarshalJSON(tc.expectedProtoResponse)
			suite.Require().NoError(err)
			suite.Require().Equal(jsonMarshalledResponse, jsonMarshalExpectedResponse)
		})
	}
}

// TestDeterministicJsonMarshal tests that we get deterministic JSON marshalled response upon
// proto struct update in the state machine.
func (suite *StargateTestSuite) TestDeterministicJsonMarshal() {
	testCases := []struct {
		name                string
		originalResponse    string
		updatedResponse     string
		queryPath           string
		responseProtoStruct interface{}
		expectedProto       func() proto.Message
	}{
		/**
		   *
		   * Origin Response
		   * 0a530a202f636f736d6f732e617574682e763162657461312e426173654163636f756e74122f0a2d636f736d6f7331346c3268686a6e676c3939367772703935673867646a6871653038326375367a7732706c686b
		   *
		   * Updated Response
		   * 0a530a202f636f736d6f732e617574682e763162657461312e426173654163636f756e74122f0a2d636f736d6f7331646a783375676866736d6b6135386676673076616a6e6533766c72776b7a6a346e6377747271122d636f736d6f7331646a783375676866736d6b6135386676673076616a6e6533766c72776b7a6a346e6377747271
		  // Origin proto
		  message QueryAccountResponse {
		    // account defines the account of the corresponding address.
		    google.protobuf.Any account = 1 [(cosmos_proto.accepts_interface) = "AccountI"];
		  }
		  // Updated proto
		  message QueryAccountResponse {
		    // account defines the account of the corresponding address.
		    google.protobuf.Any account = 1 [(cosmos_proto.accepts_interface) = "AccountI"];
		    // address is the address to query for.
		  	string address = 2;
		  }
		*/
		{
			"Query Account",
			"0a530a202f636f736d6f732e617574682e763162657461312e426173654163636f756e74122f0a2d636f736d6f733166387578756c746e3873717a687a6e72737a3371373778776171756867727367366a79766679",
			"0a530a202f636f736d6f732e617574682e763162657461312e426173654163636f756e74122f0a2d636f736d6f733166387578756c746e3873717a687a6e72737a3371373778776171756867727367366a79766679122d636f736d6f733166387578756c746e3873717a687a6e72737a3371373778776171756867727367366a79766679",
			"/cosmos.auth.v1beta1.Query/Account",
			&authtypes.QueryAccountResponse{},
			func() proto.Message {
				account := authtypes.BaseAccount{
					Address: "cosmos1f8uxultn8sqzhznrsz3q77xwaquhgrsg6jyvfy",
				}
				accountResponse, err := codectypes.NewAnyWithValue(&account)
				suite.Require().NoError(err)
				return &authtypes.QueryAccountResponse{
					Account: accountResponse,
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			binding, ok := wasmbinding.StargateWhitelist.Load(tc.queryPath)
			suite.Require().True(ok)

			originVersionBz, err := hex.DecodeString(tc.originalResponse)
			suite.Require().NoError(err)
			jsonMarshalledOriginalBz, err := wasmbinding.ConvertProtoToJSONMarshal(binding, originVersionBz, suite.app.AppCodec())
			suite.Require().NoError(err)

			newVersionBz, err := hex.DecodeString(tc.updatedResponse)
			suite.Require().NoError(err)
			jsonMarshalledUpdatedBz, err := wasmbinding.ConvertProtoToJSONMarshal(binding, newVersionBz, suite.app.AppCodec())
			suite.Require().NoError(err)

			// json marshalled bytes should be the same since we use the same proto struct for unmarshalling
			suite.Require().Equal(jsonMarshalledOriginalBz, jsonMarshalledUpdatedBz)

			// raw build also make same result
			jsonMarshalExpectedResponse, err := suite.app.AppCodec().MarshalJSON(tc.expectedProto())
			suite.Require().NoError(err)
			suite.Require().Equal(jsonMarshalledUpdatedBz, jsonMarshalExpectedResponse)
		})
	}
}
