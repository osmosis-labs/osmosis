package mempool1559

import sdk "github.com/cosmos/cosmos-sdk/types"

// Sections to this right now:
// - Maintain something thats gets parsed from chain tx execution
// update eipState according to that.
// - Every time blockheight % 1000 = 0, reset eipState to default. (B/c this isn't committed on-chain, still gives us some consistency guarantees)
// - Evaluate CheckTx/RecheckTx against this.
//
// 1000 blocks = almost 2 hours, maybe we need a smaller time for resets?
//
// PROBLEMS: Currently, a node will throw out any tx that gets under its gas bound here.
//
// Variables we can control for:
// - fees paid per unit gas
// - gas wanted per block (Ethereum)
// - gas used and gas wanted difference

var defaultBaseFee = sdk.MustNewDecFromStr("1.0")
var target_gas = int64(90_000_000)
var max_block_change_rate = sdk.NewDec(1).Quo(sdk.NewDec(16))

type eipState struct {
	// Signal when we are starting a new block
	// TODO: Or just use begin block
	lastBlockHeight         int64
	totalGasWantedThisBlock int64

	curBaseFee sdk.Dec
}

var curEipState = eipState{}

// How to get

func (e *eipState) updateBaseFee(height int64) {
	gasUsed := e.totalGasWantedThisBlock
	// obvi fix
	e.lastBlockHeight = height
	gasDiff := gasUsed - target_gas
	baseFeeMultiplier := sdk.NewDec(1).Add(sdk.NewDec((gasDiff)).Mul(max_block_change_rate))
	e.curBaseFee = e.curBaseFee.Mul(baseFeeMultiplier)
}
