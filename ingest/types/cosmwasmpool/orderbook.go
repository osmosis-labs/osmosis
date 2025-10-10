package cosmwasmpool

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	clmath "github.com/osmosis-labs/osmosis/v31/x/concentrated-liquidity/math"
)

const (
	ORDERBOOK_CONTRACT_NAME               = "crates.io:sumtree-orderbook"
	ORDERBOOK_MIN_CONTRACT_VERSION        = "0.1.0"
	ORDERBOOK_CONTRACT_VERSION_CONSTRAINT = ">= " + ORDERBOOK_MIN_CONTRACT_VERSION
)

func (model *CosmWasmPoolModel) IsOrderbook() bool {
	return model.ContractInfo.Matches(
		ORDERBOOK_CONTRACT_NAME,
		mustParseSemverConstraint(ORDERBOOK_CONTRACT_VERSION_CONSTRAINT),
	)
}

type OrderbookDirection bool

const (
	BID OrderbookDirection = true
	ASK OrderbookDirection = false
)

// RoundingDirection is used to determine the rounding direction when converting an amount of token in to the value of token out.
// We aim to always round in pool's favor, rounding down token out and rounding up token in.
type RoundingDirection bool

const (
	// ROUND_UP rounds up the result of the conversion.
	ROUND_UP RoundingDirection = true
	// ROUND_DOWN rounds down the result of the conversion.
	ROUND_DOWN RoundingDirection = false
)

func (d *OrderbookDirection) String() string {
	if *d { // BID
		return "BID"
	} else { // ASK
		return "ASK"
	}
}

func (d *OrderbookDirection) Opposite() OrderbookDirection {
	if *d { // BID
		return ASK
	} else { // ASK
		return BID
	}
}

// IterationStep returns the step to be used for iterating the orderbook.
// The orderbook ticks are ordered by tick id in ascending order.
// BID piles up on the top of the orderbook, while ASK piles up on the bottom.
// So if we want to iterate the BID orderbook, we should iterate in descending order.
// If we want to iterate the ASK orderbook, we should iterate in ascending order.
func (d *OrderbookDirection) IterationStep() (int, error) {
	if *d { // BID
		return -1, nil
	} else { // ASK
		return 1, nil
	}
}

// Converts an amount of token in to the value of token out given a tick price and target direction
func OrderbookValueInOppositeDirection(sourceDirectionAmount osmomath.BigDec, tickPrice osmomath.BigDec, targetDirection OrderbookDirection, roundingDirection RoundingDirection) osmomath.BigDec {
	switch targetDirection {
	case ASK:
		if roundingDirection == ROUND_UP {
			return sourceDirectionAmount.MulRoundUp(tickPrice)
		} else {
			return sourceDirectionAmount.MulTruncate(tickPrice)
		}
	case BID:
		if roundingDirection == ROUND_UP {
			return sourceDirectionAmount.QuoRoundUp(tickPrice)
		} else {
			return sourceDirectionAmount.QuoTruncate(tickPrice)
		}
	default:
		return osmomath.ZeroBigDec()
	}
}

// OrderbookData, since v1.0.0
type OrderbookData struct {
	QuoteDenom                     string          `json:"quote_denom"`
	BaseDenom                      string          `json:"base_denom"`
	NextBidTickIndex               int             `json:"next_bid_tick_index"`                 // tick index of the next bid tick, -1 if no bid tick
	NextAskTickIndex               int             `json:"next_ask_tick_index"`                 // tick index of the next ask tick, -1 if no ask tick
	BidAmountToExhaustAskLiquidity osmomath.BigDec `json:"bid_amount_to_exhaust_ask_liquidity"` // bid amount (in quote denom) to exhaust all ask liquidity, updated on ingest
	AskAmountToExhaustBidLiquidity osmomath.BigDec `json:"ask_amount_to_exhaust_bid_liquidity"` // ask amount (in base denom) to exhaust all bid liquidity, updated on ingest
	Ticks                          []OrderbookTick `json:"ticks"`
}

