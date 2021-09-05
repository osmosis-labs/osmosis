///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const GenesisState$json = const {
  '1': 'GenesisState',
  '2': const [
    const {'1': 'channels', '3': 1, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.IdentifiedChannel', '8': const {}, '10': 'channels'},
    const {'1': 'acknowledgements', '3': 2, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketState', '8': const {}, '10': 'acknowledgements'},
    const {'1': 'commitments', '3': 3, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketState', '8': const {}, '10': 'commitments'},
    const {'1': 'receipts', '3': 4, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketState', '8': const {}, '10': 'receipts'},
    const {'1': 'send_sequences', '3': 5, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketSequence', '8': const {}, '10': 'sendSequences'},
    const {'1': 'recv_sequences', '3': 6, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketSequence', '8': const {}, '10': 'recvSequences'},
    const {'1': 'ack_sequences', '3': 7, '4': 3, '5': 11, '6': '.ibc.core.channel.v1.PacketSequence', '8': const {}, '10': 'ackSequences'},
    const {'1': 'next_channel_sequence', '3': 8, '4': 1, '5': 4, '8': const {}, '10': 'nextChannelSequence'},
  ],
};

const PacketSequence$json = const {
  '1': 'PacketSequence',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'channelId'},
    const {'1': 'sequence', '3': 3, '4': 1, '5': 4, '10': 'sequence'},
  ],
};

