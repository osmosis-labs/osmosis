///
//  Generated code. Do not modify.
//  source: tendermint/types/types.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const BlockIDFlag$json = const {
  '1': 'BlockIDFlag',
  '2': const [
    const {'1': 'BLOCK_ID_FLAG_UNKNOWN', '2': 0, '3': const {}},
    const {'1': 'BLOCK_ID_FLAG_ABSENT', '2': 1, '3': const {}},
    const {'1': 'BLOCK_ID_FLAG_COMMIT', '2': 2, '3': const {}},
    const {'1': 'BLOCK_ID_FLAG_NIL', '2': 3, '3': const {}},
  ],
  '3': const {},
};

const SignedMsgType$json = const {
  '1': 'SignedMsgType',
  '2': const [
    const {'1': 'SIGNED_MSG_TYPE_UNKNOWN', '2': 0, '3': const {}},
    const {'1': 'SIGNED_MSG_TYPE_PREVOTE', '2': 1, '3': const {}},
    const {'1': 'SIGNED_MSG_TYPE_PRECOMMIT', '2': 2, '3': const {}},
    const {'1': 'SIGNED_MSG_TYPE_PROPOSAL', '2': 32, '3': const {}},
  ],
  '3': const {},
};

const PartSetHeader$json = const {
  '1': 'PartSetHeader',
  '2': const [
    const {'1': 'total', '3': 1, '4': 1, '5': 13, '10': 'total'},
    const {'1': 'hash', '3': 2, '4': 1, '5': 12, '10': 'hash'},
  ],
};

const Part$json = const {
  '1': 'Part',
  '2': const [
    const {'1': 'index', '3': 1, '4': 1, '5': 13, '10': 'index'},
    const {'1': 'bytes', '3': 2, '4': 1, '5': 12, '10': 'bytes'},
    const {'1': 'proof', '3': 3, '4': 1, '5': 11, '6': '.tendermint.crypto.Proof', '8': const {}, '10': 'proof'},
  ],
};

const BlockID$json = const {
  '1': 'BlockID',
  '2': const [
    const {'1': 'hash', '3': 1, '4': 1, '5': 12, '10': 'hash'},
    const {'1': 'part_set_header', '3': 2, '4': 1, '5': 11, '6': '.tendermint.types.PartSetHeader', '8': const {}, '10': 'partSetHeader'},
  ],
};

const Header$json = const {
  '1': 'Header',
  '2': const [
    const {'1': 'version', '3': 1, '4': 1, '5': 11, '6': '.tendermint.version.Consensus', '8': const {}, '10': 'version'},
    const {'1': 'chain_id', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'chainId'},
    const {'1': 'height', '3': 3, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'time', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'time'},
    const {'1': 'last_block_id', '3': 5, '4': 1, '5': 11, '6': '.tendermint.types.BlockID', '8': const {}, '10': 'lastBlockId'},
    const {'1': 'last_commit_hash', '3': 6, '4': 1, '5': 12, '10': 'lastCommitHash'},
    const {'1': 'data_hash', '3': 7, '4': 1, '5': 12, '10': 'dataHash'},
    const {'1': 'validators_hash', '3': 8, '4': 1, '5': 12, '10': 'validatorsHash'},
    const {'1': 'next_validators_hash', '3': 9, '4': 1, '5': 12, '10': 'nextValidatorsHash'},
    const {'1': 'consensus_hash', '3': 10, '4': 1, '5': 12, '10': 'consensusHash'},
    const {'1': 'app_hash', '3': 11, '4': 1, '5': 12, '10': 'appHash'},
    const {'1': 'last_results_hash', '3': 12, '4': 1, '5': 12, '10': 'lastResultsHash'},
    const {'1': 'evidence_hash', '3': 13, '4': 1, '5': 12, '10': 'evidenceHash'},
    const {'1': 'proposer_address', '3': 14, '4': 1, '5': 12, '10': 'proposerAddress'},
  ],
};

const Data$json = const {
  '1': 'Data',
  '2': const [
    const {'1': 'txs', '3': 1, '4': 3, '5': 12, '10': 'txs'},
  ],
};

