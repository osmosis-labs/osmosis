package mempool1559

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/log"

	osmomath "github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/x/txfees/types"
)

/*
   This is the logic for the Osmosis implementation for EIP-1559 fee market,
	 the goal of this code is to prevent spam by charging more for transactions when the network is busy.

	 This logic does two things:
   - Maintaining data parsed from chain transaction execution and updating eipState accordingly.
   - Resetting eipState to default every ResetInterval (6000) block height intervals to maintain consistency.

   Additionally:
   - Periodically evaluating CheckTx and RecheckTx for compliance with these parameters.

   Note: The reset interval is set to 6000 blocks, which is approximately 8.5 hours.

   Challenges:
   - Transactions falling under their gas bounds are currently discarded by nodes. This behavior can be modified for CheckTx, rather than RecheckTx.

   Global variables stored in memory:
   - DefaultBaseFee: Default base fee, initialized to 0.005.
   - MinBaseFee: Minimum base fee, initialized to 0.01.
   - MaxBaseFee: Maximum base fee, initialized to 5.
   - MaxBlockChangeRate: The maximum block change rate, initialized to 1/10.

   Global constants:
   - TargetGas: Gas wanted per block, initialized to .625 * block_gas_limt = 187.5 million.
   - ResetInterval: The interval at which eipState is reset, initialized to 6000 blocks.
   - BackupFile: File for backup, set to "eip1559state.json".
   - RecheckFeeConstant: A constant value for rechecking fees, initialized to 2.25.
*/

var (
	// We expect wallet multiplier * DefaultBaseFee < MinBaseFee * RecheckFeeConstant
	// conservatively assume a wallet multiplier of at least 7%.
	DefaultBaseFee = osmomath.MustNewDecFromStr("0.0250")
	MinBaseFee     = types.ConsensusMinFee
	MaxBaseFee     = osmomath.MustNewDecFromStr("5")
	ResetInterval  = int64(6000)

	// Max increase per block is a factor of 1.06, max decrease is 9/10
	// If recovering at ~30M gas per block, decrease is .916
	MaxBlockChangeRate      = osmomath.NewDec(1).Quo(osmomath.NewDec(10))
	TargetGas               = int64(187_500_000)
	TargetBlockSpacePercent = osmomath.MustNewDecFromStr("0.625")

	// N.B. on the reason for having two base fee constants for high and low fees:
	//
	// At higher base fees, we apply a smaller re-check factor.
	// The reason for this is that the recheck factor forces the base fee to get at minimum
	// "recheck factor" times higher than the spam rate. This leads to slow recovery
	// and a bad UX for user transactions. We aim for spam to start getting evicted from the mempool
	// sooner as to avoid more severe UX degradation for user transactions. Therefore,
	// we apply a smaller recheck factor at higher base fees.
	//
	// For low base fees:
	// In face of continuous spam, will take ~19 blocks from base fee > spam cost, to mempool eviction
	// ceil(log_{1.06}(RecheckFeeConstantLowBaseFee)) (assuming base fee not going over threshold)
	// So potentially 1.2 minutes of impaired UX from 1559 nodes on top of time to get to base fee > spam.
	RecheckFeeConstantLowBaseFee = "3"
	//
	// For high base fees:
	// In face of continuous spam, will take ~15 blocks from base fee > spam cost, to mempool eviction
	// ceil(log_{1.06}(RecheckFeeConstantHighBaseFee)) (assuming base fee surpasses threshold)
	RecheckFeeConstantHighBaseFee = "2.3"
	// Note, the choice of 0.01 was made by observing base fee metrics on mainnet and selecting
	// this value from Grafana dashboards. The observation is that below this threshold, we do not
	// observe user UX degradation. Therefore, we keep the original recheck factor.
	RecheckFeeBaseFeeThreshold = osmomath.MustNewDecFromStr("0.01")
)

var (
	RecheckFeeLowBaseFeeDec  = osmomath.MustNewDecFromStr(RecheckFeeConstantLowBaseFee)
	RecheckFeeHighBaseFeeDec = osmomath.MustNewDecFromStr(RecheckFeeConstantHighBaseFee)
)

const (
	BackupFilename = "eip1559state.json"
)

// EipState tracks the current base fee and totalGasWantedThisBlock
// this structure is never written to state
type EipState struct {
	currentBlockHeight      int64
	totalGasWantedThisBlock int64
	BackupFilePath          string
	CurBaseFee              osmomath.Dec `json:"cur_base_fee"`
}

