package simulation

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type simulatorState struct {
	opCount int

	nextValidators mockValidators

	// We technically have to store past block times for every block within the unbonding period.
	// For simplicity, we take the RAM overhead and store all past times.
	pastTimes     []time.Time
	pastVoteInfos [][]abci.VoteInfo

	// eventStats is an obviously bad design, but for now we leave it as future
	// work for us to clean up and architect well.
	// We should be collecting this raw data, and able to stream it out into a database.
	// Its fine to keep some basic aggregate statistics, but not where it should end.
	eventStats EventStats

	header tmproto.Header
}
