package cosmwasmpool_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v28/ingest/types/cosmwasmpool"

	"github.com/stretchr/testify/assert"
)

const (
	QUOTE_DENOM   = "quote"
	BASE_DENOM    = "base"
	INVALID_DENOM = "invalid"
	MIN_TICK      = -108000000
	MAX_TICK      = 182402823
	// Tick Price = 2
	LARGE_POSITIVE_TICK int64 = 1000000
	// Tick Price = 0.5
	LARGE_NEGATIVE_TICK int64 = -5000000
)

func TestGetDirection(t *testing.T) {
	tests := map[string]struct {
		tokenInDenom  string
		tokenOutDenom string
		expected      cosmwasmpool.OrderbookDirection
		expectError   error
	}{
		"BID direction": {
			tokenInDenom:  QUOTE_DENOM,
			tokenOutDenom: BASE_DENOM,
			expected:      cosmwasmpool.BID,
		},
		"ASK direction": {
			tokenInDenom:  BASE_DENOM,
			tokenOutDenom: QUOTE_DENOM,
			expected:      cosmwasmpool.ASK,
		},
		"duplicated denom": {
			tokenInDenom:  BASE_DENOM,
			tokenOutDenom: BASE_DENOM,
			expectError: cosmwasmpool.DuplicatedDenomError{
				Denom: BASE_DENOM,
			},
		},
		"invalid direction": {
			tokenInDenom:  INVALID_DENOM,
			tokenOutDenom: BASE_DENOM,
			expectError: cosmwasmpool.OrderbookUnsupportedDenomError{
				Denom:      INVALID_DENOM,
				BaseDenom:  BASE_DENOM,
				QuoteDenom: QUOTE_DENOM,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			orderbookData := cosmwasmpool.OrderbookData{}

			direction, err := orderbookData.GetDirection(tc.tokenInDenom, tc.tokenOutDenom)

			if tc.expectError != nil {
				assert.Error(err)
				assert.Equal(err, tc.expectError)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.expected, *direction)
		})
	}
}

func TestGetFillableAmount(t *testing.T) {
	tests := map[string]struct {
		input        osmomath.BigDec
		direction    cosmwasmpool.OrderbookDirection
		bidLiquidity osmomath.BigDec
		askLiquidity osmomath.BigDec
		expected     osmomath.BigDec
	}{
		"fillable amount less than tick liquidity": {
			input:        osmomath.NewBigDec(50),
			direction:    cosmwasmpool.BID,
			bidLiquidity: osmomath.NewBigDec(100),
			askLiquidity: osmomath.NewBigDec(0),
			expected:     osmomath.NewBigDec(50),
		},
		"fillable amount more than tick liquidity": {
			input:        osmomath.NewBigDec(150),
			direction:    cosmwasmpool.ASK,
			bidLiquidity: osmomath.NewBigDec(0),
			askLiquidity: osmomath.NewBigDec(100),
			expected:     osmomath.NewBigDec(100),
		},
		"fillable amount equal to tick liquidity": {
			input:        osmomath.NewBigDec(100),
			direction:    cosmwasmpool.BID,
			bidLiquidity: osmomath.NewBigDec(100),
			askLiquidity: osmomath.NewBigDec(0),
			expected:     osmomath.NewBigDec(100),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			orderbookTickLiquidity := cosmwasmpool.OrderbookTickLiquidity{
				BidLiquidity: tc.bidLiquidity,
				AskLiquidity: tc.askLiquidity,
			}

			fillableAmount := orderbookTickLiquidity.GetFillableAmount(tc.input, tc.direction)

			assert.Equal(tc.expected, fillableAmount)
		})
	}
}

