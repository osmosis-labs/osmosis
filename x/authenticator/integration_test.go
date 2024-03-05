package authenticator_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/testutils"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"testing"

	"github.com/osmosis-labs/osmosis/v23/app"
	authenticatortypes "github.com/osmosis-labs/osmosis/v23/x/authenticator/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/osmosis-labs/osmosis/v23/tests/osmosisibctesting"
)

type AuthenticatorSuite struct {
	apptesting.KeeperTestHelper

	// using ibctesting to simplify signAndDeliver abstraction
	// TODO: is there a better way to do this?
	coordinator *ibctesting.Coordinator

	chainA         *osmosisibctesting.TestChain
	app            *app.OsmosisApp
	EncodingConfig params.EncodingConfig

	PrivKeys []cryptotypes.PrivKey
	Account  authtypes.AccountI
}

type cpks = [][]cryptotypes.PrivKey
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
	s.EncodingConfig = app.MakeEncodingConfig()

	// Initialize two private keys for testing
	s.PrivKeys = make([]cryptotypes.PrivKey, 3)
	for i := 0; i < 3; i++ {
		s.PrivKeys[i] = secp256k1.GenPrivKey()
	}

	// Initialize a test account with the first private key
	s.Account = s.CreateAccount(s.PrivKeys[0], 500_000)
}

func (s *AuthenticatorSuite) CreateAccount(privKey cryptotypes.PrivKey, amount int) authtypes.AccountI {
	accountAddr := sdk.AccAddress(privKey.PubKey().Address())
	// fund the account
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(amount)))
	err := s.app.BankKeeper.SendCoins(s.chainA.GetContext(), s.chainA.SenderAccount.GetAddress(), accountAddr, coins)
	s.Require().NoError(err, "Failed to send bank tx to account")
	return s.app.AccountKeeper.GetAccount(s.chainA.GetContext(), accountAddr)
}

// TestKeyRotationStory tests the authenticator module by adding multiple SignatureVerificationAuthenticators
// to an account and sending transaction signed by those authenticators.
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
	err = s.app.AuthenticatorKeeper.AddAuthenticator(
		s.chainA.GetContext(),
		s.Account.GetAddress(),
		"SignatureVerificationAuthenticator",
		s.PrivKeys[1].PubKey().Bytes(),
	)
	s.Require().NoError(err, "Failed to add authenticator")

	// Submit a bank send tx using the second private key
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the second private key")

	// Try to send again osing the original PrivKey. This will succeed with no selected authenticator
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Sending from the original PrivKey failed. This should succeed")

	// Remove the account's authenticator
	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), 0)
	s.Require().NoError(err, "Failed to remove authenticator")

	// Sending from the default PrivKey now works again
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the first private key after removing the authenticator")
}

// TestCircuitBreaker tests the circuit breaker for the authenticator module,
// it sends transactions with the module active and inactive.
func (s *AuthenticatorSuite) TestCircuitBreaker() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	// Activate circuit breaker
	authenticatorParams := s.app.AuthenticatorKeeper.GetParams(s.chainA.GetContext())
	authenticatorParams.AreSmartAccountsActive = false
	s.app.AuthenticatorKeeper.SetParams(s.chainA.GetContext(), authenticatorParams)

	// Send msg from accounts default privkey
	_, err := s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the first private key")

	// Add signature verification authenticator
	err = s.app.AuthenticatorKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[1].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator")

	// Submit a bank send tx using the second private key
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, sendMsg)
	s.Require().Error(err, "Failed to send bank tx using the second private key")

	// Deactivate circuit breaker
	authenticatorParams.AreSmartAccountsActive = true
	s.app.AuthenticatorKeeper.SetParams(s.chainA.GetContext(), authenticatorParams)

	// ReSubmit a bank send tx using the second private key
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the second private key")
}

