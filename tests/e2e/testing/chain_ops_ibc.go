package e2eTesting

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmHash "github.com/cometbft/cometbft/crypto/tmhash"
	cmProto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmProtoVersion "github.com/cometbft/cometbft/proto/tendermint/version"
	cmTypes "github.com/cometbft/cometbft/types"
	cmVersion "github.com/cometbft/cometbft/version"
	clientTypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channelTypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	commitmentTypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibcTmTypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/stretchr/testify/require"
)

// GetMerklePrefix returns a Merkle tree prefix.
// Used to send IBC ConnectionOpenInit msg.
func (chain *TestChain) GetMerklePrefix() commitmentTypes.MerklePrefix {
	return commitmentTypes.NewMerklePrefix(
		chain.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes(),
	)
}

// GetClientState returns an IBC client state for the provided clientID.
// Used to update an IBC TM client.
func (chain *TestChain) GetClientState(clientID string) exported.ClientState {
	t := chain.t

	clientState, found := chain.app.IBCKeeper.ClientKeeper.GetClientState(chain.GetContext(), clientID)
	require.True(t, found)

	return clientState
}

// GetCurrentValSet returns a validator set for the current block height.
// Used to create an IBC TM client header.
func (chain *TestChain) GetCurrentValSet() cmTypes.ValidatorSet {
	return *chain.valSet
}

// GetValSetAtHeight returns a validator set for the specified block height.
// Used to create an IBC TM client header.
func (chain *TestChain) GetValSetAtHeight(height int64) cmTypes.ValidatorSet {
	//	t := chain.t

	// histInfo, err := chain.app.Keepers.StakingKeeper.GetHistoricalInfo(chain.GetContext(), height)
	// require.NoError(t, err)

	// validators := stakingTypes.Validators(histInfo.Valset)
	// tmValidators, err := stakingTestUtil.ToTmValidators(validators, sdk.DefaultPowerReduction)
	// require.NoError(t, err)

	// valSet := tmTypes.NewValidatorSet(tmValidators)

	// return *valSet
	return cmTypes.ValidatorSet{}
}

// GetProofAtHeight returns the proto encoded merkle proof by key for the specified height.
// Used to establish an IBC connection.
func (chain *TestChain) GetProofAtHeight(key []byte, height uint64) ([]byte, clientTypes.Height) {
	t := chain.t

	res, err := chain.app.Query(chain.app.NewContext(false), &abci.RequestQuery{
		Path:   "store/ibc/key",
		Height: int64(height) - 1,
		Data:   key,
		Prove:  true,
	})
	require.NoError(t, err)

	merkleProof, err := commitmentTypes.ConvertProofs(res.ProofOps)
	require.NoError(t, err)

	proof, err := chain.GetAppCodec().Marshal(&merkleProof)
	require.NoError(t, err)

	revision := clientTypes.ParseChainID(chain.GetChainID())

	// Proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree.

	return proof, clientTypes.NewHeight(revision, uint64(res.Height)+1)
}

// SendIBCPacket send an IBC packet and commits chain state.
func (chain *TestChain) SendIBCPacket(packet exported.PacketI) {
	t := chain.t

	require.NotNil(t, packet)

	capPath := host.ChannelCapabilityPath(packet.GetSourcePort(), packet.GetSourceChannel())
	cap, ok := chain.app.ScopedIBCKeeper.GetCapability(chain.GetContext(), capPath)
	require.True(t, ok)

	timeout := clientTypes.Height{
		RevisionNumber: packet.GetTimeoutHeight().GetRevisionNumber(),
		RevisionHeight: packet.GetTimeoutHeight().GetRevisionHeight(),
	}
	_, err := chain.app.IBCKeeper.ChannelKeeper.SendPacket(chain.GetContext(), cap, packet.GetSourcePort(), packet.GetSourceChannel(), timeout, packet.GetTimeoutTimestamp(), packet.GetData())
	require.NoError(t, err)

	chain.NextBlock(0)
}

// GetIBCPacketCommitment returns an IBC packet commitment hash (nil if not committed).
func (chain *TestChain) GetIBCPacketCommitment(packet channelTypes.Packet) []byte {
	return chain.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(
		chain.GetContext(),
		packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(),
	)
}

// GetNextIBCPacketSequence returns the next IBC packet sequence.
func (chain *TestChain) GetNextIBCPacketSequence(portID, channelID string) uint64 {
	t := chain.t

	seq, ok := chain.app.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(chain.GetContext(), portID, channelID)
	require.True(t, ok)

	return seq
}

