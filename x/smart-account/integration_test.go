package authenticator_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/osmosis-labs/osmosis/osmomath"
	txfeetypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/testutils"

	"testing"

	"github.com/osmosis-labs/osmosis/v27/app"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/tests/osmosisibctesting"
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
	Account  sdk.AccountI
}

type cpks = [][]cryptotypes.PrivKey
type pks = []cryptotypes.PrivKey

func TestAuthenticatorSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorSuite))
}

func (s *AuthenticatorSuite) SetupTest() {
	txfeetypes.ConsensusMinFee = osmomath.ZeroDec()

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

func (suite *AuthenticatorSuite) TearDownSuite() {
	for _, dir := range osmosisibctesting.TestingDirectories {
		os.RemoveAll(dir)
	}
}

func (s *AuthenticatorSuite) CreateAccount(privKey cryptotypes.PrivKey, amount int) sdk.AccountI {
	accountAddr := sdk.AccAddress(privKey.PubKey().Address())
	// fund the account
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(amount)))
	err := s.app.BankKeeper.SendCoins(s.chainA.GetContext(), s.chainA.SenderAccount.GetAddress(), accountAddr, coins)
	s.Require().NoError(err, "Failed to send bank tx to account")
	return s.app.AccountKeeper.GetAccount(s.chainA.GetContext(), accountAddr)
}

// TestKeyRotationStory tests the authenticator module by adding multiple SignatureVerifications
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
	sigVerAuthId, err := s.app.SmartAccountKeeper.AddAuthenticator(
		s.chainA.GetContext(),
		s.Account.GetAddress(),
		"SignatureVerification",
		s.PrivKeys[1].PubKey().Bytes(),
	)
	s.Require().NoError(err, "Failed to add authenticator")

	// Submit a bank send tx using the second private key
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the second private key")

	// Submit a bank send tx using the first private key. This will fail
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{1}, sendMsg)
	s.Require().Error(err)

	// Try to send again using the original PrivKey. This will succeed with no selected authenticator
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Sending from the original PrivKey failed. This should succeed")

	// Remove the account's authenticator
	err = s.app.SmartAccountKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), sigVerAuthId)
	s.Require().NoError(err, "Failed to remove authenticator")

	// Sending from the default PrivKey still works
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
	authenticatorParams := s.app.SmartAccountKeeper.GetParams(s.chainA.GetContext())
	authenticatorParams.IsSmartAccountActive = false
	s.app.SmartAccountKeeper.SetParams(s.chainA.GetContext(), authenticatorParams)

	// Send msg from accounts default privkey
	_, err := s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the first private key")

	// Add signature verification authenticator
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerification", s.PrivKeys[1].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator")

	// Submit a bank send tx using the second private key
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, sendMsg)
	s.Require().Error(err, "Failed to send bank tx using the second private key")

	// Deactivate circuit breaker
	authenticatorParams.IsSmartAccountActive = true
	s.app.SmartAccountKeeper.SetParams(s.chainA.GetContext(), authenticatorParams)

	// ReSubmit a bank send tx using the second private key
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the second private key")
}

// TestMessageFilterStory tests that the MessageFilter works as expected
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
	msgFilter := authenticator.NewMessageFilter(s.EncodingConfig)
	s.app.AuthenticatorManager.RegisterAuthenticator(msgFilter)
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(),
		"MessageFilter",
		[]byte(fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","amount": [{"denom": "%s", "amount": "50"}]}`, sdk.DefaultBondDenom)))
	s.Require().NoError(err, "Failed to add authenticator")

	// Submit a bank send tx using the second authenticator the message filter
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1}, sendMsg)
	s.Require().NoError(err, "Failed to send bank tx using the message filter")

	coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg = &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1}, sendMsg)
	s.Require().Error(err, "Message filter authenticator failed to block")
}

