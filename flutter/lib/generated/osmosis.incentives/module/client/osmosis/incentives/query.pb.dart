///
//  Generated code. Do not modify.
//  source: osmosis/incentives/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../cosmos/base/v1beta1/coin.pb.dart' as $4;
import 'gauge.pb.dart' as $7;
import '../../cosmos/base/query/v1beta1/pagination.pb.dart' as $9;
import '../../google/protobuf/duration.pb.dart' as $2;

class ModuleToDistributeCoinsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleToDistributeCoinsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  ModuleToDistributeCoinsRequest._() : super();
  factory ModuleToDistributeCoinsRequest() => create();
  factory ModuleToDistributeCoinsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleToDistributeCoinsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleToDistributeCoinsRequest clone() => ModuleToDistributeCoinsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleToDistributeCoinsRequest copyWith(void Function(ModuleToDistributeCoinsRequest) updates) => super.copyWith((message) => updates(message as ModuleToDistributeCoinsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleToDistributeCoinsRequest create() => ModuleToDistributeCoinsRequest._();
  ModuleToDistributeCoinsRequest createEmptyInstance() => create();
  static $pb.PbList<ModuleToDistributeCoinsRequest> createRepeated() => $pb.PbList<ModuleToDistributeCoinsRequest>();
  @$core.pragma('dart2js:noInline')
  static ModuleToDistributeCoinsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleToDistributeCoinsRequest>(create);
  static ModuleToDistributeCoinsRequest _defaultInstance;
}

class ModuleToDistributeCoinsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleToDistributeCoinsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  ModuleToDistributeCoinsResponse._() : super();
  factory ModuleToDistributeCoinsResponse() => create();
  factory ModuleToDistributeCoinsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleToDistributeCoinsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleToDistributeCoinsResponse clone() => ModuleToDistributeCoinsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleToDistributeCoinsResponse copyWith(void Function(ModuleToDistributeCoinsResponse) updates) => super.copyWith((message) => updates(message as ModuleToDistributeCoinsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleToDistributeCoinsResponse create() => ModuleToDistributeCoinsResponse._();
  ModuleToDistributeCoinsResponse createEmptyInstance() => create();
  static $pb.PbList<ModuleToDistributeCoinsResponse> createRepeated() => $pb.PbList<ModuleToDistributeCoinsResponse>();
  @$core.pragma('dart2js:noInline')
  static ModuleToDistributeCoinsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleToDistributeCoinsResponse>(create);
  static ModuleToDistributeCoinsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class ModuleDistributedCoinsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleDistributedCoinsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  ModuleDistributedCoinsRequest._() : super();
  factory ModuleDistributedCoinsRequest() => create();
  factory ModuleDistributedCoinsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleDistributedCoinsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleDistributedCoinsRequest clone() => ModuleDistributedCoinsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleDistributedCoinsRequest copyWith(void Function(ModuleDistributedCoinsRequest) updates) => super.copyWith((message) => updates(message as ModuleDistributedCoinsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleDistributedCoinsRequest create() => ModuleDistributedCoinsRequest._();
  ModuleDistributedCoinsRequest createEmptyInstance() => create();
  static $pb.PbList<ModuleDistributedCoinsRequest> createRepeated() => $pb.PbList<ModuleDistributedCoinsRequest>();
  @$core.pragma('dart2js:noInline')
  static ModuleDistributedCoinsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleDistributedCoinsRequest>(create);
  static ModuleDistributedCoinsRequest _defaultInstance;
}

class ModuleDistributedCoinsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ModuleDistributedCoinsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  ModuleDistributedCoinsResponse._() : super();
  factory ModuleDistributedCoinsResponse() => create();
  factory ModuleDistributedCoinsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ModuleDistributedCoinsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ModuleDistributedCoinsResponse clone() => ModuleDistributedCoinsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ModuleDistributedCoinsResponse copyWith(void Function(ModuleDistributedCoinsResponse) updates) => super.copyWith((message) => updates(message as ModuleDistributedCoinsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ModuleDistributedCoinsResponse create() => ModuleDistributedCoinsResponse._();
  ModuleDistributedCoinsResponse createEmptyInstance() => create();
  static $pb.PbList<ModuleDistributedCoinsResponse> createRepeated() => $pb.PbList<ModuleDistributedCoinsResponse>();
  @$core.pragma('dart2js:noInline')
  static ModuleDistributedCoinsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModuleDistributedCoinsResponse>(create);
  static ModuleDistributedCoinsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class GaugeByIDRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GaugeByIDRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'id', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  GaugeByIDRequest._() : super();
  factory GaugeByIDRequest() => create();
  factory GaugeByIDRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GaugeByIDRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GaugeByIDRequest clone() => GaugeByIDRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GaugeByIDRequest copyWith(void Function(GaugeByIDRequest) updates) => super.copyWith((message) => updates(message as GaugeByIDRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GaugeByIDRequest create() => GaugeByIDRequest._();
  GaugeByIDRequest createEmptyInstance() => create();
  static $pb.PbList<GaugeByIDRequest> createRepeated() => $pb.PbList<GaugeByIDRequest>();
  @$core.pragma('dart2js:noInline')
  static GaugeByIDRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GaugeByIDRequest>(create);
  static GaugeByIDRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get id => $_getI64(0);
  @$pb.TagNumber(1)
  set id($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);
}

class GaugeByIDResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GaugeByIDResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..aOM<$7.Gauge>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'gauge', subBuilder: $7.Gauge.create)
    ..hasRequiredFields = false
  ;

  GaugeByIDResponse._() : super();
  factory GaugeByIDResponse() => create();
  factory GaugeByIDResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GaugeByIDResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GaugeByIDResponse clone() => GaugeByIDResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GaugeByIDResponse copyWith(void Function(GaugeByIDResponse) updates) => super.copyWith((message) => updates(message as GaugeByIDResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GaugeByIDResponse create() => GaugeByIDResponse._();
  GaugeByIDResponse createEmptyInstance() => create();
  static $pb.PbList<GaugeByIDResponse> createRepeated() => $pb.PbList<GaugeByIDResponse>();
  @$core.pragma('dart2js:noInline')
  static GaugeByIDResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GaugeByIDResponse>(create);
  static GaugeByIDResponse _defaultInstance;

  @$pb.TagNumber(1)
  $7.Gauge get gauge => $_getN(0);
  @$pb.TagNumber(1)
  set gauge($7.Gauge v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasGauge() => $_has(0);
  @$pb.TagNumber(1)
  void clearGauge() => clearField(1);
  @$pb.TagNumber(1)
  $7.Gauge ensureGauge() => $_ensure(0);
}

class GaugesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GaugesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..aOM<$9.PageRequest>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $9.PageRequest.create)
    ..hasRequiredFields = false
  ;

  GaugesRequest._() : super();
  factory GaugesRequest() => create();
  factory GaugesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GaugesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GaugesRequest clone() => GaugesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GaugesRequest copyWith(void Function(GaugesRequest) updates) => super.copyWith((message) => updates(message as GaugesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GaugesRequest create() => GaugesRequest._();
  GaugesRequest createEmptyInstance() => create();
  static $pb.PbList<GaugesRequest> createRepeated() => $pb.PbList<GaugesRequest>();
  @$core.pragma('dart2js:noInline')
  static GaugesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GaugesRequest>(create);
  static GaugesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $9.PageRequest get pagination => $_getN(0);
  @$pb.TagNumber(1)
  set pagination($9.PageRequest v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPagination() => $_has(0);
  @$pb.TagNumber(1)
  void clearPagination() => clearField(1);
  @$pb.TagNumber(1)
  $9.PageRequest ensurePagination() => $_ensure(0);
}

class GaugesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'GaugesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$7.Gauge>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.PM, subBuilder: $7.Gauge.create)
    ..aOM<$9.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $9.PageResponse.create)
    ..hasRequiredFields = false
  ;

  GaugesResponse._() : super();
  factory GaugesResponse() => create();
  factory GaugesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GaugesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GaugesResponse clone() => GaugesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GaugesResponse copyWith(void Function(GaugesResponse) updates) => super.copyWith((message) => updates(message as GaugesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static GaugesResponse create() => GaugesResponse._();
  GaugesResponse createEmptyInstance() => create();
  static $pb.PbList<GaugesResponse> createRepeated() => $pb.PbList<GaugesResponse>();
  @$core.pragma('dart2js:noInline')
  static GaugesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GaugesResponse>(create);
  static GaugesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$7.Gauge> get data => $_getList(0);

  @$pb.TagNumber(2)
  $9.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($9.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $9.PageResponse ensurePagination() => $_ensure(1);
}

class ActiveGaugesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ActiveGaugesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..aOM<$9.PageRequest>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $9.PageRequest.create)
    ..hasRequiredFields = false
  ;

  ActiveGaugesRequest._() : super();
  factory ActiveGaugesRequest() => create();
  factory ActiveGaugesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ActiveGaugesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ActiveGaugesRequest clone() => ActiveGaugesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ActiveGaugesRequest copyWith(void Function(ActiveGaugesRequest) updates) => super.copyWith((message) => updates(message as ActiveGaugesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ActiveGaugesRequest create() => ActiveGaugesRequest._();
  ActiveGaugesRequest createEmptyInstance() => create();
  static $pb.PbList<ActiveGaugesRequest> createRepeated() => $pb.PbList<ActiveGaugesRequest>();
  @$core.pragma('dart2js:noInline')
  static ActiveGaugesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ActiveGaugesRequest>(create);
  static ActiveGaugesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $9.PageRequest get pagination => $_getN(0);
  @$pb.TagNumber(1)
  set pagination($9.PageRequest v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPagination() => $_has(0);
  @$pb.TagNumber(1)
  void clearPagination() => clearField(1);
  @$pb.TagNumber(1)
  $9.PageRequest ensurePagination() => $_ensure(0);
}

class ActiveGaugesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ActiveGaugesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$7.Gauge>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.PM, subBuilder: $7.Gauge.create)
    ..aOM<$9.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $9.PageResponse.create)
    ..hasRequiredFields = false
  ;

  ActiveGaugesResponse._() : super();
  factory ActiveGaugesResponse() => create();
  factory ActiveGaugesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ActiveGaugesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ActiveGaugesResponse clone() => ActiveGaugesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ActiveGaugesResponse copyWith(void Function(ActiveGaugesResponse) updates) => super.copyWith((message) => updates(message as ActiveGaugesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static ActiveGaugesResponse create() => ActiveGaugesResponse._();
  ActiveGaugesResponse createEmptyInstance() => create();
  static $pb.PbList<ActiveGaugesResponse> createRepeated() => $pb.PbList<ActiveGaugesResponse>();
  @$core.pragma('dart2js:noInline')
  static ActiveGaugesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ActiveGaugesResponse>(create);
  static ActiveGaugesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$7.Gauge> get data => $_getList(0);

  @$pb.TagNumber(2)
  $9.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($9.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $9.PageResponse ensurePagination() => $_ensure(1);
}

class UpcomingGaugesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'UpcomingGaugesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..aOM<$9.PageRequest>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $9.PageRequest.create)
    ..hasRequiredFields = false
  ;

  UpcomingGaugesRequest._() : super();
  factory UpcomingGaugesRequest() => create();
  factory UpcomingGaugesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpcomingGaugesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpcomingGaugesRequest clone() => UpcomingGaugesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpcomingGaugesRequest copyWith(void Function(UpcomingGaugesRequest) updates) => super.copyWith((message) => updates(message as UpcomingGaugesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static UpcomingGaugesRequest create() => UpcomingGaugesRequest._();
  UpcomingGaugesRequest createEmptyInstance() => create();
  static $pb.PbList<UpcomingGaugesRequest> createRepeated() => $pb.PbList<UpcomingGaugesRequest>();
  @$core.pragma('dart2js:noInline')
  static UpcomingGaugesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpcomingGaugesRequest>(create);
  static UpcomingGaugesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $9.PageRequest get pagination => $_getN(0);
  @$pb.TagNumber(1)
  set pagination($9.PageRequest v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPagination() => $_has(0);
  @$pb.TagNumber(1)
  void clearPagination() => clearField(1);
  @$pb.TagNumber(1)
  $9.PageRequest ensurePagination() => $_ensure(0);
}

class UpcomingGaugesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'UpcomingGaugesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$7.Gauge>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'data', $pb.PbFieldType.PM, subBuilder: $7.Gauge.create)
    ..aOM<$9.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $9.PageResponse.create)
    ..hasRequiredFields = false
  ;

  UpcomingGaugesResponse._() : super();
  factory UpcomingGaugesResponse() => create();
  factory UpcomingGaugesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UpcomingGaugesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UpcomingGaugesResponse clone() => UpcomingGaugesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UpcomingGaugesResponse copyWith(void Function(UpcomingGaugesResponse) updates) => super.copyWith((message) => updates(message as UpcomingGaugesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static UpcomingGaugesResponse create() => UpcomingGaugesResponse._();
  UpcomingGaugesResponse createEmptyInstance() => create();
  static $pb.PbList<UpcomingGaugesResponse> createRepeated() => $pb.PbList<UpcomingGaugesResponse>();
  @$core.pragma('dart2js:noInline')
  static UpcomingGaugesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UpcomingGaugesResponse>(create);
  static UpcomingGaugesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$7.Gauge> get data => $_getList(0);

  @$pb.TagNumber(2)
  $9.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($9.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $9.PageResponse ensurePagination() => $_ensure(1);
}

class RewardsEstRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RewardsEstRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'owner')
    ..p<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lockIds', $pb.PbFieldType.PU6)
    ..aInt64(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'endEpoch')
    ..hasRequiredFields = false
  ;

  RewardsEstRequest._() : super();
  factory RewardsEstRequest() => create();
  factory RewardsEstRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RewardsEstRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RewardsEstRequest clone() => RewardsEstRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RewardsEstRequest copyWith(void Function(RewardsEstRequest) updates) => super.copyWith((message) => updates(message as RewardsEstRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RewardsEstRequest create() => RewardsEstRequest._();
  RewardsEstRequest createEmptyInstance() => create();
  static $pb.PbList<RewardsEstRequest> createRepeated() => $pb.PbList<RewardsEstRequest>();
  @$core.pragma('dart2js:noInline')
  static RewardsEstRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RewardsEstRequest>(create);
  static RewardsEstRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get owner => $_getSZ(0);
  @$pb.TagNumber(1)
  set owner($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOwner() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwner() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$fixnum.Int64> get lockIds => $_getList(1);

  @$pb.TagNumber(3)
  $fixnum.Int64 get endEpoch => $_getI64(2);
  @$pb.TagNumber(3)
  set endEpoch($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasEndEpoch() => $_has(2);
  @$pb.TagNumber(3)
  void clearEndEpoch() => clearField(3);
}

class RewardsEstResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'RewardsEstResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$4.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'coins', $pb.PbFieldType.PM, subBuilder: $4.Coin.create)
    ..hasRequiredFields = false
  ;

  RewardsEstResponse._() : super();
  factory RewardsEstResponse() => create();
  factory RewardsEstResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RewardsEstResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RewardsEstResponse clone() => RewardsEstResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RewardsEstResponse copyWith(void Function(RewardsEstResponse) updates) => super.copyWith((message) => updates(message as RewardsEstResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static RewardsEstResponse create() => RewardsEstResponse._();
  RewardsEstResponse createEmptyInstance() => create();
  static $pb.PbList<RewardsEstResponse> createRepeated() => $pb.PbList<RewardsEstResponse>();
  @$core.pragma('dart2js:noInline')
  static RewardsEstResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RewardsEstResponse>(create);
  static RewardsEstResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.Coin> get coins => $_getList(0);
}

class QueryLockableDurationsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryLockableDurationsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  QueryLockableDurationsRequest._() : super();
  factory QueryLockableDurationsRequest() => create();
  factory QueryLockableDurationsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryLockableDurationsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryLockableDurationsRequest clone() => QueryLockableDurationsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryLockableDurationsRequest copyWith(void Function(QueryLockableDurationsRequest) updates) => super.copyWith((message) => updates(message as QueryLockableDurationsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryLockableDurationsRequest create() => QueryLockableDurationsRequest._();
  QueryLockableDurationsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryLockableDurationsRequest> createRepeated() => $pb.PbList<QueryLockableDurationsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryLockableDurationsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryLockableDurationsRequest>(create);
  static QueryLockableDurationsRequest _defaultInstance;
}

class QueryLockableDurationsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryLockableDurationsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.incentives'), createEmptyInstance: create)
    ..pc<$2.Duration>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'lockableDurations', $pb.PbFieldType.PM, subBuilder: $2.Duration.create)
    ..hasRequiredFields = false
  ;

  QueryLockableDurationsResponse._() : super();
  factory QueryLockableDurationsResponse() => create();
  factory QueryLockableDurationsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryLockableDurationsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryLockableDurationsResponse clone() => QueryLockableDurationsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryLockableDurationsResponse copyWith(void Function(QueryLockableDurationsResponse) updates) => super.copyWith((message) => updates(message as QueryLockableDurationsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryLockableDurationsResponse create() => QueryLockableDurationsResponse._();
  QueryLockableDurationsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryLockableDurationsResponse> createRepeated() => $pb.PbList<QueryLockableDurationsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryLockableDurationsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryLockableDurationsResponse>(create);
  static QueryLockableDurationsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.Duration> get lockableDurations => $_getList(0);
}

