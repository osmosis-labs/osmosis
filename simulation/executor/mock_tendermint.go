package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"golang.org/x/exp/maps"

	markov "github.com/osmosis-labs/osmosis/v16/simulation/simtypes/transitionmatrix"
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
func (mv mockValidators) randomProposer(r *rand.Rand) crypto.PubKey {
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

	return pk
}

func (mv mockValidators) toTmProtoValidators(proposerPubKey crypto.PubKey) (tmtypes.ValidatorSet, error) {
	var tmProtoValSet tmproto.ValidatorSet
	var tmTypesValSet *tmtypes.ValidatorSet
	// iterate through current validators and add them to the TM ValidatorSet struct
	for _, key := range mv.getKeys() {
		var validator tmproto.Validator
		mapVal := mv[key]
		validator.PubKey = mapVal.val.PubKey
		currentPubKey, err := cryptoenc.PubKeyFromProto(mapVal.val.PubKey)
		if err != nil {
			return *tmTypesValSet, err
		}
		validator.Address = currentPubKey.Address()
		tmProtoValSet.Validators = append(tmProtoValSet.Validators, &validator)
	}

	// set the proposer chosen earlier as the validator set block proposer
	var proposerVal tmtypes.Validator
	proposerVal.PubKey = proposerPubKey
	proposerVal.Address = proposerPubKey.Address()
	blockProposer, err := proposerVal.ToProto()
	if err != nil {
		return *tmTypesValSet, err
	}
	tmProtoValSet.Proposer = blockProposer

	// create a validatorSet type from the tmproto created earlier
	tmTypesValSet, err = tmtypes.ValidatorSetFromProto(&tmProtoValSet)
	return *tmTypesValSet, err
}

// updateValidators mimics Tendermint's update logic.
func updateValidators(
	r *rand.Rand,
	params simulation.Params,
	current map[string]mockValidator,
	updates []abci.ValidatorUpdate,
	// logWriter LogWriter,
	event func(route, op, evResult string),
) (map[string]mockValidator, error) {
	nextSet := mockValidators(current).Clone()
	for _, update := range updates {
		str := fmt.Sprintf("%X", update.PubKey.GetEd25519())

		if update.Power == 0 {
			if _, ok := nextSet[str]; !ok {
				return nil, fmt.Errorf("tried to delete a nonexistent validator: %s", str)
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

	return nextSet, nil
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

	voteInfos := randomVoteInfos(r, params, validators)
	evidence := randomDoubleSignEvidence(r, params, validators, pastTimes, pastVoteInfos, event, header, voteInfos)

	return abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: voteInfos,
		},
		ByzantineValidators: evidence,
	}
}

func randomVoteInfos(r *rand.Rand, simParams Params, validators mockValidators,
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

		// TODO: Do we want to log any data to statsdb here?

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
	event func(route, op, evResult string), header tmproto.Header, voteInfos []abci.VoteInfo,
) []abci.Evidence {
	evidence := []abci.Evidence{}
	// return if no past times or if only 10 validators remaining in the active set
	if len(pastTimes) == 0 {
		return evidence
	}
	var n float64 = 1
	// TODO: Change this to be markov based & clean this up
	// Right now we incrementally lower the evidence fraction to make
	// it less likely to jail many validators in one run.
	// We should also add some method of including new validators into the set
	// instead of being stuck with the ones we start with during initialization.
	for r.Float64() < (params.EvidenceFraction() / n) {
		// if only one validator remaining, don't jail any more validators
		if len(voteInfos)-int(n) <= 0 {
			return nil
		}
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
		n++
	}
	return evidence
}
