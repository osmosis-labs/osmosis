package e2eTesting

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clientTypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectionTypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channelTypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	commitmentTypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	ibcCmTypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/stretchr/testify/require"
)

type (
	// IBCPath keeps IBC endpoints in sync and provides helper relayer functions.
	// Heavily inspired by the Path from the ibc-go repo (https://github.com/cosmos/ibc-go/blob/main/testing/path.go).
	// Reasons for no using the ibc-go's version are: simplify it, custom TestChain usage, hiding some public methods
	// making it more usable without knowing how it works under the hood.
	IBCPath struct {
		t *testing.T

		a *IBCEndpoint
		b *IBCEndpoint
	}

	IBCEndpoint struct {
		t *testing.T

		srcChain    *TestChain
		dstEndpoint *IBCEndpoint

		chPort    string
		chVersion string
		chOrder   channelTypes.Order

		clientID     string
		connectionID string
		channelID    string
	}
)

// NewIBCPath builds a new IBCPath relayer creating a connection between two IBC endpoints.
func NewIBCPath(t *testing.T, srcChain, dstChain *TestChain, srcChPort, dstChPort, chVersion string, chOrder channelTypes.Order) *IBCPath {
	require.True(t, chOrder == channelTypes.ORDERED || chOrder == channelTypes.UNORDERED)

	p := IBCPath{
		t: t,
		a: &IBCEndpoint{
			t:         t,
			srcChain:  srcChain,
			chPort:    srcChPort,
			chVersion: chVersion,
			chOrder:   chOrder,
		},
		b: &IBCEndpoint{
			t:         t,
			srcChain:  dstChain,
			chPort:    dstChPort,
			chVersion: chVersion,
			chOrder:   chOrder,
		},
	}
	p.a.dstEndpoint, p.b.dstEndpoint = p.b, p.a

	// Skip a block to ensure lastHeader is set
	p.a.srcChain.NextBlock(0)
	p.b.srcChain.NextBlock(0)

	// Create clients
	p.a.createIBCClient()
	p.b.createIBCClient()

	// Create connection
	p.a.sendConnectionOpenInit()
	p.b.sendConnectionOpenTry()
	p.a.sendConnectionOpenAck()
	p.b.sendConnectionOpenConfirm()
	p.a.updateIBCClient()

	// Create channel
	p.a.sendChannelOpenInit()
	p.b.sendChannelOpenTry()
	p.a.sendChannelOpenAck()
	p.b.sendChannelOpenConfirm()
	p.a.updateIBCClient()

	return &p
}

// EndpointA returns the IBC endpoint A.
func (p *IBCPath) EndpointA() *IBCEndpoint {
	return p.a
}

// EndpointB returns the IBC endpoint A.
func (p *IBCPath) EndpointB() *IBCEndpoint {
	return p.b
}

// RelayPacket attempts to relay a packet first on EndpointA and then on EndpointB.
// Packet must be committed.
func (p *IBCPath) RelayPacket(packet channelTypes.Packet, ack []byte) {
	packetCommitmentRcvA := p.a.srcChain.GetIBCPacketCommitment(packet)
	packetCommitmentExpA := channelTypes.CommitPacket(p.a.srcChain.GetAppCodec(), packet)
	if bytes.Equal(packetCommitmentRcvA, packetCommitmentExpA) {
		// Packet found, relay from A to B
		p.b.updateIBCClient()
		p.b.sendPacketReceive(packet)
		p.a.sendPacketAck(packet, ack)
		return
	}

	packetCommitmentRcvB := p.b.srcChain.GetIBCPacketCommitment(packet)
	packetCommitmentExpB := channelTypes.CommitPacket(p.b.srcChain.GetAppCodec(), packet)
	if bytes.Equal(packetCommitmentRcvB, packetCommitmentExpB) {
		// Packet found, relay from B to A
		p.a.updateIBCClient()
		p.a.sendPacketReceive(packet)
		p.b.sendPacketAck(packet, ack)
		return
	}

	require.Fail(p.t, "packet commitment does not exist on either endpoint for provided packet")
}

