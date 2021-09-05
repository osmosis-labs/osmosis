///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgChannelOpenInit$json = const {
  '1': 'MsgChannelOpenInit',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel', '3': 2, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Channel', '8': const {}, '10': 'channel'},
    const {'1': 'signer', '3': 3, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgChannelOpenInitResponse$json = const {
  '1': 'MsgChannelOpenInitResponse',
};

const MsgChannelOpenTry$json = const {
  '1': 'MsgChannelOpenTry',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'previous_channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'previousChannelId'},
    const {'1': 'channel', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Channel', '8': const {}, '10': 'channel'},
    const {'1': 'counterparty_version', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'counterpartyVersion'},
    const {'1': 'proof_init', '3': 5, '4': 1, '5': 12, '8': const {}, '10': 'proofInit'},
    const {'1': 'proof_height', '3': 6, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'signer', '3': 7, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgChannelOpenTryResponse$json = const {
  '1': 'MsgChannelOpenTryResponse',
};

const MsgChannelOpenAck$json = const {
  '1': 'MsgChannelOpenAck',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'channelId'},
    const {'1': 'counterparty_channel_id', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'counterpartyChannelId'},
    const {'1': 'counterparty_version', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'counterpartyVersion'},
    const {'1': 'proof_try', '3': 5, '4': 1, '5': 12, '8': const {}, '10': 'proofTry'},
    const {'1': 'proof_height', '3': 6, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'signer', '3': 7, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgChannelOpenAckResponse$json = const {
  '1': 'MsgChannelOpenAckResponse',
};

const MsgChannelOpenConfirm$json = const {
  '1': 'MsgChannelOpenConfirm',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'channelId'},
    const {'1': 'proof_ack', '3': 3, '4': 1, '5': 12, '8': const {}, '10': 'proofAck'},
    const {'1': 'proof_height', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'signer', '3': 5, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgChannelOpenConfirmResponse$json = const {
  '1': 'MsgChannelOpenConfirmResponse',
};

const MsgChannelCloseInit$json = const {
  '1': 'MsgChannelCloseInit',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'channelId'},
    const {'1': 'signer', '3': 3, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgChannelCloseInitResponse$json = const {
  '1': 'MsgChannelCloseInitResponse',
};

const MsgChannelCloseConfirm$json = const {
  '1': 'MsgChannelCloseConfirm',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'channelId'},
    const {'1': 'proof_init', '3': 3, '4': 1, '5': 12, '8': const {}, '10': 'proofInit'},
    const {'1': 'proof_height', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'signer', '3': 5, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgChannelCloseConfirmResponse$json = const {
  '1': 'MsgChannelCloseConfirmResponse',
};

const MsgRecvPacket$json = const {
  '1': 'MsgRecvPacket',
  '2': const [
    const {'1': 'packet', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Packet', '8': const {}, '10': 'packet'},
    const {'1': 'proof_commitment', '3': 2, '4': 1, '5': 12, '8': const {}, '10': 'proofCommitment'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'signer', '3': 4, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgRecvPacketResponse$json = const {
  '1': 'MsgRecvPacketResponse',
};

const MsgTimeout$json = const {
  '1': 'MsgTimeout',
  '2': const [
    const {'1': 'packet', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Packet', '8': const {}, '10': 'packet'},
    const {'1': 'proof_unreceived', '3': 2, '4': 1, '5': 12, '8': const {}, '10': 'proofUnreceived'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'next_sequence_recv', '3': 4, '4': 1, '5': 4, '8': const {}, '10': 'nextSequenceRecv'},
    const {'1': 'signer', '3': 5, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgTimeoutResponse$json = const {
  '1': 'MsgTimeoutResponse',
};

const MsgTimeoutOnClose$json = const {
  '1': 'MsgTimeoutOnClose',
  '2': const [
    const {'1': 'packet', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Packet', '8': const {}, '10': 'packet'},
    const {'1': 'proof_unreceived', '3': 2, '4': 1, '5': 12, '8': const {}, '10': 'proofUnreceived'},
    const {'1': 'proof_close', '3': 3, '4': 1, '5': 12, '8': const {}, '10': 'proofClose'},
    const {'1': 'proof_height', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'next_sequence_recv', '3': 5, '4': 1, '5': 4, '8': const {}, '10': 'nextSequenceRecv'},
    const {'1': 'signer', '3': 6, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgTimeoutOnCloseResponse$json = const {
  '1': 'MsgTimeoutOnCloseResponse',
};

const MsgAcknowledgement$json = const {
  '1': 'MsgAcknowledgement',
  '2': const [
    const {'1': 'packet', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Packet', '8': const {}, '10': 'packet'},
    const {'1': 'acknowledgement', '3': 2, '4': 1, '5': 12, '10': 'acknowledgement'},
    const {'1': 'proof_acked', '3': 3, '4': 1, '5': 12, '8': const {}, '10': 'proofAcked'},
    const {'1': 'proof_height', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'signer', '3': 5, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgAcknowledgementResponse$json = const {
  '1': 'MsgAcknowledgementResponse',
};

