package cosmwasmpool_test

import (
	"fmt"
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
)

const (
	denomA        = apptesting.DefaultTransmuterDenomA
	denomB        = apptesting.DefaultTransmuterDenomB
	validCodeId   = uint64(1)
	invalidCodeId = validCodeId + 1
	defaultPoolId = uint64(1)
	nonZeroFeeStr = "0.01"
)

type PoolModuleSuite struct {
	apptesting.KeeperTestHelper
}

var (
	defaultAmount       = osmomath.NewInt(100)
	initalDefaultSupply = sdk.NewCoins(sdk.NewCoin(denomA, defaultAmount), sdk.NewCoin(denomB, defaultAmount))

	defaultDenoms = []string{denomA, denomB}
)

func TestPoolModuleSuite(t *testing.T) {
	suite.Run(t, new(PoolModuleSuite))
}

func (s *PoolModuleSuite) TestInitializePool() {
	validInstantitateMsg := s.GetTransmuterInstantiateMsgBytes(defaultDenoms, nil, sdk.AccAddress{})

	tests := map[string]struct {
		codeid            uint64
		instantiateMsg    []byte
		isInvalidPoolType bool
		isWhitelisted     bool
		expectError       bool
	}{
		"valid pool, whitelisted": {
			codeid:         validCodeId,
			instantiateMsg: validInstantitateMsg,
			isWhitelisted:  true,
		},
		"valid pool, not whitelisted": {
			codeid:         validCodeId,
			instantiateMsg: validInstantitateMsg,
			isWhitelisted:  false,
			expectError:    true,
		},
		"error: invalid code id": {
			codeid:         invalidCodeId,
			instantiateMsg: validInstantitateMsg,
			isWhitelisted:  true,
			expectError:    true,
		},
		"invalid pool type": {
			isInvalidPoolType: true,
			isWhitelisted:     true,
			expectError:       true,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.Setup()
			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			s.StoreCosmWasmPoolContractCode(apptesting.TransmuterContractName)

			var testPool poolmanagertypes.PoolI
			if !tc.isInvalidPoolType {
				testPool = model.NewCosmWasmPool(defaultPoolId, tc.codeid, tc.instantiateMsg)
			} else {
				testPool = s.PrepareConcentratedPool()
			}

			if tc.isWhitelisted {
				s.App.CosmwasmPoolKeeper.WhitelistCodeId(s.Ctx, tc.codeid)
			}

			err := cosmwasmPoolKeeper.InitializePool(s.Ctx, testPool, s.TestAccs[0])

			if tc.expectError {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			pool, err := cosmwasmPoolKeeper.GetPoolById(s.Ctx, defaultPoolId)
			s.Require().NoError(err)

			cosmWasmPool, ok := pool.(*model.Pool)
			s.Require().True(ok)

			// Check that the pool's contract address is set
			cwPoolAddress := cosmWasmPool.GetContractAddress()
			_, err = sdk.AccAddressFromBech32(cwPoolAddress)
			s.Require().NoError(err)

			// Validate the pool's instantiate msg
			s.Require().Equal(validCodeId, cosmWasmPool.GetCodeId())

			// Validate pool id
			s.Require().Equal(defaultPoolId, cosmWasmPool.GetId())

			// Validate that the wasm keeper is correctly set
			s.Require().Equal(s.App.WasmKeeper, cosmWasmPool.WasmKeeper)

			// Validate that the pool's instantiate msg is set
			s.Require().Equal(tc.instantiateMsg, cosmWasmPool.GetInstantiateMsg())
		})
	}
}

func (s *PoolModuleSuite) TestGetPoolDenoms() {
	tests := map[string]struct {
		poolId         uint64
		expectedDenoms []string
		expectError    error
	}{
		"happy path": {
			poolId:         defaultPoolId,
			expectedDenoms: defaultDenoms,
		},
		"error: invalid poold id": {
			poolId:         defaultPoolId + 1,
			expectedDenoms: defaultDenoms,
			expectError:    types.PoolNotFoundError{PoolId: defaultPoolId + 1},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.Setup()
			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			s.PrepareCosmWasmPool()

			denoms, err := cosmwasmPoolKeeper.GetPoolDenoms(s.Ctx, tc.poolId)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(defaultDenoms, denoms)
		})
	}
}

