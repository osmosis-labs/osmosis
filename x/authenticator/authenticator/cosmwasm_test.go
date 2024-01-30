package authenticator_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/app"
	"github.com/osmosis-labs/osmosis/v21/app/apptesting"
	"github.com/osmosis-labs/osmosis/v21/app/params"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"
	minttypes "github.com/osmosis-labs/osmosis/v21/x/mint/types"
)

type CosmwasmAuthenticatorTest struct {
	suite.Suite
	Ctx            sdk.Context
	OsmosisApp     *app.OsmosisApp
	Store          prefix.Store
	EncodingConfig params.EncodingConfig
	CosmwasmAuth   authenticator.CosmwasmAuthenticator
}

func TestCosmwasmAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(CosmwasmAuthenticatorTest))
}

func (s *CosmwasmAuthenticatorTest) SetupTest() {
	s.OsmosisApp = app.Setup(false)
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(10_000_000))
	s.EncodingConfig = app.MakeEncodingConfig()

	s.CosmwasmAuth = authenticator.NewCosmwasmAuthenticator(s.OsmosisApp.ContractKeeper, s.OsmosisApp.AccountKeeper, s.EncodingConfig.TxConfig.SignModeHandler(), s.OsmosisApp.AppCodec())
}

func (s *CosmwasmAuthenticatorTest) TestOnAuthenticatorAdded() {
	tests := []struct {
		name string // name
		data []byte // initData
		pass bool   // wantErr
	}{
		{"Valid Contract", []byte(`{"contract": "osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69"}`), true},
		{"Valid Contract, valid params", []byte(fmt.Sprintf(`{"contract": "osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69", "params": %s }`, toBytesString(`{ "p1": "v1", "p2": { "p21": "v21" } }`))), true},
		{"Valid Contract, invalid params", []byte(fmt.Sprintf(`{"contract": "osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69", "params": %s }`, toBytesString(`{ "p1": "v1", "p2": { "p21" "v21" } }`))), false},
		{"Missing Contract", []byte(`{}`), false},
		{"Invalid Contract", []byte(`{"contract": "invalid_address"}`), false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.CosmwasmAuth.OnAuthenticatorAdded(s.Ctx, sdk.AccAddress{}, tt.data)
			if tt.pass {
				s.Require().NoError(err, "Should succeed")
			} else {
				s.Require().Error(err, "Should fail")
			}
		})
	}
}

func (s *CosmwasmAuthenticatorTest) TestInitialize() {
	tests := []struct {
		name         string // name
		data         []byte // initData
		contractAddr string // expected address
		params       []byte // expected params
		pass         bool   // wantErr
	}{
		{
			"Valid Contract",
			[]byte(`{"contract": "osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69"}`),
			"osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69",
			nil,
			true,
		},
		{
			"Valid Contract, valid params",
			[]byte(fmt.Sprintf(`{"contract": "osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69", "params": %s }`, toBytesString(`{ "p1": "v1", "p2": { "p21": "v21" } }`))),
			"osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69",
			[]byte(`{ "p1": "v1", "p2": { "p21": "v21" } }`),
			true,
		},
		{
			"Valid Contract, invalid params",
			[]byte(fmt.Sprintf(`{"contract": "osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69", "params": %s }`, toBytesString(`{ "p1": "v1", "p2": { "p21" "v21" } }`))),
			"osmo1t3gjpqadhhqcd29v64xa06z66mmz7kazsvkp69",
			[]byte(`{ "p1": "v1", "p2": { "p21" "v21" } }`),
			false,
		},
		{
			"Missing Contract",
			[]byte(`{}`),
			"",
			nil,
			false,
		},
		{
			"Invalid Contract",
			[]byte(`{"contract": "invalid_address"}`),
			"",
			nil,
			false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			auth, err := s.CosmwasmAuth.Initialize(tt.data)
			cwa, ok := auth.(authenticator.CosmwasmAuthenticator)

			if tt.pass {
				s.Require().True(ok, "Should create valid CosmwasmAuthenticator")
				s.Require().Equal(tt.contractAddr, cwa.ContractAddress().String(), "Contract address must be initialized")
				s.Require().Equal(tt.params, cwa.Params(), "Params must be initialized")
				s.Require().NoError(err, "Should succeed")
			} else {
				s.Require().Error(err, "Should fail")
			}
		})
	}
}

type EchoInstantiateMsg struct {
	PubKey []byte `json:"pubkey"`
}

