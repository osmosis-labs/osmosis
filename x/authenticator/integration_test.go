package authenticator_test

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"time"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/testutils"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"testing"

	"github.com/osmosis-labs/osmosis/v19/app"
	authenticatortypes "github.com/osmosis-labs/osmosis/v19/x/authenticator/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/tests/osmosisibctesting"
)

type AuthenticatorSuite struct {
	apptesting.KeeperTestHelper

	// using ibctesting to simplify signAndDeliver abstraction
	// TODO: is there a better way to do this?
	coordinator *ibctesting.Coordinator

	chainA *osmosisibctesting.TestChain
	app    *app.OsmosisApp

	PrivKeys []cryptotypes.PrivKey
	Account  authtypes.AccountI
}

type pks = []cryptotypes.PrivKey

func TestAuthenticatorSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorSuite))
}

func (s *AuthenticatorSuite) SetupTest() {
	// Use the osmosis custom function for creating an osmosis app
	ibctesting.DefaultTestingAppInit = osmosisibctesting.SetupTestingApp

	// Here we create the app using ibctesting
	s.coordinator = ibctesting.NewCoordinator(s.T(), 1)
	s.chainA = &osmosisibctesting.TestChain{
		TestChain: s.coordinator.GetChain(ibctesting.GetChainID(1)),
	}
	s.app = s.chainA.GetOsmosisApp()

	// Initialize two private keys for testing
	s.PrivKeys = make([]cryptotypes.PrivKey, 3)
	for i := 0; i < 3; i++ {
		s.PrivKeys[i] = secp256k1.GenPrivKey()
	}

	// Initialize a test account with the first private key
	s.Account = s.CreateAccount(s.PrivKeys[0], 100_000)
}

func (s *AuthenticatorSuite) CreateAccount(privKey cryptotypes.PrivKey, amount int) authtypes.AccountI {
	accountAddr := sdk.AccAddress(privKey.PubKey().Address())
	// fund the account
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(amount)))
	err := s.app.BankKeeper.SendCoins(s.chainA.GetContext(), s.chainA.SenderAccount.GetAddress(), accountAddr, coins)
	s.Require().NoError(err, "Failed to send bank tx to account")
	return s.app.AccountKeeper.GetAccount(s.chainA.GetContext(), accountAddr)
}

func (s *AuthenticatorSuite) TestKeyRotationStory() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	// Send msg from accounts default privkey
	_, err := s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the first private key")

	// Change account's authenticator
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[1].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator")

	// Submit a bank send tx using the second private key
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the second private key")

	// Try to send again osing the original PrivKey. This should fail
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().Error(err, "Sending from the original PrivKey succeeded. This should fail")

	// Remove the account's authenticator
	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), 0)
	s.Require().NoError(err, "Failed to remove authenticator")

	// Sending from the default PrivKey now works again
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the first private key after removing the authenticator")

}

type SendTest struct {
	PrivKeyIndex  int
	ShouldSucceed bool
}

type KeyRotationStep struct {
	KeysToAdd              []int
	AuthenticatorsToRemove []int
	Sends                  []SendTest
}

type KeyRotationTest struct {
	Description string
	Steps       []KeyRotationStep
}