// CurEipState is a global variable used in the BeginBlock, EndBlock and
// DeliverTx (fee decorator AnteHandler) functions, it's also using when determining
// if a transaction has enough gas to successfully execute
var CurEipState = EipState{
	currentBlockHeight:      0,
	totalGasWantedThisBlock: 0,
	BackupFilePath:          "",
	CurBaseFee:              osmomath.NewDec(0),
}

// startBlock is executed at the start of each block and is responsible for resetting the state
// of the CurBaseFee when the node reaches the reset interval
func (e *EipState) startBlock(height int64, logger log.Logger) {
	e.currentBlockHeight = height
	e.totalGasWantedThisBlock = 0

	if e.CurBaseFee.Equal(osmomath.NewDec(0)) {
		// CurBaseFee has not been initialized yet. This only happens when the node has just started.
		// Try to read the previous value from the backup file and if not available, set it to the default.
		e.CurBaseFee = e.tryLoad(logger)
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
	if ctx.BlockHeight() != e.currentBlockHeight {
		ctx.Logger().Error(fmt.Sprintf("Something is off here? ctx.BlockHeight() (%d) != e.currentBlockHeight (%d)", ctx.BlockHeight(), e.currentBlockHeight))
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
	if height != e.currentBlockHeight {
		fmt.Println("Something is off here? height != e.currentBlockHeight", height, e.currentBlockHeight)
	}

	// N.B. we set the lastBlockHeight to height + 1 to avoid the case where block sdk submits a update proposal
	// tx prior to the eip startBlock being called (which is a begin block call).
	e.currentBlockHeight = height + 1

	gasUsed := e.totalGasWantedThisBlock
	gasDiff := gasUsed - TargetGas
	//  (gasUsed - targetGas) / targetGas * maxChangeRate
	baseFeeIncrement := osmomath.NewDec(gasDiff).Quo(osmomath.NewDec(TargetGas)).Mul(MaxBlockChangeRate)
	baseFeeMultiplier := osmomath.NewDec(1).Add(baseFeeIncrement)
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
	baseFee := e.CurBaseFee.Clone()

	// At higher base fees, we apply a smaller re-check factor.
	// The reason for this is that the recheck factor forces the base fee to get at minimum
	// "recheck factor" times higher than the spam rate. This leads to slow recovery
	// and a bad UX for user transactions. We aim for spam to start getting evicted from the mempool
	// sooner as to avoid more severe UX degradation for user transactions. Therefore,
	// we apply a smaller recheck factor at higher base fees.
	if baseFee.GT(RecheckFeeBaseFeeThreshold) {
		return baseFee.QuoMut(RecheckFeeHighBaseFeeDec)
	}

	return baseFee.QuoMut(RecheckFeeLowBaseFeeDec)
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
func (e *EipState) tryLoad(logger log.Logger) osmomath.Dec {
	rwMtx.Lock()
	defer rwMtx.Unlock()
	bz, err := os.ReadFile(e.BackupFilePath)
	if err != nil {
		logger.Debug("Error reading eip1559 state", "err", err)
		logger.Debug("Setting eip1559 state to default value", "MinBaseFee", MinBaseFee)
		return MinBaseFee.Clone()
	}

	var loaded EipState
	err = json.Unmarshal(bz, &loaded)
	if err != nil {
		logger.Debug("Error unmarshalling eip1559 state", "err", err)
		logger.Debug("Setting eip1559 state to default value", "MinBaseFee", MinBaseFee)
		return MinBaseFee.Clone()
	}

	logger.Info("Loaded eip1559 state", "CurBaseFee", loaded.CurBaseFee)
	if loaded.CurBaseFee.LT(MinBaseFee) {
		logger.Debug("CurBaseFee is less than MinBaseFee, setting to MinBaseFee", "CurBaseFee", loaded.CurBaseFee, "MinBaseFee", MinBaseFee)
		return MinBaseFee.Clone()
	}
	if loaded.CurBaseFee.GT(MaxBaseFee) {
		logger.Debug("CurBaseFee is greater than MaxBaseFee, setting to MaxBaseFee", "CurBaseFee", loaded.CurBaseFee, "MaxBaseFee", MaxBaseFee)
		return MaxBaseFee.Clone()
	}
	return loaded.CurBaseFee.Clone()
}