func (s *PoolModuleSuite) TestCalcOutAmtGivenIn_SwapOutAmtGivenIn() {
	tests := map[string]struct {
		initialCoins      sdk.Coins
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount osmomath.Int
		swapFee           osmomath.Dec
		isInvalidPool     bool

		expectedTokenOut     sdk.Coin
		expectedErrorMessage string
	}{
		"calc amount less than supply": {
			initialCoins:     initalDefaultSupply,
			tokenIn:          sdk.NewCoin(denomA, defaultAmount.Sub(osmomath.OneInt())),
			tokenOutDenom:    denomB,
			expectedTokenOut: sdk.NewCoin(denomB, defaultAmount.Sub(osmomath.OneInt())),
			swapFee:          osmomath.ZeroDec(),
		},
		"calc amount equal to supply": {
			initialCoins:     initalDefaultSupply,
			tokenIn:          sdk.NewCoin(denomA, defaultAmount),
			tokenOutDenom:    denomB,
			expectedTokenOut: sdk.NewCoin(denomB, defaultAmount),
			swapFee:          osmomath.ZeroDec(),
		},
		"calc amount greater than supply": {
			initialCoins:         initalDefaultSupply,
			tokenIn:              sdk.NewCoin(denomA, defaultAmount.Add(osmomath.OneInt())),
			tokenOutDenom:        denomB,
			expectedErrorMessage: fmt.Sprintf("Insufficient pool asset: required: %s, available: %s", sdk.NewCoin(denomB, defaultAmount.Add(osmomath.OneInt())), sdk.NewCoin(denomB, defaultAmount)),
		},
		"non-zero swap fee": {
			initialCoins:         initalDefaultSupply,
			tokenIn:              sdk.NewCoin(denomA, defaultAmount.Sub(osmomath.OneInt())),
			tokenOutDenom:        denomB,
			swapFee:              osmomath.MustNewDecFromStr(nonZeroFeeStr),
			expectedErrorMessage: fmt.Sprintf("Invalid swap fee: expected: %s, actual: %s", osmomath.ZeroInt(), nonZeroFeeStr),
		},
		"invalid pool given": {
			initialCoins:  sdk.NewCoins(sdk.NewCoin(denomA, defaultAmount), sdk.NewCoin(denomB, defaultAmount)),
			tokenIn:       sdk.NewCoin(denomA, defaultAmount.Sub(osmomath.OneInt())),
			tokenOutDenom: denomB,
			isInvalidPool: true,

			expectedErrorMessage: types.InvalidPoolTypeError{
				ActualPool: &clmodel.Pool{},
			}.Error(),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.Setup()

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			s.FundAcc(s.TestAccs[0], tc.initialCoins)
			// get initial denom from coins specified in the test case
			initialDenoms := []string{}
			for _, coin := range tc.initialCoins {
				initialDenoms = append(initialDenoms, coin.Denom)
			}

			// create pool
			pool := s.PrepareCustomTransmuterPool(s.TestAccs[0], initialDenoms)

			// add liquidity by joining the pool
			s.JoinTransmuterPool(s.TestAccs[0], pool.GetId(), tc.initialCoins)

			originalPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(pool.GetContractAddress()))

			var poolIn poolmanagertypes.PoolI = pool
			if tc.isInvalidPool {
				poolIn = s.PrepareConcentratedPool()
			}

			// system under test non-mutative.
			actualCalcTokenOut, err := cosmwasmPoolKeeper.CalcOutAmtGivenIn(s.Ctx, poolIn, tc.tokenIn, tc.tokenOutDenom, tc.swapFee)
			if tc.expectedErrorMessage != "" {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErrorMessage)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedTokenOut, actualCalcTokenOut)
			}

			// Assert that pool balances are unchanged
			afterCalcPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(pool.GetContractAddress()))

			s.Require().Equal(originalPoolBalances.String(), afterCalcPoolBalances.String())

			swapper := s.TestAccs[1]
			// fund swapper
			s.FundAcc(swapper, sdk.NewCoins(tc.tokenIn))

			// system under test non-mutative.
			actualSwapTokenOut, err := cosmwasmPoolKeeper.SwapExactAmountIn(s.Ctx, swapper, poolIn, tc.tokenIn, tc.tokenOutDenom, tc.tokenOutMinAmount, tc.swapFee)
			if tc.expectedErrorMessage != "" {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErrorMessage)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedTokenOut.Amount, actualSwapTokenOut)

			// Assert that pool balance is updated correctly
			expectedPoolBalances := originalPoolBalances.Add(tc.tokenIn).Sub(sdk.NewCoins(tc.expectedTokenOut)...)
			afterSwapPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(pool.GetContractAddress()))
			s.Require().Equal(expectedPoolBalances.String(), afterSwapPoolBalances.String())

			// Assert that swapper balance is updated correctly
			expectedSwapperBalances := sdk.NewCoins(tc.expectedTokenOut)
			afterSwapSwapperBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, swapper)
			s.Require().Equal(expectedSwapperBalances.String(), afterSwapSwapperBalances.String())
		})
	}
}

