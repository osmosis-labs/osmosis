package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"golang.org/x/exp/maps"

	markov "github.com/osmosis-labs/osmosis/v10/simulation/types/transitionmatrix"
)

type mockValidator struct {
	val           abci.ValidatorUpdate
	livenessState int
}

func (mv mockValidator) String() string {
	return fmt.Sprintf("mockValidator{%s power:%v state:%v}",
		mv.val.PubKey.String(),
		mv.val.Power,
		mv.livenessState)
}

type mockValidators map[string]mockValidator

// get mockValidators from abci validators
func newMockValidators(r *rand.Rand, abciVals []abci.ValidatorUpdate, params Params) mockValidators {
	validators := make(mockValidators)

	for _, validator := range abciVals {
		str := fmt.Sprintf("%X", validator.PubKey.GetEd25519())
		liveliness := markov.GetMemberOfInitialState(r, params.InitialLivenessWeightings())

		validators[str] = mockValidator{
			val:           validator,
			livenessState: liveliness,
		}
	}

	return validators
}

func (mv mockValidators) Clone() mockValidators {
	validators := make(mockValidators, len(mv))
	keys := mv.getKeys()
	for _, k := range keys {
		validators[k] = mv[k]
	}
	return validators
}

// TODO describe usage
func (mv mockValidators) getKeys() []string {
	keys := maps.Keys(mv)
	sort.Strings(keys)
	return keys
}

// randomProposer picks a random proposer from the current validator set
func (mv mockValidators) randomProposer(r *rand.Rand) tmbytes.HexBytes {
	keys := mv.getKeys()
	if len(keys) == 0 {
		return nil
	}

	key := keys[r.Intn(len(keys))]

	proposer := mv[key].val
	pk, err := cryptoenc.PubKeyFromProto(proposer.PubKey)
	if err != nil { //nolint:wsl
		panic(err)
	}

	return pk.Address()
}

// updateValidators mimics Tendermint's update logic.
func updateValidators(
	tb testing.TB,
	r *rand.Rand,
	params simulation.Params,
	current map[string]mockValidator,
	updates []abci.ValidatorUpdate,
	// logWriter LogWriter,
	event func(route, op, evResult string),
) map[string]mockValidator {
	nextSet := mockValidators(current).Clone()
	for _, update := range updates {
		str := fmt.Sprintf("%X", update.PubKey.GetEd25519())

		if update.Power == 0 {
			if _, ok := nextSet[str]; !ok {
				tb.Fatalf("tried to delete a nonexistent validator: %s", str)
			}

			// logWriter.AddEntry(NewOperationEntry())("kicked", str)
			event("end_block", "validator_updates", "kicked")
			delete(nextSet, str)
		} else if _, ok := nextSet[str]; ok {
			// validator already exists, update weight
			nextSet[str] = mockValidator{update, nextSet[str].livenessState}
			event("end_block", "validator_updates", "updated")
		} else {
			// Set this new validator
			nextSet[str] = mockValidator{
				update,
				markov.GetMemberOfInitialState(r, params.InitialLivenessWeightings()),
			}
			event("end_block", "validator_updates", "added")
		}
	}

	return nextSet
}

// RandomRequestBeginBlock generates a list of signing validators according to
// the provided list of validators, signing fraction, and evidence fraction
func RandomRequestBeginBlock(r *rand.Rand, params Params,
	validators mockValidators, pastTimes []time.Time,
	pastVoteInfos [][]abci.VoteInfo,
	event func(route, op, evResult string), header tmproto.Header,
) abci.RequestBeginBlock {
	if len(validators) == 0 {
		return abci.RequestBeginBlock{
			Header: header,
		}
	}

	voteInfos := randomVoteInfos(r, params, validators, event)
	evidence := randomDoubleSignEvidence(r, params, validators, pastTimes, pastVoteInfos, event, header, voteInfos)

	return abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: voteInfos,
		},
		ByzantineValidators: evidence,
	}
}

func randomVoteInfos(r *rand.Rand, simParams Params, validators mockValidators, event func(route, op, evResult string),
) []abci.VoteInfo {
	voteInfos := make([]abci.VoteInfo, len(validators))

	for i, key := range validators.getKeys() {
		mVal := validators[key]
		mVal.livenessState = simParams.LivenessTransitionMatrix().NextState(r, mVal.livenessState)
		signed := true

		if mVal.livenessState == 1 {
			// spotty connection, 50% probability of success
			// See https://github.com/golang/go/issues/23804#issuecomment-365370418
			// for reasoning behind computing like this
			signed = r.Int63()%2 == 0
		} else if mVal.livenessState == 2 {
			// offline
			signed = false
		}

		if signed {
			event("begin_block", "signing", "signed")
		} else {
			event("begin_block", "signing", "missed")
		}

		pubkey, err := cryptoenc.PubKeyFromProto(mVal.val.PubKey)
		if err != nil {
			panic(err)
		}

		voteInfos[i] = abci.VoteInfo{
			Validator: abci.Validator{
				Address: pubkey.Address(),
				Power:   mVal.val.Power,
			},
			SignedLastBlock: signed,
		}
	}

	return voteInfos
}

func randomDoubleSignEvidence(r *rand.Rand, params Params,
	validators mockValidators, pastTimes []time.Time,
	pastVoteInfos [][]abci.VoteInfo,
	event func(route, op, evResult string), header tmproto.Header, voteInfos []abci.VoteInfo) []abci.Evidence {
	evidence := []abci.Evidence{}
	// return if no past times
	if len(pastTimes) == 0 {
		return evidence
	}

	// TODO: Change this to be markov based & clean this up
	for r.Float64() < params.EvidenceFraction() {
		height := header.Height
		time := header.Time
		vals := voteInfos

		if r.Float64() < params.PastEvidenceFraction() && header.Height > 1 {
			height = int64(r.Intn(int(header.Height)-1)) + 1 // Tendermint starts at height 1
			// array indices offset by one
			time = pastTimes[height-1]
			vals = pastVoteInfos[height-1]
		}

		validator := vals[r.Intn(len(vals))].Validator

		var totalVotingPower int64
		for _, val := range vals {
			totalVotingPower += val.Validator.Power
		}

		evidence = append(evidence,
			abci.Evidence{
				Type:             abci.EvidenceType_DUPLICATE_VOTE,
				Validator:        validator,
				Height:           height,
				Time:             time,
				TotalVotingPower: totalVotingPower,
			},
		)

		event("begin_block", "evidence", "ok")
	}
	return evidence
}
