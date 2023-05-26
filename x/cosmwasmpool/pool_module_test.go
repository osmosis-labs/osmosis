package cosmwasmpool_test

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/cosmwasm/msg"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/cosmwasm/msg/transmuter"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/mocks"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	denomA = apptesting.DefaultTransmuterDenomA
	denomB = apptesting.DefaultTransmuterDenomB
)

type PoolModuleSuite struct {
	apptesting.KeeperTestHelper
}

var (
	defaultPoolId       = uint64(1)
	defaultAmount       = sdk.NewInt(100)
	initalDefaultSupply = sdk.NewCoins(sdk.NewCoin(denomA, defaultAmount), sdk.NewCoin(denomB, defaultAmount))
	nonZeroFeeStr       = "0.01"
)

func TestPoolModuleSuite(t *testing.T) {
	suite.Run(t, new(PoolModuleSuite))
}

func (suite *PoolModuleSuite) SetupTest() {
	suite.Setup()
}

func (s *PoolModuleSuite) TestInitializePool() {
	var (
		validTestPool = &model.Pool{
			CosmWasmPool: model.CosmWasmPool{
				PoolAddress:     poolmanagertypes.NewPoolAddress(defaultPoolId).String(),
				ContractAddress: "", // N.B.: to be set in InitializePool()
				PoolId:          defaultPoolId,
				CodeId:          1,
				InstantiateMsg:  []byte(nil),
			},
		}
	)

	tests := map[string]struct {
		mockInstantiateReturn struct {
			contractAddress sdk.AccAddress
			data            []byte
			err             error
		}
		isValidPool bool
		expectError error
	}{
		"valid pool": {
			isValidPool: true,
		},
		"invalid pool": {
			isValidPool: false,
			expectError: types.InvalidPoolTypeError{},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			var testPool poolmanagertypes.PoolI
			if tc.isValidPool {
				testPool = validTestPool

				mockContractKeeper := mocks.NewMockContractKeeper(ctrl)
				mockContractKeeper.EXPECT().Instantiate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tc.mockInstantiateReturn.contractAddress, tc.mockInstantiateReturn.data, tc.mockInstantiateReturn.err)
				cosmwasmPoolKeeper.SetContractKeeper(mockContractKeeper)
			} else {
				testPool = s.PrepareConcentratedPool()
			}

			err := cosmwasmPoolKeeper.InitializePool(s.Ctx, testPool, s.TestAccs[0])

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectError)
				return
			}
			s.Require().NoError(err)

			pool, err := cosmwasmPoolKeeper.GetPool(s.Ctx, defaultPoolId)
			s.Require().NoError(err)

			cosmWasmPool, ok := pool.(*model.Pool)
			s.Require().True(ok)

			// Check that the pool's contract address is set
			s.Require().Equal(tc.mockInstantiateReturn.contractAddress.String(), cosmWasmPool.GetContractAddress())

			// Check that the pool's data is set
			expectedPool := validTestPool
			expectedPool.ContractAddress = tc.mockInstantiateReturn.contractAddress.String()
			s.Require().Equal(expectedPool.CosmWasmPool, cosmWasmPool.CosmWasmPool)
		})
	}
}

func (s *PoolModuleSuite) TestGetPoolDenoms() {
	tests := map[string]struct {
		denoms          []string
		poolId          uint64
		isMockPool      bool
		mockErrorReturn error
		expectError     error
	}{
		"valid with 2 denoms": {
			denoms: []string{denomA, denomB},
			poolId: defaultPoolId,
		},
		"valid with 3 denoms": {
			denoms: []string{denomA, denomB, "third"},
			poolId: defaultPoolId,
		},
		"invalid number of denoms": {
			denoms:     []string{denomA},
			poolId:     defaultPoolId,
			isMockPool: true,
			expectError: types.InvalidLiquiditySetError{
				PoolId:     defaultPoolId,
				TokenCount: 1,
			},
		},
		"invalid pool id": {
			denoms: []string{denomA, denomB},
			poolId: defaultPoolId + 1,
			expectError: types.PoolNotFoundError{
				PoolId: defaultPoolId + 1,
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			if tc.isMockPool {
				ctrl := gomock.NewController(s.T())
				defer ctrl.Finish()

				// Setup byte return.

				liquidityReturn := sdk.NewCoins()
				for _, denom := range tc.denoms {
					liquidityReturn = liquidityReturn.Add(sdk.NewCoin(denom, sdk.NewInt(1)))
				}
				response := msg.GetTotalPoolLiquidityQueryMsgResponse{
					TotalPoolLiquidity: liquidityReturn,
				}
				bz, err := json.Marshal(response)
				s.Require().NoError(err)

				mockWasmKeeper := mocks.NewMockWasmKeeper(ctrl)
				mockWasmKeeper.EXPECT().QuerySmart(gomock.Any(), gomock.Any(), gomock.Any()).Return(bz, tc.mockErrorReturn)
				cosmwasmPoolKeeper.SetWasmKeeper(mockWasmKeeper)

				// Write dummy pool to store.
				cosmwasmPoolKeeper.SetPool(s.Ctx, &model.Pool{
					CosmWasmPool: model.CosmWasmPool{
						PoolId:          tc.poolId,
						ContractAddress: s.TestAccs[0].String(),
					},
				})
			} else {
				s.PrepareCustomTransmuterPool(s.TestAccs[0], tc.denoms, 1)
			}

			denoms, err := cosmwasmPoolKeeper.GetPoolDenoms(s.Ctx, tc.poolId)
			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.denoms, denoms)
		})
	}
}

