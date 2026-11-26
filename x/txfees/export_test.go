package txfees

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// ResetCachedConsParams clears the cached consensus params, so that the next call to
// CheckAndSetTargetGas behaves as it does on the first block after node start.
func ResetCachedConsParams() {
	cachedConsParams = cmtproto.ConsensusParams{}
}
