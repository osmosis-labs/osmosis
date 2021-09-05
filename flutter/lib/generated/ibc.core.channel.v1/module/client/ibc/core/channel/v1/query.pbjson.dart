///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryChannelRequest$json = const {
  '1': 'QueryChannelRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
  ],
};

const QueryChannelResponse$json = const {
  '1': 'QueryChannelResponse',
  '2': const [
    const {'1': 'channel', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Channel', '10': 'channel'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryChannelsRequest$json = const {
  '1': 'QueryChannelsRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryChannelsResponse$json = const {
  '1': 'QueryChannelsResponse',
  '2': const [
    const {'1': 'channels', '3': 1, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.IdentifiedChannel', '10': 'channels'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
    const {'1': 'height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
  ],
};

const QueryConnectionChannelsRequest$json = const {
  '1': 'QueryConnectionChannelsRequest',
  '2': const [
    const {'1': 'connection', '3': 1, '4': 1, '5': 9, '10': 'connection'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryConnectionChannelsResponse$json = const {
  '1': 'QueryConnectionChannelsResponse',
  '2': const [
    const {'1': 'channels', '3': 1, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.IdentifiedChannel', '10': 'channels'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
    const {'1': 'height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
  ],
};

const QueryChannelClientStateRequest$json = const {
  '1': 'QueryChannelClientStateRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
  ],
};

const QueryChannelClientStateResponse$json = const {
  '1': 'QueryChannelClientStateResponse',
  '2': const [
    const {'1': 'identified_client_state', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.client.v1.IdentifiedClientState', '10': 'identifiedClientState'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryChannelConsensusStateRequest$json = const {
  '1': 'QueryChannelConsensusStateRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'revision_number', '3': 3, '4': 1, '5': 4, '10': 'revisionNumber'},
    const {'1': 'revision_height', '3': 4, '4': 1, '5': 4, '10': 'revisionHeight'},
  ],
};

const QueryChannelConsensusStateResponse$json = const {
  '1': 'QueryChannelConsensusStateResponse',
  '2': const [
    const {'1': 'consensus_state', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Any', '10': 'consensusState'},
    const {'1': 'client_id', '3': 2, '4': 1, '5': 9, '10': 'clientId'},
    const {'1': 'proof', '3': 3, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryPacketCommitmentRequest$json = const {
  '1': 'QueryPacketCommitmentRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'sequence', '3': 3, '4': 1, '5': 4, '10': 'sequence'},
  ],
};

const QueryPacketCommitmentResponse$json = const {
  '1': 'QueryPacketCommitmentResponse',
  '2': const [
    const {'1': 'commitment', '3': 1, '4': 1, '5': 12, '10': 'commitment'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryPacketCommitmentsRequest$json = const {
  '1': 'QueryPacketCommitmentsRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'pagination', '3': 3, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryPacketCommitmentsResponse$json = const {
  '1': 'QueryPacketCommitmentsResponse',
  '2': const [
    const {'1': 'commitments', '3': 1, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketState', '10': 'commitments'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
    const {'1': 'height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
  ],
};

const QueryPacketReceiptRequest$json = const {
  '1': 'QueryPacketReceiptRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'sequence', '3': 3, '4': 1, '5': 4, '10': 'sequence'},
  ],
};

const QueryPacketReceiptResponse$json = const {
  '1': 'QueryPacketReceiptResponse',
  '2': const [
    const {'1': 'received', '3': 2, '4': 1, '5': 8, '10': 'received'},
    const {'1': 'proof', '3': 3, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryPacketAcknowledgementRequest$json = const {
  '1': 'QueryPacketAcknowledgementRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'sequence', '3': 3, '4': 1, '5': 4, '10': 'sequence'},
  ],
};

const QueryPacketAcknowledgementResponse$json = const {
  '1': 'QueryPacketAcknowledgementResponse',
  '2': const [
    const {'1': 'acknowledgement', '3': 1, '4': 1, '5': 12, '10': 'acknowledgement'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryPacketAcknowledgementsRequest$json = const {
  '1': 'QueryPacketAcknowledgementsRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'pagination', '3': 3, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryPacketAcknowledgementsResponse$json = const {
  '1': 'QueryPacketAcknowledgementsResponse',
  '2': const [
    const {'1': 'acknowledgements', '3': 1, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketState', '10': 'acknowledgements'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
    const {'1': 'height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
  ],
};

const QueryUnreceivedPacketsRequest$json = const {
  '1': 'QueryUnreceivedPacketsRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'packet_commitment_sequences', '3': 3, '4': 3, '5': 4, '10': 'packetCommitmentSequences'},
  ],
};

const QueryUnreceivedPacketsResponse$json = const {
  '1': 'QueryUnreceivedPacketsResponse',
  '2': const [
    const {'1': 'sequences', '3': 1, '4': 3, '5': 4, '10': 'sequences'},
    const {'1': 'height', '3': 2, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
  ],
};

const QueryUnreceivedAcksRequest$json = const {
  '1': 'QueryUnreceivedAcksRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
    const {'1': 'packet_ack_sequences', '3': 3, '4': 3, '5': 4, '10': 'packetAckSequences'},
  ],
};

const QueryUnreceivedAcksResponse$json = const {
  '1': 'QueryUnreceivedAcksResponse',
  '2': const [
    const {'1': 'sequences', '3': 1, '4': 3, '5': 4, '10': 'sequences'},
    const {'1': 'height', '3': 2, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
  ],
};

const QueryNextSequenceReceiveRequest$json = const {
  '1': 'QueryNextSequenceReceiveRequest',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '10': 'channelId'},
  ],
};

const QueryNextSequenceReceiveResponse$json = const {
  '1': 'QueryNextSequenceReceiveResponse',
  '2': const [
    const {'1': 'next_sequence_receive', '3': 1, '4': 1, '5': 4, '10': 'nextSequenceReceive'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

