package types

import (
	"fmt"
	"strings"
	"testing"
	time "time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	appParams "github.com/osmosis-labs/osmosis/v10/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// TestMsgCreatePool tests if valid/invalid create pool messages are properly validated/invalidated
func TestMsgCreatePool(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper createPool message
	createMsg := func(after func(msg MsgCreateGauge) MsgCreateGauge) MsgCreateGauge {
		distributeTo := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		}

		properMsg := *NewMsgCreateGauge(
			false,
			addr1,
			distributeTo,
			sdk.Coins{},
			time.Now(),
			2,
		)

		return after(properMsg)
	}

	// validate createPool message was created as intended
	msg := createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
		return msg
	})
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "create_gauge")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        MsgCreateGauge
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.Owner = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty distribution denom",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.DistributeTo.Denom = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid distribution denom",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.DistributeTo.Denom = "111"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lock query type",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.DistributeTo.LockQueryType = -1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lock query type",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.DistributeTo.LockQueryType = -1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid distribution start time",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.StartTime = time.Time{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.NumEpochsPaidOver = 0
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over for perpetual gauge",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.NumEpochsPaidOver = 2
				msg.IsPerpetual = true
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid num epochs paid over for perpetual gauge",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.NumEpochsPaidOver = 1
				msg.IsPerpetual = true
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

// TestMsgAddToGauge tests if valid/invalid add to gauge messages are properly validated/invalidated
func TestMsgAddToGauge(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper addToGauge message
	createMsg := func(after func(msg MsgAddToGauge) MsgAddToGauge) MsgAddToGauge {
		properMsg := *NewMsgAddToGauge(
			addr1,
			1,
			sdk.Coins{sdk.NewInt64Coin("stake", 10)},
		)

		return after(properMsg)
	}

	// validate addToGauge message was created as intended
	msg := createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
		return msg
	})
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "add_to_gauge")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        MsgAddToGauge
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
				msg.Owner = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty rewards",
			msg: createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
				msg.Rewards = sdk.Coins{}
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