func (s *AuthenticatorSuite) TestKeyRotation() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	tests := []KeyRotationTest{
		{
			Description: "Test with no keys",
			Steps: []KeyRotationStep{
				{
					KeysToAdd:              []int{},
					AuthenticatorsToRemove: []int{},
					Sends: []SendTest{
						{PrivKeyIndex: 0, ShouldSucceed: true},
					},
				},
			},
		},

		{
			Description: "Test add own key as authenticator",
			Steps: []KeyRotationStep{
				{
					KeysToAdd:              []int{0},
					AuthenticatorsToRemove: []int{},
					Sends: []SendTest{
						{PrivKeyIndex: 0, ShouldSucceed: true},
					},
				},
			},
		},

		{
			Description: "Test no authenticator change",
			Steps: []KeyRotationStep{
				{
					KeysToAdd:              []int{1},
					AuthenticatorsToRemove: []int{0},
					Sends: []SendTest{
						{PrivKeyIndex: 0, ShouldSucceed: true},
					},
				},
			},
		},

		{
			Description: "Test simple key rotation",
			Steps: []KeyRotationStep{
				{
					KeysToAdd:              []int{1},
					AuthenticatorsToRemove: []int{},
					Sends: []SendTest{
						{PrivKeyIndex: 1, ShouldSucceed: true},
						{PrivKeyIndex: 0, ShouldSucceed: false},
					},
				},
			},
		},

		{
			Description: "Test add both keys",
			Steps: []KeyRotationStep{
				{
					KeysToAdd:              []int{0, 1},
					AuthenticatorsToRemove: []int{},
					Sends: []SendTest{
						{PrivKeyIndex: 0, ShouldSucceed: true},
						{PrivKeyIndex: 1, ShouldSucceed: true},
					},
				},
			},
		},

		{
			Description: "Test complex rotations",
			Steps: []KeyRotationStep{
				{
					KeysToAdd:              []int{0},
					AuthenticatorsToRemove: []int{},
					Sends: []SendTest{ // current authenticators (id=0, key=0)
						{PrivKeyIndex: 0, ShouldSucceed: true},
						{PrivKeyIndex: 1, ShouldSucceed: false},
					},
				},
				{
					KeysToAdd:              []int{1},
					AuthenticatorsToRemove: []int{0},
					Sends: []SendTest{ // current authenticators (id=1, key=1)
						{PrivKeyIndex: 0, ShouldSucceed: false},
						{PrivKeyIndex: 1, ShouldSucceed: true},
					},
				},
				{
					KeysToAdd:              []int{0},
					AuthenticatorsToRemove: []int{1},
					Sends: []SendTest{ // current authenticators (id=2, key=0)
						{PrivKeyIndex: 0, ShouldSucceed: true},
						{PrivKeyIndex: 1, ShouldSucceed: false},
					},
				},

				{
					KeysToAdd:              []int{},
					AuthenticatorsToRemove: []int{2},
					Sends: []SendTest{ // all authenticators removed. Back to default
						{PrivKeyIndex: 0, ShouldSucceed: true},
						{PrivKeyIndex: 1, ShouldSucceed: false},
					},
				},

				{
					KeysToAdd:              []int{1, 0},
					AuthenticatorsToRemove: []int{},
					Sends: []SendTest{ // current authenticators (id=3, key=1), (id=4, key=0)
						{PrivKeyIndex: 0, ShouldSucceed: true},
						{PrivKeyIndex: 1, ShouldSucceed: true},
					},
				},

				{
					KeysToAdd:              []int{},
					AuthenticatorsToRemove: []int{4},
					Sends: []SendTest{ // current authenticators (id=3, key=1)
						{PrivKeyIndex: 0, ShouldSucceed: false},
						{PrivKeyIndex: 1, ShouldSucceed: true},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.Description, func() {
			// Reset authenticators
			s.app.AuthenticatorKeeper.SetNextAuthenticatorId(s.chainA.GetContext(), 0)
			allAuthenticators, err := s.app.AuthenticatorKeeper.GetAuthenticatorDataForAccount(s.chainA.GetContext(), s.Account.GetAddress())
			for _, authenticator := range allAuthenticators {
				err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), authenticator.Id)
				s.Require().NoError(err, "Failed to remove authenticator")
			}

			for _, step := range tc.Steps {
				// useful for debugging
				//allAuthenticators, _ := s.app.AuthenticatorKeeper.GetAuthenticatorDataForAccount(s.chainA.GetContext(), s.Account.GetAddress())
				//fmt.Println("allAuthenticators", allAuthenticators)

				// Add keys for the current step
				for _, keyIndex := range step.KeysToAdd {
					err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[keyIndex].PubKey().Bytes())
					s.Require().NoError(err, "Failed to add authenticator for key %d in %s", keyIndex, tc.Description)
				}

				// Remove keys for the current step
				for _, authenticatorId := range step.AuthenticatorsToRemove {
					err := s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), uint64(authenticatorId))
					s.Require().NoError(err, "Failed to remove authenticator with id %d in %s", authenticatorId, tc.Description)
				}

				// Send for the current step
				for _, send := range step.Sends {
					_, err := s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[send.PrivKeyIndex]}, sendMsg)
					if send.ShouldSucceed {
						s.Require().NoError(err, tc.Description)
					} else {
						s.Require().Error(err, tc.Description)
					}
				}
			}
		})
	}
}

