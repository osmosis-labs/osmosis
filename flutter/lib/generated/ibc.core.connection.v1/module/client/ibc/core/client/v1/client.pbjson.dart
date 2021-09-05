///
//  Generated code. Do not modify.
//  source: ibc/core/client/v1/client.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const IdentifiedClientState$json = const {
  '1': 'IdentifiedClientState',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'client_state', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'clientState'},
  ],
};

const ConsensusStateWithHeight$json = const {
  '1': 'ConsensusStateWithHeight',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
    const {'1': 'consensus_state', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'consensusState'},
  ],
};

const ClientConsensusStates$json = const {
  '1': 'ClientConsensusStates',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'consensus_states', '3': 2, '4': 3, '5': 11, '6': '.ibc.core.client.v1.ConsensusStateWithHeight', '8': const {}, '10': 'consensusStates'},
  ],
};

const ClientUpdateProposal$json = const {
  '1': 'ClientUpdateProposal',
  '2': const [
    const {'1': 'title', '3': 1, '4': 1, '5': 9, '10': 'title'},
    const {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    const {'1': 'client_id', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'header', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Any', '10': 'header'},
  ],
  '7': const {},
};

const Height$json = const {
  '1': 'Height',
  '2': const [
    const {'1': 'revision_number', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'revisionNumber'},
    const {'1': 'revision_height', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'revisionHeight'},
  ],
  '7': const {},
};

const Params$json = const {
  '1': 'Params',
  '2': const [
    const {'1': 'allowed_clients', '3': 1, '4': 3, '5': 9, '8': const {}, '10': 'allowedClients'},
  ],
};