// TestMessageFilterStory tests that the MessageFilterAuthenticator works as expected
func (s *AuthenticatorSuite) TestMessageFilterStory() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	// Send msg from accounts default privkey
	_, err := s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the first private key")

	// Change account's authenticator
	msgFilter := authenticator.NewMessageFilterAuthenticator(s.EncodingConfig)
	s.app.AuthenticatorManager.RegisterAuthenticator(msgFilter)
	err = s.app.AuthenticatorKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(),
		"MessageFilterAuthenticator",
		[]byte(fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","amount": [{"denom": "%s", "amount": "50"}]}`, sdk.DefaultBondDenom)))
	s.Require().NoError(err, "Failed to add authenticator")

	// Submit a bank send tx using the second authenticator the message filter
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the message filter")

	coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg = &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0}, sendMsg)
	s.Require().Error(err, "Message filter authenticator failed to block")
}

// TestKeyRotation tests an account with multiple SignatureVerificationAuthenticators
// it also checks if the account functions normally after removing authenticators
func (s *AuthenticatorSuite) TestKeyRotation() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	// Add a signature verification authenticator
	err := s.app.AuthenticatorKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[0].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator for key %d", 0)

	// Sanity check the original account with a successful message
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Bank send without authenticator should pass")

	// Sanity check the original account with a failed message
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, sendMsg)
	s.Require().Error(err, "Bank send should fail because it's signed with the incorrect private key")

	// Use the authenticator flow and verify with an authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendMsg)
	s.Require().NoError(err, "Bank send with authenticator should pass 0")

	// Add multiple keys
	err = s.app.AuthenticatorKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[1].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator for key %d", 0)

	err = s.app.AuthenticatorKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerificationAuthenticator", s.PrivKeys[2].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator for key %d", 0)

	// Use the authenticator flow and verify with an authenticator 1
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1}, sendMsg)
	s.Require().NoError(err, "Bank send with authenticator should pass 1")

	// Use the authenticator flow and verify with an authenticator 2
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[2]}, []uint64{2}, sendMsg)
	s.Require().NoError(err, "Bank send with authenticator should pass 2")

	// Fail with an incorrect authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[2]}, []uint64{1}, sendMsg)
	s.Require().Error(err, "Should fail as incorrect authenticator selected")

	// Remove an authenticator and try to send
	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), uint64(2))
	s.Require().NoError(err, "Failed to remove authenticator with id %d", 1)

	// Fail as authenticator was removed
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[2]}, []uint64{2}, sendMsg)
	s.Require().Error(err, "Should fail as authenticator was removed from store")
}

// TestingAuthenticatorState tests that the Authenticate, Track and ConfirmExecution functions work correctly,
// it increments a test authenticator state by 1 on each successful pass through the Ante and Post handlers.
func (s *AuthenticatorSuite) TestAuthenticatorState() {
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

	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0}, failSendMsg)
	s.Require().Error(err, "Succeeded sending tx that should fail")

	// Auth failed, but track still increments! Authenticate() tries to increment, but those changes are discarded.
	s.Require().Equal(1, stateful.GetValue(s.chainA.GetContext()))

	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0}, successSendMsg)
	s.Require().NoError(err, "Failed to send bank tx with enough funds")

	// Incremented by 2. Ante and Post
	s.Require().Equal(3, stateful.GetValue(s.chainA.GetContext()))
}

// TestAuthenticatorMultiMsg tests failure cases for multiple test authenticators
func (s *AuthenticatorSuite) TestAuthenticatorMultiMsg() {
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

	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0, 0}, successSendMsg, successSendMsg)
	s.Require().NoError(err)
	s.Require().Equal(int64(2_000), maxAmount.GetAmount(s.chainA.GetContext()).Int64())

	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{0, 0}, successSendMsg, successSendMsg)
	s.Require().Error(err)
	s.Require().Equal(int64(2_000), maxAmount.GetAmount(s.chainA.GetContext()).Int64())
}