// This is an experiment to determine how internal authenticator state is managed
func (s *AuthenticatorSuite) TestAuthenticatorStateExperiment() {
	successSendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)),
	}
	// This amount is too large, so the send should fail
	failSendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000_000)),
	}

	stateful := testutils.StatefulAuthenticator{KvStoreKey: s.app.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]}
	s.app.AuthenticatorManager.RegisterAuthenticator(stateful)
	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "Stateful", []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// check account balances

	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, failSendMsg)
	fmt.Println("err: ", err)
	s.Require().Error(err, "Succeeded sending tx that should fail")

	// Auth failed, so no increment
	s.Require().Equal(0, stateful.GetValue(s.chainA.GetContext()))

	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, successSendMsg)
	fmt.Println("err: ", err)
	s.Require().NoError(err, "Failed to send bank tx with enough funds")

	// Incremented by 2. Ante and Post
	s.Require().Equal(2, stateful.GetValue(s.chainA.GetContext()))
}

// TODO: Cleanup experiment tests

// This is an experiment to determine how to deal with some authenticators succeeding and others failing
func (s *AuthenticatorSuite) TestAuthenticatorMultiMsgExperiment() {
	successSendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
	}

	storeKey := s.app.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]
	maxAmount := testutils.MaxAmountAuthenticator{KvStoreKey: storeKey}
	stateful := testutils.StatefulAuthenticator{KvStoreKey: storeKey}

	s.app.AuthenticatorManager.RegisterAuthenticator(maxAmount)
	s.app.AuthenticatorManager.RegisterAuthenticator(stateful)

	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "MaxAmountAuthenticator", []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	//err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "Stateful", []byte{})
	//s.Require().NoError(err, "Failed to add authenticator")
	// check account balances

	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, successSendMsg, successSendMsg)
	fmt.Println("err: ", err)
	s.Require().NoError(err)
	s.Require().Equal(int64(2_000), maxAmount.GetAmount(s.chainA.GetContext()).Int64())

	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, successSendMsg, successSendMsg)
	fmt.Println("err: ", err)
	s.Require().Error(err)
	s.Require().Equal(int64(2_000), maxAmount.GetAmount(s.chainA.GetContext()).Int64())
}

// This is an experiment to determine how internal authenticator state is managed
func (s *AuthenticatorSuite) TestAuthenticatorGas() {
	sendFromAcc1 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
	}

	// Initialize the second account
	accountAddr := sdk.AccAddress(s.PrivKeys[1].PubKey().Address())

	// fund the account
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000))
	err := s.app.BankKeeper.SendCoins(s.chainA.GetContext(), s.chainA.SenderAccount.GetAddress(), accountAddr, coins)
	s.Require().NoError(err, "Failed to send bank tx using the first private key")

	// get the account
	account2 := s.app.AccountKeeper.GetAccount(s.chainA.GetContext(), accountAddr)

	sendFromAcc2 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", account2.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", account2.GetAddress()),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
	}

	alwaysLow := testutils.TestingAuthenticator{Approve: testutils.Always, GasConsumption: 0}
	alwaysHigh := testutils.TestingAuthenticator{Approve: testutils.Always, GasConsumption: 4_000}
	neverHigh := testutils.TestingAuthenticator{Approve: testutils.Never, GasConsumption: 8_000}

	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysLow)
	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysHigh)
	s.app.AuthenticatorManager.RegisterAuthenticator(neverHigh)

	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), alwaysLow.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// Both account 0 and account 1 can send
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendFromAcc1)
	s.Require().NoError(err)
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, sendFromAcc2)
	s.Require().NoError(err)

	// Remove account2's authenticator
	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), account2.GetAddress(), 1)
	s.Require().NoError(err, "Failed to remove authenticator")

	// Add two authenticators that are never high, and one always high.
	// This allows account2 to execute but *only* after consuming >9k gas
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), neverHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), neverHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// This should fail, as authenticating the fee payer needs to be done with low gas
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, sendFromAcc2)
	fmt.Println(err.Error())
	s.Require().Error(err)
	s.Require().ErrorContains(err, "gas")

	// This should work, since the fee payer has already been authenticated so the gas limit is raised
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0], s.PrivKeys[1]}, sendFromAcc1, sendFromAcc2)
	s.Require().NoError(err)
}

