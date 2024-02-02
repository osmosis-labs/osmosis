package authenticator_test

import (
	"github.com/osmosis-labs/osmosis/v21/app/params"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/app"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

type MessageFilterAuthenticatorTest struct {
	suite.Suite
	Ctx            sdk.Context
	OsmosisApp     *app.OsmosisApp
	EncodingConfig params.EncodingConfig

	MessageFilterAuthenticator authenticator.MessageFilterAuthenticator
}

func TestMessageFilterAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(MessageFilterAuthenticatorTest))
}

func (s *MessageFilterAuthenticatorTest) SetupTest() {
	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))
	s.EncodingConfig = app.MakeEncodingConfig()
	s.MessageFilterAuthenticator = authenticator.NewMessageFilterAuthenticator()
}

func (s *MessageFilterAuthenticatorTest) TestBankSend() {
	tests := []struct {
		name           string // name
		pattern        string
		msg            sdk.Msg
		passvalidation bool
		match          bool
	}{
		{"bank send",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo", "amount": "100"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. no amount",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to"}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. bad sender",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"someoneElse","to_address":"to"}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			false,
		},

		{"bank send. any",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. bad amount",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo", "amount": "50"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			false,
		},

		{"bank send. amount as number",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo", "amount": 100}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			false,
			false, // This fails because of floats. Should be prevented by validation
		},

		{"bank send. amount as mix string number",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo", "amount": "100"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. amount as mix string number but bad",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo", "amount": "50"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			false,
		},

		{"bank send. just denom",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. just denom",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))},
			true,
			true,
		},

		{"bank send. bad denom",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("bar", 100))},
			true,
			false,
		},

		{"bank send. any match",
			`{}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("bar", 100))},
			true,
			true,
		},

		{"bank send. using map as generic",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("bar", 100))},
			true,
			true,
		},

		{"bank send. empty array as generic for arrays",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": []}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))},
			true,
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.MessageFilterAuthenticator.OnAuthenticatorAdded(s.Ctx, sdk.AccAddress{}, []byte(tt.pattern))
			if tt.passvalidation {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}
			filter, err := s.MessageFilterAuthenticator.Initialize([]byte(tt.pattern))
			s.Require().NoError(err)

			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			tx := GenEmptyTx()
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, nil, tt.msg, tx, 0, false)
			s.Require().NoError(err)

			result := filter.Authenticate(s.Ctx, request)
			if tt.match {
				s.Require().True(result.IsAuthenticated())
			} else {
				s.Require().True(result.IsAuthenticationFailed())
			}
		})
	}
}

func (s *MessageFilterAuthenticatorTest) TestPoolManagerSwapExactAmountIn() {
	tests := []struct {
		name    string
		pattern string
		msg     sdk.Msg
		match   bool
	}{
		{"poolmanager swap exact amount in",
			`{"type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","value":{"sender":"senderAddr","token_in":{"denom":"inputDenom", "amount":"500"}, "token_out_min_amount": "100"}}`,
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            "senderAddr",
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},

		{"swap exact amount. basic match",
			`{"type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","value":{"sender":"senderAddr"}}`,
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            "senderAddr",
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},

		{"swap exact amount. match denom",
			`{"type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","value":{"token_in":{"denom":"inputDenom"}}}`,
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            "senderAddr",
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},

		{"swap exact amount. match denom and sender",
			`{"type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","value":{"token_in":{"denom":"inputDenom"}, "sender": "senderAddr"}}`,
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            "senderAddr",
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},

		{"swap exact amount. mismatch denom",
			`{"type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","value":{"token_in":{"denom":"wrongDenom"}}}`,
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            "senderAddr",
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			false,
		},

		{"swap exact amount. match with token out min amount",
			`{"type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","value":{"token_out_min_amount":"100"}}`,
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            "senderAddr",
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},

		{"swap exact amount. mismatch with token out min amount",
			`{"type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","value":{"token_out_min_amount":"200"}}`,
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            "senderAddr",
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			filter, err := s.MessageFilterAuthenticator.Initialize([]byte(tt.pattern))
			s.Require().NoError(err)

			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			var tx sdk.Tx
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, nil, tt.msg, tx, 0, false)
			s.Require().NoError(err)

			result := filter.Authenticate(s.Ctx, request)
			if tt.match {
				s.Require().True(result.IsAuthenticated())
			} else {
				s.Require().True(result.IsAuthenticationFailed())
			}
		})
	}
}
