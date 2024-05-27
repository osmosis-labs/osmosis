package cosmwasm_test

// TESTS MOVED DIRECTLY TO x/cosmwasmpool/pool_module_test.go to prevent circular imports (specifically on osmosis app for the test suite)
// import (
// 	"fmt"
// 	"os"
// 	"testing"

// 	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/suite"

// 	"github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
// 	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
// )

// type KeeperTestSuite struct {
// 	apptesting.KeeperTestHelper
// }

// func TestKeeperTestSuite(t *testing.T) {
// 	suite.Run(t, new(KeeperTestSuite))
// }

// func (s *KeeperTestSuite) TestSudoGasLimit() {
// 	// Skip test if there is system-side incompatibility
// 	s.SkipIfWSL()

// 	// We use contracts already defined in existing modules to avoid duplicate test contract code.
// 	// This is a simple counter contract that counts `Amount` times and does a state write on each iteration.
// 	// Source code can be found in x/concentrated-liquidity/testcontracts/contract-sources
// 	counterContractPath := "../../x/concentrated-liquidity/testcontracts/compiled-wasm/counter.wasm"

// 	// Message structs for the test CW contract
// 	type CountMsg struct {
// 		Amount int64 `json:"amount"`
// 	}
// 	type CountMsgResponse struct {
// 	}
// 	type CountSudoMsg struct {
// 		Count CountMsg `json:"count"`
// 	}

// 	tests := map[string]struct {
// 		wasmFile      string
// 		msg           CountSudoMsg
// 		noContractSet bool

// 		expectedError error
// 	}{
// 		"contract consumes less than limit": {
// 			wasmFile: counterContractPath,
// 			msg: CountSudoMsg{
// 				Count: CountMsg{
// 					// Consumes roughly 100k gas, which should be comfortably under the limit.
// 					Amount: 10,
// 				},
// 			},
// 		},
// 		"contract that consumes more than limit": {
// 			wasmFile: counterContractPath,
// 			msg: CountSudoMsg{
// 				Count: CountMsg{
// 					// Consumes roughly 1B gas, which is well above the 30M limit.
// 					Amount: 100000,
// 				},
// 			},
// 			expectedError: fmt.Errorf("contract call ran out of gas"),
// 		},
// 	}
// 	for name, tc := range tests {
// 		s.Run(name, func() {
// 			s.Setup()

// 			// We use a gov permissioned contract keeper to avoid having to manually set permissions
// 			contractKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)

// 			// Upload and instantiate wasm code
// 			_, cosmwasmAddressBech32 := s.uploadAndInstantiateContract(contractKeeper, tc.wasmFile)

// 			// System under test
// 			response, err := cosmwasm.Sudo[CountSudoMsg, CountMsgResponse](s.Ctx, contractKeeper, cosmwasmAddressBech32, tc.msg)

// 			if tc.expectedError != nil {
// 				s.Require().ErrorContains(err, tc.expectedError.Error())
// 				return
// 			}

// 			s.Require().NoError(err)
// 			s.Require().Equal(CountMsgResponse{}, response)
// 		})
// 	}
// }

// // uploadAndInstantiateContract is a helper function to upload and instantiate a contract from a given file path.
// // It calls an empty Instantiate message on the created contract and returns the bech32 address after instantiation.
// func (s *KeeperTestSuite) uploadAndInstantiateContract(contractKeeper *wasmkeeper.PermissionedKeeper, filePath string) (rawCWAddr sdk.AccAddress, bech32CWAddr string) {
// 	// Upload and instantiate wasm code
// 	wasmCode, err := os.ReadFile(filePath)
// 	s.Require().NoError(err)
// 	codeID, _, err := contractKeeper.Create(s.Ctx, s.TestAccs[0], wasmCode, nil)
// 	s.Require().NoError(err)
// 	rawCWAddr, _, err = contractKeeper.Instantiate(s.Ctx, codeID, s.TestAccs[0], s.TestAccs[0], []byte("{}"), "", sdk.NewCoins())
// 	s.Require().NoError(err)
// 	bech32CWAddr, err = sdk.Bech32ifyAddressBytes("osmo", rawCWAddr)
// 	s.Require().NoError(err)

// 	return rawCWAddr, bech32CWAddr
// }
