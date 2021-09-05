///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/distribution.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const Params$json = const {
  '1': 'Params',
  '2': const [
    const {'1': 'community_tax', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'communityTax'},
    const {'1': 'base_proposer_reward', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'baseProposerReward'},
    const {'1': 'bonus_proposer_reward', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'bonusProposerReward'},
    const {'1': 'withdraw_addr_enabled', '3': 4, '4': 1, '5': 8, '8': const {}, '10': 'withdrawAddrEnabled'},
  ],
  '7': const {},
};

const ValidatorHistoricalRewards$json = const {
  '1': 'ValidatorHistoricalRewards',
  '2': const [
    const {'1': 'cumulative_reward_ratio', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'cumulativeRewardRatio'},
    const {'1': 'reference_count', '3': 2, '4': 1, '5': 13, '8': const {}, '10': 'referenceCount'},
  ],
};

const ValidatorCurrentRewards$json = const {
  '1': 'ValidatorCurrentRewards',
  '2': const [
    const {'1': 'rewards', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'rewards'},
    const {'1': 'period', '3': 2, '4': 1, '5': 4, '10': 'period'},
  ],
};

const ValidatorAccumulatedCommission$json = const {
  '1': 'ValidatorAccumulatedCommission',
  '2': const [
    const {'1': 'commission', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'commission'},
  ],
};

const ValidatorOutstandingRewards$json = const {
  '1': 'ValidatorOutstandingRewards',
  '2': const [
    const {'1': 'rewards', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'rewards'},
  ],
};

const ValidatorSlashEvent$json = const {
  '1': 'ValidatorSlashEvent',
  '2': const [
    const {'1': 'validator_period', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'validatorPeriod'},
    const {'1': 'fraction', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'fraction'},
  ],
};

const ValidatorSlashEvents$json = const {
  '1': 'ValidatorSlashEvents',
  '2': const [
    const {'1': 'validator_slash_events', '3': 1, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorSlashEvent', '8': const {}, '10': 'validatorSlashEvents'},
  ],
  '7': const {},
};

const FeePool$json = const {
  '1': 'FeePool',
  '2': const [
    const {'1': 'community_pool', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'communityPool'},
  ],
};

const CommunityPoolSpendProposal$json = const {
  '1': 'CommunityPoolSpendProposal',
  '2': const [
    const {'1': 'title', '3': 1, '4': 1, '5': 9, '10': 'title'},
    const {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    const {'1': 'recipient', '3': 3, '4': 1, '5': 9, '10': 'recipient'},
    const {'1': 'amount', '3': 4, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const DelegatorStartingInfo$json = const {
  '1': 'DelegatorStartingInfo',
  '2': const [
    const {'1': 'previous_period', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'previousPeriod'},
    const {'1': 'stake', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'stake'},
    const {'1': 'height', '3': 3, '4': 1, '5': 4, '8': const {}, '10': 'height'},
  ],
};

const DelegationDelegatorReward$json = const {
  '1': 'DelegationDelegatorReward',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'reward', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'reward'},
  ],
  '7': const {},
};

const CommunityPoolSpendProposalWithDeposit$json = const {
  '1': 'CommunityPoolSpendProposalWithDeposit',
  '2': const [
    const {'1': 'title', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'title'},
    const {'1': 'description', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'description'},
    const {'1': 'recipient', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'recipient'},
    const {'1': 'amount', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'amount'},
    const {'1': 'deposit', '3': 5, '4': 1, '5': 9, '8': const {}, '10': 'deposit'},
  ],
  '7': const {},
};