func (s *AuthenticatorSuite) TestCompositeAuthenticatorIntegration() {
	// create Send Msg
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	anyOf := authenticator.NewAnyOfAuthenticator(s.app.AuthenticatorManager)
	allOf := authenticator.NewAllOfAuthenticator(s.app.AuthenticatorManager)

	s.app.AuthenticatorManager.RegisterAuthenticator(anyOf)
	s.app.AuthenticatorManager.RegisterAuthenticator(allOf)

	// construct InitializationData for each SigVerificationAuthenticator
	initDataPrivKey0 := authenticator.InitializationData{
		AuthenticatorType: "SignatureVerificationAuthenticator",
		Data:              s.PrivKeys[0].PubKey().Bytes(),
	}
	initDataPrivKey1 := authenticator.InitializationData{
		AuthenticatorType: "SignatureVerificationAuthenticator",
		Data:              s.PrivKeys[1].PubKey().Bytes(),
	}
	initDataPrivKey2 := authenticator.InitializationData{
		AuthenticatorType: "SignatureVerificationAuthenticator",
		Data:              s.PrivKeys[2].PubKey().Bytes(),
	}

	// 3. Serialize SigVerificationAuthenticator InitializationData
	dataPrivKey0, err := json.Marshal(initDataPrivKey0)
	s.Require().NoError(err)
	dataPrivKey1, err := json.Marshal(initDataPrivKey1)
	s.Require().NoError(err)

	// construct InitializationData for AnyOf authenticator
	initDataAnyOf := authenticator.InitializationData{
		AuthenticatorType: anyOf.Type(),
		Data:              append(dataPrivKey0, dataPrivKey1...),
	}

	// 5. Combine to construct the final composite for AllOf authenticator
	compositeAuthData := []authenticator.InitializationData{
		initDataAnyOf,
		initDataPrivKey2,
	}

	// serialize the AllOf InitializationData
	dataAllOf, err := json.Marshal(compositeAuthData)
	s.Require().NoError(err)

	// Set the authenticator to our account
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), allOf.Type(), dataAllOf)
	s.Require().NoError(err)

	// Current State AllOf(AnyOf(Sig0, Sig1), Sig2)

	// 0 fails
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().Error(err)

	// 2 fails
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[2]}, sendMsg)
	s.Require().Error(err)

	// 0 and 2 succeeds
	// TODO: This doesn't work right now because there are checks on the number of sigs matching
	//       senders (validation will prob fail for the same reason). We may want to test AllOf
	//       with a different authenticator (MaxAmountAuthenticator?) and use multisig instead
	//       of AllOf for validating multiple signatures
	//_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0], s.PrivKeys[2]}, sendMsg)
	//s.Require().NoError(err)
}

func (s *AuthenticatorSuite) TestSpendWithinLimit() {
	authenticatorsStoreKey := s.app.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]
	//spendLimitStore := prefix.NewStore(s.chainA.GetContext().KVStore(authenticatorsStoreKey), []byte("spendLimitAuthenticator"))

	spendLimit := authenticator.NewSpendLimitAuthenticator(authenticatorsStoreKey, "allUSD", authenticator.AbsoluteValue, s.app.BankKeeper, s.app.PoolManagerKeeper, s.app.TwapKeeper)
	s.app.AuthenticatorManager.RegisterAuthenticator(spendLimit)

	initData := []byte(`{"allowed": 1000, "period": "day"}`)
	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), spendLimit.Type(), initData)
	s.Require().NoError(err, "Failed to add authenticator")

	amountToSend := int64(500)
	// Create a send message
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, amountToSend))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
		Amount:      coins,
	}

	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "unauthorized") // Spend limit only rejects. Never authorizes

	// Add a SigVerificationAuthenticator to the account
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[0].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator")

	// sending 500 ok
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err)

	// sending 500 ok  (1000 limit reached)
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err)

	// sending again fails
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().Error(err)

	// Simulate the passage of a day
	s.coordinator.IncrementTimeBy(time.Hour * 24)
	s.coordinator.CommitBlock()

	// sending 500 ok after a day
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err)
}

