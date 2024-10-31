package authenticator_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/app/params"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"

	"github.com/stretchr/testify/suite"
)

type MessageFilterTest struct {
	BaseAuthenticatorSuite

	MessageFilter  authenticator.MessageFilter
	EncodingConfig params.EncodingConfig
}

func TestMessageFilterTest(t *testing.T) {
	suite.Run(t, new(MessageFilterTest))
}

func (s *MessageFilterTest) SetupTest() {
	s.SetupKeys()
	s.EncodingConfig = app.MakeEncodingConfig()
	s.MessageFilter = authenticator.NewMessageFilter(s.EncodingConfig)
}

func (s *MessageFilterTest) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

// TestBankSend tests the MessageFilter with multiple bank send messages
func (s *MessageFilterTest) TestBankSend() {
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
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			true,
			false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.MessageFilter.OnAuthenticatorAdded(s.Ctx, sdk.AccAddress{}, []byte(tt.pattern), "1")
			if tt.passvalidation {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}
			filter, err := s.MessageFilter.Initialize([]byte(tt.pattern))
			s.Require().NoError(err)

			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			tx, err := s.GenSimpleTx([]sdk.Msg{tt.msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, s.OsmosisApp.AppCodec(), ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], nil, sdk.NewCoins(), tt.msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			err = filter.Authenticate(s.Ctx, request)
			if tt.match {
				s.Require().True(err == nil)
			} else {
				s.Require().True(err != nil)
			}
		})
	}
}

// TestPoolManagerSwapExactAmountIn tests the MessageFilter with multiple pool manager swap messages
func (s *MessageFilterTest) TestPoolManagerSwapExactAmountIn() {
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
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. basic match",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s"}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. match denom",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_in":{"denom":"inputDenom"}}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. match denom and sender",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_in":{"denom":"inputDenom"}}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. mismatch denom",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_in":{"denom":"wrongDenom"}}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			false,
		},
		{
			"swap exact amount. match with token out min amount",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_out_min_amount":"100"}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			true,
		},
		{
			"swap exact amount. mismatch with token out min amount",
			fmt.Sprintf(`{"@type":"/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn","sender":"%s", "token_out_min_amount":"200"}`, fromAddr),
			&poolmanagertypes.MsgSwapExactAmountIn{
				Sender:            fromAddr,
				TokenIn:           sdk.NewCoin("inputDenom", osmomath.NewInt(500)),
				TokenOutMinAmount: osmomath.NewInt(100),
			},
			false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			filter, err := s.MessageFilter.Initialize([]byte(tt.pattern))
			s.Require().NoError(err)

			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			tx, err := s.GenSimpleTx([]sdk.Msg{tt.msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, s.OsmosisApp.AppCodec(), ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], nil, sdk.NewCoins(), tt.msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			err = filter.Authenticate(s.Ctx, request)
			if tt.match {
				s.Require().True(err == nil)
			} else {
				s.Require().True(err != nil)
			}
		})
	}
}