const Vote$json = const {
  '1': 'Vote',
  '2': const [
    const {'1': 'type', '3': 1, '4': 1, '5': 14, '6': '.tendermint.types.SignedMsgType', '10': 'type'},
    const {'1': 'height', '3': 2, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'round', '3': 3, '4': 1, '5': 5, '10': 'round'},
    const {'1': 'block_id', '3': 4, '4': 1, '5': 11, '6': '.tendermint.types.BlockID', '8': const {}, '10': 'blockId'},
    const {'1': 'timestamp', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
    const {'1': 'validator_address', '3': 6, '4': 1, '5': 12, '10': 'validatorAddress'},
    const {'1': 'validator_index', '3': 7, '4': 1, '5': 5, '10': 'validatorIndex'},
    const {'1': 'signature', '3': 8, '4': 1, '5': 12, '10': 'signature'},
  ],
};

const Commit$json = const {
  '1': 'Commit',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'round', '3': 2, '4': 1, '5': 5, '10': 'round'},
    const {'1': 'block_id', '3': 3, '4': 1, '5': 11, '6': '.tendermint.types.BlockID', '8': const {}, '10': 'blockId'},
    const {'1': 'signatures', '3': 4, '4': 3, '5': 11, '6': '.tendermint.types.CommitSig', '8': const {}, '10': 'signatures'},
  ],
};

const CommitSig$json = const {
  '1': 'CommitSig',
  '2': const [
    const {'1': 'block_id_flag', '3': 1, '4': 1, '5': 14, '6': '.tendermint.types.BlockIDFlag', '10': 'blockIdFlag'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 12, '10': 'validatorAddress'},
    const {'1': 'timestamp', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
    const {'1': 'signature', '3': 4, '4': 1, '5': 12, '10': 'signature'},
  ],
};

const Proposal$json = const {
  '1': 'Proposal',
  '2': const [
    const {'1': 'type', '3': 1, '4': 1, '5': 14, '6': '.tendermint.types.SignedMsgType', '10': 'type'},
    const {'1': 'height', '3': 2, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'round', '3': 3, '4': 1, '5': 5, '10': 'round'},
    const {'1': 'pol_round', '3': 4, '4': 1, '5': 5, '10': 'polRound'},
    const {'1': 'block_id', '3': 5, '4': 1, '5': 11, '6': '.tendermint.types.BlockID', '8': const {}, '10': 'blockId'},
    const {'1': 'timestamp', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
    const {'1': 'signature', '3': 7, '4': 1, '5': 12, '10': 'signature'},
  ],
};

const SignedHeader$json = const {
  '1': 'SignedHeader',
  '2': const [
    const {'1': 'header', '3': 1, '4': 1, '5': 11, '6': '.tendermint.types.Header', '10': 'header'},
    const {'1': 'commit', '3': 2, '4': 1, '5': 11, '6': '.tendermint.types.Commit', '10': 'commit'},
  ],
};

const LightBlock$json = const {
  '1': 'LightBlock',
  '2': const [
    const {'1': 'signed_header', '3': 1, '4': 1, '5': 11, '6': '.tendermint.types.SignedHeader', '10': 'signedHeader'},
    const {'1': 'validator_set', '3': 2, '4': 1, '5': 11, '6': '.tendermint.types.ValidatorSet', '10': 'validatorSet'},
  ],
};

const BlockMeta$json = const {
  '1': 'BlockMeta',
  '2': const [
    const {'1': 'block_id', '3': 1, '4': 1, '5': 11, '6': '.tendermint.types.BlockID', '8': const {}, '10': 'blockId'},
    const {'1': 'block_size', '3': 2, '4': 1, '5': 3, '10': 'blockSize'},
    const {'1': 'header', '3': 3, '4': 1, '5': 11, '6': '.tendermint.types.Header', '8': const {}, '10': 'header'},
    const {'1': 'num_txs', '3': 4, '4': 1, '5': 3, '10': 'numTxs'},
  ],
};

const TxProof$json = const {
  '1': 'TxProof',
  '2': const [
    const {'1': 'root_hash', '3': 1, '4': 1, '5': 12, '10': 'rootHash'},
    const {'1': 'data', '3': 2, '4': 1, '5': 12, '10': 'data'},
    const {'1': 'proof', '3': 3, '4': 1, '5': 11, '6': '.tendermint.crypto.Proof', '10': 'proof'},
  ],
};

