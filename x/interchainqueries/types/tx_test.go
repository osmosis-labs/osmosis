package types_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/stretchr/testify/require"

	iqtypes "github.com/osmosis-labs/osmosis/v20/x/interchainqueries/types"
)

const TestAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

func TestMsgRegisterInterchainQueryValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"invalid query type",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          "invalid_type",
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidQueryType,
		},
		{
			"invalid transactions filter format",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "&)(^Y(*&(*&(&(*",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidTransactionsFilter,
		},
		{
			"too many keys",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "[]",
					Keys:               craftKVKeys(200),
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrTooManyKVQueryKeys,
		},
		{
			"invalid update period",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       0,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidUpdatePeriod,
		},
		{
			"empty sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             "",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             "cosmos14234_invalid_address",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty connection id",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidConnectionID,
		},
		{
			"empty keys",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrEmptyKeys,
		},
		{
			"empty key path",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: ""}},
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrEmptyKeyPath,
		},
		{
			"empty key id",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               []*iqtypes.KVKey{{Key: []byte(""), Path: "path"}},
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrEmptyKeyID,
		},
		{
			"nil key",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: "path1"}, nil},
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			sdkerrors.ErrInvalidType,
		},
		{
			"duplicated keys",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: "path1"}, {Key: []byte("key1"), Path: "path1"}},
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			sdkerrors.ErrInvalidRequest,
		},
		{
			"valid",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               []*iqtypes.KVKey{{Key: []byte("key1"), Path: "path1"}},
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			nil,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

func TestMsgSubmitQueryResultValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"valid",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key: []byte{10},
							Proof: &crypto.ProofOps{Ops: []crypto.ProofOp{
								{
									Type: "type",
									Key:  []byte{10},
									Data: []byte{10},
								},
							}},
							Value:         []byte{10},
							StoragePrefix: host.KeyConnectionPrefix,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			nil,
		},
		{
			"empty result",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
					Result:   nil,
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"empty kv results and block result",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
					Result: &iqtypes.QueryResult{
						KvResults: nil,
						Block:     nil,
						Height:    100,
						Revision:  1,
					},
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"zero query id",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  0,
					Sender:   TestAddress,
					ClientId: "client-id",
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key: []byte{10},
							Proof: &crypto.ProofOps{Ops: []crypto.ProofOp{
								{
									Type: "type",
									Key:  []byte{10},
									Data: []byte{10},
								},
							}},
							Value:         []byte{10},
							StoragePrefix: ibcexported.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"empty sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   "",
					ClientId: "client-id",
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key: []byte{10},
							Proof: &crypto.ProofOps{Ops: []crypto.ProofOp{
								{
									Type: "type",
									Key:  []byte{10},
									Data: []byte{10},
								},
							}},
							Value:         []byte{10},
							StoragePrefix: ibcexported.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   "invalid_sender",
					ClientId: "client-id",
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key: []byte{10},
							Proof: &crypto.ProofOps{Ops: []crypto.ProofOp{
								{
									Type: "type",
									Key:  []byte{10},
									Data: []byte{10},
								},
							}},
							Value:         []byte{10},
							StoragePrefix: ibcexported.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty client id",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "",
					Result: &iqtypes.QueryResult{
						KvResults: nil,
						Block: &iqtypes.Block{
							NextBlockHeader: nil,
							Header:          nil,
							Tx:              nil,
						},
						Height:   100,
						Revision: 1,
					},
				}
			},
			iqtypes.ErrInvalidClientID,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

func TestMsgUpdateQueryRequestValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"valid kv",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 1,
					NewKeys: []*iqtypes.KVKey{{
						Path: "staking",
						Key:  []byte{1, 2, 3},
					}},
					NewUpdatePeriod: 10,
					Sender:          TestAddress,
				}
			},
			nil,
		},
		{
			"valid tx",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:               1,
					NewUpdatePeriod:       10,
					NewTransactionsFilter: `[{"field":"transfer.recipient","op":"eq","value":"cosmos1xxx"}]`,
					Sender:                TestAddress,
				}
			},
			nil,
		},
		{
			"both keys and filter sent",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 1,
					NewKeys: []*iqtypes.KVKey{{
						Path: "staking",
						Key:  []byte{1, 2, 3},
					}},
					NewUpdatePeriod:       0,
					NewTransactionsFilter: `{"field":"transfer.recipient","op":"eq","value":"cosmos1xxx"}`,
					Sender:                TestAddress,
				}
			},
			sdkerrors.ErrInvalidRequest,
		},
		{
			"empty keys, update_period and tx filter",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:               1,
					NewKeys:               nil,
					NewUpdatePeriod:       0,
					NewTransactionsFilter: "",
					Sender:                TestAddress,
				}
			},
			sdkerrors.ErrInvalidRequest,
		},
		{
			"empty key path",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:         1,
					NewKeys:         []*iqtypes.KVKey{{Key: []byte("key1"), Path: ""}},
					NewUpdatePeriod: 0,
					Sender:          TestAddress,
				}
			},
			iqtypes.ErrEmptyKeyPath,
		},
		{
			"empty key id",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:         1,
					NewKeys:         []*iqtypes.KVKey{{Key: []byte(""), Path: "path"}},
					NewUpdatePeriod: 0,
					Sender:          TestAddress,
				}
			},
			iqtypes.ErrEmptyKeyID,
		},
		{
			"too many keys",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId:         1,
					NewKeys:         craftKVKeys(200),
					NewUpdatePeriod: 0,
					Sender:          TestAddress,
				}
			},
			iqtypes.ErrTooManyKVQueryKeys,
		},
		{
			"invalid query id",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 0,
					NewKeys: []*iqtypes.KVKey{{
						Path: "staking",
						Key:  []byte{1, 2, 3},
					}},
					NewUpdatePeriod: 10,
					Sender:          TestAddress,
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"empty sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 1,
					NewKeys: []*iqtypes.KVKey{{
						Path: "staking",
						Key:  []byte{1, 2, 3},
					}},
					NewUpdatePeriod: 10,
					Sender:          "",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					QueryId: 1,
					NewKeys: []*iqtypes.KVKey{{
						Path: "staking",
						Key:  []byte{1, 2, 3},
					}},
					NewUpdatePeriod: 10,
					Sender:          "invalid-sender",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIsf(t, msg.ValidateBasic(), tt.expectedErr, tt.name)
		} else {
			require.NoErrorf(t, msg.ValidateBasic(), tt.name)
		}
	}
}

func TestMsgRemoveQueryRequestValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"valid",
			func() sdktypes.Msg {
				return &iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 1,
					Sender:  TestAddress,
				}
			},
			nil,
		},
		{
			"invalid query id",
			func() sdktypes.Msg {
				return &iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 0,
					Sender:  TestAddress,
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"empty sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 1,
					Sender:  "",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgRemoveInterchainQueryRequest{
					QueryId: 1,
					Sender:  "invalid-sender",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

func TestMsgRegisterInterchainQueryGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.Msg
	}{
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}

func TestMsgSubmitQueryResultGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.Msg
	}{
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key: []byte{10},
							Proof: &crypto.ProofOps{Ops: []crypto.ProofOp{
								{
									Type: "type",
									Key:  []byte{10},
									Data: []byte{10},
								},
							}},
							Value:         []byte{10},
							StoragePrefix: ibcexported.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}

func TestMsgUpdateQueryGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.Msg
	}{
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					Sender: TestAddress,
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}

func TestMsgRemoveQueryGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.Msg
	}{
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &iqtypes.MsgRemoveInterchainQueryRequest{
					Sender: TestAddress,
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}

func craftKVKeys(n uint64) []*iqtypes.KVKey {
	keys := make([]*iqtypes.KVKey, n)
	for i := uint64(0); i < n; i++ {
		keys[i] = &iqtypes.KVKey{
			Path: "path-" + strconv.FormatUint(i, 10),
			Key:  []byte(fmt.Sprintf("key-%d", i)),
		}
	}

	return keys
}
