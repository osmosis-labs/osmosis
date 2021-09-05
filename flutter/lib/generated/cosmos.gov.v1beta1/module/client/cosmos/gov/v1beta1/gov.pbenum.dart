///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/gov.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

// ignore_for_file: UNDEFINED_SHOWN_NAME
import 'dart:core' as $core;
import 'package:protobuf/protobuf.dart' as $pb;

class VoteOption extends $pb.ProtobufEnum {
  static const VoteOption VOTE_OPTION_UNSPECIFIED = VoteOption._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'VOTE_OPTION_UNSPECIFIED');
  static const VoteOption VOTE_OPTION_YES = VoteOption._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'VOTE_OPTION_YES');
  static const VoteOption VOTE_OPTION_ABSTAIN = VoteOption._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'VOTE_OPTION_ABSTAIN');
  static const VoteOption VOTE_OPTION_NO = VoteOption._(3, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'VOTE_OPTION_NO');
  static const VoteOption VOTE_OPTION_NO_WITH_VETO = VoteOption._(4, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'VOTE_OPTION_NO_WITH_VETO');

  static const $core.List<VoteOption> values = <VoteOption> [
    VOTE_OPTION_UNSPECIFIED,
    VOTE_OPTION_YES,
    VOTE_OPTION_ABSTAIN,
    VOTE_OPTION_NO,
    VOTE_OPTION_NO_WITH_VETO,
  ];

  static final $core.Map<$core.int, VoteOption> _byValue = $pb.ProtobufEnum.initByValue(values);
  static VoteOption valueOf($core.int value) => _byValue[value];

  const VoteOption._($core.int v, $core.String n) : super(v, n);
}

class ProposalStatus extends $pb.ProtobufEnum {
  static const ProposalStatus PROPOSAL_STATUS_UNSPECIFIED = ProposalStatus._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'PROPOSAL_STATUS_UNSPECIFIED');
  static const ProposalStatus PROPOSAL_STATUS_DEPOSIT_PERIOD = ProposalStatus._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'PROPOSAL_STATUS_DEPOSIT_PERIOD');
  static const ProposalStatus PROPOSAL_STATUS_VOTING_PERIOD = ProposalStatus._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'PROPOSAL_STATUS_VOTING_PERIOD');
  static const ProposalStatus PROPOSAL_STATUS_PASSED = ProposalStatus._(3, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'PROPOSAL_STATUS_PASSED');
  static const ProposalStatus PROPOSAL_STATUS_REJECTED = ProposalStatus._(4, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'PROPOSAL_STATUS_REJECTED');
  static const ProposalStatus PROPOSAL_STATUS_FAILED = ProposalStatus._(5, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'PROPOSAL_STATUS_FAILED');

  static const $core.List<ProposalStatus> values = <ProposalStatus> [
    PROPOSAL_STATUS_UNSPECIFIED,
    PROPOSAL_STATUS_DEPOSIT_PERIOD,
    PROPOSAL_STATUS_VOTING_PERIOD,
    PROPOSAL_STATUS_PASSED,
    PROPOSAL_STATUS_REJECTED,
    PROPOSAL_STATUS_FAILED,
  ];

  static final $core.Map<$core.int, ProposalStatus> _byValue = $pb.ProtobufEnum.initByValue(values);
  static ProposalStatus valueOf($core.int value) => _byValue[value];

  const ProposalStatus._($core.int v, $core.String n) : super(v, n);
}

