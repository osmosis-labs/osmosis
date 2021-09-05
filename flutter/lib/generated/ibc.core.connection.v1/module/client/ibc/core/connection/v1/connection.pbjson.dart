///
//  Generated code. Do not modify.
//  source: ibc/core/connection/v1/connection.proto
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
  ],
  '3': const {},
};

const ConnectionEnd$json = const {
  '1': 'ConnectionEnd',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'versions', '3': 2, '4': 3, '5': 11, '6': '.ibc.core.connection.v1.Version', '10': 'versions'},
    const {'1': 'state', '3': 3, '4': 1, '5': 14, '6': '.ibc.core.connection.v1.State', '10': 'state'},
    const {'1': 'counterparty', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.connection.v1.Counterparty', '8': const {}, '10': 'counterparty'},
    const {'1': 'delay_period', '3': 5, '4': 1, '5': 4, '8': const {}, '10': 'delayPeriod'},
  ],
  '7': const {},
};

const IdentifiedConnection$json = const {
  '1': 'IdentifiedConnection',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'id'},
    const {'1': 'client_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'versions', '3': 3, '4': 3, '5': 11, '6': '.ibc.core.connection.v1.Version', '10': 'versions'},
    const {'1': 'state', '3': 4, '4': 1, '5': 14, '6': '.ibc.core.connection.v1.State', '10': 'state'},
    const {'1': 'counterparty', '3': 5, '4': 1, '5': 11, '6': '.ibc.core.connection.v1.Counterparty', '8': const {}, '10': 'counterparty'},
    const {'1': 'delay_period', '3': 6, '4': 1, '5': 4, '8': const {}, '10': 'delayPeriod'},
  ],
  '7': const {},
};

const Counterparty$json = const {
  '1': 'Counterparty',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'connection_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'connectionId'},
    const {'1': 'prefix', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.commitment.v1.MerklePrefix', '8': const {}, '10': 'prefix'},
  ],
  '7': const {},
};

const ClientPaths$json = const {
  '1': 'ClientPaths',
  '2': const [
    const {'1': 'paths', '3': 1, '4': 3, '5': 9, '10': 'paths'},
  ],
};

const ConnectionPaths$json = const {
  '1': 'ConnectionPaths',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'paths', '3': 2, '4': 3, '5': 9, '10': 'paths'},
  ],
};

const Version$json = const {
  '1': 'Version',
  '2': const [
    const {'1': 'identifier', '3': 1, '4': 1, '5': 9, '10': 'identifier'},
    const {'1': 'features', '3': 2, '4': 3, '5': 9, '10': 'features'},
  ],
  '7': const {},
};

