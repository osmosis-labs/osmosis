///
//  Generated code. Do not modify.
//  source: ibc/core/client/v1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const GenesisState$json = const {
  '1': 'GenesisState',
  '2': const [
    const {'1': 'clients', '3': 1, '4': 3, '5': 11, '6': '.ibc.core.client.v1.IdentifiedClientState', '8': const {}, '10': 'clients'},
    const {'1': 'clients_consensus', '3': 2, '4': 3, '5': 11, '6': '.ibc.core.client.v1.ClientConsensusStates', '8': const {}, '10': 'clientsConsensus'},
    const {'1': 'clients_metadata', '3': 3, '4': 3, '5': 11, '6': '.ibc.core.client.v1.IdentifiedGenesisMetadata', '8': const {}, '10': 'clientsMetadata'},
    const {'1': 'params', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Params', '8': const {}, '10': 'params'},
    const {'1': 'create_localhost', '3': 5, '4': 1, '5': 8, '8': const {}, '10': 'createLocalhost'},
    const {'1': 'next_client_sequence', '3': 6, '4': 1, '5': 4, '8': const {}, '10': 'nextClientSequence'},
  ],
};

const GenesisMetadata$json = const {
  '1': 'GenesisMetadata',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'value', '3': 2, '4': 1, '5': 12, '10': 'value'},
  ],
  '7': const {},
};

const IdentifiedGenesisMetadata$json = const {
  '1': 'IdentifiedGenesisMetadata',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'client_metadata', '3': 2, '4': 3, '5': 11, '6': '.ibc.core.client.v1.GenesisMetadata', '8': const {}, '10': 'clientMetadata'},
  ],
};

