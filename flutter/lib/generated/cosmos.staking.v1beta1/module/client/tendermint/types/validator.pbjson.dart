///
//  Generated code. Do not modify.
//  source: tendermint/types/validator.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const ValidatorSet$json = const {
  '1': 'ValidatorSet',
  '2': const [
    const {'1': 'validators', '3': 1, '4': 3, '5': 11, '6': '.tendermint.types.Validator', '10': 'validators'},
    const {'1': 'proposer', '3': 2, '4': 1, '5': 11, '6': '.tendermint.types.Validator', '10': 'proposer'},
    const {'1': 'total_voting_power', '3': 3, '4': 1, '5': 3, '10': 'totalVotingPower'},
  ],
};

const Validator$json = const {
  '1': 'Validator',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 12, '10': 'address'},
    const {'1': 'pub_key', '3': 2, '4': 1, '5': 11, '6': '.tendermint.crypto.PublicKey', '8': const {}, '10': 'pubKey'},
    const {'1': 'voting_power', '3': 3, '4': 1, '5': 3, '10': 'votingPower'},
    const {'1': 'proposer_priority', '3': 4, '4': 1, '5': 3, '10': 'proposerPriority'},
  ],
};

const SimpleValidator$json = const {
  '1': 'SimpleValidator',
  '2': const [
    const {'1': 'pub_key', '3': 1, '4': 1, '5': 11, '6': '.tendermint.crypto.PublicKey', '10': 'pubKey'},
    const {'1': 'voting_power', '3': 2, '4': 1, '5': 3, '10': 'votingPower'},
  ],
};