// TestingAuthenticatorGas tests the gas limit panics when not reduced, then tests
// that the gas limit is reset after the fee payer is authenticated.
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
	alwaysHigher := testutils.TestingAuthenticator{Approve: testutils.Always, GasConsumption: 500_000}

	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysLow)
	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysHigh)
	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysHigher)

	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), alwaysLow.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// Both account 0 and account 1 can send
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendFromAcc1)
	s.Require().NoError(err)
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{1}, sendFromAcc2)
	s.Require().NoError(err)

	// Remove account2's authenticator
	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), account2.GetAddress(), 1)
	s.Require().NoError(err, "Failed to remove authenticator")

	// Add two authenticators that are never high, and one always high.
	// This allows account2 to execute but *only* after consuming >9k gas
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigher.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigher.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// This should fail, as authenticating the fee payer needs to be done with low gas
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{3}, sendFromAcc2)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "gas")

	// This should work, since the fee payer has already been authenticated so the gas limit is raised
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0], s.PrivKeys[1]}, pks{s.PrivKeys[0], s.PrivKeys[1]}, []uint64{0, 3}, sendFromAcc1, sendFromAcc2,
	)
	s.Require().NoError(err)
}

// TestCompositeAuthenticatorAnyOf tests an AnyOf authenticator with signature verification authenticators
func (s *AuthenticatorSuite) TestCompositeAuthenticatorAnyOf() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	anyOf := authenticator.NewAnyOfAuthenticator(s.app.AuthenticatorManager)

	// construct SubAuthenticatorInitData for each SigVerificationAuthenticator
	initDataPrivKey1 := authenticator.SubAuthenticatorInitData{
		AuthenticatorType: "SignatureVerificationAuthenticator",
		Data:              s.PrivKeys[1].PubKey().Bytes(),
	}
	initDataPrivKey2 := authenticator.SubAuthenticatorInitData{
		AuthenticatorType: "SignatureVerificationAuthenticator",
		Data:              s.PrivKeys[2].PubKey().Bytes(),
	}

	// 3. Serialize SigVerificationAuthenticator SubAuthenticatorInitData
	compositeData, err := json.Marshal([]authenticator.SubAuthenticatorInitData{
		initDataPrivKey1,
		initDataPrivKey2,
	})

	// Set the authenticator to our account
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), anyOf.Type(), compositeData)
	s.Require().NoError(err)

	// Send from account 1 using the AnyOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{0}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AnyOf authenticator key 1")

	// Send from account 2 using the AnyOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[2]}, pks{s.PrivKeys[2]}, []uint64{0}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AnyOf authenticator key 2")

	// Send from account 0 the account key using the AnyOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendMsg,
	)
	s.Require().Error(err, "Should be rejected because the account key is not in the AnyOf authenticator")
}

// TestCompositeAuthenticatorAllOf tests an AllOf authenticator with signature verification authenticator and message filter
func (s *AuthenticatorSuite) TestCompositeAuthenticatorAllOf() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	allOf := authenticator.NewAllOfAuthenticator(s.app.AuthenticatorManager)

	initDataPrivKey1 := authenticator.SubAuthenticatorInitData{
		AuthenticatorType: "SignatureVerificationAuthenticator",
		Data:              s.PrivKeys[1].PubKey().Bytes(),
	}

	initMessageFilter := authenticator.SubAuthenticatorInitData{
		AuthenticatorType: "MessageFilterAuthenticator",
		Data: []byte(
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","amount": [{"denom": "%s", "amount": "50"}]}`,
				sdk.DefaultBondDenom,
			)),
	}

	compositeData, err := json.Marshal([]authenticator.SubAuthenticatorInitData{
		initDataPrivKey1,
		initMessageFilter,
	})

	// Set the authenticator to our account
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), allOf.Type(), compositeData)
	s.Require().NoError(err)

	// Send from account 1 using the AllOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{0}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AllOf authenticator key 1")

	// Send from account 2 using the AllOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendMsg,
	)
	s.Require().Error(err, "Failed to authenticate using the AllOf authenticator account key")

	wrongCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	failedSendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      wrongCoins,
	}
	// Send from account 0 the account key using the AllOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{1}, failedSendMsg,
	)
	s.Require().Error(err, "Should be rejected because the message filter rejects the transaction")

	// Remove the first AllOf authenticator
	err = s.app.AuthenticatorKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), 0)
	s.Require().NoError(err, "Failed to remove authenticator")

	initDataPrivKey2 := authenticator.SubAuthenticatorInitData{
		AuthenticatorType: "SignatureVerificationAuthenticator",
		Data:              s.PrivKeys[2].PubKey().Bytes(),
	}

	// Create an AllOf authenticator with 2 signature verification authenticators
	compositeData, err = json.Marshal([]authenticator.SubAuthenticatorInitData{
		initDataPrivKey1,
		initDataPrivKey2,
	})

	// Set the authenticator to our account
	partitionedAllOf := authenticator.NewPartitionedAllOfAuthenticator(s.app.AuthenticatorManager)
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), partitionedAllOf.Type(), compositeData)
	s.Require().NoError(err)

	// We should provide only one signature (for the allOf authenticator) but the signature needs to be a compound
	// signature if privkey1 and privkey2 (json.Marshal([sig1, sig2])
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticatorAndCompoundSigs(
		pks{s.PrivKeys[1]}, cpks{{s.PrivKeys[1], s.PrivKeys[2]}}, []uint64{1}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AllOf authenticator account key")

	// Failed as composite signature does not match the AllOf data
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticatorAndCompoundSigs(
		pks{s.PrivKeys[1]}, cpks{{s.PrivKeys[1], s.PrivKeys[0]}}, []uint64{1}, sendMsg,
	)
	s.Require().Error(err, "Authenticated using the AllOf authenticator with incorrect signatures")
}

