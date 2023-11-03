package mempool1559

import (
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
)

// Sections to this right now:
// - Maintain something thats gets parsed from chain tx execution
// update eipState according to that.
// - Every time blockheight % 1000 = 0, reset eipState to default. (B/c this isn't committed on-chain, still gives us some consistency guarantees)
// - Evaluate CheckTx/RecheckTx against this.
//
// 1000 blocks = almost 2 hours, maybe we need a smaller time for resets?
// Lets say 500 blocks = 1 hour
//
// PROBLEMS: Currently, a node will throw out any tx that gets under its gas bound here.
// :OOO We can just do this on checkTx not recheck
//
// Variables we can control for:
// - fees paid per unit gas
// - gas wanted per block (Ethereum)
// - gas used and gas wanted difference
// TODO: Change this percentage update time to be faster

// TODO: Read this from config, can even make default 0, so this is only turned on by nodes who change it!
// ALt: do that with an enable/disable flag. That seems likes a better idea

var DefaultBaseFee = sdk.MustNewDecFromStr("0.0025")
var MinBaseFee = sdk.MustNewDecFromStr("0.0025")
var MaxBaseFee = sdk.MustNewDecFromStr("10")
var TargetGas = int64(60_000_000)
var MaxBlockChangeRate = sdk.NewDec(1).Quo(sdk.NewDec(16))
var ResetInterval = int64(1000)
var BackupFile = "eip1559state.json"
var RecheckFeeConstant = int64(4)

type EipState struct {
	// Signal when we are starting a new block
	// TODO: Or just use begin block
	lastBlockHeight         int64
	totalGasWantedThisBlock int64

	CurBaseFee osmomath.Dec `json:"cur_base_fee"`
}

var CurEipState = EipState{
	lastBlockHeight:         0,
	totalGasWantedThisBlock: 0,
	CurBaseFee:              sdk.NewDec(0),
}

func (e *EipState) startBlock(height int64) {
	e.lastBlockHeight = height
	e.totalGasWantedThisBlock = 0

	if e.CurBaseFee.Equal(sdk.NewDec(0)) {
		// CurBaseFee has not been initialized yet. This only happens when the node has just started.
		// Try to read the previous value from the backup file and if not available, set it to the default.
		e.CurBaseFee = e.tryLoad()
	}

	if height%ResetInterval == 0 {
		e.CurBaseFee = DefaultBaseFee.Clone()
	}
}

func (e *EipState) deliverTxCode(ctx sdk.Context, tx sdk.FeeTx) {
	if ctx.BlockHeight() != e.lastBlockHeight {
		fmt.Println("Something is off here? ctx.BlockHeight() != e.lastBlockHeight", ctx.BlockHeight(), e.lastBlockHeight)
	}
	e.totalGasWantedThisBlock += int64(tx.GetGas())
	// fmt.Println("height, tx gas, blockGas", ctx.BlockHeight(), tx.GetGas(), e.totalGasWantedThisBlock)
}

// Equation is:
// baseFeeMultiplier = 1 + (gasUsed - targetGas) / targetGas * maxChangeRate
// newBaseFee = baseFee * baseFeeMultiplier
func (e *EipState) updateBaseFee(height int64) {
	if height != e.lastBlockHeight {
		fmt.Println("Something is off here? height != e.lastBlockHeight", height, e.lastBlockHeight)
	}
	e.lastBlockHeight = height

	gasUsed := e.totalGasWantedThisBlock
	gasDiff := gasUsed - TargetGas
	//  (gasUsed - targetGas) / targetGas * maxChangeRate
	baseFeeIncrement := sdk.NewDec(gasDiff).Quo(sdk.NewDec(TargetGas)).Mul(MaxBlockChangeRate)
	baseFeeMultiplier := sdk.NewDec(1).Add(baseFeeIncrement)
	e.CurBaseFee.MulMut(baseFeeMultiplier)

	// Make a min base fee
	if e.CurBaseFee.LT(MinBaseFee) {
		e.CurBaseFee = MinBaseFee.Clone()
	}

	// Make a max base fee
	if e.CurBaseFee.GT(MaxBaseFee) {
		e.CurBaseFee = MaxBaseFee.Clone()
	}

	go e.tryPersist()
}

func (e *EipState) GetCurBaseFee() osmomath.Dec {
	return e.CurBaseFee.Clone()
}

func (e *EipState) GetCurRecheckBaseFee() osmomath.Dec {
	return e.CurBaseFee.Clone().Quo(sdk.NewDec(RecheckFeeConstant))
}

func (e *EipState) tryPersist() {
	bz, err := json.Marshal(e)
	if err != nil {
		fmt.Println("Error marshalling eip1559 state", err)
		return
	}

	err = os.WriteFile(BackupFile, bz, 0644)
	if err != nil {
		fmt.Println("Error writing eip1559 state", err)
		return
	}
}

func (e *EipState) tryLoad() osmomath.Dec {
	bz, err := os.ReadFile(BackupFile)
	if err != nil {
		fmt.Println("Error reading eip1559 state", err)
		fmt.Println("Setting eip1559 state to default value", MinBaseFee)
		return MinBaseFee.Clone()
	}

	var loaded EipState
	err = json.Unmarshal(bz, &loaded)
	if err != nil {
		fmt.Println("Error unmarshalling eip1559 state", err)
		fmt.Println("Setting eip1559 state to default value", MinBaseFee)
		return MinBaseFee.Clone()
	}

	fmt.Println("Loaded eip1559 state. CurBaseFee=", loaded.CurBaseFee)
	return loaded.CurBaseFee.Clone()
}
