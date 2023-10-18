package authenticator_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

type MessageFilterAuthenticatorTest struct {
	suite.Suite
	Ctx                        sdk.Context
	OsmosisApp                 *app.OsmosisApp
	MessageFilterAuthenticator authenticator.MessageFilterAuthenticator
}

func TestMessageFilterAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(MessageFilterAuthenticatorTest))
}

func (s *MessageFilterAuthenticatorTest) SetupTest() {
	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))
	s.MessageFilterAuthenticator = authenticator.NewMessageFilterAuthenticator(s.OsmosisApp.AppCodec())
}

func (s *MessageFilterAuthenticatorTest) TestMessageTypes() {
	tests := []struct {
		name    string // name
		pattern string
		msg     sdk.Msg
		match   bool
	}{
		{"bank send",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo", "amount": "100"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
		},

		{"bank send. no amount",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to"}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
		},

		{"bank send. bad sender",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"someoneElse","to_address":"to"}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			false,
		},

		{"bank send. any",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
		},

		{"bank send. bad amount",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo", "amount": "50"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			false,
		},

		{"bank send. just denom",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 100))},
			true,
		},

		{"bank send. just denom",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))},
			true,
		},

		{"bank send. bad denom",
			`{"type":"/cosmos.bank.v1beta1.MsgSend","value":{"from_address":"from","to_address":"to", "amount": [{"denom": "foo"}]}}`,
			&bank.MsgSend{FromAddress: "from", ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("bar", 100))},
			false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			filter, err := s.MessageFilterAuthenticator.Initialize([]byte(tt.pattern))
			s.Require().NoError(err)
			result := filter.Authenticate(s.Ctx, sdk.AccAddress{}, tt.msg, nil)
			if tt.match {
				s.Require().True(result.IsAuthenticated())
			} else {
				s.Require().True(result.IsAuthenticationFailed())
			}
		})
	}
}
