package authenticator_test

import (
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	authenticatortypes "github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
	minttypes "github.com/osmosis-labs/osmosis/v23/x/mint/types"
)

type SpendLimitAuthenticatorTest struct {
	BaseAuthenticatorSuite

	Store      prefix.Store
	SpendLimit authenticator.SpendLimitAuthenticator
}

func TestSpendLimitAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(SpendLimitAuthenticatorTest))
}

func (s *SpendLimitAuthenticatorTest) SetupTest() {
	s.SetupKeys()

	authenticatorsStoreKey := s.OsmosisApp.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]
	//s.Store = prefix.NewStore(s.Ctx.KVStore(authenticatorsStoreKey), []byte("spendLimitAuthenticator"))
	s.SpendLimit = authenticator.NewSpendLimitAuthenticator(authenticatorsStoreKey, "uosmo", authenticator.AbsoluteValue, s.OsmosisApp.BankKeeper, s.OsmosisApp.PoolManagerKeeper, s.OsmosisApp.TwapKeeper)
}

func (s *SpendLimitAuthenticatorTest) TestInitialize() {
	tests := []struct {
		name string // name
		data []byte // initData
		pass bool   // wantErr
	}{
		{"Valid day", []byte(`{"allowed": 100, "period": "day"}`), true},
		{"Valid month", []byte(`{"allowed": 100, "period": "week"}`), true},
		{"Neg allowed", []byte(`{"allowed": -100, "period": "year"}`), false},
		{"Invalid period", []byte(`{"allowed": 100, "period": "decade"}`), false},
		{"Missing allowed", []byte(`{"period": "day"}`), false},
		{"Missing period", []byte(`{"allowed": 100}`), false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := s.SpendLimit.Initialize(tt.data)
			if tt.pass {
				s.Require().NoError(err, "Should succeed")
			} else {
				s.Require().Error(err, "Should fail")
			}
		})
	}
}

func (s *SpendLimitAuthenticatorTest) TestPeriodTransition() {
	// Mock an account
	account, err := sdk.AccAddressFromBech32("osmo1s43st0ev6zuvu8ck64jumtjsz06tzqvqqmfspg")
	accountSet := authtypes.NewBaseAccount(account, nil, 0, 0)
	s.OsmosisApp.AccountKeeper.SetAccount(s.Ctx, accountSet)

	supply := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(2_000_000_000)))
	err = s.OsmosisApp.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, supply)
	s.Require().NoError(err)
	initialBalance := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1_000)))
	err = s.OsmosisApp.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, account, initialBalance)
	s.Require().NoError(err)

	tests := []struct {
		name string    // name
		data []byte    // initData
		t1   time.Time // initial time
		t2   time.Time // time after transition
		amt  int64     // spending amount
		pass bool      // expect block
	}{
		{"Day Dec31 to Jan1", []byte(`{"allowed": 100, "period": "day"}`),
			time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC), 50, true},

		{"Week Dec to Jan", []byte(`{"allowed": 100, "period": "week"}`),
			time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			time.Date(2024, 1, 7, 0, 0, 1, 0, time.UTC), 101, true},

		{"Year Dec31 to Jan1", []byte(`{"allowed": 100, "period": "year"}`),
			time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC), 50, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Initialize SpendLimitAuthenticator
			spendLimit, err := s.SpendLimit.Initialize(tt.data)
			s.Require().NoError(err, "Initialization failed")

			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
			// sample msg
			msg := &bank.MsgSend{FromAddress: s.TestAccAddress[0].String(), ToAddress: "to", Amount: sdk.NewCoins(sdk.NewInt64Coin("foo", 1))}
			// sample tx
			tx, err := s.GenSimpleTx([]sdk.Msg{msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
			s.Require().NoError(err)
			request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, s.TestAccAddress[0], s.TestAccAddress[0], msg, tx, 0, false, authenticator.SequenceMatch)
			s.Require().NoError(err)

			// Set initial time
			s.Ctx = s.Ctx.WithBlockTime(tt.t1)
			spendLimit.Authenticate(s.Ctx, request)

			// simulate spending
			err = s.OsmosisApp.BankKeeper.SendCoins(s.Ctx, account, account, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(tt.amt))))
			s.Require().NoError(err)

			// Simulate time transition
			s.Ctx = s.Ctx.WithBlockTime(tt.t2)

			// Execute ConfirmExecution and check if it's confirmed or blocked
			err = spendLimit.ConfirmExecution(s.Ctx, request)
			if tt.pass {
				s.Require().NoError(err, "Should succeed")
			} else {
				s.Require().Error(err, "Should fail")
			}
		})
	}
}