// TimeoutPacket timeouts a packet for the specified endpoint skipping block/time.
// Packet must be committed.
func (p *IBCPath) TimeoutPacket(packet channelTypes.Packet, endpoint *IBCEndpoint) {
	t := p.t

	require.True(t, p.a == endpoint || p.b == endpoint)

	// Skip blocks
	if targetBlock := int64(packet.TimeoutHeight.RevisionHeight); targetBlock > 0 {
		curBlock := endpoint.srcChain.GetBlockHeight()
		require.Greater(t, targetBlock, curBlock)

		p.SkipBlocks(targetBlock - curBlock)
	}

	// Skip time
	if targetTime := int64(packet.TimeoutTimestamp); targetTime > 0 {
		curTime := endpoint.srcChain.GetBlockTime().UnixNano()
		require.Greater(t, targetTime, curTime)

		p.SkipTime(endpoint, time.Duration(targetTime-curTime))
	}

	endpoint.sendPacketTimeout(packet)
}

// SkipBlocks skips a number of blocks for both endpoints updating IBC clients.
func (p *IBCPath) SkipBlocks(n int64) {
	for i := int64(0); i < n; i++ {
		p.a.updateIBCClient()
		p.b.updateIBCClient()
	}
}

// SkipTime skips a period for both endpoints updating IBC clients.
func (p *IBCPath) SkipTime(endpoint *IBCEndpoint, dur time.Duration) {
	t := p.t

	require.True(t, p.a == endpoint || p.b == endpoint)

	startTime := endpoint.srcChain.GetBlockTime().UnixNano()
	for {
		endpoint.dstEndpoint.updateIBCClient()
		endpoint.updateIBCClient()

		curTime := endpoint.srcChain.GetBlockTime().UnixNano()
		if curTime > startTime+int64(dur) {
			break
		}
	}
}

// ChannelID returns the endpoint IBC channel ID.
func (e *IBCEndpoint) ChannelID() string {
	return e.channelID
}

// sendPacket sends an IBC packet through the channel keeper and updates the counterparty client.
//
// nolint: unused
func (e *IBCEndpoint) sendPacket(packet exported.PacketI) {
	e.srcChain.SendIBCPacket(packet)
	e.dstEndpoint.updateIBCClient()
	e.updateIBCClient()
}

// sendPacketReceive receives a packet on the source chain and updates the counterparty client.
func (e *IBCEndpoint) sendPacketReceive(packet channelTypes.Packet) {
	srcChain, dstChain := e.srcChain, e.dstEndpoint.srcChain

	packetKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	proof, proofHeight := dstChain.GetProofAtHeight(packetKey, uint64(dstChain.GetBlockHeight()))

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := channelTypes.NewMsgRecvPacket(
		packet,
		proof,
		proofHeight,
		srcChainSenderAcc.Address.String(),
	)

	_, _, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)

	e.dstEndpoint.updateIBCClient()
}

// sendPacketAck sends a MsgAcknowledgement to the channel associated with the source chain.
func (e *IBCEndpoint) sendPacketAck(packet channelTypes.Packet, ack []byte) {
	srcChain, dstChain := e.srcChain, e.dstEndpoint.srcChain

	packetKey := host.PacketAcknowledgementKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	proof, proofHeight := dstChain.GetProofAtHeight(packetKey, uint64(dstChain.GetBlockHeight()))

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := channelTypes.NewMsgAcknowledgement(
		packet,
		ack,
		proof,
		proofHeight,
		srcChainSenderAcc.Address.String(),
	)

	_, _, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)
}

// sendPacketTimeout sends a MsgTimeout to the channel associated with the source chain.
func (e *IBCEndpoint) sendPacketTimeout(packet channelTypes.Packet) {
	srcChain, dstChain := e.srcChain, e.dstEndpoint.srcChain
	srcPortID, srcChannelID := e.chPort, e.channelID

	var packetKey []byte
	switch e.chOrder {
	case channelTypes.ORDERED:
		packetKey = host.NextSequenceRecvKey(packet.GetDestPort(), packet.GetDestChannel())
	case channelTypes.UNORDERED:
		packetKey = host.PacketReceiptKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	}

	proof, proofHeight := dstChain.GetProofAtHeight(packetKey, uint64(dstChain.GetBlockHeight()))
	nextSeqRecv := dstChain.GetNextIBCPacketSequence(srcPortID, srcChannelID)

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := channelTypes.NewMsgTimeout(
		packet,
		nextSeqRecv,
		proof,
		proofHeight,
		srcChainSenderAcc.Address.String(),
	)

	_, _, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)
}

