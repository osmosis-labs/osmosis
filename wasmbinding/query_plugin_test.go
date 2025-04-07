package wasmbinding_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	proto "github.com/golang/protobuf/proto" //nolint:staticcheck // we're intentionally using this deprecated package to be compatible with cosmos protos
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammv2types "github.com/osmosis-labs/osmosis/v27/x/gamm/v2types"

	"github.com/osmosis-labs/osmosis/v27/app"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	"github.com/osmosis-labs/osmosis/v27/wasmbinding"
)

type StargateTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     *app.OsmosisApp
	HomeDir string
}

func (suite *StargateTestSuite) SetupTestInternal() {
	suite.HomeDir = fmt.Sprintf("%d", rand.Int())
	suite.app = app.SetupWithCustomHome(false, suite.HomeDir)
	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
}

func (suite *StargateTestSuite) TearDownTestInternal() {
	os.RemoveAll(suite.HomeDir)
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
		responseProtoStruct    proto.Message
		expectedQuerierError   bool
		expectedUnMarshalError bool
		resendRequest          bool
		checkResponseStruct    bool
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
			name: "happy path gamm spot price",
			path: "/osmosis.gamm.v2.Query/SpotPrice",
			testSetup: func() {
				pk := ed25519.GenPrivKey().PubKey()
				sender := sdk.AccAddress(pk.Address())
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, sender, apptesting.DefaultAcctFunds)
				suite.Require().NoError(err)
				msg := balancer.NewMsgCreateBalancerPool(sender,
					balancer.NewPoolParams(osmomath.ZeroDec(), osmomath.ZeroDec(), nil),
					apptesting.DefaultPoolAssets, "")
				_, err = suite.app.PoolManagerKeeper.CreatePool(suite.ctx, msg)
				suite.NoError(err)
			},
			requestData: func() []byte {
				queryrequest := gammv2types.QuerySpotPriceRequest{ //nolint:staticcheck // we're intentionally using this deprecated package for testing
					PoolId:          1,
					BaseAssetDenom:  "bar",
					QuoteAssetDenom: appparams.BaseCoinUnit,
				}
				bz, err := proto.Marshal(&queryrequest)
				suite.Require().NoError(err)
				return bz
			},
			checkResponseStruct: true,
			responseProtoStruct: &gammv2types.QuerySpotPriceResponse{ //nolint:staticcheck // we're intentionally using this deprecated package for testing
				SpotPrice: osmomath.NewDecWithPrec(5, 1).String(),
			},
		},
		{
			name: "happy path pool manager",
			path: "/osmosis.poolmanager.v1beta1.Query/SpotPrice",
			testSetup: func() {
				pk := ed25519.GenPrivKey().PubKey()
				sender := sdk.AccAddress(pk.Address())
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, sender, apptesting.DefaultAcctFunds)
				suite.Require().NoError(err)
				msg := balancer.NewMsgCreateBalancerPool(sender,
					balancer.NewPoolParams(osmomath.ZeroDec(), osmomath.ZeroDec(), nil),
					apptesting.DefaultPoolAssets, "")
				_, err = suite.app.PoolManagerKeeper.CreatePool(suite.ctx, msg)
				suite.NoError(err)
			},
			requestData: func() []byte {
				queryrequest := gammv2types.QuerySpotPriceRequest{ //nolint:staticcheck // we're intentionally using this deprecated package for testing
					PoolId:          1,
					BaseAssetDenom:  "bar",
					QuoteAssetDenom: appparams.BaseCoinUnit,
				}
				bz, err := proto.Marshal(&queryrequest)
				suite.Require().NoError(err)
				return bz
			},
			checkResponseStruct: true,
			responseProtoStruct: &gammv2types.QuerySpotPriceResponse{ //nolint:staticcheck // we're intentionally using this deprecated package for testing
				SpotPrice: osmomath.NewDecWithPrec(5, 1).String(),
			},
		},
		{
			name: "unregistered path(not whitelisted)",
			path: "/osmosis.lockup.Query/AccountLockedLongerDuration",
			requestData: func() []byte {
				request := lockuptypes.AccountLockedLongerDurationRequest{}
				bz, err := proto.Marshal(&request)
				suite.Require().NoError(err)
				return bz
			},
			expectedQuerierError: true,
		},
		{
			name: "test query using iterator",
			testSetup: func() {
				accAddr, err := sdk.AccAddressFromBech32("osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44")
				suite.Require().NoError(err)

				// fund account to receive non-empty response
				err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, accAddr, sdk.Coins{sdk.NewCoin("stake", osmomath.NewInt(10))})
				suite.Require().NoError(err)

				wasmbinding.SetWhitelistedQuery("/cosmos.bank.v1beta1.Query/AllBalances", &banktypes.QueryAllBalancesResponse{})
			},
			path: "/cosmos.bank.v1beta1.Query/AllBalances",
			requestData: func() []byte {
				bankrequest := banktypes.QueryAllBalancesRequest{
					Address: "osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44",
				}
				bz, err := proto.Marshal(&bankrequest)
				suite.Require().NoError(err)
				return bz
			},
			responseProtoStruct: &banktypes.QueryAllBalancesResponse{},
		},
		{
			name: "edge case: resending request",
			testSetup: func() {
				accAddr, err := sdk.AccAddressFromBech32("osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44")
				suite.Require().NoError(err)

				// fund account to receive non-empty response
				err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, accAddr, sdk.Coins{sdk.NewCoin("stake", osmomath.NewInt(10))})
				suite.Require().NoError(err)

				wasmbinding.SetWhitelistedQuery("/cosmos.bank.v1beta1.Query/AllBalances", &banktypes.QueryAllBalancesResponse{})
			},
			path: "/cosmos.bank.v1beta1.Query/AllBalances",
			requestData: func() []byte {
				bankrequest := banktypes.QueryAllBalancesRequest{
					Address: "osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44",
				}
				bz, err := proto.Marshal(&bankrequest)
				suite.Require().NoError(err)
				return bz
			},
			responseProtoStruct: &banktypes.QueryAllBalancesResponse{},
			resendRequest:       true,
		},
		{
			name: "invalid query router route",
			testSetup: func() {
				wasmbinding.SetWhitelistedQuery("invalid/query/router/route", &epochtypes.QueryEpochsInfoRequest{})
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
				wasmbinding.SetWhitelistedQuery("/osmosis.epochs.v1beta1.Query/EpochInfos",
					&banktypes.QueryAllBalancesResponse{})
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
				wasmbinding.SetWhitelistedQuery("/cosmos.bank.v1beta1.Query/AllBalances", &banktypes.QueryAllBalancesRequest{})
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
			suite.SetupTestInternal()
			defer suite.TearDownTestInternal()
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
			}
			if tc.checkResponseStruct {
				expectedResponse, err := proto.Marshal(tc.responseProtoStruct)
				suite.Require().NoError(err)
				expJsonResp, err := wasmbinding.ConvertProtoToJSONMarshal(tc.responseProtoStruct, expectedResponse, suite.app.AppCodec())
				suite.Require().NoError(err)
				suite.Require().Equal(expJsonResp, stargateResponse)
			}

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

			if tc.resendRequest {
				stargateQuerier = wasmbinding.StargateQuerier(*suite.app.GRPCQueryRouter(), suite.app.AppCodec())
				stargateRequest = &wasmvmtypes.StargateQuery{
					Path: tc.path,
					Data: tc.requestData(),
				}
				resendResponse, err := stargateQuerier(suite.ctx, stargateRequest)
				suite.Require().NoError(err)
				suite.Require().Equal(stargateResponse, resendResponse)
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
				Balances: sdk.NewCoins(sdk.NewCoin("bar", osmomath.NewInt(30))),
				Pagination: &query.PageResponse{
					NextKey: []byte("foo"),
				},
			},
		},
		{
			name:                "invalid proto response struct",
			queryPath:           "/cosmos.bank.v1beta1.Query/AllBalances",
			originalResponse:    "0a090a036261721202333012050a03666f6f",
			protoResponseStruct: &epochtypes.QueryCurrentEpochResponse{},
			expectedError:       true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTestInternal()
			defer suite.TearDownTestInternal()

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
		testSetup           func()
		originalResponsebz  []byte
		updatedResponsebz   []byte
		queryPath           string
		responseProtoStruct interface{}
		expectedProto       func() proto.Message
	}{
		/**
		   * Origin Response
		   * balances:<denom:"bar" amount:"30" > pagination:<next_key:"foo" >
		   * New Version Response
		   * The binary built from the proto response with additional field address
		   * balances:<denom:"bar" amount:"30" > pagination:<next_key:"foo" > address:"cosmos1j6j5tsquq2jlw2af7l3xekyaq7zg4l8jsufu78"
		   // Origin proto
		   message QueryAllBalancesResponse {
		  	// balances is the balances of all the coins.
		  	repeated cosmos.base.v1beta1.Coin balances = 1
		  	[(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
		  	// pagination defines the pagination in the response.
		  	cosmos.base.query.v1beta1.PageResponse pagination = 2;
		  }
		  // Updated proto
		  message QueryAllBalancesResponse {
		  	// balances is the balances of all the coins.
		  	repeated cosmos.base.v1beta1.Coin balances = 1
		  	[(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
		  	// pagination defines the pagination in the response.
		  	cosmos.base.query.v1beta1.PageResponse pagination = 2;
		  	// address is the address to query all balances for.
		  	string address = 3;
		  }
		*/
		{
			"Query All Balances",
			func() {
				wasmbinding.SetWhitelistedQuery("/cosmos.bank.v1beta1.Query/AllBalances", &banktypes.QueryAllBalancesResponse{})
			},
			[]byte{10, 9, 10, 3, 98, 97, 114, 18, 2, 51, 48, 18, 5, 10, 3, 102, 111, 111},
			[]byte{
				10, 9, 10, 3, 98, 97, 114, 18, 2, 51, 48, 18, 5, 10, 3, 102, 111, 111, 26, 45, 99, 111, 115, 109, 111, 115, 49, 106,
				54, 106, 53, 116, 115, 113, 117, 113, 50, 106, 108, 119, 50, 97, 102, 55, 108, 51, 120, 101, 107, 121, 97, 113, 55, 122, 103,
				52, 108, 56, 106, 115, 117, 102, 117, 55, 56,
			},
			"/cosmos.bank.v1beta1.Query/AllBalances",
			&banktypes.QueryAllBalancesResponse{},
			func() proto.Message {
				return &banktypes.QueryAllBalancesResponse{
					Balances: sdk.NewCoins(sdk.NewCoin("bar", osmomath.NewInt(30))),
					Pagination: &query.PageResponse{
						NextKey: []byte("foo"),
					},
				}
			},
		},
		/**
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
			nil,
			[]byte{
				10, 83, 10, 32, 47, 99, 111, 115, 109, 111, 115, 46, 97, 117, 116, 104, 46, 118, 49, 98, 101, 116, 97, 49, 46, 66, 97, 115,
				101, 65, 99, 99, 111, 117, 110, 116, 18, 47, 10, 45, 99, 111, 115, 109, 111, 115, 49, 102, 56, 117, 120, 117, 108, 116, 110, 56,
				115, 113, 122, 104, 122, 110, 114, 115, 122, 51, 113, 55, 55, 120, 119, 97, 113, 117, 104, 103, 114, 115, 103, 54, 106, 121, 118, 102, 121,
			},
			[]byte{
				10, 83, 10, 32, 47, 99, 111, 115, 109, 111, 115, 46, 97, 117, 116, 104, 46, 118, 49, 98, 101, 116, 97, 49, 46, 66, 97, 115,
				101, 65, 99, 99, 111, 117, 110, 116, 18, 47, 10, 45, 99, 111, 115, 109, 111, 115, 49, 102, 56, 117, 120, 117, 108, 116, 110, 56,
				115, 113, 122, 104, 122, 110, 114, 115, 122, 51, 113, 55, 55, 120, 119, 97, 113, 117, 104, 103, 114, 115, 103, 54, 106, 121, 118, 102, 121,
				18, 45, 99, 111, 115, 109, 111, 115, 49, 102, 56, 117, 120, 117, 108, 116, 110, 56, 115, 113, 122, 104, 122, 110, 114, 115, 122, 51, 113, 55,
				55, 120, 119, 97, 113, 117, 104, 103, 114, 115, 103, 54, 106, 121, 118, 102, 121,
			},
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
			suite.SetupTestInternal()
			defer suite.TearDownTestInternal()

			if tc.testSetup != nil {
				tc.testSetup()
			}

			binding, err := wasmbinding.GetWhitelistedQuery(tc.queryPath)
			suite.Require().Nil(err)

			suite.Require().NoError(err)
			jsonMarshalledOriginalBz, err := wasmbinding.ConvertProtoToJSONMarshal(binding, tc.originalResponsebz, suite.app.AppCodec())
			suite.Require().NoError(err)

			jsonMarshalledUpdatedBz, err := wasmbinding.ConvertProtoToJSONMarshal(binding, tc.updatedResponsebz, suite.app.AppCodec())
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