// GetTMClientLastHeader creates an IBC TM client header from the last committed block.
// Used to create a new client.
func (chain *TestChain) GetTMClientLastHeader() ibcTmTypes.Header {
	return chain.createTMClientHeader(
		chain.GetChainID(),
		chain.lastHeader.Height,
		clientTypes.Height{},
		chain.lastHeader.Time,
		chain.valSet,
		nil,
		chain.valSigners,
	)
}

// GetTMClientHeaderUpdate creates an IBC TM client header to update an existing client on the source chain.
func (chain *TestChain) GetTMClientHeaderUpdate(counterpartyChain *TestChain, clientID string, blockHeightTrusted clientTypes.Height) ibcTmTypes.Header {
	t := chain.t

	header := counterpartyChain.GetTMClientLastHeader()

	if blockHeightTrusted.IsZero() {
		blockHeightTrusted = chain.GetClientState(clientID).GetLatestHeight().(clientTypes.Height)
	}

	// Once we get TrustedHeight from client, we must query the validators from the counterparty chain
	// If the LatestHeight == LastHeader.Height, then TrustedValidators are current validators
	// If LatestHeight < LastHeader.Height, we can query the historical validator set from HistoricalInfo
	var valSetTrusted cmTypes.ValidatorSet
	if blockHeightTrusted == header.GetHeight() {
		valSetTrusted = counterpartyChain.GetCurrentValSet()
	} else {
		// NOTE: We need to get validators from counterparty at height: trustedHeight+1
		// since the last trusted validators for a header at height h
		// is the NextValidators at h+1 committed to in header h by
		// NextValidatorsHash
		valSetTrusted = counterpartyChain.GetValSetAtHeight(int64(blockHeightTrusted.RevisionHeight + 1))
	}

	// Update trusted fields assuming revision number is 0
	header.TrustedHeight = blockHeightTrusted

	valSetTrustedProto, err := valSetTrusted.ToProto()
	require.NoError(t, err)
	header.TrustedValidators = valSetTrustedProto

	return header
}

// createTMClientHeader creates a valid TM client header.
func (chain *TestChain) createTMClientHeader(chainID string, blockHeight int64, blockHeightTrusted clientTypes.Height, blockTime time.Time, valSet, valSetTrusted *cmTypes.ValidatorSet, valSigners []cmTypes.PrivValidator) ibcTmTypes.Header {
	t := chain.t

	require.NotNil(t, valSet)

	valSetHash := valSet.Hash()

	header := cmTypes.Header{
		Version: cmProtoVersion.Consensus{Block: cmVersion.BlockProtocol, App: 2},
		ChainID: chainID,
		Height:  blockHeight,
		Time:    blockTime,
		LastBlockID: cmTypes.BlockID{
			Hash: make([]byte, cmHash.Size),
			PartSetHeader: cmTypes.PartSetHeader{
				Total: 10_000,
				Hash:  make([]byte, cmHash.Size),
			},
		},
		LastCommitHash:     chain.app.LastCommitID().Hash,
		DataHash:           cmHash.Sum([]byte("data_hash")),
		ValidatorsHash:     valSetHash,
		NextValidatorsHash: valSetHash,
		ConsensusHash:      cmHash.Sum([]byte("consensus_hash")),
		AppHash:            chain.lastHeader.AppHash,
		LastResultsHash:    cmHash.Sum([]byte("last_results_hash")),
		EvidenceHash:       cmHash.Sum([]byte("evidence_hash")),
		ProposerAddress:    valSet.Proposer.Address,
	}

	blockID := cmTypes.BlockID{
		Hash: header.Hash(),
		PartSetHeader: cmTypes.PartSetHeader{
			Total: 3,
			Hash:  cmHash.Sum([]byte("part_set")),
		},
	}
	voteSet := cmTypes.NewVoteSet(chainID, blockHeight, 1, cmProto.PrecommitType, valSet)

	commit, err := cmTypes.MakeExtCommit(blockID, blockHeight, 1, voteSet, valSigners, blockTime, false)
	require.NoError(t, err)

	valSetProto, err := valSet.ToProto()
	require.NoError(t, err)

	var valSetTrustedProto *cmProto.ValidatorSet
	if valSetTrusted != nil {
		valSetTrustedProto, err = valSetTrusted.ToProto()
		require.NoError(t, err)
	}

	return ibcTmTypes.Header{
		SignedHeader: &cmProto.SignedHeader{
			Header: header.ToProto(),
			Commit: commit.ToCommit().ToProto(),
		},
		ValidatorSet:      valSetProto,
		TrustedHeight:     blockHeightTrusted,
		TrustedValidators: valSetTrustedProto,
	}
}
