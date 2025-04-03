package authenticator_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"cosmossdk.io/store/prefix"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/app/params"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
)

type CosmwasmAuthenticatorTest struct {
	suite.Suite
	Ctx            sdk.Context
	OsmosisApp     *app.OsmosisApp
	Store          prefix.Store
	EncodingConfig params.EncodingConfig
	CosmwasmAuth   authenticator.CosmwasmAuthenticator
	HomeDir        string
}

func TestCosmwasmAuthenticatorTest(t *testing.T) {
	suite.Run(t, new(CosmwasmAuthenticatorTest))
}

func (s *CosmwasmAuthenticatorTest) SetupTest() {
	s.HomeDir = fmt.Sprintf("%d", rand.Int())
	s.OsmosisApp = app.SetupWithCustomHome(false, s.HomeDir)
	s.Ctx = s.OsmosisApp.NewContextLegacy(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(10_000_000))
	s.EncodingConfig = app.MakeEncodingConfig()

	s.CosmwasmAuth = authenticator.NewCosmwasmAuthenticator(s.OsmosisApp.ContractKeeper, s.OsmosisApp.AccountKeeper, s.OsmosisApp.AppCodec())
}

func (s *CosmwasmAuthenticatorTest) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

func (s *CosmwasmAuthenticatorTest) TestOnAuthenticatorAdded() {

	// Generate a private key for signing
	priv := secp256k1.GenPrivKey()

	// Set up the contract
	s.StoreContractCode("../testutils/bytecode/echo.wasm")
	instantiateMsg := EchoInstantiateMsg{PubKey: priv.PubKey().Bytes()}
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	contractAddr := s.InstantiateContract(string(instantiateMsgBz), 1)

	// create new account
	acc := apptesting.CreateRandomAccounts(1)[0]

	tests := []struct {
		name string // name
		data []byte // initData
		pass bool   // wantErr
	}{
		{"Valid Contract, valid params", []byte(fmt.Sprintf(`{"contract": "%s", "params": %s }`, contractAddr, toBytesString(`{ "label": "test" }`))), true},
		{"Valid Contract, unexpected params", []byte(fmt.Sprintf(`{"contract": "%s", "params": %s }`, contractAddr, toBytesString(`{ "unexpected": "json" }`))), false},
		{"Valid Contract, malform json params", []byte(fmt.Sprintf(`{"contract": "%s", "params": %s }`, contractAddr, toBytesString(`{ malform json }`))), false},
		{"Valid Contract, missing authenticator params (required by contract)", []byte(fmt.Sprintf(`{"contract": "%s"}`, contractAddr)), false},
		{"Missing Contract", []byte(`{}`), false},
		{"Invalid Contract Address", []byte(`{"contract": "invalid_address"}`), false},
		{"Valid address but non-existing contract", []byte(`{"contract": "osmo175dck737jmvr9mw34pqs7y5fv0umnak3vrsj3mjxg75cnkmyulfs0c3sxr"}`), false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.CosmwasmAuth.OnAuthenticatorAdded(s.Ctx.WithBlockTime(time.Now()), acc, tt.data, "1")

			if tt.pass {
				s.Require().NoError(err, "Should succeed")

				msg := s.QueryLatestSudoCall(contractAddr)

				// unmashal the initData as CosmWasmAuthenticatorInitData
				var initData authenticator.CosmwasmAuthenticatorInitData
				err = json.Unmarshal(tt.data, &initData)
				s.Require().NoError(err, "Should unmarshall data successfully")

				expectedMsg := authenticator.SudoMsg{
					OnAuthenticatorAdded: &authenticator.OnAuthenticatorAddedRequest{
						Account:             acc,
						AuthenticatorParams: initData.Params,
						AuthenticatorId:     "1",
					},
				}

				s.Require().Equal(expectedMsg, msg, "Should match latest sudo msg")
			} else {
				s.Require().Error(err, "Should fail")
			}
		})
	}
}

