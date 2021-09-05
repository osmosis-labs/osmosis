///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryProposalRequest$json = const {
  '1': 'QueryProposalRequest',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '10': 'proposalId'},
  ],
};

const QueryProposalResponse$json = const {
  '1': 'QueryProposalResponse',
  '2': const [
    const {'1': 'proposal', '3': 1, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.Proposal', '8': const {}, '10': 'proposal'},
  ],
};

const QueryProposalsRequest$json = const {
  '1': 'QueryProposalsRequest',
  '2': const [
    const {'1': 'proposal_status', '3': 1, '4': 1, '5': 14, '6': '.cosmos.gov.v1beta1.ProposalStatus', '10': 'proposalStatus'},
    const {'1': 'voter', '3': 2, '4': 1, '5': 9, '10': 'voter'},
    const {'1': 'depositor', '3': 3, '4': 1, '5': 9, '10': 'depositor'},
    const {'1': 'pagination', '3': 4, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
  '7': const {},
};

const QueryProposalsResponse$json = const {
  '1': 'QueryProposalsResponse',
  '2': const [
    const {'1': 'proposals', '3': 1, '4': 3, '5': 11, '6': '.cosmos.gov.v1beta1.Proposal', '8': const {}, '10': 'proposals'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryVoteRequest$json = const {
  '1': 'QueryVoteRequest',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '10': 'proposalId'},
    const {'1': 'voter', '3': 2, '4': 1, '5': 9, '10': 'voter'},
  ],
  '7': const {},
};

const QueryVoteResponse$json = const {
  '1': 'QueryVoteResponse',
  '2': const [
    const {'1': 'vote', '3': 1, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.Vote', '8': const {}, '10': 'vote'},
  ],
};

const QueryVotesRequest$json = const {
  '1': 'QueryVotesRequest',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '10': 'proposalId'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryVotesResponse$json = const {
  '1': 'QueryVotesResponse',
  '2': const [
    const {'1': 'votes', '3': 1, '4': 3, '5': 11, '6': '.cosmos.gov.v1beta1.Vote', '8': const {}, '10': 'votes'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryParamsRequest$json = const {
  '1': 'QueryParamsRequest',
  '2': const [
    const {'1': 'params_type', '3': 1, '4': 1, '5': 9, '10': 'paramsType'},
  ],
};

const QueryParamsResponse$json = const {
  '1': 'QueryParamsResponse',
  '2': const [
    const {'1': 'voting_params', '3': 1, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.VotingParams', '8': const {}, '10': 'votingParams'},
    const {'1': 'deposit_params', '3': 2, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.DepositParams', '8': const {}, '10': 'depositParams'},
    const {'1': 'tally_params', '3': 3, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.TallyParams', '8': const {}, '10': 'tallyParams'},
  ],
};

const QueryDepositRequest$json = const {
  '1': 'QueryDepositRequest',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '10': 'proposalId'},
    const {'1': 'depositor', '3': 2, '4': 1, '5': 9, '10': 'depositor'},
  ],
  '7': const {},
};

const QueryDepositResponse$json = const {
  '1': 'QueryDepositResponse',
  '2': const [
    const {'1': 'deposit', '3': 1, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.Deposit', '8': const {}, '10': 'deposit'},
  ],
};

const QueryDepositsRequest$json = const {
  '1': 'QueryDepositsRequest',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '10': 'proposalId'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryDepositsResponse$json = const {
  '1': 'QueryDepositsResponse',
  '2': const [
    const {'1': 'deposits', '3': 1, '4': 3, '5': 11, '6': '.cosmos.gov.v1beta1.Deposit', '8': const {}, '10': 'deposits'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryTallyResultRequest$json = const {
  '1': 'QueryTallyResultRequest',
  '2': const [
    const {'1': 'proposal_id', '3': 1, '4': 1, '5': 4, '10': 'proposalId'},
  ],
};

const QueryTallyResultResponse$json = const {
  '1': 'QueryTallyResultResponse',
  '2': const [
    const {'1': 'tally', '3': 1, '4': 1, '5': 11, '6': '.cosmos.gov.v1beta1.TallyResult', '8': const {}, '10': 'tally'},
  ],
};

