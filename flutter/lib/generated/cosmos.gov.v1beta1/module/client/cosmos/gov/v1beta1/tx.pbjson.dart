///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgSubmitProposal$json = const {
  '1': 'MsgSubmitProposal',
  '2': const [
    const {'1': 'content', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'content'},
    const {'1': 'initial_deposit', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'initialDeposit'},
    const {'1': 'proposer', '3': 3, '4': 1, '5': 9, '10': 'proposer'},
  ],
  '7': const {},
};

const MsgSubmitProposalResponse$json = const {
  '1': 'MsgSubmitProposalResponse',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'proposalId'},
  ],
};

const MsgVote$json = const {
  '1': 'MsgVote',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'proposalId'},
    const {'1': 'voter', '3': 2, '4': 1, '5': 9, '10': 'voter'},
    const {'1': 'option', '3': 3, '4': 1, '5': 14, '6': '.cosmos.gov.v1beta1.VoteOption', '10': 'option'},
  ],
  '7': const {},
};

const MsgVoteResponse$json = const {
  '1': 'MsgVoteResponse',
};

const MsgDeposit$json = const {
  '1': 'MsgDeposit',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'proposalId'},
    const {'1': 'depositor', '3': 2, '4': 1, '5': 9, '10': 'depositor'},
    const {'1': 'amount', '3': 3, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const MsgDepositResponse$json = const {
  '1': 'MsgDepositResponse',
};