func (s *CosmwasmAuthenticatorTest) TestOnAuthenticatorRemoved() {

	// Generate a private key for signing
	priv := secp256k1.GenPrivKey()

	// Set up the contract
	s.StoreContractCode("../testutils/bytecode/echo.wasm")
	instantiateMsg := EchoInstantiateMsg{PubKey: priv.PubKey().Bytes()}
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	contractAddr := s.InstantiateContract(string(instantiateMsgBz), 1)

	// create new account
	acc := apptesting.CreateRandomAccounts(1)[0]

	tests := []struct {
		name string // name
		data []byte // initData
		pass bool   // wantErr
	}{
		{"Valid Contract, valid params", []byte(fmt.Sprintf(`{"contract": "%s", "params": %s }`, contractAddr, toBytesString(`{ "label": "test" }`))), true},
		{"Valid Contract, unexpected params", []byte(fmt.Sprintf(`{"contract": "%s", "params": %s }`, contractAddr, toBytesString(`{ "unexpected": "json" }`))), false},
		{"Valid Contract, malform json params", []byte(fmt.Sprintf(`{"contract": "%s", "params": %s }`, contractAddr, toBytesString(`{ malform json }`))), false},
		{"Valid Contract, missing authenticator params (required by contract)", []byte(fmt.Sprintf(`{"contract": "%s"}`, contractAddr)), false},
		{"Missing Contract", []byte(`{}`), false},
		{"Invalid Contract Address", []byte(`{"contract": "invalid_address"}`), false},
		{"Valid address but non-existing contract", []byte(`{"contract": "osmo175dck737jmvr9mw34pqs7y5fv0umnak3vrsj3mjxg75cnkmyulfs0c3sxr"}`), false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.CosmwasmAuth.OnAuthenticatorRemoved(s.Ctx.WithBlockTime(time.Now()), acc, tt.data, "1")
			if tt.pass {
				s.Require().NoError(err, "Should succeed")

				msg := s.QueryLatestSudoCall(contractAddr)

				// unmashal the initData as CosmWasmAuthenticatorInitData
				var initData authenticator.CosmwasmAuthenticatorInitData
				err = json.Unmarshal(tt.data, &initData)
				s.Require().NoError(err, "Should unmarshall data successfully")

				expectedMsg := authenticator.SudoMsg{
					OnAuthenticatorRemoved: &authenticator.OnAuthenticatorRemovedRequest{
						Account:             acc,
						AuthenticatorParams: initData.Params,
						AuthenticatorId:     "1",
					},
				}

				s.Require().Equal(expectedMsg, msg, "Should match latest sudo msg")
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
	s.StoreContractCode("../testutils/bytecode/echo.wasm")
	instantiateMsg := EchoInstantiateMsg{PubKey: priv.PubKey().Bytes()}
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	addr := s.InstantiateContract(string(instantiateMsgBz), 1)

	params := `{ "label": "test" }`
	initData := []byte(fmt.Sprintf(`{"contract": "%s", "params": %s}`, addr, toBytesString(params)))
	err = s.CosmwasmAuth.OnAuthenticatorAdded(s.Ctx.WithBlockTime(time.Now()), accounts[0], initData, "1")
	s.Require().NoError(err, "OnAuthenticator added should succeed")

	msg := s.QueryLatestSudoCall(addr)
	s.Require().Equal(authenticator.SudoMsg{
		OnAuthenticatorAdded: &authenticator.OnAuthenticatorAddedRequest{
			Account:             accounts[0],
			AuthenticatorParams: []byte(params),
			AuthenticatorId:     "1",
		},
	}, msg, "Should match latest sudo msg ")

	auth, err := s.CosmwasmAuth.Initialize(initData)
	s.Require().NoError(err, "Initialize should succeed")

	tx, _ := GenTx(
		s.Ctx,
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

	ak := s.OsmosisApp.AccountKeeper
	sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
	request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, s.OsmosisApp.AppCodec(), ak, sigModeHandler, accounts[0], accounts[0], nil, feeCoins, testMsg, tx, 0, false, authenticator.SequenceMatch)
	s.Require().NoError(err)
	request.AuthenticatorId = "0"

	// Test with valid signature
	err = auth.Authenticate(s.Ctx.WithBlockTime(time.Now()), request)
	s.Require().NoError(err, "Should be authenticated")

	msg = s.QueryLatestSudoCall(addr)
	request.AuthenticatorParams = []byte(params)
	request.FeeGranter = sdk.AccAddress{}
	s.Require().Equal(authenticator.SudoMsg{
		Authenticate: &request,
	}, msg, "Should match latest sudo msg ")

	err = auth.Track(s.Ctx.WithBlockTime(time.Now()), request)
	s.Require().NoError(err, "Track should succeed")

	encodedMsg, err := codectypes.NewAnyWithValue(testMsg)
	s.Require().NoError(err, "Should encode Any value successfully")

	msg = s.QueryLatestSudoCall(addr)
	s.Require().Equal(authenticator.SudoMsg{
		Track: &authenticator.TrackRequest{
			AuthenticatorId: "0",
			Account:         accounts[0],
			FeePayer:        accounts[0],
			FeeGranter:      sdk.AccAddress{},
			Fee:             feeCoins,
			Msg: authenticator.LocalAny{
				TypeURL: encodedMsg.TypeUrl,
				Value:   encodedMsg.Value,
			},
			AuthenticatorParams: []byte(params),
		},
	}, msg, "Should match latest sudo msg")

	err = auth.ConfirmExecution(s.Ctx.WithBlockTime(time.Now()), request)
	s.Require().NoError(err, "Execution should be confirmed")

	msg = s.QueryLatestSudoCall(addr)
	s.Require().Equal(authenticator.SudoMsg{
		ConfirmExecution: &authenticator.ConfirmExecutionRequest{
			AuthenticatorId: "0",
			Account:         accounts[0],
			FeePayer:        accounts[0],
			FeeGranter:      sdk.AccAddress{},
			Fee:             feeCoins,
			Msg: authenticator.LocalAny{
				TypeURL: encodedMsg.TypeUrl,
				Value:   encodedMsg.Value,
			},
			AuthenticatorParams: []byte(params),
		},
	}, msg, "Should match latest sudo msg")

	// Test with an invalid signature
	request.Signature = []byte("invalid")
	err = auth.Authenticate(s.Ctx.WithBlockTime(time.Now()), request)
	s.Require().Error(err, "Should not be authenticated")
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
	s.StoreContractCode("../testutils/bytecode/cosigner_authenticator.wasm")
	instantiateMsg := CosignerInstantiateMsg{PubKeys: [][]byte{priv.PubKey().Bytes()}}
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	addr := s.InstantiateContract(string(instantiateMsgBz), 1)

	auth, err := s.CosmwasmAuth.Initialize([]byte(
		fmt.Sprintf(`{"contract": "%s"}`, addr)))
	s.Require().NoError(err, "Should succeed")

	tx, _ := GenTx(
		s.Ctx,
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
	s.T().Skip("TODO: this currently fails as signatures are stripped from the tx. Should we add them or maybe do a better cosigner implementation later?")
	ak := s.OsmosisApp.AccountKeeper
	sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()
	request, err := authenticator.GenerateAuthenticationRequest(s.Ctx, s.OsmosisApp.AppCodec(), ak, sigModeHandler, accounts[0], accounts[0], nil, sdk.NewCoins(), testMsg, tx, 0, false, authenticator.SequenceMatch)
	s.Require().NoError(err)

	status := auth.Authenticate(s.Ctx.WithBlockTime(time.Now()), request)
	fmt.Println(status)
	//TODO: review this after full refactor
	//s.Require().True(status.IsAuthenticated(), "Should be authenticated")

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

func (s *CosmwasmAuthenticatorTest) QueryContract(msg string, contractAddr sdk.AccAddress) []byte {
	// Query the contract
	osmosisApp := s.OsmosisApp
	res, err := osmosisApp.WasmKeeper.QuerySmart(s.Ctx.WithBlockTime(time.Now()), contractAddr, []byte(msg))
	s.Require().NoError(err)

	return res
}

func (s *CosmwasmAuthenticatorTest) QueryLatestSudoCall(contractAddr sdk.AccAddress) authenticator.SudoMsg {
	// Query the contract
	osmosisApp := s.OsmosisApp
	res, err := osmosisApp.WasmKeeper.QuerySmart(s.Ctx.WithBlockTime(time.Now()), contractAddr, []byte(`{"latest_sudo_call": {}}`))
	s.Require().NoError(err)

	// unmarshal the call as SudoMsg
	msg := authenticator.SudoMsg{}
	err = json.Unmarshal(res, &msg)
	s.Require().NoError(err, "Should unmarshall latest sudo msg successfully")

	return msg
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
