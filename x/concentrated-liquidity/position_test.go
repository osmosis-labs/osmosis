package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestInitOrUpdatePosition() {
	const validPoolId = 1
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)
	type param struct {
		poolId         uint64
		lowerTick      int64
		upperTick      int64
		frozenUntil    time.Time
		liquidityDelta sdk.Dec
		liquidityIn    sdk.Dec
	}

	tests := []struct {
		name              string
		param             param
		positionExists    bool
		expectedLiquidity sdk.Dec
		expectedErr       error
	}{
		{
			name: "Init position from -50 to 50 with DefaultLiquidityAmt liquidity and no freeze duration",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
			},
			positionExists:    false,
			expectedLiquidity: DefaultLiquidityAmt,
		},
		{
			name: "Update position from -50 to 50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
			},
			positionExists:    true,
			expectedLiquidity: DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
		},
		{
			name: "Update position from -50 to 50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity with an hour freeze duration",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				frozenUntil:    defaultFrozenUntil,
				liquidityDelta: DefaultLiquidityAmt,
			},
			positionExists:    true,
			expectedLiquidity: DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
		},
		{
			name: "Init position for non-existing pool",
			param: param{
				poolId:         2,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
			},
			positionExists: false,
			expectedErr:    types.PoolNotFoundError{PoolId: 2},
		},
		{
			name: "Init position from -50 to 50 with negative DefaultLiquidityAmt liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt.Neg(),
			},
			positionExists: false,
			expectedErr:    types.NegativeLiquidityError{Liquidity: DefaultLiquidityAmt.Neg()},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a default CL pool
			s.PrepareConcentratedPool()

			// If positionExists set, initialize the specified position with defaultLiquidityAmt
			preexistingLiquidity := sdk.ZeroDec()
			if test.positionExists {
				err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta, test.param.frozenUntil)
				s.Require().NoError(err)
				preexistingLiquidity = DefaultLiquidityAmt
			}

			// Get the position info for poolId 1
			positionInfo, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, validPoolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.frozenUntil)
			if test.positionExists {
				// If we had a position before, ensure the position info displays proper liquidity
				s.Require().NoError(err)
				s.Require().Equal(preexistingLiquidity, positionInfo.Liquidity)
			} else {
				// If we did not have a position before, ensure getting the non-existent position returns an error
				s.Require().Error(err)
				s.Require().ErrorContains(err, types.PositionNotFoundError{PoolId: validPoolId, LowerTick: test.param.lowerTick, UpperTick: test.param.upperTick}.Error())
			}

			// System under test. Initialize or update the position according to the test case
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta, test.param.frozenUntil)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
				return
			}
			s.Require().NoError(err)

			// Get the tick info for poolId 1
			positionInfo, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, validPoolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.frozenUntil)
			s.Require().NoError(err)

			// Check that the initialized or updated position matches our expectation
			s.Require().Equal(test.expectedLiquidity, positionInfo.Liquidity)
		})
	}
}

func (s *KeeperTestSuite) TestGetPosition() {
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)
	tests := []struct {
		name             string
		poolToGet        uint64
		ownerIndex       uint64
		lowerTick        int64
		upperTick        int64
		frozenUntil      time.Time
		expectedPosition *model.Position
		expectedErr      error
	}{
		{
			name:             "Get position info on existing pool and existing position",
			poolToGet:        validPoolId,
			lowerTick:        DefaultLowerTick,
			upperTick:        DefaultUpperTick,
			frozenUntil:      defaultFrozenUntil,
			expectedPosition: &model.Position{Liquidity: DefaultLiquidityAmt, FrozenUntil: defaultFrozenUntil},
		},
		{
			name:        "Get position info on existing pool and existing position but wrong owner",
			poolToGet:   validPoolId,
			ownerIndex:  1,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick, FrozenUntil: defaultFrozenUntil},
		},
		{
			name:        "Get position info on existing pool and existing position but wrong frozenUntil time",
			poolToGet:   validPoolId,
			ownerIndex:  1,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil.Add(time.Second),
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick, FrozenUntil: defaultFrozenUntil.Add(time.Second)},
		},
		{
			name:        "Get position info on existing pool with no existing position",
			poolToGet:   validPoolId,
			lowerTick:   DefaultLowerTick - 1,
			upperTick:   DefaultUpperTick + 1,
			frozenUntil: defaultFrozenUntil,
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick - 1, UpperTick: DefaultUpperTick + 1, FrozenUntil: defaultFrozenUntil},
		},
		{
			name:        "Get position info on a non-existing pool with no existing position",
			poolToGet:   2,
			lowerTick:   DefaultLowerTick - 1,
			upperTick:   DefaultUpperTick + 1,
			frozenUntil: defaultFrozenUntil,
			expectedErr: types.PositionNotFoundError{PoolId: 2, LowerTick: DefaultLowerTick - 1, UpperTick: DefaultUpperTick + 1, FrozenUntil: defaultFrozenUntil},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a default CL pool
			s.PrepareConcentratedPool()

			// Set up a default initialized position
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt, defaultFrozenUntil)

			// System under test
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, test.poolToGet, s.TestAccs[test.ownerIndex], test.lowerTick, test.upperTick, test.frozenUntil)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Nil(position)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.expectedPosition, position)
			}

		})
	}
}

