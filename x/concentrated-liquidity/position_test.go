package concentrated_liquidity_test

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestInitOrUpdatePosition() {
	const validPoolId = 1
	type param struct {
		poolId         uint64
		lowerTick      int64
		upperTick      int64
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
			name: "Init position from -50 to 50 with DefaultLiquidityAmt liquidity",
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
				err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta)
				s.Require().NoError(err)
				preexistingLiquidity = DefaultLiquidityAmt
			}

			// Get the position info for poolId 1
			positionInfo, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, validPoolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick)
			if test.positionExists {
				// If we had a position before, ensure the position info displays proper liquidity
				s.Require().NoError(err)
				s.Require().Equal(preexistingLiquidity, positionInfo.Liquidity)

				// Ensure JoinTime tracker has logged existing liquidity in sumtree
				liqAtBeforeCurTime := s.App.ConcentratedLiquidityKeeper.GetLiquidityBeforeOrAtJoinTime(s.Ctx, validPoolId, s.Ctx.BlockTime())
				s.Require().Equal(preexistingLiquidity, liqAtBeforeCurTime)
			} else {
				// If we did not have a position before, ensure getting the non-existent position returns an error
				s.Require().Error(err)
				s.Require().ErrorContains(err, types.PositionNotFoundError{PoolId: validPoolId, LowerTick: test.param.lowerTick, UpperTick: test.param.upperTick}.Error())

				// Ensure JoinTime tracker has *not* logged existing liquidity in sumtree
				liqAtBeforeCurTime := s.App.ConcentratedLiquidityKeeper.GetLiquidityBeforeOrAtJoinTime(s.Ctx, validPoolId, s.Ctx.BlockTime())
				s.Require().Equal(preexistingLiquidity, liqAtBeforeCurTime)
			}

			// Move block time up by 5 seconds, saving old and new join times for testing
			oldJoinTime := s.Ctx.BlockTime()
			newJoinTime := s.Ctx.BlockTime().Add(5 * time.Second)
			newCtx := s.Ctx.WithBlockTime(newJoinTime)
			s.Ctx = newCtx

			// System under test. Initialize or update the position according to the test case
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
				return
			}
			s.Require().NoError(err)

			// Get the tick info for poolId 1
			positionInfo, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, validPoolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick)
			s.Require().NoError(err)

			// Check that the initialized or updated position matches our expectation
			s.Require().Equal(test.expectedLiquidity, positionInfo.Liquidity)

			// Liquidity at & before old join time should be cleared
			liqAtBeforeOldJoinTime := s.App.ConcentratedLiquidityKeeper.GetLiquidityBeforeOrAtJoinTime(s.Ctx, validPoolId, oldJoinTime)
			s.Require().Equal(sdk.ZeroDec(), liqAtBeforeOldJoinTime)

			// Liquidity at & before new join time should be equal to position's liquidity
			liqAtBeforeNewJoinTime := s.App.ConcentratedLiquidityKeeper.GetLiquidityBeforeOrAtJoinTime(s.Ctx, validPoolId, newJoinTime)
			s.Require().Equal(positionInfo.Liquidity, liqAtBeforeNewJoinTime)
		})
	}
}

func (s *KeeperTestSuite) TestGetPosition() {
	tests := []struct {
		name             string
		poolToGet        uint64
		ownerIndex       uint64
		lowerTick        int64
		upperTick        int64
		expectedPosition *model.Position
		expectedErr      error
	}{
		{
			name:             "Get position info on existing pool and existing position",
			poolToGet:        validPoolId,
			lowerTick:        DefaultLowerTick,
			upperTick:        DefaultUpperTick,
			expectedPosition: &model.Position{Liquidity: DefaultLiquidityAmt},
		},
		{
			name:        "Get position info on existing pool and existing position but wrong owner",
			poolToGet:   validPoolId,
			ownerIndex:  1,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick},
		},
		{
			name:        "Get position info on existing pool with no existing position",
			poolToGet:   validPoolId,
			lowerTick:   DefaultLowerTick - 1,
			upperTick:   DefaultUpperTick + 1,
			expectedErr: types.PositionNotFoundError{PoolId: validPoolId, LowerTick: DefaultLowerTick - 1, UpperTick: DefaultUpperTick + 1},
		},
		{
			name:        "Get position info on a non-existing pool with no existing position",
			poolToGet:   2,
			lowerTick:   DefaultLowerTick - 1,
			upperTick:   DefaultUpperTick + 1,
			expectedErr: types.PositionNotFoundError{PoolId: 2, LowerTick: DefaultLowerTick - 1, UpperTick: DefaultUpperTick + 1},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Update expectedPos time to current context's blocktime
			if test.expectedPosition != nil {
				test.expectedPosition.JoinTime = s.Ctx.BlockTime()
			}

			// Create a default CL pool
			s.PrepareConcentratedPool()

			// Set up a default initialized position
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt)

			// System under test
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, test.poolToGet, s.TestAccs[test.ownerIndex], test.lowerTick, test.upperTick)
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

