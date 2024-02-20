package authenticator_test

import (
	"fmt"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/osmosis-labs/osmosis/v21/app"
	"github.com/osmosis-labs/osmosis/v21/app/params"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"

	"github.com/stretchr/testify/suite"
)

type MessageFilterAuthenticatorTest struct {
	BaseAuthenticatorSuite

	MessageFilterAuthenticator authenticator.MessageFilterAuthenticator
	EncodingConfig             params.EncodingConfig
}

func TestMessageFilterAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(MessageFilterAuthenticatorTest))
}

func (s *MessageFilterAuthenticatorTest) SetupTest() {
	s.SetupKeys()
	s.EncodingConfig = app.MakeEncodingConfig()
	s.MessageFilterAuthenticator = authenticator.NewMessageFilterAuthenticator(s.EncodingConfig)
}

// TestBankSend tests the MessageFilterAuthenticator with multiple bank send messages
func (s *MessageFilterAuthenticatorTest) TestBankSend() {
	fromAddr := s.TestAccAddress[0].String()
	tests := []struct {
		name           string // name
		pattern        string
		msg            sdk.Msg
		passvalidation bool
		match          bool
	}{
		{"bank send",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend", "from_address":"%s","to_address":"to", "amount": [{"denom": "foo", "amount": "100"}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. no amount",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to"}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. bad sender",
			`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"someoneElse","to_address":"to"}`,
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			false,
		},

		{"bank send. any",
			`{"@type":"/cosmos.bank.v1beta1.MsgSend"}`,
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. bad amount",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": [{"denom": "foo", "amount": "50"}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			false,
		},

		{"bank send. amount as number",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": [{"denom": "foo", "amount": 100}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			false,
			false, // This fails because of floats. Should be prevented by validation
		},

		{"bank send. amount as mix string number",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": [{"denom": "foo", "amount": "100"}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. amount as mix string number but bad",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": [{"denom": "foo", "amount": "50"}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			false,
		},

		{"bank send. just denom",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": [{"denom": "foo"}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},

		{"bank send. just denom",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": [{"denom": "foo"}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))},
			true,
			true,
		},

		{"bank send. bad denom",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": [{"denom": "foo"}]}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("bar", 100))},
			true,
			false,
		},

		{"bank send. any match",
			`{"@type":"/cosmos.bank.v1beta1.MsgSend"}`,
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("bar", 100))},
			true,
			true,
		},

		{"bank send. using map as generic",
			`{"@type":"/cosmos.bank.v1beta1.MsgSend"}`,
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("bar", 100))},
			true,
			true,
		},

		{"bank send. empty array as generic for arrays",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": []}`, fromAddr),
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))},
			true,
			true,
		},
		{"bank send. fail on different message type",
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"to", "amount": []}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
			false,
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
			tx, err := s.GenSimpleTx([]sdk.Msg{tt.msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], tt.msg, tx, 0, false, authenticator.SequenceMatch)
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

// TestPoolManagerSwapExactAmountIn tests the MessageFilterAuthenticator with multiple pool manager swap messages
func (s *MessageFilterAuthenticatorTest) TestPoolManagerSwapExactAmountIn() {
	fromAddr := s.TestAccAddress[0].String()
	tests := []struct {
		name    string
		pattern string
		msg     sdk.Msg
		match   bool
	}{
		{
			"poolmanager swap exact amount in",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s","token_in":{"denom":"inputDenom", "amount":"500"}, "token_out_min_amount": "100"}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. basic match",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s"}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. match denom",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_in":{"denom":"inputDenom"}}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. match denom and sender",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_in":{"denom":"inputDenom"}}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. mismatch denom",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_in":{"denom":"wrongDenom"}}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			false,
		},
		{
			"swap exact amount. match with token out min amount",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_out_min_amount":"100"}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. mismatch with token out min amount",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_out_min_amount":"200"}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
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
			tx, err := s.GenSimpleTx([]sdk.Msg{tt.msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], tt.msg, tx, 0, false, authenticator.SequenceMatch)
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