func (s *PoolModuleSuite) TestCalcOutAmtGivenIn_SwapOutAmtGivenIn() {
	tests := map[string]struct {
		initialCoins      sdk.Coins
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount sdk.Int
		swapFee           sdk.Dec
		isInvalidPool     bool

		expectedTokenOut     sdk.Coin
		expectedErrorMessage string
	}{
		"calc amount less than supply": {
			initialCoins:     initalDefaultSupply,
			tokenIn:          sdk.NewCoin(denomA, defaultAmount.Sub(sdk.OneInt())),
			tokenOutDenom:    denomB,
			expectedTokenOut: sdk.NewCoin(denomB, defaultAmount.Sub(sdk.OneInt())),
			swapFee:          sdk.ZeroDec(),
		},
		"calc amount equal to supply": {
			initialCoins:     initalDefaultSupply,
			tokenIn:          sdk.NewCoin(denomA, defaultAmount),
			tokenOutDenom:    denomB,
			expectedTokenOut: sdk.NewCoin(denomB, defaultAmount),
			swapFee:          sdk.ZeroDec(),
		},
		"calc amount greater than supply": {
			initialCoins:         initalDefaultSupply,
			tokenIn:              sdk.NewCoin(denomA, defaultAmount.Add(sdk.OneInt())),
			tokenOutDenom:        denomB,
			expectedErrorMessage: fmt.Sprintf("Insufficient fund: required: %s, available: %s", sdk.NewCoin(denomB, defaultAmount.Add(sdk.OneInt())), sdk.NewCoin(denomB, defaultAmount)),
		},
		"non-zero swap fee": {
			initialCoins:         initalDefaultSupply,
			tokenIn:              sdk.NewCoin(denomA, defaultAmount.Sub(sdk.OneInt())),
			tokenOutDenom:        denomB,
			swapFee:              sdk.MustNewDecFromStr(nonZeroFeeStr),
			expectedErrorMessage: fmt.Sprintf("Invalid swap fee: expected: %s, actual: %s", sdk.ZeroInt(), nonZeroFeeStr),
		},
		"invalid pool given": {
			initialCoins:  sdk.NewCoins(sdk.NewCoin(denomA, defaultAmount), sdk.NewCoin(denomB, defaultAmount)),
			tokenIn:       sdk.NewCoin(denomA, defaultAmount.Sub(sdk.OneInt())),
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
			s.SetupTest()

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			// fund pool joiner
			s.FundAcc(s.TestAccs[0], tc.initialCoins)

			// get initial denom from coins specified in the test case
			initialDenoms := []string{}
			for _, coin := range tc.initialCoins {
				initialDenoms = append(initialDenoms, coin.Denom)
			}

			// create pool
			pool := s.PrepareCustomTransmuterPool(s.TestAccs[0], initialDenoms, 1)

			// add liquidity by joining the pool
			request := transmuter.JoinPoolExecuteMsgRequest{}
			cosmwasm.MustExecute[transmuter.JoinPoolExecuteMsgRequest, msg.EmptyStruct](s.Ctx, s.App.ContractKeeper, pool.GetContractAddress(), s.TestAccs[0], tc.initialCoins, request)

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
			expectedPoolBalances := originalPoolBalances.Add(tc.tokenIn).Sub(sdk.NewCoins(tc.expectedTokenOut))
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
		tokenInMaxAmount sdk.Int
		swapFee          sdk.Dec
		isInvalidPool    bool

		expectedTokenIn      sdk.Coin
		expectedErrorMessage string
	}{
		"calc amount less than supply": {
			initialCoins:     initalDefaultSupply,
			tokenOut:         sdk.NewCoin(denomA, defaultAmount.Sub(sdk.OneInt())),
			tokenInDenom:     denomB,
			expectedTokenIn:  sdk.NewCoin(denomB, defaultAmount.Sub(sdk.OneInt())),
			tokenInMaxAmount: defaultAmount,
			swapFee:          sdk.ZeroDec(),
		},
		"calc amount equal to supply": {
			initialCoins:     initalDefaultSupply,
			tokenOut:         sdk.NewCoin(denomA, defaultAmount),
			tokenInDenom:     denomB,
			expectedTokenIn:  sdk.NewCoin(denomB, defaultAmount),
			tokenInMaxAmount: defaultAmount,
			swapFee:          sdk.ZeroDec(),
		},
		"calc amount greater than supply": {
			initialCoins:         initalDefaultSupply,
			tokenOut:             sdk.NewCoin(denomA, defaultAmount.Add(sdk.OneInt())),
			tokenInDenom:         denomB,
			expectedErrorMessage: fmt.Sprintf("Insufficient fund: required: %s, available: %s", sdk.NewCoin(denomA, defaultAmount.Add(sdk.OneInt())), sdk.NewCoin(denomA, defaultAmount)),
		},
		"non-zero swap fee": {
			initialCoins:         initalDefaultSupply,
			tokenOut:             sdk.NewCoin(denomA, defaultAmount.Sub(sdk.OneInt())),
			tokenInDenom:         denomB,
			swapFee:              sdk.MustNewDecFromStr(nonZeroFeeStr),
			expectedErrorMessage: fmt.Sprintf("Invalid swap fee: expected: %s, actual: %s", sdk.ZeroInt(), nonZeroFeeStr),
		},
		"invalid pool given": {
			initialCoins:  sdk.NewCoins(sdk.NewCoin(denomA, defaultAmount), sdk.NewCoin(denomB, defaultAmount)),
			tokenOut:      sdk.NewCoin(denomA, defaultAmount.Sub(sdk.OneInt())),
			tokenInDenom:  denomB,
			isInvalidPool: true,

			expectedErrorMessage: types.InvalidPoolTypeError{
				ActualPool: &clmodel.Pool{},
			}.Error(),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			// fund pool joiner
			s.FundAcc(s.TestAccs[0], tc.initialCoins)

			// get initial denom from coins specified in the test case
			initialDenoms := []string{}
			for _, coin := range tc.initialCoins {
				initialDenoms = append(initialDenoms, coin.Denom)
			}

			// create pool
			pool := s.PrepareCustomTransmuterPool(s.TestAccs[0], initialDenoms, 1)

			// add liquidity by joining the pool
			request := transmuter.JoinPoolExecuteMsgRequest{}
			cosmwasm.MustExecute[transmuter.JoinPoolExecuteMsgRequest, msg.EmptyStruct](s.Ctx, s.App.ContractKeeper, pool.GetContractAddress(), s.TestAccs[0], tc.initialCoins, request)

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
			if !tc.expectedTokenIn.IsNil() {
				// Fund with expected token in
				s.FundAcc(swapper, sdk.NewCoins(tc.expectedTokenIn))
			} else {
				// Fund with pool reserve of token in denom
				// This case happens in the error case, and we want
				// to make sure that the error that we get is not
				// due to insufficient funds.
				s.FundAcc(swapper, sdk.NewCoins(sdk.NewCoin(tc.tokenInDenom, defaultAmount)))
			}

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
			expectedPoolBalances := originalPoolBalances.Add(tc.expectedTokenIn).Sub(sdk.NewCoins(tc.tokenOut))
			afterSwapPoolBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(pool.GetContractAddress()))
			s.Require().Equal(expectedPoolBalances.String(), afterSwapPoolBalances.String())

			// Assert that swapper balance is updated correctly
			expectedSwapperBalances := sdk.NewCoins(tc.tokenOut)
			afterSwapSwapperBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, swapper)
			s.Require().Equal(expectedSwapperBalances.String(), afterSwapSwapperBalances.String())
		})
	}
}