func (s *KeeperTestSuite) TestMultiplePositionsFuzz() {
	// Init suite for each test.
	s.Setup()

	// Create a default CL pool
	s.PrepareConcentratedPool()

	// Test 100 random positions in same pool at random jointimes
	// Note: can fuzz these values in the future as well
	numPositions := 100
	validPoolId := uint64(1)

	// Max lower and upper ticks
	tickLowerBound := int64(100)
	tickUpperBound := int64(100)

	// Max time elapsed between positions (in seconds)
	timeElapsedUpperBound := 1000

	// We generate a new test account per position for robustness
	testAccounts := osmoutils.CreateRandomAccounts(numPositions)

	// We use the reference time to check if our sumtree properly accumulates liquidity
	// Max seconds set at 32 bits & max nanoseconds at 30 bits to comply with bounds
	// specified in the Time package
	referenceTime := time.Unix(int64(rand.Int31()), int64(rand.Int31()) % (1 << 30))
	maxElapsedTimeFromStart := timeElapsedUpperBound * numPositions

	// Set start time = ref time - one tenth of max time elapsed
	// This should mean we end up with at least some liquidity on each side of the
	// ref time, ensuring testability while still having some variance in join times
	startTime := referenceTime.Add(time.Duration(-1 * maxElapsedTimeFromStart / 10) * time.Second)
	s.Ctx = s.Ctx.WithBlockTime(startTime)

	// Trackers for total liquidity on either side of our reference join time
	expLiqBeforeOrAtRefTime := sdk.ZeroDec()
	expLiqAfterRefTime := sdk.ZeroDec()

	for i := 0; i < numPositions; i++ {
		// Generate random time elapsed since last join and update block time
		// Half the time, we simply add to the same join time as the previous position (time elapsed = 0)
		addToSameJoinTime := int64(rand.Int() % 2)
		timeElapsed := time.Duration((int64(rand.Int31()) % int64(timeElapsedUpperBound)) * addToSameJoinTime) * time.Second

		joinTime := s.Ctx.BlockTime().Add(timeElapsed)
		s.Ctx = s.Ctx.WithBlockTime(joinTime)

		// Generate random lower and upper ticks
		lowerTick := (-1) * (rand.Int63() % tickLowerBound)
		upperTick := rand.Int63() % tickUpperBound

		// Generate random liquidity to join with
		// Note: might need to make these int32 due to overflows
		baseAmt := rand.Int63()
		prec := rand.Int63() % 18
		joinAmt := sdk.NewDecWithPrec(baseAmt, prec)

		// Track total liq entered before and after reference time
		if joinTime.After(referenceTime) {
			expLiqAfterRefTime = expLiqAfterRefTime.Add(joinAmt)
		} else {
			expLiqBeforeOrAtRefTime = expLiqBeforeOrAtRefTime.Add(joinAmt)
		}

		// Create position with account `i` given our randomly generated constraints
		err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, testAccounts[i], lowerTick, upperTick, joinAmt)
		s.Require().NoError(err)

		// Debug prints:
		// fmt.Println("--------------------------------")
		// fmt.Println("Position number: ", i)
		// fmt.Println("Liquidity added: ", joinAmt)
		// fmt.Println("Accumulated liq before ref time: ", expLiqBeforeOrAtRefTime)
		// fmt.Println("Accumulated liq after ref time: ", expLiqAfterRefTime)
		// fmt.Println("Liquidity in tree before ref time: ", s.App.ConcentratedLiquidityKeeper.GetLiquidityBeforeOrAtJoinTime(s.Ctx, validPoolId, referenceTime))
		// fmt.Println("Liquidity in tree after ref time: ", s.App.ConcentratedLiquidityKeeper.GetLiquidityAfterJoinTime(s.Ctx, validPoolId, referenceTime))
	}

	// Get liquidity at or before ref time and ensure sumtree tracked it correctly
	liqBeforeOrAtRefTime := s.App.ConcentratedLiquidityKeeper.GetLiquidityBeforeOrAtJoinTime(s.Ctx, validPoolId, referenceTime)
	s.Require().Equal(expLiqBeforeOrAtRefTime, liqBeforeOrAtRefTime)

	// Get liquidity after ref time and ensure sumtree tracked it correctly
	liqAfterRefTime := s.App.ConcentratedLiquidityKeeper.GetLiquidityAfterJoinTime(s.Ctx, validPoolId, referenceTime)
	s.Require().Equal(expLiqAfterRefTime, liqAfterRefTime)
}