package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAggregateVoteHash(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
	}

	aggregateVoteHash := types.GetAggregateVoteHash("salt", "100ukrw,200uusd", sdk.ValAddress(addrs[0]))
	hexStr := hex.EncodeToString(aggregateVoteHash)
	aggregateVoteHashRes, err := types.AggregateVoteHashFromHexString(hexStr)
	require.NoError(t, err)
	require.Equal(t, aggregateVoteHash, aggregateVoteHashRes)
	require.Equal(t, aggregateVoteHash, aggregateVoteHash)
	require.True(t, types.AggregateVoteHash([]byte{}).Empty())

	got, _ := yaml.Marshal(&aggregateVoteHash)
	require.Equal(t, aggregateVoteHash.String()+"\n", string(got))

	res := types.AggregateVoteHash{}
	testMarshal(t, &aggregateVoteHash, &res, aggregateVoteHash.MarshalJSON, (&res).UnmarshalJSON)
	testMarshal(t, &aggregateVoteHash, &res, aggregateVoteHash.Marshal, (&res).Unmarshal)
}

func testMarshal(t *testing.T, original interface{}, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	require.Nil(t, err)
	err = unmarshal(bz)
	require.Nil(t, err)
	require.Equal(t, original, res)
}