// Determines order direction for the current orderbook given token in and out denoms
// Returns:
// - BID (true) if the order is a bid (buying token out)
// - ASK (false) if the order is an ask (selling token out)
// - 0 if the order is not valid
func (d *OrderbookData) GetDirection(tokenInDenom, tokenOutDenom string) (*OrderbookDirection, error) {
	if tokenInDenom == tokenOutDenom {
		return nil, DuplicatedDenomError{Denom: tokenInDenom}
	}

	if tokenInDenom == d.BaseDenom && tokenOutDenom == d.QuoteDenom {
		dir := ASK
		return &dir, nil
	} else if tokenInDenom == d.QuoteDenom && tokenOutDenom == d.BaseDenom {
		dir := BID
		return &dir, nil
	} else if tokenInDenom != d.BaseDenom && tokenInDenom != d.QuoteDenom {
		return nil, OrderbookUnsupportedDenomError{Denom: tokenInDenom, QuoteDenom: d.QuoteDenom, BaseDenom: d.BaseDenom}
	} else {
		return nil, OrderbookUnsupportedDenomError{Denom: tokenOutDenom, QuoteDenom: d.QuoteDenom, BaseDenom: d.BaseDenom}
	}
}

// Get the index for the tick state array for the starting index given direction
func (d *OrderbookData) GetStartTickIndex(direction OrderbookDirection) (int, error) {
	if direction == ASK {
		return d.NextAskTickIndex, nil
	} else { // BID
		return d.NextBidTickIndex, nil
	}
}

// Represents Total Amount of Liquidity at tick (TAL) of a specific price tick in a liquidity pool.
// - Every limit order placement increments this value.
// - Every swap at this tick decrements this value.
// - Every cancellation decrements this value.
//
// It is split into two parts for the ask and bid directions.
type OrderbookTickLiquidity struct {
	// Total Amount of Liquidity at tick (TAL) for the bid direction of the tick
	BidLiquidity osmomath.BigDec `json:"bid_liquidity"`
	// Total Amount of Liquidity at tick (TAL) for the ask direction of the tick
	AskLiquidity osmomath.BigDec `json:"ask_liquidity"`
}

// Returns the related liquidity for a given direction on the current tick
func (tl *OrderbookTickLiquidity) ByDirection(direction OrderbookDirection) osmomath.BigDec {
	if direction == ASK {
		return tl.AskLiquidity
	} else { // BID
		return tl.BidLiquidity
	}
}

// Determines how much of a given amount can be filled by the current tick state (independent for each direction)
func (tl *OrderbookTickLiquidity) GetFillableAmount(input osmomath.BigDec, direction OrderbookDirection) osmomath.BigDec {
	tickLiquidity := tl.ByDirection(direction)
	if input.LT(tickLiquidity) {
		return input
	}
	return tickLiquidity
}

type OrderbookTick struct {
	TickId        int64                  `json:"tick_id"`
	TickLiquidity OrderbookTickLiquidity `json:"tick_liquidity"`
}

// CalcAmountInToExhaustOrderbookLiquidity calculates the amount of token in needed to exhaust all liquidity in the orderbook.
// - orderDirection is the direction of the order to be placed and liquidity to be exhausted will be in the opposite direction.
// - startingIndex is the index of the tick to start the calculation from. This should be starting index form the opposite direction from the order direction.
// - ticks is the list of ticks in the orderbook, assumed to be ordered by tick id in ascending order.
func CalcAmountInToExhaustOrderbookLiquidity(directionIn OrderbookDirection, startingIndex int, ticks []OrderbookTick) (osmomath.BigDec, error) {
	directionOut := directionIn.Opposite()

	directionOutIterationStep, err := directionOut.IterationStep()
	if err != nil {
		return osmomath.ZeroBigDec(), err
	}

	requiredAmountIn := osmomath.ZeroBigDec()

	// Iterate over the ticks in the orderbook
	// starting from the given starting index
	// iterate in the orderbook side that liquidity will be drained
	// which is the opposite of the order direction
	i := startingIndex
	for i >= 0 && i < len(ticks) {
		tick := ticks[i]
		tickLiquidity := tick.TickLiquidity.ByDirection(directionOut)
		tickPrice, err := clmath.TickToPrice(tick.TickId)
		if err != nil {
			return osmomath.ZeroBigDec(), err
		}

		// convert current tick liquidity to value required in the order direction
		tickRequiredAmountIn := OrderbookValueInOppositeDirection(tickLiquidity, tickPrice, directionOut, ROUND_UP)

		// accumulate the required amount in
		requiredAmountIn = requiredAmountIn.Add(tickRequiredAmountIn)

		// move to the next tick based on orderbook direction
		i += directionOutIterationStep
	}

	return requiredAmountIn, nil
}