// createIBCClient creates a new IBC TM client.
func (e *IBCEndpoint) createIBCClient() {
	const (
		tmClientTrustPeriod                  = 7 * 24 * time.Hour
		tmClientMaxClockDrift                = 10 * time.Second
		tmClientAllowUpdateAfterExpiry       = false
		tmClientAllowUpdateAfterMisbehaviour = false
	)
	var (
		tmClientTrustLevel  = ibcCmTypes.DefaultTrustLevel
		tmClientUpgradePath = []string{"upgrade", "upgradedIBCState"}
	)

	t, srcChain, dstChain := e.t, e.srcChain, e.dstEndpoint.srcChain

	dstChainLastTMHeader := dstChain.GetTMClientLastHeader()

	clientState := ibcCmTypes.NewClientState(
		dstChain.GetChainID(),
		tmClientTrustLevel,
		tmClientTrustPeriod,
		dstChain.GetUnbondingTime(),
		tmClientMaxClockDrift,
		dstChainLastTMHeader.GetHeight().(clientTypes.Height),
		commitmentTypes.GetSDKSpecs(),
		tmClientUpgradePath,
	)

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg, err := clientTypes.NewMsgCreateClient(
		clientState,
		dstChainLastTMHeader.ConsensusState(),
		srcChainSenderAcc.Address.String(),
	)
	require.NoError(t, err)

	_, res, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(t, err)

	clientID := GetStringEventAttribute(res.Events, clientTypes.EventTypeCreateClient, clientTypes.AttributeKeyClientID)
	require.NotEmpty(t, clientID)
	e.clientID = clientID
}

// updateIBCClient updates an IBC TM client.
func (e *IBCEndpoint) updateIBCClient() {
	t, srcChain, dstChain := e.t, e.srcChain, e.dstEndpoint.srcChain
	srcChainClientID := e.clientID

	dstChain.NextBlock(0)

	header := srcChain.GetTMClientHeaderUpdate(dstChain, srcChainClientID, clientTypes.ZeroHeight())

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg, err := clientTypes.NewMsgUpdateClient(
		e.clientID,
		&header,
		srcChainSenderAcc.Address.String(),
	)
	require.NoError(t, err)

	_, _, _, err = srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)
}

// sendConnectionOpenInit sends a ConnectionOpenInit message to the source chain.
func (e *IBCEndpoint) sendConnectionOpenInit() {
	const (
		defDelayPeriod uint64 = 0
	)
	version := connectionTypes.GetCompatibleVersions()[0]

	t, srcChain, dstChain := e.t, e.srcChain, e.dstEndpoint.srcChain
	srcChainClientID, dstChainClientID := e.clientID, e.dstEndpoint.clientID

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := connectionTypes.NewMsgConnectionOpenInit(
		srcChainClientID,
		dstChainClientID,
		dstChain.GetMerklePrefix(),
		version,
		defDelayPeriod,
		srcChainSenderAcc.Address.String(),
	)

	_, res, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(t, err)

	e.connectionID = GetStringEventAttribute(res.Events, connectionTypes.EventTypeConnectionOpenInit, connectionTypes.AttributeKeyConnectionID)
	require.NotEmpty(t, e.connectionID)
}

