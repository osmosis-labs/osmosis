package mempool1559

import (
	"testing"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/assert"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/noapptest"
)

// TestUpdateBaseFee simulates the update of a base fee in Osmosis.
// It employs the following equation to calculate the new base fee:
//
//	baseFeeMultiplier = 1 + (gasUsed - targetGas) / targetGas * maxChangeRate
//	newBaseFee = baseFee * baseFeeMultiplier
//
// The function iterates through a series of simulated blocks and transactions,
// updating and validating the base fee at each step to ensure it follows the equation.
func TestUpdateBaseFee(t *testing.T) {
	// Create an instance of eipState
	eip := &EipState{
		currentBlockHeight:      0,
		totalGasWantedThisBlock: 0,
		CurBaseFee:              DefaultBaseFee.Clone(),
	}

	// we iterate over 1000 blocks as the reset happens after 1000 blocks
	for i := 1; i <= 1002; i++ {
		// create a new block
		ctx := sdk.NewContext(nil, tmproto.Header{Height: int64(i)}, false, log.NewNopLogger())

		// start the new block
		eip.startBlock(int64(i), ctx.Logger())

		// generate transactions
		if i%10 == 0 {
			for j := 1; j <= 3; j++ {
				tx := GenTx(uint64(500000000 + i))
				eip.deliverTxCode(ctx, tx.(sdk.FeeTx))
			}
		}
		baseFeeBeforeUpdate := eip.GetCurBaseFee()

		// update base fee
		eip.updateBaseFee(int64(i))

		// calculate the base fees
		expectedBaseFee := calculateBaseFee(eip.totalGasWantedThisBlock, baseFeeBeforeUpdate)

		// Assert that the actual result matches the expected result
		assert.DeepEqual(t, expectedBaseFee, eip.CurBaseFee)
	}
}

// calculateBaseFee is the same as in is defined on the eip1559 code
func calculateBaseFee(totalGasWantedThisBlock int64, eipStateCurBaseFee osmomath.Dec) (expectedBaseFee osmomath.Dec) {
	gasUsed := totalGasWantedThisBlock
	gasDiff := gasUsed - TargetGas

	baseFeeIncrement := osmomath.NewDec(gasDiff).Quo(osmomath.NewDec(TargetGas)).Mul(MaxBlockChangeRate)
	expectedBaseFeeMultiplier := osmomath.NewDec(1).Add(baseFeeIncrement)
	expectedBaseFee = eipStateCurBaseFee.MulMut(expectedBaseFeeMultiplier)

	if expectedBaseFee.LT(MinBaseFee) {
		expectedBaseFee = MinBaseFee
	}

	if expectedBaseFee.GT(MaxBaseFee) {
		expectedBaseFee = MaxBaseFee.Clone()
	}

	return expectedBaseFee
}

// GenTx generates a mock gas transaction.
func GenTx(gas uint64) sdk.Tx {
	gen := noapptest.MakeTestEncodingConfig().TxConfig
	txBuilder := gen.NewTxBuilder()
	txBuilder.SetGasLimit(gas)
	return txBuilder.GetTx()
}
