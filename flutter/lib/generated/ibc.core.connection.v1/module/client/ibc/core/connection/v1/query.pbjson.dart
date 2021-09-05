///
//  Generated code. Do not modify.
//  source: ibc/core/connection/v1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryConnectionRequest$json = const {
  '1': 'QueryConnectionRequest',
  '2': const [
    const {'1': 'connection_id', '3': 1, '4': 1, '5': 9, '10': 'connectionId'},
  ],
};

const QueryConnectionResponse$json = const {
  '1': 'QueryConnectionResponse',
  '2': const [
    const {'1': 'connection', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.connection.v1.ConnectionEnd', '10': 'connection'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryConnectionsRequest$json = const {
  '1': 'QueryConnectionsRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryConnectionsResponse$json = const {
  '1': 'QueryConnectionsResponse',
  '2': const [
    const {'1': 'connections', '3': 1, '4': 3, '5': 11, '6': '.ibc.core.connection.v1.IdentifiedConnection', '10': 'connections'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
    const {'1': 'height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'height'},
  ],
};

const QueryClientConnectionsRequest$json = const {
  '1': 'QueryClientConnectionsRequest',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

const QueryClientConnectionsResponse$json = const {
  '1': 'QueryClientConnectionsResponse',
  '2': const [
    const {'1': 'connection_paths', '3': 1, '4': 3, '5': 9, '10': 'connectionPaths'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryConnectionClientStateRequest$json = const {
  '1': 'QueryConnectionClientStateRequest',
  '2': const [
    const {'1': 'connection_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'connectionId'},
  ],
};

const QueryConnectionClientStateResponse$json = const {
  '1': 'QueryConnectionClientStateResponse',
  '2': const [
    const {'1': 'identified_client_state', '3': 1, '4': 1, '5': 11, '6': '.ibc.core.client.v1.IdentifiedClientState', '10': 'identifiedClientState'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

const QueryConnectionConsensusStateRequest$json = const {
  '1': 'QueryConnectionConsensusStateRequest',
  '2': const [
    const {'1': 'connection_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'connectionId'},
    const {'1': 'revision_number', '3': 2, '4': 1, '5': 4, '10': 'revisionNumber'},
    const {'1': 'revision_height', '3': 3, '4': 1, '5': 4, '10': 'revisionHeight'},
  ],
};

const QueryConnectionConsensusStateResponse$json = const {
  '1': 'QueryConnectionConsensusStateResponse',
  '2': const [
    const {'1': 'consensus_state', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Any', '10': 'consensusState'},
    const {'1': 'client_id', '3': 2, '4': 1, '5': 9, '10': 'clientId'},
    const {'1': 'proof', '3': 3, '4': 1, '5': 12, '10': 'proof'},
    const {'1': 'proof_height', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
  ],
};

