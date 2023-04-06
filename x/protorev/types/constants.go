package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OsmosisDenomination stores the native denom name for Osmosis on chain used for route building
var OsmosisDenomination string = "uosmo"

// ----------------- Module Execution Time Constants ----------------- //

// MaxInputAmount is the upper bound index for finding the optimal in amount when determining route profitability (2 ^ 14) = 16,384
var MaxInputAmount = sdk.NewInt(16_384)

// ExtendedMaxInputAmount is the upper bound index for finding the optimal in amount
// when determining route profitability for an arb that's above the default range (2 ^ 17) = 131,072
var ExtendedMaxInputAmount = sdk.NewInt(131_072)

// Max iterations for binary search (log2(131_072) = 17)
const MaxIterations int = 17

// Max number of pool points that can be consumed per tx. This roughly corresponds
// to the maximum execution time (in ms) of protorev per tx
const MaxPoolPointsPerTx uint64 = 50

// Max number of pool points that can be consumed per block (default of 100). This roughly corresponds
// to the maximum execution time (in ms) of protorev per block
const MaxPoolPointsPerBlock uint64 = 200

// ---------------- Module Profit Splitting Constants ---------------- //

// Year 1 (20% of total profit)
const Phase1Length uint64 = 365
const ProfitSplitPhase1 int64 = 20

// Year 2 (10% of total profit)
const Phase2Length uint64 = 730
const ProfitSplitPhase2 int64 = 10

// All other years (5% of total profit)
const ProfitSplitPhase3 int64 = 5
