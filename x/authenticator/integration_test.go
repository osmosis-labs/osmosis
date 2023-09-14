package authenticator_test

import (
	"fmt"

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
	s.PrivKeys = make([]cryptotypes.PrivKey, 2)
	for i := 0; i < 2; i++ {
		s.PrivKeys[i] = secp256k1.GenPrivKey()
	}

	// Initialize a test account with the first private key
	accountAddr := sdk.AccAddress(s.PrivKeys[0].PubKey().Address())

	// fund the account
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000))
	err := s.app.BankKeeper.SendCoins(s.chainA.GetContext(), s.chainA.SenderAccount.GetAddress(), accountAddr, coins)
	s.Require().NoError(err, "Failed to send bank tx using the first private key")

	// get the account
	s.Account = s.app.AccountKeeper.GetAccount(s.chainA.GetContext(), accountAddr)
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

	stateful := StatefulAuthenticator{KvStoreKey: s.app.GetKVStoreKey()[authenticatortypes.StoreKey]}
	s.app.AuthenticatorManager.RegisterAuthenticator(stateful)
	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "Stateful", []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// check account balances

	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, failSendMsg)
	fmt.Println("err: ", err)
	s.Require().Error(err, "Succeeded sending tx that should fail")

	// Incremented by one. Only on the ante.
	s.Require().Equal(1, stateful.GetValue(s.chainA.GetContext()))

	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, successSendMsg)
	fmt.Println("err: ", err)
	s.Require().NoError(err, "Failed to send bank tx with enough funds")

	// Incremented by 2. Ante and Post
	s.Require().Equal(3, stateful.GetValue(s.chainA.GetContext()))
}

// TODO: Cleanup experiment tests

// This is an experiment to determine how to deal with some authenticators succeeding and others failing
func (s *AuthenticatorSuite) TestAuthenticatorMultiMsgExperiment() {
	successSendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
	}

	storeKey := s.app.GetKVStoreKey()[authenticatortypes.StoreKey]
	maxAmount := MaxAmountAuthenticator{KvStoreKey: storeKey}
	stateful := StatefulAuthenticator{KvStoreKey: storeKey}

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

	alwaysLow := TestingAuthenticator{Approve: Always, GasConsumption: 0}
	alwaysHigh := TestingAuthenticator{Approve: Always, GasConsumption: 4_000}
	neverHigh := TestingAuthenticator{Approve: Never, GasConsumption: 8_000}

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
