///
//  Generated code. Do not modify.
//  source: cosmos/gov/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'gov.pb.dart' as $6;
import '../../base/query/v1beta1/pagination.pb.dart' as $8;

import 'gov.pbenum.dart' as $6;

class QueryProposalRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryProposalRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryProposalRequest._() : super();
  factory QueryProposalRequest() => create();
  factory QueryProposalRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryProposalRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryProposalRequest clone() => QueryProposalRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryProposalRequest copyWith(void Function(QueryProposalRequest) updates) => super.copyWith((message) => updates(message as QueryProposalRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryProposalRequest create() => QueryProposalRequest._();
  QueryProposalRequest createEmptyInstance() => create();
  static $pb.PbList<QueryProposalRequest> createRepeated() => $pb.PbList<QueryProposalRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryProposalRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryProposalRequest>(create);
  static QueryProposalRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);
}

class QueryProposalResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryProposalResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOM<$6.Proposal>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposal', subBuilder: $6.Proposal.create)
    ..hasRequiredFields = false
  ;

  QueryProposalResponse._() : super();
  factory QueryProposalResponse() => create();
  factory QueryProposalResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryProposalResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryProposalResponse clone() => QueryProposalResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryProposalResponse copyWith(void Function(QueryProposalResponse) updates) => super.copyWith((message) => updates(message as QueryProposalResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryProposalResponse create() => QueryProposalResponse._();
  QueryProposalResponse createEmptyInstance() => create();
  static $pb.PbList<QueryProposalResponse> createRepeated() => $pb.PbList<QueryProposalResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryProposalResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryProposalResponse>(create);
  static QueryProposalResponse _defaultInstance;

  @$pb.TagNumber(1)
  $6.Proposal get proposal => $_getN(0);
  @$pb.TagNumber(1)
  set proposal($6.Proposal v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposal() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposal() => clearField(1);
  @$pb.TagNumber(1)
  $6.Proposal ensureProposal() => $_ensure(0);
}

class QueryProposalsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryProposalsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..e<$6.ProposalStatus>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalStatus', $pb.PbFieldType.OE, defaultOrMaker: $6.ProposalStatus.PROPOSAL_STATUS_UNSPECIFIED, valueOf: $6.ProposalStatus.valueOf, enumValues: $6.ProposalStatus.values)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'voter')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'depositor')
    ..aOM<$8.PageRequest>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryProposalsRequest._() : super();
  factory QueryProposalsRequest() => create();
  factory QueryProposalsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryProposalsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryProposalsRequest clone() => QueryProposalsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryProposalsRequest copyWith(void Function(QueryProposalsRequest) updates) => super.copyWith((message) => updates(message as QueryProposalsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryProposalsRequest create() => QueryProposalsRequest._();
  QueryProposalsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryProposalsRequest> createRepeated() => $pb.PbList<QueryProposalsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryProposalsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryProposalsRequest>(create);
  static QueryProposalsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $6.ProposalStatus get proposalStatus => $_getN(0);
  @$pb.TagNumber(1)
  set proposalStatus($6.ProposalStatus v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalStatus() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get voter => $_getSZ(1);
  @$pb.TagNumber(2)
  set voter($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasVoter() => $_has(1);
  @$pb.TagNumber(2)
  void clearVoter() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get depositor => $_getSZ(2);
  @$pb.TagNumber(3)
  set depositor($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasDepositor() => $_has(2);
  @$pb.TagNumber(3)
  void clearDepositor() => clearField(3);

  @$pb.TagNumber(4)
  $8.PageRequest get pagination => $_getN(3);
  @$pb.TagNumber(4)
  set pagination($8.PageRequest v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasPagination() => $_has(3);
  @$pb.TagNumber(4)
  void clearPagination() => clearField(4);
  @$pb.TagNumber(4)
  $8.PageRequest ensurePagination() => $_ensure(3);
}

class QueryProposalsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryProposalsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..pc<$6.Proposal>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposals', $pb.PbFieldType.PM, subBuilder: $6.Proposal.create)
    ..aOM<$8.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageResponse.create)
    ..hasRequiredFields = false
  ;

  QueryProposalsResponse._() : super();
  factory QueryProposalsResponse() => create();
  factory QueryProposalsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryProposalsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryProposalsResponse clone() => QueryProposalsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryProposalsResponse copyWith(void Function(QueryProposalsResponse) updates) => super.copyWith((message) => updates(message as QueryProposalsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryProposalsResponse create() => QueryProposalsResponse._();
  QueryProposalsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryProposalsResponse> createRepeated() => $pb.PbList<QueryProposalsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryProposalsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryProposalsResponse>(create);
  static QueryProposalsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$6.Proposal> get proposals => $_getList(0);

  @$pb.TagNumber(2)
  $8.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($8.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $8.PageResponse ensurePagination() => $_ensure(1);
}

class QueryVoteRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryVoteRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'voter')
    ..hasRequiredFields = false
  ;

  QueryVoteRequest._() : super();
  factory QueryVoteRequest() => create();
  factory QueryVoteRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryVoteRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryVoteRequest clone() => QueryVoteRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryVoteRequest copyWith(void Function(QueryVoteRequest) updates) => super.copyWith((message) => updates(message as QueryVoteRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryVoteRequest create() => QueryVoteRequest._();
  QueryVoteRequest createEmptyInstance() => create();
  static $pb.PbList<QueryVoteRequest> createRepeated() => $pb.PbList<QueryVoteRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryVoteRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryVoteRequest>(create);
  static QueryVoteRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get voter => $_getSZ(1);
  @$pb.TagNumber(2)
  set voter($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasVoter() => $_has(1);
  @$pb.TagNumber(2)
  void clearVoter() => clearField(2);
}

class QueryVoteResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryVoteResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOM<$6.Vote>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'vote', subBuilder: $6.Vote.create)
    ..hasRequiredFields = false
  ;

  QueryVoteResponse._() : super();
  factory QueryVoteResponse() => create();
  factory QueryVoteResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryVoteResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryVoteResponse clone() => QueryVoteResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryVoteResponse copyWith(void Function(QueryVoteResponse) updates) => super.copyWith((message) => updates(message as QueryVoteResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryVoteResponse create() => QueryVoteResponse._();
  QueryVoteResponse createEmptyInstance() => create();
  static $pb.PbList<QueryVoteResponse> createRepeated() => $pb.PbList<QueryVoteResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryVoteResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryVoteResponse>(create);
  static QueryVoteResponse _defaultInstance;

  @$pb.TagNumber(1)
  $6.Vote get vote => $_getN(0);
  @$pb.TagNumber(1)
  set vote($6.Vote v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasVote() => $_has(0);
  @$pb.TagNumber(1)
  void clearVote() => clearField(1);
  @$pb.TagNumber(1)
  $6.Vote ensureVote() => $_ensure(0);
}

class QueryVotesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryVotesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$8.PageRequest>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryVotesRequest._() : super();
  factory QueryVotesRequest() => create();
  factory QueryVotesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryVotesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryVotesRequest clone() => QueryVotesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryVotesRequest copyWith(void Function(QueryVotesRequest) updates) => super.copyWith((message) => updates(message as QueryVotesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryVotesRequest create() => QueryVotesRequest._();
  QueryVotesRequest createEmptyInstance() => create();
  static $pb.PbList<QueryVotesRequest> createRepeated() => $pb.PbList<QueryVotesRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryVotesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryVotesRequest>(create);
  static QueryVotesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $8.PageRequest get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($8.PageRequest v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $8.PageRequest ensurePagination() => $_ensure(1);
}

class QueryVotesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryVotesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..pc<$6.Vote>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votes', $pb.PbFieldType.PM, subBuilder: $6.Vote.create)
    ..aOM<$8.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageResponse.create)
    ..hasRequiredFields = false
  ;

  QueryVotesResponse._() : super();
  factory QueryVotesResponse() => create();
  factory QueryVotesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryVotesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryVotesResponse clone() => QueryVotesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryVotesResponse copyWith(void Function(QueryVotesResponse) updates) => super.copyWith((message) => updates(message as QueryVotesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryVotesResponse create() => QueryVotesResponse._();
  QueryVotesResponse createEmptyInstance() => create();
  static $pb.PbList<QueryVotesResponse> createRepeated() => $pb.PbList<QueryVotesResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryVotesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryVotesResponse>(create);
  static QueryVotesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$6.Vote> get votes => $_getList(0);

  @$pb.TagNumber(2)
  $8.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($8.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $8.PageResponse ensurePagination() => $_ensure(1);
}

class QueryParamsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryParamsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'paramsType')
    ..hasRequiredFields = false
  ;

  QueryParamsRequest._() : super();
  factory QueryParamsRequest() => create();
  factory QueryParamsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryParamsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryParamsRequest clone() => QueryParamsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryParamsRequest copyWith(void Function(QueryParamsRequest) updates) => super.copyWith((message) => updates(message as QueryParamsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryParamsRequest create() => QueryParamsRequest._();
  QueryParamsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryParamsRequest> createRepeated() => $pb.PbList<QueryParamsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryParamsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryParamsRequest>(create);
  static QueryParamsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get paramsType => $_getSZ(0);
  @$pb.TagNumber(1)
  set paramsType($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasParamsType() => $_has(0);
  @$pb.TagNumber(1)
  void clearParamsType() => clearField(1);
}

class QueryParamsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryParamsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOM<$6.VotingParams>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'votingParams', subBuilder: $6.VotingParams.create)
    ..aOM<$6.DepositParams>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'depositParams', subBuilder: $6.DepositParams.create)
    ..aOM<$6.TallyParams>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tallyParams', subBuilder: $6.TallyParams.create)
    ..hasRequiredFields = false
  ;

  QueryParamsResponse._() : super();
  factory QueryParamsResponse() => create();
  factory QueryParamsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryParamsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryParamsResponse clone() => QueryParamsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryParamsResponse copyWith(void Function(QueryParamsResponse) updates) => super.copyWith((message) => updates(message as QueryParamsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryParamsResponse create() => QueryParamsResponse._();
  QueryParamsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryParamsResponse> createRepeated() => $pb.PbList<QueryParamsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryParamsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryParamsResponse>(create);
  static QueryParamsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $6.VotingParams get votingParams => $_getN(0);
  @$pb.TagNumber(1)
  set votingParams($6.VotingParams v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasVotingParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearVotingParams() => clearField(1);
  @$pb.TagNumber(1)
  $6.VotingParams ensureVotingParams() => $_ensure(0);

  @$pb.TagNumber(2)
  $6.DepositParams get depositParams => $_getN(1);
  @$pb.TagNumber(2)
  set depositParams($6.DepositParams v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasDepositParams() => $_has(1);
  @$pb.TagNumber(2)
  void clearDepositParams() => clearField(2);
  @$pb.TagNumber(2)
  $6.DepositParams ensureDepositParams() => $_ensure(1);

  @$pb.TagNumber(3)
  $6.TallyParams get tallyParams => $_getN(2);
  @$pb.TagNumber(3)
  set tallyParams($6.TallyParams v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasTallyParams() => $_has(2);
  @$pb.TagNumber(3)
  void clearTallyParams() => clearField(3);
  @$pb.TagNumber(3)
  $6.TallyParams ensureTallyParams() => $_ensure(2);
}

class QueryDepositRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDepositRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'depositor')
    ..hasRequiredFields = false
  ;

  QueryDepositRequest._() : super();
  factory QueryDepositRequest() => create();
  factory QueryDepositRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDepositRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDepositRequest clone() => QueryDepositRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDepositRequest copyWith(void Function(QueryDepositRequest) updates) => super.copyWith((message) => updates(message as QueryDepositRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDepositRequest create() => QueryDepositRequest._();
  QueryDepositRequest createEmptyInstance() => create();
  static $pb.PbList<QueryDepositRequest> createRepeated() => $pb.PbList<QueryDepositRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryDepositRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDepositRequest>(create);
  static QueryDepositRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get depositor => $_getSZ(1);
  @$pb.TagNumber(2)
  set depositor($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDepositor() => $_has(1);
  @$pb.TagNumber(2)
  void clearDepositor() => clearField(2);
}

class QueryDepositResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDepositResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOM<$6.Deposit>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'deposit', subBuilder: $6.Deposit.create)
    ..hasRequiredFields = false
  ;

  QueryDepositResponse._() : super();
  factory QueryDepositResponse() => create();
  factory QueryDepositResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDepositResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDepositResponse clone() => QueryDepositResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDepositResponse copyWith(void Function(QueryDepositResponse) updates) => super.copyWith((message) => updates(message as QueryDepositResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDepositResponse create() => QueryDepositResponse._();
  QueryDepositResponse createEmptyInstance() => create();
  static $pb.PbList<QueryDepositResponse> createRepeated() => $pb.PbList<QueryDepositResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryDepositResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDepositResponse>(create);
  static QueryDepositResponse _defaultInstance;

  @$pb.TagNumber(1)
  $6.Deposit get deposit => $_getN(0);
  @$pb.TagNumber(1)
  set deposit($6.Deposit v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasDeposit() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeposit() => clearField(1);
  @$pb.TagNumber(1)
  $6.Deposit ensureDeposit() => $_ensure(0);
}

class QueryDepositsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDepositsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$8.PageRequest>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryDepositsRequest._() : super();
  factory QueryDepositsRequest() => create();
  factory QueryDepositsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDepositsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDepositsRequest clone() => QueryDepositsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDepositsRequest copyWith(void Function(QueryDepositsRequest) updates) => super.copyWith((message) => updates(message as QueryDepositsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDepositsRequest create() => QueryDepositsRequest._();
  QueryDepositsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryDepositsRequest> createRepeated() => $pb.PbList<QueryDepositsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryDepositsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDepositsRequest>(create);
  static QueryDepositsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);

  @$pb.TagNumber(2)
  $8.PageRequest get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($8.PageRequest v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $8.PageRequest ensurePagination() => $_ensure(1);
}

class QueryDepositsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDepositsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..pc<$6.Deposit>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'deposits', $pb.PbFieldType.PM, subBuilder: $6.Deposit.create)
    ..aOM<$8.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageResponse.create)
    ..hasRequiredFields = false
  ;

  QueryDepositsResponse._() : super();
  factory QueryDepositsResponse() => create();
  factory QueryDepositsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDepositsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDepositsResponse clone() => QueryDepositsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDepositsResponse copyWith(void Function(QueryDepositsResponse) updates) => super.copyWith((message) => updates(message as QueryDepositsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDepositsResponse create() => QueryDepositsResponse._();
  QueryDepositsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryDepositsResponse> createRepeated() => $pb.PbList<QueryDepositsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryDepositsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDepositsResponse>(create);
  static QueryDepositsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$6.Deposit> get deposits => $_getList(0);

  @$pb.TagNumber(2)
  $8.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($8.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $8.PageResponse ensurePagination() => $_ensure(1);
}

class QueryTallyResultRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryTallyResultRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proposalId', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryTallyResultRequest._() : super();
  factory QueryTallyResultRequest() => create();
  factory QueryTallyResultRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryTallyResultRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryTallyResultRequest clone() => QueryTallyResultRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryTallyResultRequest copyWith(void Function(QueryTallyResultRequest) updates) => super.copyWith((message) => updates(message as QueryTallyResultRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryTallyResultRequest create() => QueryTallyResultRequest._();
  QueryTallyResultRequest createEmptyInstance() => create();
  static $pb.PbList<QueryTallyResultRequest> createRepeated() => $pb.PbList<QueryTallyResultRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryTallyResultRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryTallyResultRequest>(create);
  static QueryTallyResultRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get proposalId => $_getI64(0);
  @$pb.TagNumber(1)
  set proposalId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasProposalId() => $_has(0);
  @$pb.TagNumber(1)
  void clearProposalId() => clearField(1);
}

class QueryTallyResultResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryTallyResultResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.gov.v1beta1'), createEmptyInstance: create)
    ..aOM<$6.TallyResult>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tally', subBuilder: $6.TallyResult.create)
    ..hasRequiredFields = false
  ;

  QueryTallyResultResponse._() : super();
  factory QueryTallyResultResponse() => create();
  factory QueryTallyResultResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryTallyResultResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryTallyResultResponse clone() => QueryTallyResultResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryTallyResultResponse copyWith(void Function(QueryTallyResultResponse) updates) => super.copyWith((message) => updates(message as QueryTallyResultResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryTallyResultResponse create() => QueryTallyResultResponse._();
  QueryTallyResultResponse createEmptyInstance() => create();
  static $pb.PbList<QueryTallyResultResponse> createRepeated() => $pb.PbList<QueryTallyResultResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryTallyResultResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryTallyResultResponse>(create);
  static QueryTallyResultResponse _defaultInstance;

  @$pb.TagNumber(1)
  $6.TallyResult get tally => $_getN(0);
  @$pb.TagNumber(1)
  set tally($6.TallyResult v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasTally() => $_has(0);
  @$pb.TagNumber(1)
  void clearTally() => clearField(1);
  @$pb.TagNumber(1)
  $6.TallyResult ensureTally() => $_ensure(0);
}