func (s *MessageFilterTest) TestLimitOrder() {
	fromAddr := s.TestAccAddress[0].String()
	tests := []struct {
		name           string // name
		pattern        string
		msg            sdk.Msg
		passvalidation bool
		match          bool
	}{
		{"place limit order, simple message filter, no contract address",
			fmt.Sprintf(`{"@type":"/cosmwasm.wasm.v1.MsgExecuteContract", "msg": {"place_limit": {}}}`),
			&types.MsgExecuteContract{
				Contract: "osmo16xcfxjd8263srfqhl5stru49y2w3u7dllugn9dkdrlrhfaeu523s85htxv",
				Msg:      []byte(fmt.Sprintf(`{"place_limit": { "claim_bounty": "%s", "order_direction": "%s", "quantity": "%s", "tick_id": %d}}`, "0.0001", "bid", "47612515", -5257343)),
				Sender:   fromAddr,
				Funds:    sdk.NewCoins(sdk.NewCoin("inputDenom", osmomath.NewInt(100))),
			},
			true,
			true,
		},
		{"place limit order, complex message filter, no contract address",
			fmt.Sprintf(`{"@type":"/cosmwasm.wasm.v1.MsgExecuteContract", "sender":"%s", "msg": {"place_limit": {}}, "contract": ""}`, fromAddr),
			&types.MsgExecuteContract{
				Contract: "",
				Msg:      []byte(fmt.Sprintf(`{"place_limit": { "claim_bounty": "%s", "order_direction": "%s", "quantity": "%s", "tick_id": %d}}`, "0.0001", "bid", "47612515", -5257343)),
				Sender:   fromAddr, Funds: sdk.NewCoins(sdk.NewCoin("inputDenom", osmomath.NewInt(100)))},
			true,
			true,
		},
		{"place limit order, only bid, no contract address",
			fmt.Sprintf(`{"@type":"/cosmwasm.wasm.v1.MsgExecuteContract", "sender":"%s", "msg": {"place_limit": { "order_direction": "bid"}}, "contract": ""}`, fromAddr),
			&types.MsgExecuteContract{
				Contract: "",
				Msg:      []byte(fmt.Sprintf(`{"place_limit": { "claim_bounty": "%s", "order_direction": "%s", "quantity": "%s", "tick_id": %d}}`, "0.0001", "bid", "47612515", -5257343)),
				Sender:   fromAddr, Funds: sdk.NewCoins(sdk.NewCoin("inputDenom", osmomath.NewInt(100)))},
			true,
			true,
		},
		{"place limit order, only ask, with contract address",
			fmt.Sprintf(`{"@type":"/cosmwasm.wasm.v1.MsgExecuteContract", "sender":"%s", "msg": {"place_limit": { "order_direction": "bid"}}, "contract": "osmo1aufrskevnmtvafnvflea9ypllkqqz333tzfcs3cwsg9mfk7946yqprspug"}`, fromAddr),
			&types.MsgExecuteContract{
				Contract: "osmo1aufrskevnmtvafnvflea9ypllkqqz333tzfcs3cwsg9mfk7946yqprspug",
				Msg:      []byte(fmt.Sprintf(`{"place_limit": { "claim_bounty": "%s", "order_direction": "%s", "quantity": "%s", "tick_id": %d}}`, "0.0001", "bid", "47612515", -5257343)),
				Sender:   fromAddr, Funds: sdk.NewCoins(sdk.NewCoin("inputDenom", osmomath.NewInt(100)))},
			true,
			true,
		},
		{"place limit order error, no contract address",
			fmt.Sprintf(`{"@type":"/cosmwasm.wasm.v1.MsgExecuteContract", "sender":"%s", "msg": {"place_limit": { "claim_bounty": "", "order_direction": ""}}, "contract": ""}`, fromAddr),
			&types.MsgExecuteContract{
				Contract: "",
				Msg:      []byte(fmt.Sprintf(`{"place_limit": { "claim_bounty": "%s", "order_direction": "%s", "quantity": "%s", "tick_id": %d}}`, "0.0001", "bid", "47612515", -5257343)),
				Sender:   fromAddr, Funds: sdk.NewCoins(sdk.NewCoin("inputDenom", osmomath.NewInt(100)))},
			true,
			false,
		},
		{"place limit order error, restricted message",
			fmt.Sprintf(`{"@type":"/cosmwasm.wasm.v1.MsgExecuteContract", "msg": {"place_limit": { "claim_bounty": "", "order_direction": "", "quantity": "", "tick_id": "0"}}}`),
			&types.MsgExecuteContract{
				Contract: "",
				Msg:      []byte(fmt.Sprintf(`{"place_limit": { "claim_bounty": "%s", "order_direction": "%s", "quantity": "%s", "tick_id": %d}}`, "0.0001", "bid", "47612515", -5257343)),
				Sender:   fromAddr, Funds: sdk.NewCoins(sdk.NewCoin("inputDenom", osmomath.NewInt(100)))},
			true,
			false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.MessageFilter.OnAuthenticatorAdded(s.Ctx, sdk.AccAddress{}, []byte(tt.pattern), "1")
			if tt.passvalidation {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}
			filter, err := s.MessageFilter.Initialize([]byte(tt.pattern))
			s.Require().NoError(err)

			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			tx, err := s.GenSimpleTx([]sdk.Msg{tt.msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, s.OsmosisApp.AppCodec(), ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], nil, sdk.NewCoins(), tt.msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			err = filter.Authenticate(s.Ctx, request)
			if tt.match {
				s.Require().True(err == nil)
			} else {
				s.Require().True(err != nil)
			}
		})
	}
}