// sendConnectionOpenTry sends a ConnectionOpenTry message to the source chain.
func (e *IBCEndpoint) sendConnectionOpenTry() {
	const (
		defDelayPeriod uint64 = 0
	)
	version := connectionTypes.GetCompatibleVersions()[0]

	t, srcChain, dstChain := e.t, e.srcChain, e.dstEndpoint.srcChain
	srcChainClientID, dstChainClientID, srcChainConnectionID, dstChainConnectionID := e.clientID, e.dstEndpoint.clientID, e.connectionID, e.dstEndpoint.connectionID

	e.updateIBCClient()

	clientState, proofClient, proofConsensus, proofInit, consensusHeight, proofHeight := e.getConnectionHandshakeProof()

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := connectionTypes.NewMsgConnectionOpenTry(
		srcChainClientID,
		dstChainConnectionID,
		dstChainClientID,
		clientState,
		dstChain.GetMerklePrefix(),
		[]*connectionTypes.Version{version},
		defDelayPeriod,
		proofInit,
		proofClient,
		proofConsensus,
		proofHeight,
		consensusHeight,
		srcChainSenderAcc.Address.String(),
	)

	_, res, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(t, err)

	if srcChainConnectionID == "" {
		e.connectionID = GetStringEventAttribute(res.Events, connectionTypes.EventTypeConnectionOpenTry, connectionTypes.AttributeKeyConnectionID)
		require.NotEmpty(t, e.connectionID)
	}
}

// sendConnectionOpenAck sends a ConnectionOpenAck message to the source chain.
func (e *IBCEndpoint) sendConnectionOpenAck() {
	version := connectionTypes.GetCompatibleVersions()[0]

	srcChain := e.srcChain
	srcChainConnectionID, dstChainConnectionID := e.connectionID, e.dstEndpoint.connectionID

	e.updateIBCClient()

	clientState, proofClient, proofConsensus, proofTry, consensusHeight, proofHeight := e.getConnectionHandshakeProof()

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := connectionTypes.NewMsgConnectionOpenAck(
		srcChainConnectionID,
		dstChainConnectionID,
		clientState,
		proofTry,
		proofClient,
		proofConsensus,
		proofHeight,
		consensusHeight,
		version,
		srcChainSenderAcc.Address.String(),
	)

	_, _, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)
}

// sendConnectionOpenConfirm sends a ConnectionOpenConfirm message to the source chain.
func (e *IBCEndpoint) sendConnectionOpenConfirm() {
	srcChain, dstChain := e.srcChain, e.dstEndpoint.srcChain
	srcChainConnectionID, dstChainConnectionID := e.connectionID, e.dstEndpoint.connectionID

	e.updateIBCClient()

	connectionKey := host.ConnectionKey(dstChainConnectionID)
	proof, height := dstChain.GetProofAtHeight(connectionKey, uint64(dstChain.GetBlockHeight()))

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := connectionTypes.NewMsgConnectionOpenConfirm(
		srcChainConnectionID,
		proof,
		height,
		srcChainSenderAcc.Address.String(),
	)

	_, _, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)
}

// sendChannelOpenInit sends a ChannelOpenInit message to the source chain.
func (e *IBCEndpoint) sendChannelOpenInit() {
	t, srcChain, srcConnectionID := e.t, e.srcChain, e.connectionID
	srcChPort, dstChPort, chVersion, chOrder := e.chPort, e.dstEndpoint.chPort, e.chVersion, e.chOrder

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := channelTypes.NewMsgChannelOpenInit(
		srcChPort,
		chVersion,
		chOrder,
		[]string{srcConnectionID},
		dstChPort,
		srcChainSenderAcc.Address.String(),
	)

	_, res, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(t, err)

	e.channelID = GetStringEventAttribute(res.Events, channelTypes.EventTypeChannelOpenInit, channelTypes.AttributeKeyChannelID)
	require.NotEmpty(t, e.channelID)
}

// sendChannelOpenTry sends a ChannelOpenTry message to the source chain.
func (e *IBCEndpoint) sendChannelOpenTry() {
	e.updateIBCClient()

	t, srcChain, dstChain, srcConnectionID := e.t, e.srcChain, e.dstEndpoint.srcChain, e.connectionID
	srcChPort, dstChPort, dstChannelID, chVersion, chOrder := e.chPort, e.dstEndpoint.chPort, e.dstEndpoint.channelID, e.chVersion, e.chOrder

	dstChannelKey := host.ChannelKey(dstChPort, dstChannelID)
	proof, height := dstChain.GetProofAtHeight(dstChannelKey, uint64(dstChain.GetBlockHeight()))

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := channelTypes.NewMsgChannelOpenTry(
		srcChPort,
		chVersion,
		chOrder,
		[]string{srcConnectionID},
		dstChPort,
		dstChannelID,
		chVersion,
		proof,
		height,
		srcChainSenderAcc.Address.String(),
	)

	_, res, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(t, err)

	if e.channelID == "" {
		e.channelID = GetStringEventAttribute(res.Events, channelTypes.EventTypeChannelOpenTry, channelTypes.AttributeKeyChannelID)
		require.NotEmpty(t, e.channelID)
	}
}

