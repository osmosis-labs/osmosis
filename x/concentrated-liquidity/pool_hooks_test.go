package concentrated_liquidity_test

import (
	"encoding/json"
	"os"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

var (
	validCosmwasmAddress   = "osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9"
	invalidCosmwasmAddress = "osmo1{}{}4hj2tfpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9"
	validActionPrefix      = "beforeSwapExactAmountIn"
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
		"error: invalid hook action": {
			cosmwasmAddress: invalidCosmwasmAddress,
			actionPrefix:    "invalidActionPrefix",
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
			_, cosmwasmAddressBech32 := s.uploadAndInstantiateContract(tc.wasmFile)

			// Set pool hook contract to the newly instantiated contract
			err := s.Clk.SetPoolHookContract(s.Ctx, validPoolId, validActionPrefix, cosmwasmAddressBech32)
			s.Require().NoError(err)

			// Marshal test case msg to pass into contract
			msgBuilderFn := func(uint64) ([]byte, error) {
				msgBz, err := json.Marshal(tc.msg)
				s.Require().NoError(err)
				return msgBz, nil
			}

			// --- System under test ---

			err = s.Clk.CallPoolActionListener(s.Ctx, msgBuilderFn, validPoolId, validActionPrefix)

			// --- Assertions ---

			if tc.expectedError != nil {
				s.Require().ErrorIs(err, tc.expectedError)
				return
			}

			s.Require().NoError(err)
		})
	}
}

// Pool hook tests
// General testing strategy:
//  1. Build a pre-defined contract that defines the following behavior for all hooks:
//     if triggered, transfer 1 token with denom corresponding to the action prefix
//     e.g. if action prefix is "beforeSwap", transfer 1 token with denom "beforeSwap"
//  2. Set this contract for all hooks defined by the test case (each case should have a list
//     of action prefixes it wants to "activate")
//  3. Run a series of actions that would trigger all the hooks (create, withdraw from, swap against a position),
//     and ensure that the correct denoms are in the account balance after each action/at the end.
//
// NOTE: we assume that set contracts have valid implementations for all hooks and that this is validated
// at the contract setting stage at a higher level of abstraction. Thus, this class of errors is not covered
// by these tests.
func (s *KeeperTestSuite) TestPoolHooks() {
	hookContractFilePath := "./testcontracts/compiled-wasm/hooks.wasm"

	allBeforeHooks := []string{
		before(types.CreatePositionPrefix),
		before(types.WithdrawPositionPrefix),
		before(types.SwapExactAmountInPrefix),
		before(types.SwapExactAmountOutPrefix),
	}

	allAfterHooks := []string{
		after(types.CreatePositionPrefix),
		after(types.WithdrawPositionPrefix),
		after(types.SwapExactAmountInPrefix),
		after(types.SwapExactAmountOutPrefix),
	}

	allHooks := append(allBeforeHooks, allAfterHooks...)

	testCases := map[string]struct {
		actionPrefixes []string
	}{
		"single hook: before create position": {
			actionPrefixes: []string{before(types.CreatePositionPrefix)},
		},
		"all before hooks": {
			actionPrefixes: allBeforeHooks,
		},
		"all after hooks": {
			actionPrefixes: allAfterHooks,
		},
		"all hooks": {
			actionPrefixes: allHooks,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			clPool := s.PrepareConcentratedPool()

			// Upload and instantiate wasm code
			rawCosmwasmAddress, cosmwasmAddressBech32 := s.uploadAndInstantiateContract(hookContractFilePath)

			// Fund the contract with tokens for all action prefixes using a helper
			for _, actionPrefix := range tc.actionPrefixes {
				s.FundAcc(rawCosmwasmAddress, sdk.NewCoins(sdk.NewCoin(actionPrefix, osmomath.NewInt(10))))
			}

			// Set the contract for all hooks as defined by tc.actionPrefixes
			for _, actionPrefix := range tc.actionPrefixes {
				// We use the bech32 address here since the set function expects it for security reasons
				err := s.Clk.SetPoolHookContract(s.Ctx, validPoolId, actionPrefix, cosmwasmAddressBech32)
				s.Require().NoError(err)
			}

			// --- Execute a series of actions that trigger all supported hooks if set ---

			// Create position
			_, positionId := s.SetupPosition(clPool.GetId(), s.TestAccs[0], DefaultCoins, types.MinInitializedTick, types.MaxTick, true)

			// Withdraw from position
			_, _, err := s.Clk.WithdrawPosition(s.Ctx, s.TestAccs[0], positionId, osmomath.NewDec(100))
			s.Require().NoError(err)

			// Execute swap (SwapExactAmountIn)
			s.FundAcc(rawCosmwasmAddress, sdk.NewCoins(sdk.NewCoin(types.SwapExactAmountInPrefix, osmomath.NewInt(10))))
			_, err = s.Clk.SwapExactAmountIn(s.Ctx, s.TestAccs[0], clPool, sdk.NewCoin(ETH, osmomath.NewInt(1)), USDC, osmomath.ZeroInt(), DefaultZeroSpreadFactor)
			s.Require().NoError(err)

			// Execute swap (SwapExactAmountOut)
			s.FundAcc(rawCosmwasmAddress, sdk.NewCoins(sdk.NewCoin(types.SwapExactAmountOutPrefix, osmomath.NewInt(10))))
			_, err = s.Clk.SwapExactAmountOut(s.Ctx, s.TestAccs[0], clPool, ETH, osmomath.NewInt(100), sdk.NewCoin(USDC, osmomath.NewInt(10)), DefaultZeroSpreadFactor)
			s.Require().NoError(err)

			// Check that each set hook was successfully triggered.
			// These assertions lean on the test construction defined in the comments for these tests.
			// In short, each hook trigger is expected to transfer 1 token with denom corresponding to the
			// action that triggered it.
			expectedTriggers := sdk.NewCoins()
			for _, actionPrefix := range tc.actionPrefixes {
				expectedTriggers = expectedTriggers.Add(sdk.NewCoin(actionPrefix, osmomath.NewInt(1)))
			}

			// Ensure that correct hooks were triggered
			balances := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			s.Require().True(expectedTriggers.DenomsSubsetOf(balances), "expected balance to include: %s, actual balances: %s", expectedTriggers, balances)

			// Derive actions that should not have been triggered
			notTriggeredActions := osmoutils.Filter(func(s string) bool { return osmoutils.Contains(tc.actionPrefixes, s) }, allHooks)

			// Ensure that hooks that weren't set weren't triggered
			for _, action := range notTriggeredActions {
				s.Require().False(osmoutils.Contains(balances, sdk.NewCoin(action, osmomath.NewInt(1))), "expected balance to not include: %s, actual balances: %s", action, balances)
			}
		})
	}
}

// Adds "before" prefix to action (helper for test readability)
func before(action string) string {
	return types.BeforeActionPrefix(action)
}

// Adds "after" prefix to action (helper for test readability)
func after(action string) string {
	return types.AfterActionPrefix(action)
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
