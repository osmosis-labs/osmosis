package wasmbinding_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v10/app"
	"github.com/osmosis-labs/osmosis/v10/wasmbinding"
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

func (suite *StargateTestSuite) TestDeterministicJsonMarshal() {
	testCases := []struct {
		name             string
		originalResponse string
		updatedResponse  string
		queryPath        string
		expectedProto    proto.Message
	}{

		/**
		   * Origin Response
		   * balances:<denom:"bar" amount:"30" > pagination:<next_key:"foo" >
		   * "0a090a036261721202333012050a03666f6f"
		   *
		   * New Version Response
		   * The binary built from the proto response with additional field address
		   * balances:<denom:"bar" amount:"30" > pagination:<next_key:"foo" > address:"cosmos1j6j5tsquq2jlw2af7l3xekyaq7zg4l8jsufu78"
		   * "0a090a036261721202333012050a03666f6f1a2d636f736d6f73316a366a357473717571326a6c77326166376c3378656b796171377a67346c386a737566753738"
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
			"0a090a036261721202333012050a03666f6f",
			"0a090a036261721202333012050a03666f6f1a2d636f736d6f73316a366a357473717571326a6c77326166376c3378656b796171377a67346c386a737566753738",
			"/cosmos.bank.v1beta1.Query/AllBalances",
			&banktypes.QueryAllBalancesResponse{
				Balances: sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(30))),
				Pagination: &query.PageResponse{
					NextKey: []byte("foo"),
				},
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
			jsonMarshalledOriginalBz, err := wasmbinding.NormalizeReponseAndJsonMarshal(binding, originVersionBz, suite.app.AppCodec())
			suite.Require().NoError(err)

			newVersionBz, err := hex.DecodeString(tc.updatedResponse)
			suite.Require().NoError(err)
			jsonMarshalledUpdatedBz, err := wasmbinding.NormalizeReponseAndJsonMarshal(binding, newVersionBz, suite.app.AppCodec())
			suite.Require().NoError(err)

			// json marshalled bytes should be the same since we use the same proto sturct for unmarshalling
			suite.Require().Equal(jsonMarshalledOriginalBz, jsonMarshalledUpdatedBz)

			// raw build also make same result
			jsonMarshalExpectedResponse, err := suite.app.AppCodec().MarshalJSON(tc.expectedProto)
			suite.Require().NoError(err)
			suite.Require().Equal(jsonMarshalledUpdatedBz, jsonMarshalExpectedResponse)
		})
	}
}
