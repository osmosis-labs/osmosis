package concentrated_liquidity_test

import (
	"encoding/json"
	"os"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
)

var (
	validCosmwasmAddress   = "osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9"
	invalidCosmwasmAddress = "osmo1{}{}4hj2tfpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9"
	validActionPrefix      = "beforeSwap"
	counterContractPath    = "./testcontracts/compiled-wasm/counter.wasm"
)

// Message structs for the test CW contract, which is simply a counter
// that counts `Amount` times and does a state write on each iteration.
type CountMsg struct {
	Amount int64 `json:"amount"`
}
type CountSudoMsg struct {
	Count CountMsg `json:"count"`
}

// TestSetAndGetPoolHookContract tests the basic functionality of setting and getting a CW contract address for a specific hook type from state.
func (s *KeeperTestSuite) TestSetAndGetPoolHookContract() {
	tests := map[string]struct {
		cosmwasmAddress string
		actionPrefix    string
		poolId          uint64

		// We do boolean checks instead of exact error checks because any
		// expected errors would come from lower level calls that don't
		// conform to our error types.
		expectErrOnSet bool
	}{
		"basic valid test": {
			// Random correctly constructed address (we do not check contract existence at the layer)
			cosmwasmAddress: validCosmwasmAddress,
			actionPrefix:    validActionPrefix,
			poolId:          validPoolId,
		},
		"attempt to delete non-existent address": {
			// Should fail quietly and return nil
			cosmwasmAddress: "",
			actionPrefix:    validActionPrefix,
			poolId:          validPoolId,
		},
		"error: incorrectly constructed address": {
			cosmwasmAddress: invalidCosmwasmAddress,
			actionPrefix:    validActionPrefix,
			poolId:          validPoolId,

			expectErrOnSet: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// Set contract address using SetPoolHookContract
			err := s.Clk.SetPoolHookContract(s.Ctx, 1, tc.actionPrefix, tc.cosmwasmAddress)

			// If expect error on set, check here
			if tc.expectErrOnSet {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// Get contract address and ensure it was the one we set in state
			contractAddress := s.Clk.GetPoolHookContract(s.Ctx, 1, tc.actionPrefix)
			s.Require().Equal(tc.cosmwasmAddress, contractAddress)

			// Delete contract address
			err = s.Clk.SetPoolHookContract(s.Ctx, 1, tc.actionPrefix, "")
			s.Require().NoError(err)

			// Ensure contract was correctly removed from state
			contractAddress = s.Clk.GetPoolHookContract(s.Ctx, 1, tc.actionPrefix)
			s.Require().Equal("", contractAddress)
		})
	}
}

// TestCallPoolActionListener tests the high level functionality of CallPoolActionListener,
// which is the helper that calls CW contracts as part of hooks.
//
// Since the function is quite general, we only test basic flows and run a sanity check on the
// gas limit to ensure it is metered correctly and can't run unboundedly.
//
// A basic CW contract is used for testing, which simply counts up to `Amount` and does a state write on each iteration.
func (s *KeeperTestSuite) TestCallPoolActionListener() {
	// Skip test if there is system-side incompatibility
	s.SkipIfWSL()

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
		"no contract set in state": {
			wasmFile: counterContractPath,
			msg: CountSudoMsg{
				Count: CountMsg{
					// Consumes roughly 100k gas, which should be comfortably under the limit.
					Amount: 10,
				},
			},

			// We expect this to fail quietly and be a no-op
			noContractSet: true,
		},
		"empty message": {
			wasmFile: counterContractPath,
			// We expect this to be a no-op without error
			msg: CountSudoMsg{},
		},
		"error: contract that consumes more than limit": {
			wasmFile: counterContractPath,
			msg: CountSudoMsg{
				Count: CountMsg{
					// Each loop in the contract consumes on the order of 1k-10k gas,
					// so this should push consumed gas over the limit.
					Amount: int64(types.DefaultContractHookGasLimit) / 1000,
				},
			},

			expectedError: types.ContractHookOutOfGasError{GasLimit: types.DefaultContractHookGasLimit},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// --- Setup ---

			// Upload and instantiate wasm code
			cosmwasmAddressBech32 := s.uploadAndInstantiateContract(tc.wasmFile)

			// Set pool hook contract to the newly instantiated contract
			err := s.Clk.SetPoolHookContract(s.Ctx, validPoolId, validActionPrefix, cosmwasmAddressBech32)
			s.Require().NoError(err)

			// Marshal test case msg to pass into contract
			msgBz, err := json.Marshal(tc.msg)
			s.Require().NoError(err)

			// --- System under test ---

			err = s.Clk.CallPoolActionListener(s.Ctx, msgBz, validPoolId, validActionPrefix)

			// --- Assertions ---

			if tc.expectedError != nil {
				s.Require().ErrorIs(err, tc.expectedError)
				return
			}

			s.Require().NoError(err)
		})
	}
}

// uploadAndInstantiateContract is a helper function to upload and instantiate a contract from a given file path.
// It calls an empty Instantiate message on the created contract and returns the bech32 address after instantiation.
func (s *KeeperTestSuite) uploadAndInstantiateContract(filePath string) (cosmwasmAddressBech32 string) {
	// We use a gov permissioned contract keeper to avoid having to manually set permissions
	contractKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)

	// Upload and instantiate wasm code
	wasmCode, err := os.ReadFile(filePath)
	s.Require().NoError(err)
	codeID, _, err := contractKeeper.Create(s.Ctx, s.TestAccs[0], wasmCode, nil)
	s.Require().NoError(err)
	cosmwasmAddress, _, err := contractKeeper.Instantiate(s.Ctx, codeID, s.TestAccs[0], s.TestAccs[0], []byte("{}"), "", sdk.NewCoins())
	s.Require().NoError(err)
	cosmwasmAddressBech32, err = sdk.Bech32ifyAddressBytes("osmo", cosmwasmAddress)
	s.Require().NoError(err)

	return cosmwasmAddressBech32
}