// TODO: We have discovered an issue with the authz integration that prevents this test from if
//
//	the spendLimitAuthenticator doesn't keep it's own store (only a store key)
//	This test is modified for now and will issue a hotfix later
func (s *AuthenticatorSuite) TestSpendWithinLimitWithAuthz() {
	// Setup and register the authenticator
	authenticatorsStoreKey := s.app.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]
	spendLimit := authenticator.NewSpendLimitAuthenticator(authenticatorsStoreKey, "allUSD", authenticator.AbsoluteValue, s.app.BankKeeper, s.app.PoolManagerKeeper, s.app.TwapKeeper)
	s.app.AuthenticatorManager.RegisterAuthenticator(spendLimit)

	// Add the authenticator to account1
	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[0].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator")
	initData := []byte(`{"allowed": 1000, "period": "day"}`)
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), spendLimit.Type(), initData)
	s.Require().NoError(err, "Failed to add authenticator")

	// Create SendAuthorization
	sendAuth := &banktypes.SendAuthorization{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000)),
	}

	// Pack the SendAuthorization into an Any type
	sendAuthAny, err := types.NewAnyWithValue(sendAuth)
	s.Require().NoError(err)

	// Create Grant
	grant := authz.Grant{
		Authorization: sendAuthAny,
		Expiration:    time.Now().Add(time.Hour * 24 * 10),
	}

	// Grant Send Authorization from s.PrivKeys[0] to s.PrivKeys[1]
	grantMsg := &authz.MsgGrant{
		Granter: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Grantee: sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
		Grant:   grant,
	}

	// Create account for the second private key. This is needed for executing the grant
	s.CreateAccount(s.PrivKeys[1], 50_000)

	// Store the grant
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, grantMsg)
	s.Require().NoError(err)

	// Prepare sends
	amountToSend := int64(500)
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, amountToSend))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
		Amount:      coins,
	}

	// Pack the MsgSend into an Any type
	sendMsgAny, err := types.NewAnyWithValue(sendMsg)
	s.Require().NoError(err)

	// Create a MsgExec to send coins from s.PrivKeys[1] using the granted authz
	execMsg := &authz.MsgExec{
		Grantee: sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
		Msgs:    []*types.Any{sendMsgAny},
	}

	// sending 500 ok. This send is without authz; both are tracked
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err)

	// sending 500 ok (1000 limit reached)
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, execMsg)
	s.Require().NoError(err)

	// sending again fails
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, execMsg)
	s.Require().Error(err)

	// Sending without authz also fails
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().Error(err)

	// Simulate the passage of a day
	s.coordinator.IncrementTimeBy(time.Hour * 24)
	s.coordinator.CommitBlock()

	// sending 500 ok after a day
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err)

	// and sending 500 with authz also ok after a day
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, execMsg)
	s.Require().NoError(err)

}

func (s *AuthenticatorSuite) TestAuthenticatorAddRemove() {
	// Register the authenticators
	blockAdd := testutils.TestingAuthenticator{BlockAddition: true}
	allowAdd := testutils.TestingAuthenticator{}
	blockRemove := testutils.TestingAuthenticator{BlockRemoval: true}
	allowRemove := testutils.TestingAuthenticator{}

	s.app.AuthenticatorManager.RegisterAuthenticator(blockAdd)
	s.app.AuthenticatorManager.RegisterAuthenticator(allowAdd)
	s.app.AuthenticatorManager.RegisterAuthenticator(blockRemove)
	s.app.AuthenticatorManager.RegisterAuthenticator(allowRemove)

	// Initialize an account
	accountAddr := sdk.AccAddress(s.PrivKeys[0].PubKey().Address())

	// Test authenticator that blocks addition
	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, blockAdd.Type(), []byte{})
	s.Require().Error(err, "Authenticator should not be added")
	s.Require().ErrorContains(err, "authenticator could not be added")

	// Test authenticator that allows addition
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, allowAdd.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// Test authenticator that blocks removal
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, blockRemove.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), accountAddr, 1)
	s.Require().Error(err, "Authenticator should not be removed")
	s.Require().ErrorContains(err, "authenticator could not be removed")

	// Test authenticator that allows removal
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, allowRemove.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), accountAddr, 2)
	s.Require().NoError(err, "Failed to remove authenticator")
}