// TestSpendWithinLimit test the spend limit authenticator
func (s *AuthenticatorSuite) TestSpendWithinLimit() {
	authenticatorsStoreKey := s.app.GetKVStoreKey()[authenticatortypes.AuthenticatorStoreKey]
	//spendLimitStore := prefix.NewStore(s.chainA.GetContext().KVStore(authenticatorsStoreKey), []byte("spendLimitAuthenticator"))

	spendLimit := authenticator.NewSpendLimitAuthenticator(
		authenticatorsStoreKey, "allUSD", authenticator.AbsoluteValue, s.app.BankKeeper, s.app.PoolManagerKeeper, s.app.TwapKeeper,
	)
	s.app.AuthenticatorManager.RegisterAuthenticator(spendLimit)

	initData := []byte(`{"allowed": 1000, "period": "day"}`)
	err := s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), spendLimit.Type(), initData)
	s.Require().NoError(err, "Failed to add authenticator")

	amountToSend := int64(500)
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, amountToSend))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.PrivKeys[1].PubKey().Address()),
		Amount:      coins,
	}

	// Add the spend limit as part of an AnyOf
	anyOf := authenticator.NewAnyOfAuthenticator(s.app.AuthenticatorManager)
	s.app.AuthenticatorManager.RegisterAuthenticator(anyOf)

	internalData, err := json.Marshal([]authenticator.SubAuthenticatorInitData{
		{
			AuthenticatorType: "SignatureVerificationAuthenticator",
			Data:              s.PrivKeys[0].PubKey().Bytes(),
		},
		{
			AuthenticatorType: spendLimit.Type(),
			Data:              initData,
		},
	})

	// Add a SigVerificationAuthenticator to the account
	err = s.app.AuthenticatorKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), anyOf.Type(), internalData)
	s.Require().NoError(err, "Failed to add authenticator")

	// sending 500 ok
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendMsg,
	)
	s.Require().NoError(err, "Spend limit failed when it should have passed 500")

	// sending 500 ok  (1000 limit reached)
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendMsg,
	)
	s.Require().NoError(err, "Spend limit failed when it should have passed 1000")

	// sending again fails
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendMsg,
	)
	s.Require().Error(err, "Spend limit should have blocked the transaction")

	// Simulate the passage of a day
	s.coordinator.IncrementTimeBy(time.Hour * 24)
	s.coordinator.CommitBlock()

	// sending 500 ok after a day
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{0}, sendMsg,
	)
	s.Require().NoError(err, "Spend limit should have been reset")
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