// sendChannelOpenAck sends a ChannelOpenAck message to the source chain.
func (e *IBCEndpoint) sendChannelOpenAck() {
	e.updateIBCClient()

	srcChain, dstChain := e.srcChain, e.dstEndpoint.srcChain
	srcChPort, dstChPort, srcChannelID, dstChannelID, chVersion := e.chPort, e.dstEndpoint.chPort, e.channelID, e.dstEndpoint.channelID, e.chVersion

	dstChannelKey := host.ChannelKey(dstChPort, dstChannelID)
	proof, height := dstChain.GetProofAtHeight(dstChannelKey, uint64(dstChain.GetBlockHeight()))

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := channelTypes.NewMsgChannelOpenAck(
		srcChPort,
		srcChannelID,
		dstChannelID,
		chVersion,
		proof,
		height,
		srcChainSenderAcc.Address.String(),
	)

	_, _, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)
}

// sendChannelOpenConfirm sends a ChannelOpenConfirm message to the source chain.
func (e *IBCEndpoint) sendChannelOpenConfirm() {
	e.updateIBCClient()

	srcChain, dstChain := e.srcChain, e.dstEndpoint.srcChain
	srcChPort, dstChPort, srcChannelID, dstChannelID := e.chPort, e.dstEndpoint.chPort, e.channelID, e.dstEndpoint.channelID

	dstChannelKey := host.ChannelKey(dstChPort, dstChannelID)
	proof, height := dstChain.GetProofAtHeight(dstChannelKey, uint64(dstChain.GetBlockHeight()))

	srcChainSenderAcc := srcChain.GetAccount(0)
	msg := channelTypes.NewMsgChannelOpenConfirm(
		srcChPort,
		srcChannelID,
		proof,
		height,
		srcChainSenderAcc.Address.String(),
	)

	_, _, _, err := srcChain.SendMsgs(srcChainSenderAcc, true, []sdk.Msg{msg})
	require.NoError(e.t, err)
}

// getConnectionHandshakeProof returns all the proofs necessary to execute OpenTry or OpenAck during the handshake
// from the counterparty chain.
func (e *IBCEndpoint) getConnectionHandshakeProof() (
	dstClientState exported.ClientState,
	dstProofClient, dstProofConsensus, dstProofConnection []byte,
	dstConsensusHeight, dstProofHeight clientTypes.Height,
) {
	srcChain, dstChain, srcChainClientID, dstChainClientID, dstChainConnectionID := e.srcChain, e.dstEndpoint.srcChain, e.clientID, e.dstEndpoint.clientID, e.dstEndpoint.connectionID

	srcClientState := srcChain.GetClientState(srcChainClientID)
	dstClientState = dstChain.GetClientState(dstChainClientID)

	srcClientKey := host.FullClientStateKey(srcChainClientID)
	dstProofClient, dstProofHeight = dstChain.GetProofAtHeight(srcClientKey, srcClientState.GetLatestHeight().GetRevisionHeight())

	dstConsensusHeight = dstClientState.GetLatestHeight().(clientTypes.Height)

	dstConsensusKey := host.FullConsensusStateKey(dstChainClientID, dstConsensusHeight)
	dstProofConsensus, _ = dstChain.GetProofAtHeight(dstConsensusKey, dstProofHeight.GetRevisionHeight())

	dstConnectionKey := host.ConnectionKey(dstChainConnectionID)
	dstProofConnection, _ = dstChain.GetProofAtHeight(dstConnectionKey, dstProofHeight.GetRevisionHeight())

	return
}
