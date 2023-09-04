package authenticator_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"

	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/tests/osmosisibctesting"
	"github.com/stretchr/testify/suite"
)

type AuthenticatorSuite struct {
	apptesting.KeeperTestHelper

	// using ibctesting to simplify signAndDeliver abstraction
	// TODO: is there a better way to do this?
	coordinator *ibctesting.Coordinator

	chainA *osmosisibctesting.TestChain

	SenderPrivKeys []cryptotypes.PrivKey
	AccountAddr    sdk.AccAddress
}

func TestAuthenticatorSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorSuite))
}

func (suite *AuthenticatorSuite) SetupTest() {
	// NOTE: do we want to setup a second testing app?
	// suite.Setup()

	// Use the osmosis custom function for creating an osmosis app
	ibctesting.DefaultTestingAppInit = osmosisibctesting.SetupTestingApp

	// Here we create the app using ibctesting
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1)
	suite.chainA = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(1)),
	}

	// Initialize two private keys for testing
	suite.SenderPrivKeys = make([]cryptotypes.PrivKey, 2)
	for i := 0; i < 2; i++ {
		suite.SenderPrivKeys[i] = secp256k1.GenPrivKey()
	}

	// NOTE: osmosis app != ibctesting app
	// suite.App != suite.chainA

	// Initialize a test account with the first private key
	suite.AccountAddr = sdk.AccAddress(suite.SenderPrivKeys[0].PubKey().Address())
}

func (suite *AuthenticatorSuite) TestKeyRotation() {
	// NOTE: do we need to call this here, this will set up a third app?
	//suite.SetupTest()

	osmosisApp := suite.chainA.GetOsmosisApp()
	// fund the account
	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000))
	err := osmosisApp.BankKeeper.SendCoins(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), suite.AccountAddr, coins)
	suite.Require().NoError(err, "Failed to send bank tx using the first private key")

	coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", suite.AccountAddr.Bytes()),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", suite.AccountAddr.Bytes()),
		Amount:      coins,
	}

	// can't initialise here as this is where the error stems from
	// osmosisApp.AccountKeeper = suite.App.AccountKeeper

	result, err := suite.chainA.SendMsgsFromPrivKey(1, 1, suite.SenderPrivKeys[0], sendMsg)
	suite.Require().NoError(err, "Failed to send bank tx using the first private key")
	fmt.Println(result)

	// Step 2: Change account's authenticator
	// TODO: Your logic to change the authenticator

	// Step 3: Submit a bank send tx using the second private key
}