func (s *PoolModuleSuite) TestCalcInAmtGivenOut_SwapInAmtGivenOut() {
	tests := map[string]struct {
		initialCoins     sdk.Coins
		tokenOut         sdk.Coin
		tokenInDenom     string
		tokenInMaxAmount osmomath.Int
		swapFee          osmomath.Dec
		isInvalidPool    bool

		expectedTokenIn      sdk.Coin
		expectedErrorMessage string
	}{
		"calc amount less than supply": {
			initialCoins:     initalDefaultSupply,
			tokenOut:         sdk.NewCoin(denomA, defaultAmount.Sub(osmomath.OneInt())),
			tokenInDenom:     denomB,
			expectedTokenIn:  sdk.NewCoin(denomB, defaultAmount.Sub(osmomath.OneInt())),
			tokenInMaxAmount: defaultAmount,
			swapFee:          osmomath.ZeroDec(),
		},
		"calc amount equal to supply": {
			initialCoins:     initalDefaultSupply,
			tokenOut:         sdk.NewCoin(denomA, defaultAmount),
			tokenInDenom:     denomB,
			expectedTokenIn:  sdk.NewCoin(denomB, defaultAmount),
			tokenInMaxAmount: defaultAmount,
			swapFee:          osmomath.ZeroDec(),
		},
		"calc amount greater than supply": {
			initialCoins:         initalDefaultSupply,
			tokenOut:             sdk.NewCoin(denomA, defaultAmount.Add(osmomath.OneInt())),
			tokenInDenom:         denomB,
			tokenInMaxAmount:     defaultAmount.Sub(osmomath.OneInt()),
			expectedErrorMessage: fmt.Sprintf("Insufficient pool asset: required: %s, available: %s", sdk.NewCoin(denomA, defaultAmount.Add(osmomath.OneInt())), sdk.NewCoin(denomA, defaultAmount)),
		},
		"non-zero swap fee": {
			initialCoins:         initalDefaultSupply,
			tokenOut:             sdk.NewCoin(denomA, defaultAmount.Sub(osmomath.OneInt())),
			tokenInDenom:         denomB,
			tokenInMaxAmount:     defaultAmount.Sub(osmomath.OneInt()),
			swapFee:              osmomath.MustNewDecFromStr(nonZeroFeeStr),
			expectedErrorMessage: fmt.Sprintf("Invalid swap fee: expected: %s, actual: %s", osmomath.ZeroInt(), nonZeroFeeStr),
		},
		"invalid pool given": {
			initialCoins:     sdk.NewCoins(sdk.NewCoin(denomA, defaultAmount), sdk.NewCoin(denomB, defaultAmount)),
			tokenOut:         sdk.NewCoin(denomA, defaultAmount.Sub(osmomath.OneInt())),
			tokenInDenom:     denomB,
			tokenInMaxAmount: defaultAmount.Sub(osmomath.OneInt()),
			isInvalidPool:    true,

			expectedErrorMessage: types.InvalidPoolTypeError{
				ActualPool: &clmodel.Pool{},
			}.Error(),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.Setup()

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			// fund pool joiner
			s.FundAcc(s.TestAccs[0], tc.initialCoins)

			// get initial denom from coins specified in the test case
			initialDenoms := []string{}
			for _, coin := range tc.initialCoins {
				initialDenoms = append(initialDenoms, coin.Denom)
			}

			// create pool
			pool := s.PrepareCustomTransmuterPool(s.TestAccs[0], initialDenoms)

			// add liquidity by joining the pool
			s.JoinTransmuterPool(s.TestAccs[0], pool.GetId(), tc.initialCoins)

			originalPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(pool.GetContractAddress()))

			var poolIn poolmanagertypes.PoolI = pool
			if tc.isInvalidPool {
				poolIn = s.PrepareConcentratedPool()
			}

			// system under test non-mutative.
			actualCalcTokenOut, err := cosmwasmPoolKeeper.CalcInAmtGivenOut(s.Ctx, poolIn, tc.tokenOut, tc.tokenInDenom, tc.swapFee)
			if tc.expectedErrorMessage != "" {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErrorMessage)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedTokenIn, actualCalcTokenOut)
			}

			// Assert that pool balances are unchanged
			afterCalcPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(pool.GetContractAddress()))

			s.Require().Equal(originalPoolBalances.String(), afterCalcPoolBalances.String())

			swapper := s.TestAccs[1]

			// fund swapper
			s.FundAcc(swapper, sdk.NewCoins(sdk.NewCoin(tc.tokenInDenom, tc.tokenInMaxAmount)))

			beforeSwapSwapperBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, swapper)

			// system under test non-mutative.
			actualSwapTokenIn, err := cosmwasmPoolKeeper.SwapExactAmountOut(s.Ctx, swapper, poolIn, tc.tokenInDenom, tc.tokenInMaxAmount, tc.tokenOut, tc.swapFee)
			if tc.expectedErrorMessage != "" {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErrorMessage)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedTokenIn.Amount, actualSwapTokenIn)

			// Assert that pool balance is updated correctly
			expectedPoolBalances := originalPoolBalances.Add(tc.expectedTokenIn).Sub(sdk.NewCoins(tc.tokenOut)...)
			afterSwapPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(pool.GetContractAddress()))
			s.Require().Equal(expectedPoolBalances.String(), afterSwapPoolBalances.String())

			// Assert that swapper balance is updated correctly
			expectedSwapperBalances := beforeSwapSwapperBalances.Sub(sdk.NewCoins(tc.expectedTokenIn)...).Add(tc.tokenOut)
			afterSwapSwapperBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, swapper)
			s.Require().Equal(expectedSwapperBalances.String(), afterSwapSwapperBalances.String())
		})
	}
}

