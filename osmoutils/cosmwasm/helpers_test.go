package cosmwasm_test

import (
	"os"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// uploadAndInstantiateContract is a helper function to upload and instantiate a contract from a given file path.
// It calls an empty Instantiate message on the created contract and returns the bech32 address after instantiation.
func (s *KeeperTestSuite) uploadAndInstantiateContract(filePath string) (rawCWAddr sdk.AccAddress, bech32CWAddr string) {
	// We use a gov permissioned contract keeper to avoid having to manually set permissions
	contractKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)

	// Upload and instantiate wasm code
	wasmCode, err := os.ReadFile(filePath)
	s.Require().NoError(err)
	codeID, _, err := contractKeeper.Create(s.Ctx, s.TestAccs[0], wasmCode, nil)
	s.Require().NoError(err)
	rawCWAddr, _, err = contractKeeper.Instantiate(s.Ctx, codeID, s.TestAccs[0], s.TestAccs[0], []byte("{}"), "", sdk.NewCoins())
	s.Require().NoError(err)
	bech32CWAddr, err = sdk.Bech32ifyAddressBytes("osmo", rawCWAddr)
	s.Require().NoError(err)

	return rawCWAddr, bech32CWAddr
}

func (s *KeeperTestSuite) TestSudoHelper() {
	// Skip test if there is system-side incompatibility
	s.SkipIfWSL()

	// We use contracts already defined in existing modules to avoid duplicate test contract code.
	// This is a simple counter contract that counts `Amount` times and does a state write on each iteration.
	// Source code can be found in x/concentrated-liquidity/testcontracts/contract-sources
	counterContractPath := "../../x/concentrated-liquidity/testcontracts/compiled-wasm/counter.wasm"

	// Message structs for the test CW contract
	type CountMsg struct {
		Amount int64 `json:"amount"`
	}
	type CountSudoMsg struct {
		Count CountMsg `json:"count"`
	}

	tests := map[string]struct {
		wasmFile      string
		msg           CountSudoMsg
		noContractSet bool

		expectedError error
	}{
		"valid contract that consumes less than limit": {
			wasmFile: counterContractPath,
			msg: CountSudoMsg{
				Count: CountMsg{
					// Consumes roughly 100k gas, which should be comfortably under the limit.
					Amount: 10,
				},
			},
		},
		// "no contract set in state": {
		// 	wasmFile: counterContractPath,
		// 	msg: CountSudoMsg{
		// 		Count: CountMsg{
		// 			// Consumes roughly 100k gas, which should be comfortably under the limit.
		// 			Amount: 10,
		// 		},
		// 	},

		// 	// We expect this to fail quietly and be a no-op
		// 	noContractSet: true,
		// },
		// "empty message": {
		// 	wasmFile: counterContractPath,
		// 	// We expect this to be a no-op without error
		// 	msg: CountSudoMsg{},
		// },
		// "error: contract that consumes more than limit": {
		// 	wasmFile: counterContractPath,
		// 	msg: CountSudoMsg{
		// 		Count: CountMsg{
		// 			// Each loop in the contract consumes on the order of 1k-10k gas,
		// 			// so this should push consumed gas over the limit.
		// 			Amount: int64(types.DefaultContractHookGasLimit) / 1000,
		// 		},
		// 	},

		// 	expectedError: types.ContractHookOutOfGasError{GasLimit: types.DefaultContractHookGasLimit},
		// },
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.Setup()
			// Upload and instantiate wasm code
			_, cosmwasmAddressBech32 := s.uploadAndInstantiateContract(tc.wasmFile)

			if tc.expectedError != nil {
				s.Require().ErrorIs(err, tc.expectedError)
				return
			}

			s.Require().NoError(err)
		})
	}
}