func (s *KeeperTestSuite) TestGetAllUserPositions() {
	defaultAddress := s.TestAccs[0]
	secondAddress := s.TestAccs[1]

	type position struct {
		poolId    uint64
		acc       sdk.AccAddress
		coin0     sdk.Coin
		coin1     sdk.Coin
		lowerTick int64
		upperTick int64
	}

	tests := []struct {
		name           string
		sender         sdk.AccAddress
		setupPositions []position
		expectedErr    error
	}{
		{
			name:   "Get current user one position",
			sender: defaultAddress,
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
			},
		},
		{
			name:   "Get current users multiple position same pool",
			sender: defaultAddress,
			setupPositions: []position{
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 1, DefaultUpperTick + 1},
				{1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 2, DefaultUpperTick + 2},
			},
		},
		{
			name:   "Get current users multiple position multiple pools",
			sender: secondAddress,
			setupPositions: []position{
				{1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{2, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 1, DefaultUpperTick + 1},
				{3, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 2, DefaultUpperTick + 2},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a default CL pools
			s.PrepareMultipleConcentratedPools(3)

			expectedUserPositions := []types.FullPositionByOwnerResult{}
			for _, pos := range test.setupPositions {
				// if position does not exist this errors
				position := s.SetupPosition(pos.poolId, pos.acc, pos.coin0, pos.coin1, pos.lowerTick, pos.upperTick, s.Ctx.BlockTime().Add(DefaultFreezeDuration))
				if pos.acc.Equals(pos.acc) {
					expectedUserPositions = append(expectedUserPositions, types.FullPositionByOwnerResult{
						PoolId:      pos.poolId,
						LowerTick:   pos.lowerTick,
						UpperTick:   pos.upperTick,
						FrozenUntil: s.Ctx.BlockTime().Add(DefaultFreezeDuration),
						Liquidity:   position.Liquidity,
					})
				}
			}

			// System under test
			position, err := s.App.ConcentratedLiquidityKeeper.GetUserPositions(s.Ctx, test.sender)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Nil(position)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(expectedUserPositions, position)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDeletePosition() {
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)
	tests := []struct {
		name        string
		poolToGet   uint64
		ownerIndex  uint64
		lowerTick   int64
		upperTick   int64
		frozenUntil time.Time
		expectedErr error
	}{
		{
			name:        "Delete position info on existing pool and existing position",
			poolToGet:   validPoolId,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
		},
		{
			name:        "Delete position on existing pool and existing position but wrong owner",
			poolToGet:   validPoolId,
			ownerIndex:  1,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick, FrozenUntil: defaultFrozenUntil},
		},
		{
			name:        "Delete position on existing pool and existing position but wrong frozenUntil time",
			poolToGet:   validPoolId,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil.Add(time.Second),
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick, FrozenUntil: defaultFrozenUntil.Add(time.Second)},
		},
		{
			name:        "Delete position on existing pool with no existing position",
			poolToGet:   validPoolId,
			lowerTick:   DefaultLowerTick - 1,
			upperTick:   DefaultUpperTick + 1,
			frozenUntil: defaultFrozenUntil,
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick - 1, UpperTick: DefaultUpperTick + 1, FrozenUntil: defaultFrozenUntil},
		},
		{
			name:        "Delete position on a non-existing pool with no existing position",
			poolToGet:   2,
			lowerTick:   DefaultLowerTick - 1,
			upperTick:   DefaultUpperTick + 1,
			frozenUntil: defaultFrozenUntil,
			expectedErr: types.PositionNotFoundError{PoolId: 2, LowerTick: DefaultLowerTick - 1, UpperTick: DefaultUpperTick + 1, FrozenUntil: defaultFrozenUntil},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a default CL pool
			s.PrepareConcentratedPool()

			// Set up a default initialized position
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt, defaultFrozenUntil)
			s.Require().NoError(err)

			err = s.App.ConcentratedLiquidityKeeper.DeletePosition(s.Ctx, test.poolToGet, s.TestAccs[test.ownerIndex], test.lowerTick, test.upperTick, test.frozenUntil)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
			} else {
				s.Require().NoError(err)

				// Since the position is deleted, retrieving it should return an error.
				position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, test.poolToGet, s.TestAccs[test.ownerIndex], test.lowerTick, test.upperTick, test.frozenUntil)
				s.Require().Error(err)
				s.Require().ErrorIs(err, types.PositionNotFoundError{PoolId: test.poolToGet, LowerTick: test.lowerTick, UpperTick: test.upperTick, FrozenUntil: test.frozenUntil})
				s.Require().Nil(position)
			}
		})
	}
}