func (s *SpendLimitAuthenticatorTest) TestPeriodTransitionWithAccumulatedSpends() {
	// Mock an account
	account := s.TestAccAddress[0]
	receiver := s.TestAccAddress[1]

	supply := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(2_000_000_000)))
	err := s.OsmosisApp.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, supply)
	s.Require().NoError(err)
	initialBalance := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10_000)))
	err = s.OsmosisApp.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, account, initialBalance)
	s.Require().NoError(err)

	tests := []struct {
		name             string
		initData         []byte
		timeSpendingList []struct {
			timePoint    time.Time
			spendingAmt  int64
			expectToPass bool
		}
	}{
		{
			name:     "Day with accumulated spendings",
			initData: []byte(`{"allowed": 100, "period": "day"}`),
			timeSpendingList: []struct {
				timePoint    time.Time
				spendingAmt  int64
				expectToPass bool
			}{
				{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 150, false},
				{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 30, true},
				{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), 40, true},
				{time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), 31, true},
				{time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC), 50, true},
				{time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC), 50, true},
				{time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC), 50, true},
				{time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC), 1, false},
				{time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC), 1, true},
			},
		},
		{
			name:     "Week with accumulated spendings",
			initData: []byte(`{"allowed": 200, "period": "week"}`),
			timeSpendingList: []struct {
				timePoint    time.Time
				spendingAmt  int64
				expectToPass bool
			}{
				{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 100, true},
				{time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), 50, true},
				{time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), 51, false},
				{time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), 150, true},
				{time.Date(2024, 2, 8, 0, 0, 0, 0, time.UTC), 200, true},
				{time.Date(2024, 2, 11, 15, 0, 6, 0, time.UTC), 200, false},
				{time.Date(2024, 2, 12, 15, 0, 6, 0, time.UTC), 200, true},
			},
		},
		{
			name:     "Month with accumulated spendings",
			initData: []byte(`{"allowed": 300, "period": "month"}`),
			timeSpendingList: []struct {
				timePoint    time.Time
				spendingAmt  int64
				expectToPass bool
			}{
				{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 100, true},
				{time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 100, true},
				{time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), 101, false},
				{time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), 150, true},
				{time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), 300, true},
			},
		},
		{
			name:     "Year with accumulated spendings",
			initData: []byte(`{"allowed": 500, "period": "year"}`),
			timeSpendingList: []struct {
				timePoint    time.Time
				spendingAmt  int64
				expectToPass bool
			}{
				{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 200, true},
				{time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), 200, true},
				{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 101, false},
				{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 300, false},
				{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 99, true},
				{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), 300, true},
				{time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC), 500, true},
				{time.Date(2028, 6, 10, 0, 0, 0, 0, time.UTC), 1, false},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Reset gas
			s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))

			// Initialize SpendLimitAuthenticator
			spendLimit, err := s.SpendLimit.Initialize(tt.initData)
			s.Require().NoError(err, "Initialization failed")

			for _, pair := range tt.timeSpendingList {
				// Simulate time transition
				s.Ctx = s.Ctx.WithBlockTime(pair.timePoint)

				ak := s.OsmosisApp.AccountKeeper
				sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
				// sample msg
				msg := &bank.MsgSend{FromAddress: account.String(), ToAddress: receiver.String(), Amount: sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(pair.spendingAmt)))}
				// sample tx
				tx, err := s.GenSimpleTx([]sdk.Msg{msg}, []cryptotypes.PrivKey{s.TestPrivKeys[0]})
				s.Require().NoError(err)
				request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, account, account, msg, tx, 0, false, authenticator.SequenceMatch)
				s.Require().NoError(err)

				spendLimit.Authenticate(s.Ctx, request)
				err = spendLimit.Track(s.Ctx, account, account, nil, 0, "0")
				s.Require().NoError(err)

				// Simulate spending
				err = s.OsmosisApp.BankKeeper.SendCoins(s.Ctx, account, receiver, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(pair.spendingAmt))))
				s.Require().NoError(err)

				// Execute ConfirmExecution and check if it's confirmed or blocked
				result := spendLimit.ConfirmExecution(s.Ctx, request)
				if pair.expectToPass {
					s.Require().NoError(result, "Should succeed")
				} else {
					s.Require().Error(result, "Should fail")
				}
			}
		})
	}
}
