package osmocli

import (
	"testing"

	"github.com/stretchr/testify/require"

	clqueryproto "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/client/queryproto"
	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
)

// test-specific helper descriptor
type TestDescriptor interface {
	Descriptor
	GetUse() string
}

type TestQueryDescriptor struct {
	*QueryDescriptor
}

func (tqd *TestQueryDescriptor) GetUse() string {
	return tqd.Use
}

type TestTxCliDescriptor struct {
	*TxCliDesc
}

func (ttxcd *TestTxCliDescriptor) GetUse() string {
	return ttxcd.Use
}

func TestAttachFieldsToUse(t *testing.T) {
	tests := map[string]struct {
		desc        TestDescriptor
		attachFunc  func(Descriptor)
		expectedUse string
	}{
		"basic/TxCliDesc/attaches_2_args": {
			desc: &TestTxCliDescriptor{
				&TxCliDesc{
					Use:   "set-reward-receiver-address",
					Short: "sets reward receiver address for the designated lock id",
					Long:  "sets reward receiver address for the designated lock id",
				},
			},
			attachFunc:  attachFieldsToUse[*lockuptypes.MsgSetRewardReceiverAddress],
			expectedUse: "set-reward-receiver-address [lock-id] [reward-receiver]",
		},
		"basic/QueryDescriptor/attaches_1_arg": {
			desc: &TestQueryDescriptor{
				&QueryDescriptor{
					Use:   "pool-accumulator-rewards",
					Short: "Query pool accumulator rewards",
					Long: `{{.Short}}{{.ExampleHeader}}
			{{.CommandPrefix}} pool-accumulator-rewards 1`,
				},
			},
			attachFunc:  attachFieldsToUse[*clqueryproto.PoolAccumulatorRewardsRequest],
			expectedUse: "pool-accumulator-rewards [pool-id]",
		},
		"ignore_pagination/QueryDescriptor/no_args": {
			desc: &TestQueryDescriptor{
				&QueryDescriptor{
					Use:   "pools",
					Short: "Query pools",
					Long: `{{.Short}}{{.ExampleHeader}}
			{{.CommandPrefix}} pools`,
				},
			},
			attachFunc:  attachFieldsToUse[*clqueryproto.PoolsRequest],
			expectedUse: "pools",
		},
		"ignore_owner/TxCliDesc/attach_2_args": {
			desc: &TestTxCliDescriptor{
				&TxCliDesc{
					Use:   "lock-tokens",
					Short: "lock tokens into lockup pool from user account",
				},
			},
			attachFunc:  attachFieldsToUse[*lockuptypes.MsgLockTokens],
			expectedUse: "lock-tokens [duration] [coins]", // in osmosis, this command takes duration from a flag, but here it is just for testing purposes
		},
		"ignore_sender/TxCliDesc/attach_5_args": { // also tests that args are shown in kebab-case
			desc: &TestTxCliDescriptor{
				&TxCliDesc{
					Use:     "add-to-position",
					Short:   "add to an existing concentrated liquidity position",
					Example: "osmosisd tx concentratedliquidity add-to-position 10 1000000000uosmo 10000000uion --from val --chain-id localosmosis -b block --keyring-backend test --fees 1000000uosmo",
				},
			},
			attachFunc:  attachFieldsToUse[*cltypes.MsgAddToPosition],
			expectedUse: "add-to-position [position-id] [amount0] [amount1] [token-min-amount0] [token-min-amount1]",
		},
		"ignore_custom_flag_overrides/TxCliDesc/": {
			desc: &TestTxCliDescriptor{
				&TxCliDesc{
					Use:   "lock-tokens",
					Short: "lock tokens into lockup pool from user account",
					CustomFlagOverrides: map[string]string{
						"duration": "duration",
					},
				},
			},
			attachFunc:  attachFieldsToUse[*lockuptypes.MsgLockTokens],
			expectedUse: "lock-tokens [coins]",
		},
	}

	for _, tt := range tests {
		tt.attachFunc(tt.desc)
		require.Equal(t, tt.desc.GetUse(), tt.expectedUse)
	}
}