func (s *AuthenticatorSuite) TestSpendWithinLimitWithAuthzTableTest() {
	// Setup and register the authenticator
	authenticatorsStoreKey := s.app.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]
	spendLimit := authenticator.NewSpendLimitAuthenticator(authenticatorsStoreKey, "allUSD", authenticator.AbsoluteValue, s.app.BankKeeper, s.app.PoolManagerKeeper, s.app.TwapKeeper)
	s.app.AuthenticatorManager.RegisterAuthenticator(spendLimit)

	// Add the authenticator to account1
	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[0].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator")
	initData := []byte(`{"allowed": 1000, "period": "day"}`)
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), spendLimit.Type(), initData)
	s.Require().NoError(err, "Failed to add authenticator")

	// Create SendAuthorization
	sendAuth := &banktypes.SendAuthorization{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000)),
	}

	// Pack the SendAuthorization into an Any type
	sendAuthAny, err := types.NewAnyWithValue(sendAuth)
	s.Require().NoError(err)

	// Create Grant
	grant := authz.Grant{
		Authorization: sendAuthAny,
		Expiration:    time.Now().Add(time.Hour * 24 * 10),
	}

	// Grant Send Authorization from s.PrivKeys[0] to s.PrivKeys[1]
	grantMsg := &authz.MsgGrant{
		Granter: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Grantee: sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
		Grant:   grant,
	}

	// Create account for the second private key. This is needed for executing the grant
	s.CreateAccount(s.PrivKeys[1], 50_000)

	// Store the grant
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, grantMsg)
	s.Require().NoError(err)

	// Define a struct for each send operation within a test case
	type sendOperation struct {
		txType      string // "execMsg" or "sendMsg"
		amount      int64  // Amount to send
		pass        bool   // Whether this send should pass or fail
		advanceDays int64
	}

	// Define a struct for test cases
	type testCase struct {
		name  string          // Name of the test case for easier identification
		sends []sendOperation // A slice of send operations
	}

	testCases := []testCase{
		{name: "Case 1",
			sends: []sendOperation{
				{"sendMsg", 500, true, 0},
				{"execMsg", 500, true, 0},
				{"sendMsg", 500, false, 0},
				{"execMsg", 500, false, 1},
				{"execMsg", 500, true, 0},
			}},

		{name: "Case 2",
			sends: []sendOperation{
				{"sendMsg", 1001, false, 0},
				{"execMsg", 999, true, 0},
				{"sendMsg", 2, false, 0},
				{"execMsg", 1, true, 0},
				{"execMsg", 1, false, 1},
				{"execMsg", 1000, true, 1},
				{"execMsg", 1001, false, 0},
				{"execMsg", 1000, true, 1},
				{"sendMsg", 1001, false, 0},
				{"sendMsg", 1000, true, 0},
			}},

		{name: "Case 3",
			sends: []sendOperation{
				{"sendMsg", 900, true, 0},
				{"execMsg", 50, true, 0},
				{"execMsg", 500, false, 0},
				{"sendMsg", 1100, false, 0},
				{"sendMsg", 100, false, 0},
				{"execMsg", 1, true, 0},
				{"execMsg", 49, true, 0},
				{"execMsg", 1, false, 1},
				{"execMsg", 1, true, 0},
			}},

		{name: "Case 4",
			sends: []sendOperation{
				{"sendMsg", 250, true, 0},
				{"execMsg", 250, true, 0},
				{"sendMsg", 250, true, 0},
				{"execMsg", 250, true, 0},
				{"sendMsg", 100, false, 0},
				{"execMsg", 1, false, 0},
				{"execMsg", 1, false, 1},
				{"execMsg", 1, true, 0},
				{"sendMsg", 1, true, 0},
			}},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			fmt.Println("Running test case: ", tc.name)
			s.coordinator.IncrementTimeBy(time.Hour * 48)
			s.coordinator.CommitBlock()

			for _, send := range tc.sends {
				coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, send.amount))
				var msg sdk.Msg
				if send.txType == "sendMsg" {
					msg = &banktypes.MsgSend{
						FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
						ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
						Amount:      coins,
					}
				} else if send.txType == "execMsg" {
					sendMsg := &banktypes.MsgSend{
						FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
						ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
						Amount:      coins,
					}
					sendMsgAny, err := types.NewAnyWithValue(sendMsg)
					s.Require().NoError(err)
					msg = &authz.MsgExec{
						Grantee: sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
						Msgs:    []*types.Any{sendMsgAny},
					}
				}

				var pk cryptotypes.PrivKey
				if send.txType == "sendMsg" {
					pk = s.PrivKeys[0]
				} else {
					pk = s.PrivKeys[1]
				}

				_, err := s.chainA.SendMsgsFromPrivKeys(pks{pk}, msg)
				if send.pass {
					s.Require().NoError(err, "Failed in test case: %s, txType: %s, amount: %d", tc.name, send.txType, send.amount)
				} else {
					s.Require().Error(err, "Unexpected pass in test case: %s, txType: %s, amount: %d", tc.name, send.txType, send.amount)
				}
				if send.advanceDays > 0 {
					s.coordinator.IncrementTimeBy(time.Hour * 24 * time.Duration(send.advanceDays))
					s.coordinator.CommitBlock()
				}
			}
		})
	}
}
