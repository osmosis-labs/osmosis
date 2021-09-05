///
//  Generated code. Do not modify.
//  source: ibc/core/connection/v1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgConnectionOpenInit$json = const {
  '1': 'MsgConnectionOpenInit',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'counterparty', '3': 2, '4': 1, '5': 11, '6': '.ibc.core.connection.v1.Counterparty', '8': const {}, '10': 'counterparty'},
    const {'1': 'version', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.connection.v1.Version', '10': 'version'},
    const {'1': 'delay_period', '3': 4, '4': 1, '5': 4, '8': const {}, '10': 'delayPeriod'},
    const {'1': 'signer', '3': 5, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgConnectionOpenInitResponse$json = const {
  '1': 'MsgConnectionOpenInitResponse',
};

const MsgConnectionOpenTry$json = const {
  '1': 'MsgConnectionOpenTry',
  '2': const [
    const {'1': 'client_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'clientId'},
    const {'1': 'previous_connection_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'previousConnectionId'},
    const {'1': 'client_state', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'clientState'},
    const {'1': 'counterparty', '3': 4, '4': 1, '5': 11, '6': '.ibc.core.connection.v1.Counterparty', '8': const {}, '10': 'counterparty'},
    const {'1': 'delay_period', '3': 5, '4': 1, '5': 4, '8': const {}, '10': 'delayPeriod'},
    const {'1': 'counterparty_versions', '3': 6, '4': 3, '5': 11, '6': '.ibc.core.connection.v1.Version', '8': const {}, '10': 'counterpartyVersions'},
    const {'1': 'proof_height', '3': 7, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'proof_init', '3': 8, '4': 1, '5': 12, '8': const {}, '10': 'proofInit'},
    const {'1': 'proof_client', '3': 9, '4': 1, '5': 12, '8': const {}, '10': 'proofClient'},
    const {'1': 'proof_consensus', '3': 10, '4': 1, '5': 12, '8': const {}, '10': 'proofConsensus'},
    const {'1': 'consensus_height', '3': 11, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'consensusHeight'},
    const {'1': 'signer', '3': 12, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgConnectionOpenTryResponse$json = const {
  '1': 'MsgConnectionOpenTryResponse',
};

const MsgConnectionOpenAck$json = const {
  '1': 'MsgConnectionOpenAck',
  '2': const [
    const {'1': 'connection_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'connectionId'},
    const {'1': 'counterparty_connection_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'counterpartyConnectionId'},
    const {'1': 'version', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.connection.v1.Version', '10': 'version'},
    const {'1': 'client_state', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'clientState'},
    const {'1': 'proof_height', '3': 5, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'proof_try', '3': 6, '4': 1, '5': 12, '8': const {}, '10': 'proofTry'},
    const {'1': 'proof_client', '3': 7, '4': 1, '5': 12, '8': const {}, '10': 'proofClient'},
    const {'1': 'proof_consensus', '3': 8, '4': 1, '5': 12, '8': const {}, '10': 'proofConsensus'},
    const {'1': 'consensus_height', '3': 9, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'consensusHeight'},
    const {'1': 'signer', '3': 10, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgConnectionOpenAckResponse$json = const {
  '1': 'MsgConnectionOpenAckResponse',
};

const MsgConnectionOpenConfirm$json = const {
  '1': 'MsgConnectionOpenConfirm',
  '2': const [
    const {'1': 'connection_id', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'connectionId'},
    const {'1': 'proof_ack', '3': 2, '4': 1, '5': 12, '8': const {}, '10': 'proofAck'},
    const {'1': 'proof_height', '3': 3, '4': 1, '5': 11, '6': '.ibc.core.client.v1.Height', '8': const {}, '10': 'proofHeight'},
    const {'1': 'signer', '3': 4, '4': 1, '5': 9, '10': 'signer'},
  ],
  '7': const {},
};

const MsgConnectionOpenConfirmResponse$json = const {
  '1': 'MsgConnectionOpenConfirmResponse',
};

