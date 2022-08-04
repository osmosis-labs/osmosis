package wasmbinding_test

import (
	"encoding/hex"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v10/wasmbinding"
)

/**
 * Origin Response
 * balances:<denom:"bar" amount:"30" > pagination:<next_key:"foo" >
 * "0a090a036261721202333012050a03666f6f"
 *
 * New Version Response
 * The binary built from the proto response with additional field address
 * balances:<denom:"bar" amount:"30" > pagination:<next_key:"foo" > address:"cosmos1j6j5tsquq2jlw2af7l3xekyaq7zg4l8jsufu78"
 * "0a090a036261721202333012050a03666f6f1a2d636f736d6f73316a366a357473717571326a6c77326166376c3378656b796171377a67346c386a737566753738"
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
// Origin proto
message QueryAllBalancesResponse {
	// balances is the balances of all the coins.
	repeated cosmos.base.v1beta1.Coin balances = 1
	[(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
	// pagination defines the pagination in the response.
	cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
*/

func TestDeterministic_AllBalances(t *testing.T) {
	originVersionBz, err := hex.DecodeString("0a090a036261721202333012050a03666f6f")
	require.NoError(t, err)

	newVersionBz, err := hex.DecodeString("0a090a036261721202333012050a03666f6f1a2d636f736d6f73316a366a357473717571326a6c77326166376c3378656b796171377a67346c386a737566753738")
	require.NoError(t, err)

	binding, ok := wasmbinding.StargateWhitelist.Load("/cosmos.bank.v1beta1.Query/AllBalances")
	require.True(t, ok)

	// new version response should be changed into origin version response
	normalizedBz, err := wasmbinding.NormalizeReponse(binding, newVersionBz)
	require.NoError(t, err)

	require.Equal(t, originVersionBz, normalizedBz)
	require.NotEqual(t, newVersionBz, normalizedBz)

	// raw build also make same result
	expectedResponse := banktypes.QueryAllBalancesResponse{
		Balances: sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(30))),
		Pagination: &query.PageResponse{
			NextKey: []byte("foo"),
		},
	}
	expectedResponseBz, err := proto.Marshal(&expectedResponse)
	require.NoError(t, err)
	require.Equal(t, expectedResponseBz, normalizedBz)

	// should be cleared
	data := binding.(*banktypes.QueryAllBalancesResponse)
	require.Empty(t, data.Balances)
	require.Empty(t, data.Pagination)
}

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

func TestDeterministic_Account(t *testing.T) {
	originVersionBz, err := hex.DecodeString("0a530a202f636f736d6f732e617574682e763162657461312e426173654163636f756e74122f0a2d636f736d6f733166387578756c746e3873717a687a6e72737a3371373778776171756867727367366a79766679")
	require.NoError(t, err)

	newVersionBz, err := hex.DecodeString("0a530a202f636f736d6f732e617574682e763162657461312e426173654163636f756e74122f0a2d636f736d6f733166387578756c746e3873717a687a6e72737a3371373778776171756867727367366a79766679122d636f736d6f733166387578756c746e3873717a687a6e72737a3371373778776171756867727367366a79766679")
	require.NoError(t, err)

	binding, ok := wasmbinding.StargateWhitelist.Load("/cosmos.auth.v1beta1.Query/Account")
	require.True(t, ok)

	// new version response should be changed into origin version response
	normalizedBz, err := wasmbinding.NormalizeReponse(binding, newVersionBz)
	require.NoError(t, err)

	require.Equal(t, originVersionBz, normalizedBz)
	require.NotEqual(t, newVersionBz, normalizedBz)

	// should be cleared
	data := binding.(*authtypes.QueryAccountResponse)
	require.Empty(t, data.Account)
}
