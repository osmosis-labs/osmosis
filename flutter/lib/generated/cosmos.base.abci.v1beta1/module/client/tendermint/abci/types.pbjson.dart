///
//  Generated code. Do not modify.
//  source: tendermint/abci/types.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const CheckTxType$json = const {
  '1': 'CheckTxType',
  '2': const [
    const {'1': 'NEW', '2': 0, '3': const {}},
    const {'1': 'RECHECK', '2': 1, '3': const {}},
  ],
};

const EvidenceType$json = const {
  '1': 'EvidenceType',
  '2': const [
    const {'1': 'UNKNOWN', '2': 0},
    const {'1': 'DUPLICATE_VOTE', '2': 1},
    const {'1': 'LIGHT_CLIENT_ATTACK', '2': 2},
  ],
};

const Request$json = const {
  '1': 'Request',
  '2': const [
    const {'1': 'echo', '3': 1, '4': 1, '5': 11, '6': '.tendermint.abci.RequestEcho', '9': 0, '10': 'echo'},
    const {'1': 'flush', '3': 2, '4': 1, '5': 11, '6': '.tendermint.abci.RequestFlush', '9': 0, '10': 'flush'},
    const {'1': 'info', '3': 3, '4': 1, '5': 11, '6': '.tendermint.abci.RequestInfo', '9': 0, '10': 'info'},
    const {'1': 'set_option', '3': 4, '4': 1, '5': 11, '6': '.tendermint.abci.RequestSetOption', '9': 0, '10': 'setOption'},
    const {'1': 'init_chain', '3': 5, '4': 1, '5': 11, '6': '.tendermint.abci.RequestInitChain', '9': 0, '10': 'initChain'},
    const {'1': 'query', '3': 6, '4': 1, '5': 11, '6': '.tendermint.abci.RequestQuery', '9': 0, '10': 'query'},
    const {'1': 'begin_block', '3': 7, '4': 1, '5': 11, '6': '.tendermint.abci.RequestBeginBlock', '9': 0, '10': 'beginBlock'},
    const {'1': 'check_tx', '3': 8, '4': 1, '5': 11, '6': '.tendermint.abci.RequestCheckTx', '9': 0, '10': 'checkTx'},
    const {'1': 'deliver_tx', '3': 9, '4': 1, '5': 11, '6': '.tendermint.abci.RequestDeliverTx', '9': 0, '10': 'deliverTx'},
    const {'1': 'end_block', '3': 10, '4': 1, '5': 11, '6': '.tendermint.abci.RequestEndBlock', '9': 0, '10': 'endBlock'},
    const {'1': 'commit', '3': 11, '4': 1, '5': 11, '6': '.tendermint.abci.RequestCommit', '9': 0, '10': 'commit'},
    const {'1': 'list_snapshots', '3': 12, '4': 1, '5': 11, '6': '.tendermint.abci.RequestListSnapshots', '9': 0, '10': 'listSnapshots'},
    const {'1': 'offer_snapshot', '3': 13, '4': 1, '5': 11, '6': '.tendermint.abci.RequestOfferSnapshot', '9': 0, '10': 'offerSnapshot'},
    const {'1': 'load_snapshot_chunk', '3': 14, '4': 1, '5': 11, '6': '.tendermint.abci.RequestLoadSnapshotChunk', '9': 0, '10': 'loadSnapshotChunk'},
    const {'1': 'apply_snapshot_chunk', '3': 15, '4': 1, '5': 11, '6': '.tendermint.abci.RequestApplySnapshotChunk', '9': 0, '10': 'applySnapshotChunk'},
  ],
  '8': const [
    const {'1': 'value'},
  ],
};