// TestKeyRotation tests an account with multiple SignatureVerifications
// it also checks if the account functions normally after removing authenticators
func (s *AuthenticatorSuite) TestKeyRotation() {
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", s.Account.GetAddress()),
		Amount:      coins,
	}

	// Add a signature verification authenticator
	_, err := s.app.SmartAccountKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerification", s.PrivKeys[0].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator for key %d", 0)

	// Sanity check the original account with a successful message
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[0]}, sendMsg)
	s.Require().NoError(err, "Bank send without authenticator should pass")

	// Sanity check the original account with a failed message
	_, err = s.chainA.SendMsgsFromPrivKeys(pks{s.PrivKeys[1]}, sendMsg)
	s.Require().Error(err, "Bank send should fail because it's signed with the incorrect private key")

	// Use the authenticator flow and verify with an authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{1}, sendMsg)
	s.Require().NoError(err, "Bank send with authenticator should pass 0")

	// Add multiple keys
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerification", s.PrivKeys[1].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator for key %d", 0)

	_, err = s.app.SmartAccountKeeper.AddAuthenticator(
		s.chainA.GetContext(), s.Account.GetAddress(), "SignatureVerification", s.PrivKeys[2].PubKey().Bytes())
	s.Require().NoError(err, "Failed to add authenticator for key %d", 0)

	// Use the authenticator flow and verify with an authenticator 1
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{2}, sendMsg)
	s.Require().NoError(err, "Bank send with authenticator should pass 1")

	// Use the authenticator flow and verify with an authenticator 2
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[2]}, []uint64{3}, sendMsg)
	s.Require().NoError(err, "Bank send with authenticator should pass 2")

	// Fail with an incorrect authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[2]}, []uint64{2}, sendMsg)
	s.Require().Error(err, "Should fail as incorrect authenticator selected")

	// Remove an authenticator and try to send
	err = s.app.SmartAccountKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), uint64(3))
	s.Require().NoError(err, "Failed to remove authenticator with id %d", 1)

	// Fail as authenticator was removed
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[2]}, []uint64{3}, sendMsg)
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

	stateful := testutils.StatefulAuthenticator{KvStoreKey: s.app.GetKVStoreKey()[smartaccounttypes.StoreKey]}
	s.app.AuthenticatorManager.RegisterAuthenticator(stateful)
	_, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "Stateful", []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1}, failSendMsg)
	s.Require().Error(err, "Succeeded sending tx that should fail")

	// Auth failed, but track still increments! Authenticate() tries to increment, but those changes are discarded.
	s.Require().Equal(1, stateful.GetValue(s.chainA.GetContext()))

	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1}, successSendMsg)
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

	storeKey := s.app.GetKVStoreKey()[smartaccounttypes.StoreKey]
	maxAmount := testutils.MaxAmountAuthenticator{KvStoreKey: storeKey}
	stateful := testutils.StatefulAuthenticator{KvStoreKey: storeKey}

	s.app.AuthenticatorManager.RegisterAuthenticator(maxAmount)
	s.app.AuthenticatorManager.RegisterAuthenticator(stateful)

	_, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), "MaxAmountAuthenticator", []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// Note that we are sending 2 messages here, so the amount should be 2_000 (2*1_000)
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1, 1}, successSendMsg, successSendMsg)
	s.Require().NoError(err)
	s.Require().Equal(int64(2_000), maxAmount.GetAmount(s.chainA.GetContext()).Int64())

	// This should now fail, as the max amount has been reached
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[1]}, []uint64{1, 1}, successSendMsg, successSendMsg)
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

	// Will always approve and will consume 0 gas
	alwaysLow := testutils.TestingAuthenticator{Approve: testutils.Always, GasConsumption: 0}
	// Will always approve and will consume 4k gas
	alwaysHigh := testutils.TestingAuthenticator{Approve: testutils.Always, GasConsumption: 4_000}
	// Will always approve and will consume 500k gas
	alwaysHigher := testutils.TestingAuthenticator{Approve: testutils.Always, GasConsumption: 500_000}

	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysLow)
	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysHigh)
	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysHigher)

	_, err = s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), alwaysLow.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	acc2authId, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// Both account 0 and account 1 can send
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{1}, sendFromAcc1)
	s.Require().NoError(err)
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{2}, sendFromAcc2)
	s.Require().NoError(err)

	// Remove account2's authenticator
	err = s.app.SmartAccountKeeper.RemoveAuthenticator(s.chainA.GetContext(), account2.GetAddress(), acc2authId)
	s.Require().NoError(err, "Failed to remove authenticator")

	// Add two authenticators that are always higher, and one always high.
	// This allows account2 to execute but *only* after consuming >9k gas
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigher.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigher.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), account2.GetAddress(), alwaysHigh.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// This should fail, as authenticating the fee payer needs to be done with low gas
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{4}, sendFromAcc2)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "gas")

	// This should work, since the fee payer has already been authenticated so the gas limit is raised
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0], s.PrivKeys[1]}, pks{s.PrivKeys[0], s.PrivKeys[1]}, []uint64{1, 4}, sendFromAcc1, sendFromAcc2,
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

	anyOf := authenticator.NewAnyOf(s.app.AuthenticatorManager)

	// construct SubAuthenticatorInitData for each SigVerificationAuthenticator
	initDataPrivKey1 := authenticator.SubAuthenticatorInitData{
		Type:   "SignatureVerification",
		Config: s.PrivKeys[1].PubKey().Bytes(),
	}
	initDataPrivKey2 := authenticator.SubAuthenticatorInitData{
		Type:   "SignatureVerification",
		Config: s.PrivKeys[2].PubKey().Bytes(),
	}

	// 3. Serialize SigVerificationAuthenticator SubAuthenticatorInitData
	compositeData, err := json.Marshal([]authenticator.SubAuthenticatorInitData{
		initDataPrivKey1,
		initDataPrivKey2,
	})

	// Set the authenticator to our account
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), anyOf.Type(), compositeData)
	s.Require().NoError(err)

	// Send from account 1 using the AnyOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{1}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AnyOf authenticator key 1")

	// Send from account 2 using the AnyOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[2]}, pks{s.PrivKeys[2]}, []uint64{1}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AnyOf authenticator key 2")

	// Send from account 0 the account key using the AnyOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{1}, sendMsg,
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

	allOf := authenticator.NewAllOf(s.app.AuthenticatorManager)

	initDataPrivKey1 := authenticator.SubAuthenticatorInitData{
		Type:   "SignatureVerification",
		Config: s.PrivKeys[1].PubKey().Bytes(),
	}

	initMessageFilter := authenticator.SubAuthenticatorInitData{
		Type: "MessageFilter",
		Config: []byte(
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.MsgSend","amount": [{"denom": "%s", "amount": "50"}]}`,
				sdk.DefaultBondDenom,
			)),
	}

	compositeData, err := json.Marshal([]authenticator.SubAuthenticatorInitData{
		initDataPrivKey1,
		initMessageFilter,
	})

	// Set the authenticator to our account
	allOfAuthId, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), allOf.Type(), compositeData)
	s.Require().NoError(err)

	// Send from account 1 using the AllOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[1]}, pks{s.PrivKeys[1]}, []uint64{1}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AllOf authenticator key 1")

	// Send from account 2 using the AllOf authenticator
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticator(
		pks{s.PrivKeys[0]}, pks{s.PrivKeys[0]}, []uint64{1}, sendMsg,
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
	s.Require().ErrorContains(err, "message does not match pattern")

	// Remove the first AllOf authenticator
	err = s.app.SmartAccountKeeper.RemoveAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), allOfAuthId)
	s.Require().NoError(err, "Failed to remove authenticator")

	initDataPrivKey2 := authenticator.SubAuthenticatorInitData{
		Type:   "SignatureVerification",
		Config: s.PrivKeys[2].PubKey().Bytes(),
	}

	// Create an AllOf authenticator with 2 signature verification authenticators
	compositeData, err = json.Marshal([]authenticator.SubAuthenticatorInitData{
		initDataPrivKey1,
		initDataPrivKey2,
	})

	// Set the authenticator to our account
	partitionedAllOf := authenticator.NewPartitionedAllOf(s.app.AuthenticatorManager)
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), s.Account.GetAddress(), partitionedAllOf.Type(), compositeData)
	s.Require().NoError(err)

	// We should provide only one signature (for the allOf authenticator) but the signature needs to be a compound
	// signature if privkey1 and privkey2 (json.Marshal([sig1, sig2])
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticatorAndCompoundSigs(
		pks{s.PrivKeys[1]}, cpks{{s.PrivKeys[1], s.PrivKeys[2]}}, []uint64{2}, sendMsg,
	)
	s.Require().NoError(err, "Failed to authenticate using the AllOf authenticator account key")

	// Failed as composite signature does not match the AllOf data
	_, err = s.chainA.SendMsgsFromPrivKeysWithAuthenticatorAndCompoundSigs(
		pks{s.PrivKeys[1]}, cpks{{s.PrivKeys[1], s.PrivKeys[0]}}, []uint64{2}, sendMsg,
	)
	s.Require().Error(err, "Authenticated using the AllOf authenticator with incorrect signatures")
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
	_, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, blockAdd.Type(), []byte{})
	s.Require().Error(err, "Authenticator should not be added")
	s.Require().ErrorContains(err, "authenticator could not be added")

	// Test authenticator that allows addition
	_, err = s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, allowAdd.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	// Test authenticator that blocks removal
	blockRemoveId, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, blockRemove.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	err = s.app.SmartAccountKeeper.RemoveAuthenticator(s.chainA.GetContext(), accountAddr, blockRemoveId)
	s.Require().Error(err, "Authenticator should not be removed")
	s.Require().ErrorContains(err, "authenticator could not be removed")

	// Test authenticator that allows removal
	allowRemoveId, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), accountAddr, allowRemove.Type(), []byte{})
	s.Require().NoError(err, "Failed to add authenticator")

	err = s.app.SmartAccountKeeper.RemoveAuthenticator(s.chainA.GetContext(), accountAddr, allowRemoveId)
	s.Require().NoError(err, "Failed to remove authenticator")
}

