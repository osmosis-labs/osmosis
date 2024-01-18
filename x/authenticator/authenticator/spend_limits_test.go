package authenticator_test

import (
	"encoding/json"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/app"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"
	authenticatortypes "github.com/osmosis-labs/osmosis/v21/x/authenticator/types"
	minttypes "github.com/osmosis-labs/osmosis/v21/x/mint/types"
)

type SpendLimitAuthenticatorTest struct {
	CosmwasmAuthenticatorTest

	SpendLimit   authenticator.SpendLimitAuthenticator
	ContractAddr sdk.AccAddress
}

func TestSpendLimitAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(SpendLimitAuthenticatorTest))
}

func (s *SpendLimitAuthenticatorTest) SetupTest() {
	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))

	authenticatorsStoreKey := s.OsmosisApp.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]
	s.SpendLimit = authenticator.NewSpendLimitAuthenticator(authenticatorsStoreKey, "uosmo", authenticator.AbsoluteValue, s.OsmosisApp.BankKeeper, s.OsmosisApp.PoolManagerKeeper, s.OsmosisApp.TwapKeeper)
	s.StoreContractCode("../testutils/contracts/spend-limit/artifacts/spend-limit-aarch64.wasm")
	s.ContractAddr = s.InstantiateContract("{}", 1)

}

func (s *SpendLimitAuthenticatorTest) InitializeContract(initData authenticator.CosmwasmAuthenticatorInitData) authenticator.CosmwasmAuthenticator {
	initDataBz, err := json.Marshal(initData)
	s.Require().NoError(err, "Initialization failed")
	cw, err := s.OsmosisApp.AuthenticatorManager.GetAuthenticatorByType("CosmwasmAuthenticatorV1").Initialize(initDataBz)
	s.Require().NoError(err, "Initialization failed")
	return cw.(authenticator.CosmwasmAuthenticator)
}

func (s *SpendLimitAuthenticatorTest) TestInitialize() {
	tests := []struct {
		name   string // name
		params []byte // initData
		pass   bool   // wantErr
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
			initData := authenticator.CosmwasmAuthenticatorInitData{Contract: s.ContractAddr.String()}
			a11r := s.InitializeContract(initData)
			_, err := a11r.Initialize(tt.params)
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

			// Set initial time
			s.Ctx = s.Ctx.WithBlockTime(tt.t1)
			spendLimit.Authenticate(s.Ctx, account, nil, nil)

			// simulate spending
			err = s.OsmosisApp.BankKeeper.SendCoins(s.Ctx, account, account, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(tt.amt))))
			s.Require().NoError(err)

			// Simulate time transition
			s.Ctx = s.Ctx.WithBlockTime(tt.t2)

			// Execute ConfirmExecution and check if it's confirmed or blocked
			result := spendLimit.ConfirmExecution(s.Ctx, account, nil, nil)
			s.Require().Equal(tt.pass, result.IsConfirm())
		})
	}
}

func (s *SpendLimitAuthenticatorTest) TestPeriodTransitionWithAccumulatedSpends() {
	// Mock an account
	account, err := sdk.AccAddressFromBech32("osmo1s43st0ev6zuvu8ck64jumtjsz06tzqvqqmfspg")
	accountSet := authtypes.NewBaseAccount(account, nil, 0, 0)
	s.OsmosisApp.AccountKeeper.SetAccount(s.Ctx, accountSet)

	receiver, err := sdk.AccAddressFromBech32("osmo1f3cwcxwmpzjm56zavvv4xat43jxeyk0du4hqfj")
	accountSet = authtypes.NewBaseAccount(receiver, nil, 0, 0)
	s.OsmosisApp.AccountKeeper.SetAccount(s.Ctx, accountSet)

	supply := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(2_000_000_000)))
	err = s.OsmosisApp.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, supply)
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

				spendLimit.Authenticate(s.Ctx, account, nil, nil)
				err := spendLimit.Track(s.Ctx, account, nil)
				s.Require().NoError(err)

				// Simulate spending
				err = s.OsmosisApp.BankKeeper.SendCoins(s.Ctx, account, receiver, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(pair.spendingAmt))))
				s.Require().NoError(err)

				// Execute ConfirmExecution and check if it's confirmed or blocked
				result := spendLimit.ConfirmExecution(s.Ctx, account, nil, nil)
				s.Require().Equal(pair.expectToPass, result.IsConfirm())
			}
		})
	}
}
