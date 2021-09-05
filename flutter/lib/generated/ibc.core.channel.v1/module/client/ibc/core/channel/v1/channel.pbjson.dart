///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/channel.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const State$json = const {
  '1': 'State',
  '2': const [
    const {'1': 'STATE_UNINITIALIZED_UNSPECIFIED', '2': 0, '3': const {}},
    const {'1': 'STATE_INIT', '2': 1, '3': const {}},
    const {'1': 'STATE_TRYOPEN', '2': 2, '3': const {}},
    const {'1': 'STATE_OPEN', '2': 3, '3': const {}},
    const {'1': 'STATE_CLOSED', '2': 4, '3': const {}},
  ],
  '3': const {},
};

const Order$json = const {
  '1': 'Order',
  '2': const [
    const {'1': 'ORDER_NONE_UNSPECIFIED', '2': 0, '3': const {}},
    const {'1': 'ORDER_UNORDERED', '2': 1, '3': const {}},
    const {'1': 'ORDER_ORDERED', '2': 2, '3': const {}},
  ],
  '3': const {},
};

const Channel$json = const {
  '1': 'Channel',
  '2': const [
    const {'1': 'state', '3': 1, '4': 1, '5': 14, '6': '.ibc.core.channel.v1.State', '10': 'state'},
    const {'1': 'ordering', '3': 2, '4': 1, '5': 14, '6': '.ibc.core.channel.v1.Order', '10': 'ordering'},
    const {'1': 'counterparty', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Counterparty', '8': const {}, '10': 'counterparty'},
    const {'1': 'connection_hops', '3': 4, '4': 3, '5': 9, '8': const {}, '10': 'connectionHops'},
    const {'1': 'version', '3': 5, '4': 1, '5': 9, '10': 'version'},
  ],
  '7': const {},
};

const IdentifiedChannel$json = const {
  '1': 'IdentifiedChannel',
  '2': const [
    const {'1': 'state', '3': 1, '4': 1, '5': 14, '6': '.ibc.core.channel.v1.State', '10': 'state'},
    const {'1': 'ordering', '3': 2, '4': 1, '5': 14, '6': '.ibc.core.channel.v1.Order', '10': 'ordering'},
    const {'1': 'counterparty', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.channel.v1.Counterparty', '8': const {}, '10': 'counterparty'},
    const {'1': 'connection_hops', '3': 4, '4': 3, '5': 9, '8': const {}, '10': 'connectionHops'},
    const {'1': 'version', '3': 5, '4': 1, '5': 9, '10': 'version'},
    const {'1': 'port_id', '3': 6, '4': 1, '5': 9, '10': 'portId'},
    const {'1': 'channel_id', '3': 7, '4': 1, '5': 9, '10': 'channelId'},
  ],
  '7': const {},
};

const Counterparty$json = const {
  '1': 'Counterparty',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'channelId'},
  ],
  '7': const {},
};

const Packet$json = const {
  '1': 'Packet',
  '2': const [
    const {'1': 'sequence', '3': 1, '4': 1, '5': 4, '10': 'sequence'},
    const {'1': 'source_port', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'sourcePort'},
    const {'1': 'source_channel', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'sourceChannel'},
    const {'1': 'destination_port', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'destinationPort'},
    const {'1': 'destination_channel', '3': 5, '4': 1, '5': 9, '8': const {}, '10': 'destinationChannel'},
    const {'1': 'data', '3': 6, '4': 1, '5': 12, '10': 'data'},
    const {'1': 'timeout_height', '3': 7, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'timeoutHeight'},
    const {'1': 'timeout_timestamp', '3': 8, '4': 1, '5': 4, '8': const {}, '10': 'timeoutTimestamp'},
  ],
  '7': const {},
};

const PacketState$json = const {
  '1': 'PacketState',
  '2': const [
    const {'1': 'port_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'portId'},
    const {'1': 'channel_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'channelId'},
    const {'1': 'sequence', '3': 3, '4': 1, '5': 4, '10': 'sequence'},
    const {'1': 'data', '3': 4, '4': 1, '5': 12, '10': 'data'},
  ],
  '7': const {},
};

const Acknowledgement$json = const {
  '1': 'Acknowledgement',
  '2': const [
    const {'1': 'result', '3': 21, '4': 1, '5': 12, '9': 0, '10': 'result'},
    const {'1': 'error', '3': 22, '4': 1, '5': 9, '9': 0, '10': 'error'},
  ],
  '8': const [
    const {'1': 'response'},
  ],
};