// // Test authz serialize and de-serializes for incentives msg.
func TestAuthzMsg(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("stake", sdk.NewInt(1))
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)

	const (
		mockGranter string = "cosmos1abc"
		mockGrantee string = "cosmos1xyz"
	)

	testCases := []struct {
		name                       string
		expectedGrantSignByteMsg   string
		expectedRevokeSignByteMsg  string
		expectedExecStrSignByteMsg string
		incentivesMsg              sdk.Msg
	}{
		{
			name: "MsgAddToGauge",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
				   "amount":[
					  
				   ],
				   "gas":"0"
				},
				"memo":"memo",
				"msgs":[
				   {
					  "type":"cosmos-sdk/MsgGrant",
					  "value":{
						 "grant":{
							"authorization":{
							   "type":"cosmos-sdk/GenericAuthorization",
							   "value":{
								  "msg":"/osmosis.incentives.MsgAddToGauge"
							   }
							},
							"expiration":"0001-01-01T02:01:01.000000001Z"
						 },
						 "grantee":"%s",
						 "granter":"%s"
					  }
				   }
				],
				"sequence":"1",
				"timeout_height":"1"
			 }`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
					"amount":[
						
					],
					"gas":"0"
				},
				"memo":"memo",
				"msgs":[
					{
						"type":"cosmos-sdk/MsgRevoke",
						"value":{
							"grantee":"%s",
							"granter":"%s",
							"msg_type_url":"/osmosis.incentives.MsgAddToGauge"
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
					"amount":[
						
					],
					"gas":"0"
				},
				"memo":"memo",
				"msgs":[
					{
						"type":"cosmos-sdk/MsgExec",
						"value":{
							"grantee":"%s",
							"msgs":[
								{
									"type":"osmosis/incentives/add-to-gauge",
									"value":{
										"gauge_id":"1",
										"owner":"%s",
										"rewards":[
											{
												"amount":"1",
												"denom":"stake"
											}
										]
									}
								}
							]
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, addr1),
			incentivesMsg: &MsgAddToGauge{
				Owner:   addr1,
				GaugeId: 1,
				Rewards: sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgCreateGauge",
			expectedGrantSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
					"amount":[
						
					],
					"gas":"0"
				},
				"memo":"memo",
				"msgs":[
					{
						"type":"cosmos-sdk/MsgGrant",
						"value":{
							"grant":{
								"authorization":{
									"type":"cosmos-sdk/GenericAuthorization",
									"value":{
										"msg":"/osmosis.incentives.MsgCreateGauge"
									}
								},
								"expiration":"0001-01-01T02:01:01.000000001Z"
							},
							"grantee":"%s",
							"granter":"%s"
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, mockGranter),
			expectedRevokeSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
					"amount":[
						
					],
					"gas":"0"
				},
				"memo":"memo",
				"msgs":[
					{
						"type":"cosmos-sdk/MsgRevoke",
						"value":{
							"grantee":"%s",
							"granter":"%s",
							"msg_type_url":"/osmosis.incentives.MsgCreateGauge"
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, mockGranter),
			expectedExecStrSignByteMsg: fmt.Sprintf(`{
				"account_number":"1",
				"chain_id":"foo",
				"fee":{
					"amount":[
						
					],
					"gas":"0"
				},
				"memo":"memo",
				"msgs":[
					{
						"type":"cosmos-sdk/MsgExec",
						"value":{
							"grantee":"%s",
							"msgs":[
								{
									"type":"osmosis/incentives/create-gauge",
									"value":{
										"coins":[
											{
												"amount":"1",
												"denom":"stake"
											}
										],
										"distribute_to":{
											"denom":"lptoken",
											"duration":"1000000000",
											"timestamp":"0001-01-01T00:00:00Z"
										},
										"num_epochs_paid_over":"1",
										"owner":"%s",
										"start_time":"0001-01-01T01:01:01.000000001Z"
									}
								}
							]
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, addr1),
			incentivesMsg: &MsgCreateGauge{
				IsPerpetual: false,
				Owner:       addr1,
				DistributeTo: lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         "lptoken",
					Duration:      time.Second,
				},
				Coins:             sdk.NewCoins(coin),
				StartTime:         someDate,
				NumEpochsPaidOver: 1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Authz: Grant Msg
			typeURL := sdk.MsgTypeURL(tc.incentivesMsg)
			grant, err := authz.NewGrant(someDate, authz.NewGenericAuthorization(typeURL), someDate.Add(time.Hour))
			require.NoError(t, err)
			msgGrant := &authz.MsgGrant{Granter: mockGranter, Grantee: mockGrantee, Grant: grant}

			require.Equal(t,
				formatJsonStr(tc.expectedGrantSignByteMsg),
				string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgGrant}, "memo")),
			)

			// Authz: Revoke Msg
			msgRevoke := &authz.MsgRevoke{Granter: mockGranter, Grantee: mockGrantee, MsgTypeUrl: typeURL}

			require.Equal(t,
				formatJsonStr(tc.expectedRevokeSignByteMsg),
				string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgRevoke}, "memo")),
			)

			// Authz: Exec Msg
			msgAny, _ := cdctypes.NewAnyWithValue(tc.incentivesMsg)
			msgExec := &authz.MsgExec{Grantee: mockGrantee, Msgs: []*cdctypes.Any{msgAny}}

			require.Equal(t,
				formatJsonStr(tc.expectedExecStrSignByteMsg),
				string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgExec}, "memo")),
			)
		})
	}
}

func formatJsonStr(jsonStrMsg string) string {
	ans := strings.ReplaceAll(jsonStrMsg, "\n", "")
	ans = strings.ReplaceAll(ans, "\t", "")
	ans = strings.ReplaceAll(ans, " ", "")

	return ans
}