func (s *CosmwasmAuthenticatorTest) TestGeneral() {
	accounts := apptesting.CreateRandomAccounts(2)
	for _, acc := range accounts {
		someCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
		err := s.OsmosisApp.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, someCoins)
		s.Require().NoError(err)
		err = s.OsmosisApp.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, someCoins)
		s.Require().NoError(err)
	}

	// Mocking some data for the GenTx function based on PassKeyTests
	osmoToken := "osmo"
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Create a test message for signing
	testMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, accounts[0]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, accounts[1]),
		Amount:      feeCoins,
	}
	msgs := []sdk.Msg{testMsg}

	// Account numbers and sequences
	accNums := []uint64{0}
	accSeqs := []uint64{0}

	// Generate a private key for signing
	priv := secp256k1.GenPrivKey()
	signers := []cryptotypes.PrivKey{priv}
	signatures := []cryptotypes.PrivKey{priv}

	// Define encoding config if not already defined
	encodingConfig := app.MakeEncodingConfig() // Assuming the app has a method called MakeEncodingConfig

	// Set up the contract
	s.StoreContractCode("../testutils/contracts/echo/artifacts/echo-aarch64.wasm")
	instantiateMsg := EchoInstantiateMsg{PubKey: priv.PubKey().Bytes()}
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	addr := s.InstantiateContract(string(instantiateMsgBz), 1)

	auth, err := s.CosmwasmAuth.Initialize([]byte(
		fmt.Sprintf(`{"contract": "%s"}`, addr)))
	s.Require().NoError(err, "Should succeed")

	tx, _ := GenTx(
		encodingConfig.TxConfig,
		msgs,
		feeCoins,
		300000,
		"",
		accNums,
		accSeqs,
		signers,
		signatures,
	)

	authData, err := auth.GetAuthenticationData(s.Ctx, tx, -1, false)
	s.Require().NoError(err, "Should succeed")

	status := auth.Authenticate(s.Ctx.WithBlockTime(time.Now()), accounts[0], testMsg, authData)
	s.Require().True(status.IsAuthenticated(), "Should be authenticated")

	authData.(authenticator.SignatureData).Signatures[0].Data = &txsigning.SingleSignatureData{
		SignMode:  0,
		Signature: []byte("invalid"),
	}
	status = auth.Authenticate(s.Ctx.WithBlockTime(time.Now()), accounts[0], testMsg, authData)
	s.Require().False(status.IsAuthenticated(), "Should not be authenticated")
}

type CosignerInstantiateMsg struct {
	PubKeys [][]byte `json:"pubkeys"`
}

func (s *CosmwasmAuthenticatorTest) TestCosignerContract() {
	accounts := apptesting.CreateRandomAccounts(2)
	for _, acc := range accounts {
		someCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
		err := s.OsmosisApp.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, someCoins)
		s.Require().NoError(err)
		err = s.OsmosisApp.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, someCoins)
		s.Require().NoError(err)
	}

	// Mocking some data for the GenTx function based on PassKeyTests
	osmoToken := "osmo"
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Create a test message for signing
	testMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, accounts[0]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, accounts[1]),
		Amount:      feeCoins,
	}

	// Generate a private key for signing
	msgs := []sdk.Msg{testMsg}

	// Account numbers and sequences
	accNums := []uint64{0, 0}
	accSeqs := []uint64{0, 0}

	// Generate a private key for signing
	priv := secp256k1.GenPrivKey()
	cosigner := secp256k1.GenPrivKey()
	signers := []cryptotypes.PrivKey{priv, cosigner}
	signatures := []cryptotypes.PrivKey{priv, cosigner}

	// Define encoding config if not already defined
	encodingConfig := app.MakeEncodingConfig() // Assuming the app has a method called MakeEncodingConfig

	// Set up the contract
	s.StoreContractCode("../testutils/contracts/cosigner-authenticator/artifacts/cosigner_authenticator-aarch64.wasm")
	instantiateMsg := CosignerInstantiateMsg{PubKeys: [][]byte{priv.PubKey().Bytes()}}
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	addr := s.InstantiateContract(string(instantiateMsgBz), 1)

	auth, err := s.CosmwasmAuth.Initialize([]byte(
		fmt.Sprintf(`{"contract": "%s"}`, addr)))
	s.Require().NoError(err, "Should succeed")

	tx, _ := GenTx(
		encodingConfig.TxConfig,
		msgs,
		feeCoins,
		300000,
		"",
		accNums,
		accSeqs,
		signers,
		signatures,
	)

	// TODO: this currently fails as signatures are stripped from the tx. Should we add them or maybe do a better
	//  cosigner implementation later?
	authData, err := auth.GetAuthenticationData(s.Ctx, tx, -1, false)
	s.Require().NoError(err, "Should succeed")

	status := auth.Authenticate(s.Ctx.WithBlockTime(time.Now()), accounts[0], testMsg, authData)
	s.Require().True(status.IsAuthenticated(), "Should be authenticated")

}

func (s *CosmwasmAuthenticatorTest) StoreContractCode(path string) uint64 {
	osmosisApp := s.OsmosisApp
	govKeeper := wasmkeeper.NewGovPermissionKeeper(osmosisApp.WasmKeeper)
	creator := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	wasmCode, err := os.ReadFile(path)
	s.Require().NoError(err)
	accessEveryone := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody}
	codeID, _, err := govKeeper.Create(s.Ctx.WithBlockTime(time.Now()), creator, wasmCode, &accessEveryone)
	s.Require().NoError(err)
	return codeID
}

func (s *CosmwasmAuthenticatorTest) InstantiateContract(msg string, codeID uint64) sdk.AccAddress {
	osmosisApp := s.OsmosisApp
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	creator := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(s.Ctx.WithBlockTime(time.Now()), codeID, creator, creator, []byte(msg), "contract", nil)
	s.Require().NoError(err)
	return addr
}

func toBytesString(s string) string {
	bytes := []byte(s)
	bytesString := "["
	for i, b := range bytes {
		if i != 0 {
			bytesString += ","
		}
		bytesString += fmt.Sprintf("%d", b)
	}
	bytesString += "]"

	return bytesString
}