func (s *AuthenticatorSuite) TestFeeDeduction() {
	address1 := sdk.AccAddress(s.PrivKeys[0].PubKey().Address())
	address2 := sdk.AccAddress(s.PrivKeys[1].PubKey().Address())
	s.CreateAccount(s.PrivKeys[1], 500_000)

	alwaysAuth := testutils.TestingAuthenticator{Approve: testutils.Always}
	neverAuth := testutils.TestingAuthenticator{Approve: testutils.Never}

	s.app.AuthenticatorManager.RegisterAuthenticator(alwaysAuth)
	s.app.AuthenticatorManager.RegisterAuthenticator(neverAuth)

	payerYes, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), address1, alwaysAuth.Type(), []byte{})
	s.Require().NoError(err)
	payerNo, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), address1, neverAuth.Type(), []byte{})
	s.Require().NoError(err)
	otherYes, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), address2, alwaysAuth.Type(), []byte{})
	s.Require().NoError(err)
	otherNo, err := s.app.SmartAccountKeeper.AddAuthenticator(s.chainA.GetContext(), address2, neverAuth.Type(), []byte{})
	s.Require().NoError(err)

	testCases := []struct {
		name                   string
		signers                []cryptotypes.PrivKey
		messages               []sdk.Msg
		selectedAuthenticators []uint64
		expectedError          bool
		expectedErrorMsg       string
	}{

		{
			name:                   "single message, authenticated, fee deducted + tx succeeded",
			signers:                []cryptotypes.PrivKey{s.PrivKeys[0]},
			selectedAuthenticators: []uint64{payerYes},
			messages: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address1),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address1),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
				},
			},
			expectedError: false,
		},

		{
			name:                   "single message, not authenticated, fee not deducted + tx failed",
			signers:                []cryptotypes.PrivKey{s.PrivKeys[0]},
			selectedAuthenticators: []uint64{payerNo},
			messages: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address1),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address1),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
				},
			},
			expectedError:    true,
			expectedErrorMsg: "unauthorized",
		},

		{
			name:                   "multiple messages, all authenticated, fee deducted + tx succeeded",
			signers:                []cryptotypes.PrivKey{s.PrivKeys[0], s.PrivKeys[1]},
			selectedAuthenticators: []uint64{payerYes, otherYes},
			messages: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address1),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address1),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)),
				},
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address2),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address2),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)),
				},
			},
			expectedError: false,
		},

		{
			name:                   "multiple messages, fee payer authenticated but other not, fee deducted + tx failed",
			signers:                []cryptotypes.PrivKey{s.PrivKeys[0], s.PrivKeys[1]},
			selectedAuthenticators: []uint64{payerYes, otherNo},
			messages: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address1),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address1),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)),
				},
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address2),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address2),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)),
				},
			},
			expectedError:    true,
			expectedErrorMsg: "unauthorized",
		},

		{
			name:                   "multiple messages, fee payer not authenticated, fee not deducted + tx failed",
			signers:                []cryptotypes.PrivKey{s.PrivKeys[0], s.PrivKeys[1]},
			selectedAuthenticators: []uint64{payerNo, otherNo},
			messages: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address1),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address1),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)),
				},
				&banktypes.MsgSend{
					FromAddress: sdk.MustBech32ifyAddressBytes("osmo", address2),
					ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", address2),
					Amount:      sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)),
				},
			},
			expectedError:    true,
			expectedErrorMsg: "unauthorized",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			initialBalance := s.app.BankKeeper.GetAllBalances(s.chainA.GetContext(), sdk.AccAddress(tc.signers[0].PubKey().Address()))
			_, err := s.chainA.SendMsgsFromPrivKeysWithAuthenticator(tc.signers, tc.signers, tc.selectedAuthenticators, tc.messages...)
			if tc.expectedError {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErrorMsg)
			} else {
				s.Require().NoError(err)
			}
			finalBalance := s.app.BankKeeper.GetAllBalances(s.chainA.GetContext(), sdk.AccAddress(tc.signers[0].PubKey().Address()))
			fee := sdk.NewInt64Coin(sdk.DefaultBondDenom, 25000)
			expectedBalance := initialBalance.Sub(fee)
			if tc.selectedAuthenticators[0] == payerYes {
				s.Require().True(expectedBalance.Equal(finalBalance), "Fee should be deducted")
			} else {
				s.Require().True(initialBalance.Equal(finalBalance), "Fee should not be deducted")
			}
		})
	}
}
