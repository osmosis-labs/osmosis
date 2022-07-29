package types

import (
	fmt "fmt"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appParams "github.com/osmosis-labs/osmosis/v10/app/params"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// // Test authz serialize and de-serializes for superfluid msg.
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
		msg                        sdk.Msg
	}{
		{
			name: "MsgLockAndSuperfluidDelegate",
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
										"msg":"/osmosis.superfluid.MsgLockAndSuperfluidDelegate"
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
							"msg_type_url":"/osmosis.superfluid.MsgLockAndSuperfluidDelegate"
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
									"type":"osmosis/lock-and-superfluid-delegate",
									"value":{
										"coins":[
											{
												"amount":"1",
												"denom":"stake"
											}
										],
										"sender":"%s",
										"val_addr":"valoper1xyz"
									}
								}
							]
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, addr1),
			msg: &MsgLockAndSuperfluidDelegate{
				Sender:  addr1,
				Coins:   sdk.NewCoins(coin),
				ValAddr: "valoper1xyz",
			},
		},
		{
			name: "MsgSuperfluidDelegate",
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
										"msg":"/osmosis.superfluid.MsgSuperfluidDelegate"
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
							"msg_type_url":"/osmosis.superfluid.MsgSuperfluidDelegate"
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
									"type":"osmosis/superfluid-delegate",
									"value":{
										"lock_id":"1",
										"sender":"%s",
										"val_addr":"valoper1xyz"
									}
								}
							]
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, addr1),
			msg: &MsgSuperfluidDelegate{
				Sender:  addr1,
				LockId:  1,
				ValAddr: "valoper1xyz",
			},
		},
		{
			name: "MsgSuperfluidUnbondLock",
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
										"msg":"/osmosis.superfluid.MsgSuperfluidUnbondLock"
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
							"msg_type_url":"/osmosis.superfluid.MsgSuperfluidUnbondLock"
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
									"type":"osmosis/superfluid-unbond-lock",
									"value":{
										"lock_id":"1",
										"sender":"%s"
									}
								}
							]
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, addr1),
			msg: &MsgSuperfluidUnbondLock{
				Sender: addr1,
				LockId: 1,
			},
		},
		{
			name: "MsgSuperfluidUndelegate",
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
										"msg":"/osmosis.superfluid.MsgSuperfluidUndelegate"
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
							"msg_type_url":"/osmosis.superfluid.MsgSuperfluidUndelegate"
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
									"type":"osmosis/superfluid-undelegate",
									"value":{
										"lock_id":"1",
										"sender":"%s"
									}
								}
							]
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, addr1),
			msg: &MsgSuperfluidUndelegate{
				Sender: addr1,
				LockId: 1,
			},
		},
		{
			name: "MsgUnPoolWhitelistedPool",
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
										"msg":"/osmosis.superfluid.MsgUnPoolWhitelistedPool"
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
							"msg_type_url":"/osmosis.superfluid.MsgUnPoolWhitelistedPool"
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
									"type":"osmosis/unpool-whitelisted-pool",
									"value":{
										"pool_id":"1",
										"sender":"%s"
									}
								}
							]
						}
					}
				],
				"sequence":"1",
				"timeout_height":"1"
			}`, mockGrantee, addr1),
			msg: &MsgUnPoolWhitelistedPool{
				Sender: addr1,
				PoolId: 1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Authz: Grant Msg
			typeURL := sdk.MsgTypeURL(tc.msg)
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
			msgAny, _ := cdctypes.NewAnyWithValue(tc.msg)
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