const RequestEcho$json = const {
  '1': 'RequestEcho',
  '2': const [
    const {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

const RequestFlush$json = const {
  '1': 'RequestFlush',
};

const RequestInfo$json = const {
  '1': 'RequestInfo',
  '2': const [
    const {'1': 'version', '3': 1, '4': 1, '5': 9, '10': 'version'},
    const {'1': 'block_version', '3': 2, '4': 1, '5': 4, '10': 'blockVersion'},
    const {'1': 'p2p_version', '3': 3, '4': 1, '5': 4, '10': 'p2pVersion'},
  ],
};

const RequestSetOption$json = const {
  '1': 'RequestSetOption',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    const {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
};

const RequestInitChain$json = const {
  '1': 'RequestInitChain',
  '2': const [
    const {'1': 'time', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'time'},
    const {'1': 'chain_id', '3': 2, '4': 1, '5': 9, '10': 'chainId'},
    const {'1': 'consensus_params', '3': 3, '4': 1, '5': 11, '6': '.tendermint.abci.ConsensusParams', '10': 'consensusParams'},
    const {'1': 'validators', '3': 4, '4': 3, '5': 11, '6': '.tendermint.abci.ValidatorUpdate', '8': const {}, '10': 'validators'},
    const {'1': 'app_state_bytes', '3': 5, '4': 1, '5': 12, '10': 'appStateBytes'},
    const {'1': 'initial_height', '3': 6, '4': 1, '5': 3, '10': 'initialHeight'},
  ],
};

const RequestQuery$json = const {
  '1': 'RequestQuery',
  '2': const [
    const {'1': 'data', '3': 1, '4': 1, '5': 12, '10': 'data'},
    const {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
    const {'1': 'height', '3': 3, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'prove', '3': 4, '4': 1, '5': 8, '10': 'prove'},
  ],
};

const RequestBeginBlock$json = const {
  '1': 'RequestBeginBlock',
  '2': const [
    const {'1': 'hash', '3': 1, '4': 1, '5': 12, '10': 'hash'},
    const {'1': 'header', '3': 2, '4': 1, '5': 11, '6': '.tendermint.types.Header', '8': const {}, '10': 'header'},
    const {'1': 'last_commit_info', '3': 3, '4': 1, '5': 11, '6': '.tendermint.abci.LastCommitInfo', '8': const {}, '10': 'lastCommitInfo'},
    const {'1': 'byzantine_validators', '3': 4, '4': 3, '5': 11, '6': '.tendermint.abci.Evidence', '8': const {}, '10': 'byzantineValidators'},
  ],
};

const RequestCheckTx$json = const {
  '1': 'RequestCheckTx',
  '2': const [
    const {'1': 'tx', '3': 1, '4': 1, '5': 12, '10': 'tx'},
    const {'1': 'type', '3': 2, '4': 1, '5': 14, '6': '.tendermint.abci.CheckTxType', '10': 'type'},
  ],
};

const RequestDeliverTx$json = const {
  '1': 'RequestDeliverTx',
  '2': const [
    const {'1': 'tx', '3': 1, '4': 1, '5': 12, '10': 'tx'},
  ],
};

const RequestEndBlock$json = const {
  '1': 'RequestEndBlock',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 3, '10': 'height'},
  ],
};

const RequestCommit$json = const {
  '1': 'RequestCommit',
};

const RequestListSnapshots$json = const {
  '1': 'RequestListSnapshots',
};

const RequestOfferSnapshot$json = const {
  '1': 'RequestOfferSnapshot',
  '2': const [
    const {'1': 'snapshot', '3': 1, '4': 1, '5': 11, '6': '.tendermint.abci.Snapshot', '10': 'snapshot'},
    const {'1': 'app_hash', '3': 2, '4': 1, '5': 12, '10': 'appHash'},
  ],
};

const RequestLoadSnapshotChunk$json = const {
  '1': 'RequestLoadSnapshotChunk',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 4, '10': 'height'},
    const {'1': 'format', '3': 2, '4': 1, '5': 13, '10': 'format'},
    const {'1': 'chunk', '3': 3, '4': 1, '5': 13, '10': 'chunk'},
  ],
};

const RequestApplySnapshotChunk$json = const {
  '1': 'RequestApplySnapshotChunk',
  '2': const [
    const {'1': 'index', '3': 1, '4': 1, '5': 13, '10': 'index'},
    const {'1': 'chunk', '3': 2, '4': 1, '5': 12, '10': 'chunk'},
    const {'1': 'sender', '3': 3, '4': 1, '5': 9, '10': 'sender'},
  ],
};

const Response$json = const {
  '1': 'Response',
  '2': const [
    const {'1': 'exception', '3': 1, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseException', '9': 0, '10': 'exception'},
    const {'1': 'echo', '3': 2, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseEcho', '9': 0, '10': 'echo'},
    const {'1': 'flush', '3': 3, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseFlush', '9': 0, '10': 'flush'},
    const {'1': 'info', '3': 4, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseInfo', '9': 0, '10': 'info'},
    const {'1': 'set_option', '3': 5, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseSetOption', '9': 0, '10': 'setOption'},
    const {'1': 'init_chain', '3': 6, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseInitChain', '9': 0, '10': 'initChain'},
    const {'1': 'query', '3': 7, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseQuery', '9': 0, '10': 'query'},
    const {'1': 'begin_block', '3': 8, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseBeginBlock', '9': 0, '10': 'beginBlock'},
    const {'1': 'check_tx', '3': 9, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseCheckTx', '9': 0, '10': 'checkTx'},
    const {'1': 'deliver_tx', '3': 10, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseDeliverTx', '9': 0, '10': 'deliverTx'},
    const {'1': 'end_block', '3': 11, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseEndBlock', '9': 0, '10': 'endBlock'},
    const {'1': 'commit', '3': 12, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseCommit', '9': 0, '10': 'commit'},
    const {'1': 'list_snapshots', '3': 13, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseListSnapshots', '9': 0, '10': 'listSnapshots'},
    const {'1': 'offer_snapshot', '3': 14, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseOfferSnapshot', '9': 0, '10': 'offerSnapshot'},
    const {'1': 'load_snapshot_chunk', '3': 15, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseLoadSnapshotChunk', '9': 0, '10': 'loadSnapshotChunk'},
    const {'1': 'apply_snapshot_chunk', '3': 16, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseApplySnapshotChunk', '9': 0, '10': 'applySnapshotChunk'},
  ],
  '8': const [
    const {'1': 'value'},
  ],
};

const ResponseException$json = const {
  '1': 'ResponseException',
  '2': const [
    const {'1': 'error', '3': 1, '4': 1, '5': 9, '10': 'error'},
  ],
};

const ResponseEcho$json = const {
  '1': 'ResponseEcho',
  '2': const [
    const {'1': 'message', '3': 1, '4': 1, '5': 9, '10': 'message'},
  ],
};

const ResponseFlush$json = const {
  '1': 'ResponseFlush',
};

const ResponseInfo$json = const {
  '1': 'ResponseInfo',
  '2': const [
    const {'1': 'data', '3': 1, '4': 1, '5': 9, '10': 'data'},
    const {'1': 'version', '3': 2, '4': 1, '5': 9, '10': 'version'},
    const {'1': 'app_version', '3': 3, '4': 1, '5': 4, '10': 'appVersion'},
    const {'1': 'last_block_height', '3': 4, '4': 1, '5': 3, '10': 'lastBlockHeight'},
    const {'1': 'last_block_app_hash', '3': 5, '4': 1, '5': 12, '10': 'lastBlockAppHash'},
  ],
};

const ResponseSetOption$json = const {
  '1': 'ResponseSetOption',
  '2': const [
    const {'1': 'code', '3': 1, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'log', '3': 3, '4': 1, '5': 9, '10': 'log'},
    const {'1': 'info', '3': 4, '4': 1, '5': 9, '10': 'info'},
  ],
};

const ResponseInitChain$json = const {
  '1': 'ResponseInitChain',
  '2': const [
    const {'1': 'consensus_params', '3': 1, '4': 1, '5': 11, '6': '.tendermint.abci.ConsensusParams', '10': 'consensusParams'},
    const {'1': 'validators', '3': 2, '4': 3, '5': 11, '6': '.tendermint.abci.ValidatorUpdate', '8': const {}, '10': 'validators'},
    const {'1': 'app_hash', '3': 3, '4': 1, '5': 12, '10': 'appHash'},
  ],
};

const ResponseQuery$json = const {
  '1': 'ResponseQuery',
  '2': const [
    const {'1': 'code', '3': 1, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'log', '3': 3, '4': 1, '5': 9, '10': 'log'},
    const {'1': 'info', '3': 4, '4': 1, '5': 9, '10': 'info'},
    const {'1': 'index', '3': 5, '4': 1, '5': 3, '10': 'index'},
    const {'1': 'key', '3': 6, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'value', '3': 7, '4': 1, '5': 12, '10': 'value'},
    const {'1': 'proof_ops', '3': 8, '4': 1, '5': 11, '6': '.tendermint.crypto.ProofOps', '10': 'proofOps'},
    const {'1': 'height', '3': 9, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'codespace', '3': 10, '4': 1, '5': 9, '10': 'codespace'},
  ],
};

const ResponseBeginBlock$json = const {
  '1': 'ResponseBeginBlock',
  '2': const [
    const {'1': 'events', '3': 1, '4': 3, '5': 11, '6': '.tendermint.abci.Event', '8': const {}, '10': 'events'},
  ],
};

const ResponseCheckTx$json = const {
  '1': 'ResponseCheckTx',
  '2': const [
    const {'1': 'code', '3': 1, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'data', '3': 2, '4': 1, '5': 12, '10': 'data'},
    const {'1': 'log', '3': 3, '4': 1, '5': 9, '10': 'log'},
    const {'1': 'info', '3': 4, '4': 1, '5': 9, '10': 'info'},
    const {'1': 'gas_wanted', '3': 5, '4': 1, '5': 3, '10': 'gas_wanted'},
    const {'1': 'gas_used', '3': 6, '4': 1, '5': 3, '10': 'gas_used'},
    const {'1': 'events', '3': 7, '4': 3, '5': 11, '6': '.tendermint.abci.Event', '8': const {}, '10': 'events'},
    const {'1': 'codespace', '3': 8, '4': 1, '5': 9, '10': 'codespace'},
  ],
};

const ResponseDeliverTx$json = const {
  '1': 'ResponseDeliverTx',
  '2': const [
    const {'1': 'code', '3': 1, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'data', '3': 2, '4': 1, '5': 12, '10': 'data'},
    const {'1': 'log', '3': 3, '4': 1, '5': 9, '10': 'log'},
    const {'1': 'info', '3': 4, '4': 1, '5': 9, '10': 'info'},
    const {'1': 'gas_wanted', '3': 5, '4': 1, '5': 3, '10': 'gas_wanted'},
    const {'1': 'gas_used', '3': 6, '4': 1, '5': 3, '10': 'gas_used'},
    const {'1': 'events', '3': 7, '4': 3, '5': 11, '6': '.tendermint.abci.Event', '8': const {}, '10': 'events'},
    const {'1': 'codespace', '3': 8, '4': 1, '5': 9, '10': 'codespace'},
  ],
};

const ResponseEndBlock$json = const {
  '1': 'ResponseEndBlock',
  '2': const [
    const {'1': 'validator_updates', '3': 1, '4': 3, '5': 11, '6': '.tendermint.abci.ValidatorUpdate', '8': const {}, '10': 'validatorUpdates'},
    const {'1': 'consensus_param_updates', '3': 2, '4': 1, '5': 11, '6': '.tendermint.abci.ConsensusParams', '10': 'consensusParamUpdates'},
    const {'1': 'events', '3': 3, '4': 3, '5': 11, '6': '.tendermint.abci.Event', '8': const {}, '10': 'events'},
  ],
};

const ResponseCommit$json = const {
  '1': 'ResponseCommit',
  '2': const [
    const {'1': 'data', '3': 2, '4': 1, '5': 12, '10': 'data'},
    const {'1': 'retain_height', '3': 3, '4': 1, '5': 3, '10': 'retainHeight'},
  ],
};

const ResponseListSnapshots$json = const {
  '1': 'ResponseListSnapshots',
  '2': const [
    const {'1': 'snapshots', '3': 1, '4': 3, '5': 11, '6': '.tendermint.abci.Snapshot', '10': 'snapshots'},
  ],
};

const ResponseOfferSnapshot$json = const {
  '1': 'ResponseOfferSnapshot',
  '2': const [
    const {'1': 'result', '3': 1, '4': 1, '5': 14, '6': '.tendermint.abci.ResponseOfferSnapshot.Result', '10': 'result'},
  ],
  '4': const [ResponseOfferSnapshot_Result$json],
};

const ResponseOfferSnapshot_Result$json = const {
  '1': 'Result',
  '2': const [
    const {'1': 'UNKNOWN', '2': 0},
    const {'1': 'ACCEPT', '2': 1},
    const {'1': 'ABORT', '2': 2},
    const {'1': 'REJECT', '2': 3},
    const {'1': 'REJECT_FORMAT', '2': 4},
    const {'1': 'REJECT_SENDER', '2': 5},
  ],
};

const ResponseLoadSnapshotChunk$json = const {
  '1': 'ResponseLoadSnapshotChunk',
  '2': const [
    const {'1': 'chunk', '3': 1, '4': 1, '5': 12, '10': 'chunk'},
  ],
};

const ResponseApplySnapshotChunk$json = const {
  '1': 'ResponseApplySnapshotChunk',
  '2': const [
    const {'1': 'result', '3': 1, '4': 1, '5': 14, '6': '.tendermint.abci.ResponseApplySnapshotChunk.Result', '10': 'result'},
    const {'1': 'refetch_chunks', '3': 2, '4': 3, '5': 13, '10': 'refetchChunks'},
    const {'1': 'reject_senders', '3': 3, '4': 3, '5': 9, '10': 'rejectSenders'},
  ],
  '4': const [ResponseApplySnapshotChunk_Result$json],
};

const ResponseApplySnapshotChunk_Result$json = const {
  '1': 'Result',
  '2': const [
    const {'1': 'UNKNOWN', '2': 0},
    const {'1': 'ACCEPT', '2': 1},
    const {'1': 'ABORT', '2': 2},
    const {'1': 'RETRY', '2': 3},
    const {'1': 'RETRY_SNAPSHOT', '2': 4},
    const {'1': 'REJECT_SNAPSHOT', '2': 5},
  ],
};

const ConsensusParams$json = const {
  '1': 'ConsensusParams',
  '2': const [
    const {'1': 'block', '3': 1, '4': 1, '5': 11, '6': '.tendermint.abci.BlockParams', '10': 'block'},
    const {'1': 'evidence', '3': 2, '4': 1, '5': 11, '6': '.tendermint.types.EvidenceParams', '10': 'evidence'},
    const {'1': 'validator', '3': 3, '4': 1, '5': 11, '6': '.tendermint.types.ValidatorParams', '10': 'validator'},
    const {'1': 'version', '3': 4, '4': 1, '5': 11, '6': '.tendermint.types.VersionParams', '10': 'version'},
  ],
};

const BlockParams$json = const {
  '1': 'BlockParams',
  '2': const [
    const {'1': 'max_bytes', '3': 1, '4': 1, '5': 3, '10': 'maxBytes'},
    const {'1': 'max_gas', '3': 2, '4': 1, '5': 3, '10': 'maxGas'},
  ],
};

const LastCommitInfo$json = const {
  '1': 'LastCommitInfo',
  '2': const [
    const {'1': 'round', '3': 1, '4': 1, '5': 5, '10': 'round'},
    const {'1': 'votes', '3': 2, '4': 3, '5': 11, '6': '.tendermint.abci.VoteInfo', '8': const {}, '10': 'votes'},
  ],
};

const Event$json = const {
  '1': 'Event',
  '2': const [
    const {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    const {'1': 'attributes', '3': 2, '4': 3, '5': 11, '6': '.tendermint.abci.EventAttribute', '8': const {}, '10': 'attributes'},
  ],
};

const EventAttribute$json = const {
  '1': 'EventAttribute',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'value', '3': 2, '4': 1, '5': 12, '10': 'value'},
    const {'1': 'index', '3': 3, '4': 1, '5': 8, '10': 'index'},
  ],
};

const TxResult$json = const {
  '1': 'TxResult',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'index', '3': 2, '4': 1, '5': 13, '10': 'index'},
    const {'1': 'tx', '3': 3, '4': 1, '5': 12, '10': 'tx'},
    const {'1': 'result', '3': 4, '4': 1, '5': 11, '6': '.tendermint.abci.ResponseDeliverTx', '8': const {}, '10': 'result'},
  ],
};

const Validator$json = const {
  '1': 'Validator',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 12, '10': 'address'},
    const {'1': 'power', '3': 3, '4': 1, '5': 3, '10': 'power'},
  ],
};

const ValidatorUpdate$json = const {
  '1': 'ValidatorUpdate',
  '2': const [
    const {'1': 'pub_key', '3': 1, '4': 1, '5': 11, '6': '.tendermint.crypto.PublicKey', '8': const {}, '10': 'pubKey'},
    const {'1': 'power', '3': 2, '4': 1, '5': 3, '10': 'power'},
  ],
};

const VoteInfo$json = const {
  '1': 'VoteInfo',
  '2': const [
    const {'1': 'validator', '3': 1, '4': 1, '5': 11, '6': '.tendermint.abci.Validator', '8': const {}, '10': 'validator'},
    const {'1': 'signed_last_block', '3': 2, '4': 1, '5': 8, '10': 'signedLastBlock'},
  ],
};

const Evidence$json = const {
  '1': 'Evidence',
  '2': const [
    const {'1': 'type', '3': 1, '4': 1, '5': 14, '6': '.tendermint.abci.EvidenceType', '10': 'type'},
    const {'1': 'validator', '3': 2, '4': 1, '5': 11, '6': '.tendermint.abci.Validator', '8': const {}, '10': 'validator'},
    const {'1': 'height', '3': 3, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'time', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'time'},
    const {'1': 'total_voting_power', '3': 5, '4': 1, '5': 3, '10': 'totalVotingPower'},
  ],
};

const Snapshot$json = const {
  '1': 'Snapshot',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 4, '10': 'height'},
    const {'1': 'format', '3': 2, '4': 1, '5': 13, '10': 'format'},
    const {'1': 'chunks', '3': 3, '4': 1, '5': 13, '10': 'chunks'},
    const {'1': 'hash', '3': 4, '4': 1, '5': 12, '10': 'hash'},
    const {'1': 'metadata', '3': 5, '4': 1, '5': 12, '10': 'metadata'},
  ],
};

