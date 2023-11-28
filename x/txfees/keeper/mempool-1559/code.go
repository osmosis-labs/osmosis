package mempool1559

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
)

/*
   This is the logic for the Osmosis implementation for EIP-1559 fee market,
	 the goal of this code is to prevent spam by charging more for transactions when the network is busy.

	 This logic does two things:
   - Maintaining data parsed from chain transaction execution and updating eipState accordingly.
   - Resetting eipState to default every ResetInterval (1000) block height intervals to maintain consistency.

   Additionally:
   - Periodically evaluating CheckTx and RecheckTx for compliance with these parameters.

   Note: The reset interval is set to 2000 blocks, which is approximately 4 hours. Consider adjusting for a smaller time interval (e.g., 500 blocks = 1 hour) if necessary.

   Challenges:
   - Transactions falling under their gas bounds are currently discarded by nodes. This behavior can be modified for CheckTx, rather than RecheckTx.

   Global variables stored in memory:
   - DefaultBaseFee: Default base fee, initialized to 0.01.
   - MinBaseFee: Minimum base fee, initialized to 0.0025.
   - MaxBaseFee: Maximum base fee, initialized to 10.
   - MaxBlockChangeRate: The maximum block change rate, initialized to 1/10.

   Global constants:
   - TargetGas: Gas wanted per block, initialized to 70,000,000.
   - ResetInterval: The interval at which eipState is reset, initialized to 1000 blocks.
   - BackupFile: File for backup, set to "eip1559state.json".
   - RecheckFeeConstant: A constant value for rechecking fees, initialized to 4.
*/

var (
	DefaultBaseFee = sdk.MustNewDecFromStr("0.01")
	MinBaseFee     = sdk.MustNewDecFromStr("0.0025")
	MaxBaseFee     = sdk.MustNewDecFromStr("10")

	// Max increase per block is a factor of 15/14, max decrease is 9/10
	MaxBlockChangeRate = sdk.NewDec(1).Quo(sdk.NewDec(10))
	TargetGas          = int64(70_000_000)
	// In face of continuous spam, will take ~21 blocks from base fee > spam cost, to mempool eviction
	// ceil(log_{15/14}(RecheckFee mnConstant))
	// So potentially 2 minutes of impaired UX from 1559 nodes on top of time to get to base fee > spam.
	RecheckFeeConstant = int64(4)
	ResetInterval      = int64(2000)
)

const (
	BackupFilename = "eip1559state.json"
)

// EipState tracks the current base fee and totalGasWantedThisBlock
// this structure is never written to state
type EipState struct {
	lastBlockHeight         int64
	totalGasWantedThisBlock int64
	BackupFilePath          string
	CurBaseFee              osmomath.Dec `json:"cur_base_fee"`
}

// CurEipState is a global variable used in the BeginBlock, EndBlock and
// DeliverTx (fee decorator AnteHandler) functions, it's also using when determining
// if a transaction has enough gas to successfully execute
var CurEipState = EipState{
	lastBlockHeight:         0,
	totalGasWantedThisBlock: 0,
	BackupFilePath:          "",
	CurBaseFee:              sdk.NewDec(0),
}

// startBlock is executed at the start of each block and is responsible for resetting the state
// of the CurBaseFee when the node reaches the reset interval
func (e *EipState) startBlock(height int64) {
	e.lastBlockHeight = height
	e.totalGasWantedThisBlock = 0

	if e.CurBaseFee.Equal(sdk.NewDec(0)) {
		// CurBaseFee has not been initialized yet. This only happens when the node has just started.
		// Try to read the previous value from the backup file and if not available, set it to the default.
		e.CurBaseFee = e.tryLoad()
	}

	// we reset the CurBaseFee every ResetInterval
	if height%ResetInterval == 0 {
		e.CurBaseFee = DefaultBaseFee.Clone()
	}
}

func (e EipState) Clone() EipState {
	e.CurBaseFee = e.CurBaseFee.Clone()
	return e
}

// deliverTxCode runs on every transaction in the feedecorator ante handler and sums the gas of each transaction
func (e *EipState) deliverTxCode(ctx sdk.Context, tx sdk.FeeTx) {
	if ctx.BlockHeight() != e.lastBlockHeight {
		ctx.Logger().Error("Something is off here? ctx.BlockHeight() != e.lastBlockHeight", ctx.BlockHeight(), e.lastBlockHeight)
	}
	e.totalGasWantedThisBlock += int64(tx.GetGas())
}

// updateBaseFee updates of a base fee in Osmosis.
// It employs the following equation to calculate the new base fee:
//
//	baseFeeMultiplier = 1 + (gasUsed - targetGas) / targetGas * maxChangeRate
//	newBaseFee = baseFee * baseFeeMultiplier
//
// updateBaseFee runs at the end of every block
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

	// Enforce the minimum base fee by resetting the CurBaseFee is it drops below the MinBaseFee
	if e.CurBaseFee.LT(MinBaseFee) {
		e.CurBaseFee = MinBaseFee.Clone()
	}

	// Enforce the maximum base fee by resetting the CurBaseFee is it goes above the MaxBaseFee
	if e.CurBaseFee.GT(MaxBaseFee) {
		e.CurBaseFee = MaxBaseFee.Clone()
	}

	go e.Clone().tryPersist()
}

// GetCurBaseFee returns a clone of the CurBaseFee to avoid overwriting the initial value in
// the EipState, we use this in the AnteHandler to Check transactions
func (e *EipState) GetCurBaseFee() osmomath.Dec {
	return e.CurBaseFee.Clone()
}

// GetCurRecheckBaseFee returns a clone of the CurBaseFee / RecheckFeeConstant to account for
// rechecked transactions in the feedecorator ante handler
func (e *EipState) GetCurRecheckBaseFee() osmomath.Dec {
	return e.CurBaseFee.Clone().Quo(sdk.NewDec(RecheckFeeConstant))
}

var rwMtx = sync.Mutex{}

// tryPersist persists the eip1559 state to disk in the form of a json file
// we do this in case a node stops and it can continue functioning as normal
func (e EipState) tryPersist() {
	bz, err := json.Marshal(e)
	if err != nil {
		fmt.Println("Error marshalling eip1559 state", err)
		return
	}
	rwMtx.Lock()
	defer rwMtx.Unlock()

	err = os.WriteFile(e.BackupFilePath, bz, 0644)
	if err != nil {
		fmt.Println("Error writing eip1559 state", err)
		return
	}
}

// tryLoad reads eip1559 state from disk and initializes the CurEipState to
// the previous state when a node is restarted
func (e *EipState) tryLoad() osmomath.Dec {
	rwMtx.Lock()
	defer rwMtx.Unlock()
	bz, err := os.ReadFile(e.BackupFilePath)
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
