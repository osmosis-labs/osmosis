package authenticator_test

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

type MessageFilterAuthenticatorTest struct {
	BaseAuthenticatorSuite

	MessageFilterAuthenticator authenticator.MessageFilterAuthenticator
}

func TestMessageFilterAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(MessageFilterAuthenticatorTest))
}

func (s *MessageFilterAuthenticatorTest) SetupTest() {
	s.SetupKeys()
	s.MessageFilterAuthenticator = authenticator.NewMessageFilterAuthenticator()
}

func (s *MessageFilterAuthenticatorTest) TestBankSend() {
	fromAddr := s.TestAccAddress[0].String()
	tests := []struct {
		name           string // name
		pattern        string
		msg            sdk.Msg
		passvalidation bool
		match          bool
	}{
		{"PASS: bank send",
			"/cosmos.bank.v1beta1.MsgSend",
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},
		{"PASS: bank send, multimsg",
			"/cosmos.bank.v1beta1.MsgSend, /osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn",
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},
		{"PASS: msg swap, multimsg",
			"/cosmos.bank.v1beta1.MsgSend, /osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn",
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			true,
		},
		{
			"FAIL: swap message, poolmanager swap exact amount in",
			"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn",
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			true,
			true,
		},
		{"FAIL: bank send. incorrect message",
			"/cosmos.bank.v1beta1.MgSnd",
			&bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
			false,
		},
		{
			"FAiL: empty data",
			"",
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", sdk.NewInt(500)),
				TokenOutMinAmount: sdk.NewInt(100),
			},
			false,
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
