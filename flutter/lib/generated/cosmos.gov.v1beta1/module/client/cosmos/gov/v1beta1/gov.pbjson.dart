///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/gov.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const VoteOption$json = const {
  '1': 'VoteOption',
  '2': const [
    const {'1': 'VOTE_OPTION_UNSPECIFIED', '2': 0, '3': const {}},
    const {'1': 'VOTE_OPTION_YES', '2': 1, '3': const {}},
    const {'1': 'VOTE_OPTION_ABSTAIN', '2': 2, '3': const {}},
    const {'1': 'VOTE_OPTION_NO', '2': 3, '3': const {}},
    const {'1': 'VOTE_OPTION_NO_WITH_VETO', '2': 4, '3': const {}},
  ],
  '3': const {},
};

const ProposalStatus$json = const {
  '1': 'ProposalStatus',
  '2': const [
    const {'1': 'PROPOSAL_STATUS_UNSPECIFIED', '2': 0, '3': const {}},
    const {'1': 'PROPOSAL_STATUS_DEPOSIT_PERIOD', '2': 1, '3': const {}},
    const {'1': 'PROPOSAL_STATUS_VOTING_PERIOD', '2': 2, '3': const {}},
    const {'1': 'PROPOSAL_STATUS_PASSED', '2': 3, '3': const {}},
    const {'1': 'PROPOSAL_STATUS_REJECTED', '2': 4, '3': const {}},
    const {'1': 'PROPOSAL_STATUS_FAILED', '2': 5, '3': const {}},
  ],
  '3': const {},
};

const TextProposal$json = const {
  '1': 'TextProposal',
  '2': const [
    const {'1': 'title', '3': 1, '4': 1, '5': 9, '10': 'title'},
    const {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
  ],
  '7': const {},
};

const Deposit$json = const {
  '1': 'Deposit',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'proposalId'},
    const {'1': 'depositor', '3': 2, '4': 1, '5': 9, '10': 'depositor'},
    const {'1': 'amount', '3': 3, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const Proposal$json = const {
  '1': 'Proposal',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'proposalId'},
    const {'1': 'content', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'content'},
    const {'1': 'status', '3': 3, '4': 1, '5': 14, '6': '.cosmos.gov.v1beta1.ProposalStatus', '8': const {}, '10': 'status'},
    const {'1': 'final_tally_result', '3': 4, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.TallyResult', '8': const {}, '10': 'finalTallyResult'},
    const {'1': 'submit_time', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'submitTime'},
    const {'1': 'deposit_end_time', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'depositEndTime'},
    const {'1': 'total_deposit', '3': 7, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'totalDeposit'},
    const {'1': 'voting_start_time', '3': 8, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'votingStartTime'},
    const {'1': 'voting_end_time', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'votingEndTime'},
  ],
  '7': const {},
};

const TallyResult$json = const {
  '1': 'TallyResult',
  '2': const [
    const {'1': 'yes', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'yes'},
    const {'1': 'abstain', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'abstain'},
    const {'1': 'no', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'no'},
    const {'1': 'no_with_veto', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'noWithVeto'},
  ],
  '7': const {},
};

const Vote$json = const {
  '1': 'Vote',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'proposalId'},
    const {'1': 'voter', '3': 2, '4': 1, '5': 9, '10': 'voter'},
    const {'1': 'option', '3': 3, '4': 1, '5': 14, '6': '.cosmos.gov.v1beta1.VoteOption', '10': 'option'},
  ],
  '7': const {},
};

const DepositParams$json = const {
  '1': 'DepositParams',
  '2': const [
    const {'1': 'min_deposit', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'minDeposit'},
    const {'1': 'max_deposit_period', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'maxDepositPeriod'},
  ],
};

const VotingParams$json = const {
  '1': 'VotingParams',
  '2': const [
    const {'1': 'voting_period', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'votingPeriod'},
  ],
};

const TallyParams$json = const {
  '1': 'TallyParams',
  '2': const [
    const {'1': 'quorum', '3': 1, '4': 1, '5': 12, '8': const {}, '10': 'quorum'},
    const {'1': 'threshold', '3': 2, '4': 1, '5': 12, '8': const {}, '10': 'threshold'},
    const {'1': 'veto_threshold', '3': 3, '4': 1, '5': 12, '8': const {}, '10': 'vetoThreshold'},
  ],
};