func TestCalcAmountInToExhaustOrderbookLiquidity(t *testing.T) {
	tests := map[string]struct {
		orderDirection           cosmwasmpool.OrderbookDirection
		startingIndex            int
		ticks                    []cosmwasmpool.OrderbookTick
		expectedRequiredAmountIn osmomath.BigDec
		expectError              error
	}{
		"no liquidity to exhaust": {
			orderDirection: cosmwasmpool.BID,
			startingIndex:  0,
			ticks: []cosmwasmpool.OrderbookTick{
				{
					TickId: 0,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.ZeroBigDec(),
						AskLiquidity: osmomath.ZeroBigDec(),
					},
				},
			},
			expectedRequiredAmountIn: osmomath.ZeroBigDec(),
		},
		"exhausting all bid liquidity, single tick": {
			orderDirection: cosmwasmpool.ASK,
			startingIndex:  0,
			ticks: []cosmwasmpool.OrderbookTick{
				{
					TickId: 0,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.NewBigDec(150),
						AskLiquidity: osmomath.ZeroBigDec(),
					},
				},
			},
			expectedRequiredAmountIn: osmomath.NewBigDec(150),
		},
		"exhausting all bid liquidity, single tick, non 0": {
			orderDirection: cosmwasmpool.ASK,
			startingIndex:  0,
			ticks: []cosmwasmpool.OrderbookTick{
				{
					TickId: LARGE_POSITIVE_TICK,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.NewBigDec(150),
						AskLiquidity: osmomath.ZeroBigDec(),
					},
				},
			},
			expectedRequiredAmountIn: osmomath.NewBigDec(300),
		},
		"exhausting all bid liquidity, multiple ticks": {
			orderDirection: cosmwasmpool.ASK,
			startingIndex:  1,
			ticks: []cosmwasmpool.OrderbookTick{
				{
					TickId: 0,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.NewBigDec(50),
						AskLiquidity: osmomath.ZeroBigDec(),
					},
				},
				{
					TickId: LARGE_POSITIVE_TICK,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.NewBigDec(100),
						AskLiquidity: osmomath.ZeroBigDec(),
					},
				},
			},
			expectedRequiredAmountIn: osmomath.NewBigDec(250),
		},
		"exhausting all ask liquidity, single tick": {
			orderDirection: cosmwasmpool.BID,
			startingIndex:  0,
			ticks: []cosmwasmpool.OrderbookTick{
				{
					TickId: 0,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.ZeroBigDec(),
						AskLiquidity: osmomath.NewBigDec(150),
					},
				},
			},
			expectedRequiredAmountIn: osmomath.NewBigDec(150),
		},
		"exhausting all ask liquidity, single tick, non 0": {
			orderDirection: cosmwasmpool.BID,
			startingIndex:  0,
			ticks: []cosmwasmpool.OrderbookTick{
				{
					TickId: LARGE_NEGATIVE_TICK,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.ZeroBigDec(),
						AskLiquidity: osmomath.NewBigDec(150),
					},
				},
			},
			expectedRequiredAmountIn: osmomath.NewBigDec(300),
		},
		"exhausting all ask liquidity, multiple ticks": {
			orderDirection: cosmwasmpool.BID,
			startingIndex:  0,
			ticks: []cosmwasmpool.OrderbookTick{
				{
					TickId: LARGE_NEGATIVE_TICK,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.ZeroBigDec(),
						AskLiquidity: osmomath.NewBigDec(100),
					},
				},
				{
					TickId: 0,
					TickLiquidity: cosmwasmpool.OrderbookTickLiquidity{
						BidLiquidity: osmomath.ZeroBigDec(),
						AskLiquidity: osmomath.NewBigDec(50),
					},
				},
			},
			expectedRequiredAmountIn: osmomath.NewBigDec(250),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			amountIn, err := cosmwasmpool.CalcAmountInToExhaustOrderbookLiquidity(tc.orderDirection, tc.startingIndex, tc.ticks)

			if tc.expectError != nil {
				assert.Error(err)
				assert.Equal(tc.expectError, err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.expectedRequiredAmountIn, amountIn)
		})
	}
}
