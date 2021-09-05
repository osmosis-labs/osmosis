///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'distribution.pb.dart' as $3;
import '../../base/query/v1beta1/pagination.pb.dart' as $5;
import '../../base/v1beta1/coin.pb.dart' as $2;

class QueryParamsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryParamsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
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
}

class QueryParamsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryParamsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOM<$3.Params>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'params', subBuilder: $3.Params.create)
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
  $3.Params get params => $_getN(0);
  @$pb.TagNumber(1)
  set params($3.Params v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearParams() => clearField(1);
  @$pb.TagNumber(1)
  $3.Params ensureParams() => $_ensure(0);
}

class QueryValidatorOutstandingRewardsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryValidatorOutstandingRewardsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..hasRequiredFields = false
  ;

  QueryValidatorOutstandingRewardsRequest._() : super();
  factory QueryValidatorOutstandingRewardsRequest() => create();
  factory QueryValidatorOutstandingRewardsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryValidatorOutstandingRewardsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryValidatorOutstandingRewardsRequest clone() => QueryValidatorOutstandingRewardsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryValidatorOutstandingRewardsRequest copyWith(void Function(QueryValidatorOutstandingRewardsRequest) updates) => super.copyWith((message) => updates(message as QueryValidatorOutstandingRewardsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryValidatorOutstandingRewardsRequest create() => QueryValidatorOutstandingRewardsRequest._();
  QueryValidatorOutstandingRewardsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryValidatorOutstandingRewardsRequest> createRepeated() => $pb.PbList<QueryValidatorOutstandingRewardsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryValidatorOutstandingRewardsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryValidatorOutstandingRewardsRequest>(create);
  static QueryValidatorOutstandingRewardsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);
}

class QueryValidatorOutstandingRewardsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryValidatorOutstandingRewardsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOM<$3.ValidatorOutstandingRewards>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', subBuilder: $3.ValidatorOutstandingRewards.create)
    ..hasRequiredFields = false
  ;

  QueryValidatorOutstandingRewardsResponse._() : super();
  factory QueryValidatorOutstandingRewardsResponse() => create();
  factory QueryValidatorOutstandingRewardsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryValidatorOutstandingRewardsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryValidatorOutstandingRewardsResponse clone() => QueryValidatorOutstandingRewardsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryValidatorOutstandingRewardsResponse copyWith(void Function(QueryValidatorOutstandingRewardsResponse) updates) => super.copyWith((message) => updates(message as QueryValidatorOutstandingRewardsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryValidatorOutstandingRewardsResponse create() => QueryValidatorOutstandingRewardsResponse._();
  QueryValidatorOutstandingRewardsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryValidatorOutstandingRewardsResponse> createRepeated() => $pb.PbList<QueryValidatorOutstandingRewardsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryValidatorOutstandingRewardsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryValidatorOutstandingRewardsResponse>(create);
  static QueryValidatorOutstandingRewardsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $3.ValidatorOutstandingRewards get rewards => $_getN(0);
  @$pb.TagNumber(1)
  set rewards($3.ValidatorOutstandingRewards v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasRewards() => $_has(0);
  @$pb.TagNumber(1)
  void clearRewards() => clearField(1);
  @$pb.TagNumber(1)
  $3.ValidatorOutstandingRewards ensureRewards() => $_ensure(0);
}

class QueryValidatorCommissionRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryValidatorCommissionRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..hasRequiredFields = false
  ;

  QueryValidatorCommissionRequest._() : super();
  factory QueryValidatorCommissionRequest() => create();
  factory QueryValidatorCommissionRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryValidatorCommissionRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryValidatorCommissionRequest clone() => QueryValidatorCommissionRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryValidatorCommissionRequest copyWith(void Function(QueryValidatorCommissionRequest) updates) => super.copyWith((message) => updates(message as QueryValidatorCommissionRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryValidatorCommissionRequest create() => QueryValidatorCommissionRequest._();
  QueryValidatorCommissionRequest createEmptyInstance() => create();
  static $pb.PbList<QueryValidatorCommissionRequest> createRepeated() => $pb.PbList<QueryValidatorCommissionRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryValidatorCommissionRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryValidatorCommissionRequest>(create);
  static QueryValidatorCommissionRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);
}

class QueryValidatorCommissionResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryValidatorCommissionResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOM<$3.ValidatorAccumulatedCommission>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commission', subBuilder: $3.ValidatorAccumulatedCommission.create)
    ..hasRequiredFields = false
  ;

  QueryValidatorCommissionResponse._() : super();
  factory QueryValidatorCommissionResponse() => create();
  factory QueryValidatorCommissionResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryValidatorCommissionResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryValidatorCommissionResponse clone() => QueryValidatorCommissionResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryValidatorCommissionResponse copyWith(void Function(QueryValidatorCommissionResponse) updates) => super.copyWith((message) => updates(message as QueryValidatorCommissionResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryValidatorCommissionResponse create() => QueryValidatorCommissionResponse._();
  QueryValidatorCommissionResponse createEmptyInstance() => create();
  static $pb.PbList<QueryValidatorCommissionResponse> createRepeated() => $pb.PbList<QueryValidatorCommissionResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryValidatorCommissionResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryValidatorCommissionResponse>(create);
  static QueryValidatorCommissionResponse _defaultInstance;

  @$pb.TagNumber(1)
  $3.ValidatorAccumulatedCommission get commission => $_getN(0);
  @$pb.TagNumber(1)
  set commission($3.ValidatorAccumulatedCommission v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasCommission() => $_has(0);
  @$pb.TagNumber(1)
  void clearCommission() => clearField(1);
  @$pb.TagNumber(1)
  $3.ValidatorAccumulatedCommission ensureCommission() => $_ensure(0);
}

class QueryValidatorSlashesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryValidatorSlashesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'startingHeight', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'endingHeight', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<$5.PageRequest>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $5.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryValidatorSlashesRequest._() : super();
  factory QueryValidatorSlashesRequest() => create();
  factory QueryValidatorSlashesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryValidatorSlashesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryValidatorSlashesRequest clone() => QueryValidatorSlashesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryValidatorSlashesRequest copyWith(void Function(QueryValidatorSlashesRequest) updates) => super.copyWith((message) => updates(message as QueryValidatorSlashesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryValidatorSlashesRequest create() => QueryValidatorSlashesRequest._();
  QueryValidatorSlashesRequest createEmptyInstance() => create();
  static $pb.PbList<QueryValidatorSlashesRequest> createRepeated() => $pb.PbList<QueryValidatorSlashesRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryValidatorSlashesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryValidatorSlashesRequest>(create);
  static QueryValidatorSlashesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get validatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set validatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasValidatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearValidatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get startingHeight => $_getI64(1);
  @$pb.TagNumber(2)
  set startingHeight($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasStartingHeight() => $_has(1);
  @$pb.TagNumber(2)
  void clearStartingHeight() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get endingHeight => $_getI64(2);
  @$pb.TagNumber(3)
  set endingHeight($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasEndingHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearEndingHeight() => clearField(3);

  @$pb.TagNumber(4)
  $5.PageRequest get pagination => $_getN(3);
  @$pb.TagNumber(4)
  set pagination($5.PageRequest v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasPagination() => $_has(3);
  @$pb.TagNumber(4)
  void clearPagination() => clearField(4);
  @$pb.TagNumber(4)
  $5.PageRequest ensurePagination() => $_ensure(3);
}

class QueryValidatorSlashesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryValidatorSlashesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$3.ValidatorSlashEvent>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'slashes', $pb.PbFieldType.PM, subBuilder: $3.ValidatorSlashEvent.create)
    ..aOM<$5.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $5.PageResponse.create)
    ..hasRequiredFields = false
  ;

  QueryValidatorSlashesResponse._() : super();
  factory QueryValidatorSlashesResponse() => create();
  factory QueryValidatorSlashesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryValidatorSlashesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryValidatorSlashesResponse clone() => QueryValidatorSlashesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryValidatorSlashesResponse copyWith(void Function(QueryValidatorSlashesResponse) updates) => super.copyWith((message) => updates(message as QueryValidatorSlashesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryValidatorSlashesResponse create() => QueryValidatorSlashesResponse._();
  QueryValidatorSlashesResponse createEmptyInstance() => create();
  static $pb.PbList<QueryValidatorSlashesResponse> createRepeated() => $pb.PbList<QueryValidatorSlashesResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryValidatorSlashesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryValidatorSlashesResponse>(create);
  static QueryValidatorSlashesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$3.ValidatorSlashEvent> get slashes => $_getList(0);

  @$pb.TagNumber(2)
  $5.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($5.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $5.PageResponse ensurePagination() => $_ensure(1);
}

class QueryDelegationRewardsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegationRewardsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validatorAddress')
    ..hasRequiredFields = false
  ;

  QueryDelegationRewardsRequest._() : super();
  factory QueryDelegationRewardsRequest() => create();
  factory QueryDelegationRewardsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegationRewardsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegationRewardsRequest clone() => QueryDelegationRewardsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegationRewardsRequest copyWith(void Function(QueryDelegationRewardsRequest) updates) => super.copyWith((message) => updates(message as QueryDelegationRewardsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegationRewardsRequest create() => QueryDelegationRewardsRequest._();
  QueryDelegationRewardsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryDelegationRewardsRequest> createRepeated() => $pb.PbList<QueryDelegationRewardsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegationRewardsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegationRewardsRequest>(create);
  static QueryDelegationRewardsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get validatorAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set validatorAddress($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasValidatorAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearValidatorAddress() => clearField(2);
}

class QueryDelegationRewardsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegationRewardsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.DecCoin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  QueryDelegationRewardsResponse._() : super();
  factory QueryDelegationRewardsResponse() => create();
  factory QueryDelegationRewardsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegationRewardsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegationRewardsResponse clone() => QueryDelegationRewardsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegationRewardsResponse copyWith(void Function(QueryDelegationRewardsResponse) updates) => super.copyWith((message) => updates(message as QueryDelegationRewardsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegationRewardsResponse create() => QueryDelegationRewardsResponse._();
  QueryDelegationRewardsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryDelegationRewardsResponse> createRepeated() => $pb.PbList<QueryDelegationRewardsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegationRewardsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegationRewardsResponse>(create);
  static QueryDelegationRewardsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.DecCoin> get rewards => $_getList(0);
}

class QueryDelegationTotalRewardsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegationTotalRewardsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..hasRequiredFields = false
  ;

  QueryDelegationTotalRewardsRequest._() : super();
  factory QueryDelegationTotalRewardsRequest() => create();
  factory QueryDelegationTotalRewardsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegationTotalRewardsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegationTotalRewardsRequest clone() => QueryDelegationTotalRewardsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegationTotalRewardsRequest copyWith(void Function(QueryDelegationTotalRewardsRequest) updates) => super.copyWith((message) => updates(message as QueryDelegationTotalRewardsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegationTotalRewardsRequest create() => QueryDelegationTotalRewardsRequest._();
  QueryDelegationTotalRewardsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryDelegationTotalRewardsRequest> createRepeated() => $pb.PbList<QueryDelegationTotalRewardsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegationTotalRewardsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegationTotalRewardsRequest>(create);
  static QueryDelegationTotalRewardsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);
}

class QueryDelegationTotalRewardsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegationTotalRewardsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$3.DelegationDelegatorReward>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'rewards', $pb.PbFieldType.PM, subBuilder: $3.DelegationDelegatorReward.create)
    ..pc<$2.DecCoin>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'total', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  QueryDelegationTotalRewardsResponse._() : super();
  factory QueryDelegationTotalRewardsResponse() => create();
  factory QueryDelegationTotalRewardsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegationTotalRewardsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegationTotalRewardsResponse clone() => QueryDelegationTotalRewardsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegationTotalRewardsResponse copyWith(void Function(QueryDelegationTotalRewardsResponse) updates) => super.copyWith((message) => updates(message as QueryDelegationTotalRewardsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegationTotalRewardsResponse create() => QueryDelegationTotalRewardsResponse._();
  QueryDelegationTotalRewardsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryDelegationTotalRewardsResponse> createRepeated() => $pb.PbList<QueryDelegationTotalRewardsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegationTotalRewardsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegationTotalRewardsResponse>(create);
  static QueryDelegationTotalRewardsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$3.DelegationDelegatorReward> get rewards => $_getList(0);

  @$pb.TagNumber(2)
  $core.List<$2.DecCoin> get total => $_getList(1);
}

class QueryDelegatorValidatorsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegatorValidatorsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..hasRequiredFields = false
  ;

  QueryDelegatorValidatorsRequest._() : super();
  factory QueryDelegatorValidatorsRequest() => create();
  factory QueryDelegatorValidatorsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegatorValidatorsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegatorValidatorsRequest clone() => QueryDelegatorValidatorsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegatorValidatorsRequest copyWith(void Function(QueryDelegatorValidatorsRequest) updates) => super.copyWith((message) => updates(message as QueryDelegatorValidatorsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorValidatorsRequest create() => QueryDelegatorValidatorsRequest._();
  QueryDelegatorValidatorsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryDelegatorValidatorsRequest> createRepeated() => $pb.PbList<QueryDelegatorValidatorsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorValidatorsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegatorValidatorsRequest>(create);
  static QueryDelegatorValidatorsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);
}

class QueryDelegatorValidatorsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegatorValidatorsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pPS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'validators')
    ..hasRequiredFields = false
  ;

  QueryDelegatorValidatorsResponse._() : super();
  factory QueryDelegatorValidatorsResponse() => create();
  factory QueryDelegatorValidatorsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegatorValidatorsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegatorValidatorsResponse clone() => QueryDelegatorValidatorsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegatorValidatorsResponse copyWith(void Function(QueryDelegatorValidatorsResponse) updates) => super.copyWith((message) => updates(message as QueryDelegatorValidatorsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorValidatorsResponse create() => QueryDelegatorValidatorsResponse._();
  QueryDelegatorValidatorsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryDelegatorValidatorsResponse> createRepeated() => $pb.PbList<QueryDelegatorValidatorsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorValidatorsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegatorValidatorsResponse>(create);
  static QueryDelegatorValidatorsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.String> get validators => $_getList(0);
}

class QueryDelegatorWithdrawAddressRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegatorWithdrawAddressRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'delegatorAddress')
    ..hasRequiredFields = false
  ;

  QueryDelegatorWithdrawAddressRequest._() : super();
  factory QueryDelegatorWithdrawAddressRequest() => create();
  factory QueryDelegatorWithdrawAddressRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegatorWithdrawAddressRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegatorWithdrawAddressRequest clone() => QueryDelegatorWithdrawAddressRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegatorWithdrawAddressRequest copyWith(void Function(QueryDelegatorWithdrawAddressRequest) updates) => super.copyWith((message) => updates(message as QueryDelegatorWithdrawAddressRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorWithdrawAddressRequest create() => QueryDelegatorWithdrawAddressRequest._();
  QueryDelegatorWithdrawAddressRequest createEmptyInstance() => create();
  static $pb.PbList<QueryDelegatorWithdrawAddressRequest> createRepeated() => $pb.PbList<QueryDelegatorWithdrawAddressRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorWithdrawAddressRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegatorWithdrawAddressRequest>(create);
  static QueryDelegatorWithdrawAddressRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get delegatorAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set delegatorAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasDelegatorAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearDelegatorAddress() => clearField(1);
}

class QueryDelegatorWithdrawAddressResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryDelegatorWithdrawAddressResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'withdrawAddress')
    ..hasRequiredFields = false
  ;

  QueryDelegatorWithdrawAddressResponse._() : super();
  factory QueryDelegatorWithdrawAddressResponse() => create();
  factory QueryDelegatorWithdrawAddressResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryDelegatorWithdrawAddressResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryDelegatorWithdrawAddressResponse clone() => QueryDelegatorWithdrawAddressResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryDelegatorWithdrawAddressResponse copyWith(void Function(QueryDelegatorWithdrawAddressResponse) updates) => super.copyWith((message) => updates(message as QueryDelegatorWithdrawAddressResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorWithdrawAddressResponse create() => QueryDelegatorWithdrawAddressResponse._();
  QueryDelegatorWithdrawAddressResponse createEmptyInstance() => create();
  static $pb.PbList<QueryDelegatorWithdrawAddressResponse> createRepeated() => $pb.PbList<QueryDelegatorWithdrawAddressResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryDelegatorWithdrawAddressResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryDelegatorWithdrawAddressResponse>(create);
  static QueryDelegatorWithdrawAddressResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get withdrawAddress => $_getSZ(0);
  @$pb.TagNumber(1)
  set withdrawAddress($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasWithdrawAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearWithdrawAddress() => clearField(1);
}

class QueryCommunityPoolRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryCommunityPoolRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  QueryCommunityPoolRequest._() : super();
  factory QueryCommunityPoolRequest() => create();
  factory QueryCommunityPoolRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryCommunityPoolRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryCommunityPoolRequest clone() => QueryCommunityPoolRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryCommunityPoolRequest copyWith(void Function(QueryCommunityPoolRequest) updates) => super.copyWith((message) => updates(message as QueryCommunityPoolRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryCommunityPoolRequest create() => QueryCommunityPoolRequest._();
  QueryCommunityPoolRequest createEmptyInstance() => create();
  static $pb.PbList<QueryCommunityPoolRequest> createRepeated() => $pb.PbList<QueryCommunityPoolRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryCommunityPoolRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryCommunityPoolRequest>(create);
  static QueryCommunityPoolRequest _defaultInstance;
}

class QueryCommunityPoolResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryCommunityPoolResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'cosmos.distribution.v1beta1'), createEmptyInstance: create)
    ..pc<$2.DecCoin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pool', $pb.PbFieldType.PM, subBuilder: $2.DecCoin.create)
    ..hasRequiredFields = false
  ;

  QueryCommunityPoolResponse._() : super();
  factory QueryCommunityPoolResponse() => create();
  factory QueryCommunityPoolResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryCommunityPoolResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryCommunityPoolResponse clone() => QueryCommunityPoolResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryCommunityPoolResponse copyWith(void Function(QueryCommunityPoolResponse) updates) => super.copyWith((message) => updates(message as QueryCommunityPoolResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryCommunityPoolResponse create() => QueryCommunityPoolResponse._();
  QueryCommunityPoolResponse createEmptyInstance() => create();
  static $pb.PbList<QueryCommunityPoolResponse> createRepeated() => $pb.PbList<QueryCommunityPoolResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryCommunityPoolResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryCommunityPoolResponse>(create);
  static QueryCommunityPoolResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.DecCoin> get pool => $_getList(0);
}