func (s *PoolModuleSuite) TestGetTotalPoolLiquidity() {
	tests := map[string]struct {
		poolId               uint64
		initialCoins         sdk.Coins
		expectedErrorMessage string
	}{
		"happy path": {
			poolId:       defaultPoolId,
			initialCoins: initalDefaultSupply,
		},
		"unhappy path: invalid pool id": {
			poolId:       defaultPoolId + 1,
			initialCoins: initalDefaultSupply,

			expectedErrorMessage: types.PoolNotFoundError{
				PoolId: defaultPoolId + 1,
			}.Error(),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.Setup()

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			// fund pool joiner
			s.FundAcc(s.TestAccs[0], tc.initialCoins)

			// get initial denom from coins specified in the test case
			initialDenoms := []string{}
			for _, coin := range tc.initialCoins {
				initialDenoms = append(initialDenoms, coin.Denom)
			}

			// create pool
			pool := s.PrepareCustomTransmuterPool(s.TestAccs[0], initialDenoms)

			// add liquidity by joining the pool
			s.JoinTransmuterPool(s.TestAccs[0], pool.GetId(), tc.initialCoins)

			// system under test.
			actualLiquidity, err := cosmwasmPoolKeeper.GetTotalPoolLiquidity(s.Ctx, tc.poolId)
			if tc.expectedErrorMessage != "" {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedErrorMessage)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.initialCoins, actualLiquidity)
		})
	}
}

func (s *PoolModuleSuite) TestSudoGasLimit() {
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
	type CountMsgResponse struct {
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
		"contract consumes less than limit": {
			wasmFile: counterContractPath,
			msg: CountSudoMsg{
				Count: CountMsg{
					// Consumes roughly 100k gas, which should be comfortably under the limit.
					Amount: 10,
				},
			},
		},
		"contract that consumes more than limit": {
			wasmFile: counterContractPath,
			msg: CountSudoMsg{
				Count: CountMsg{
					// Consumes roughly 1B gas, which is well above the 30M limit.
					Amount: 100000,
				},
			},
			expectedError: fmt.Errorf("contract call ran out of gas"),
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.Setup()

			// We use a gov permissioned contract keeper to avoid having to manually set permissions
			contractKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)

			// Upload and instantiate wasm code
			_, cosmwasmAddressBech32 := s.uploadAndInstantiateContract(contractKeeper, tc.wasmFile)

			// System under test
			response, err := cosmwasm.Sudo[CountSudoMsg, CountMsgResponse](s.Ctx, contractKeeper, cosmwasmAddressBech32, tc.msg)

			if tc.expectedError != nil {
				s.Require().ErrorContains(err, tc.expectedError.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(CountMsgResponse{}, response)
		})
	}
}

// uploadAndInstantiateContract is a helper function to upload and instantiate a contract from a given file path.
// It calls an empty Instantiate message on the created contract and returns the bech32 address after instantiation.
func (s *PoolModuleSuite) uploadAndInstantiateContract(contractKeeper *wasmkeeper.PermissionedKeeper, filePath string) (rawCWAddr sdk.AccAddress, bech32CWAddr string) {
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
