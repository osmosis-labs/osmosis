///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/channel.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

// ignore_for_file: UNDEFINED_SHOWN_NAME
import 'dart:core' as $core;
import 'package:protobuf/protobuf.dart' as $pb;

class State extends $pb.ProtobufEnum {
  static const State STATE_UNINITIALIZED_UNSPECIFIED = State._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'STATE_UNINITIALIZED_UNSPECIFIED');
  static const State STATE_INIT = State._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'STATE_INIT');
  static const State STATE_TRYOPEN = State._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'STATE_TRYOPEN');
  static const State STATE_OPEN = State._(3, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'STATE_OPEN');
  static const State STATE_CLOSED = State._(4, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'STATE_CLOSED');

  static const $core.List<State> values = <State> [
    STATE_UNINITIALIZED_UNSPECIFIED,
    STATE_INIT,
    STATE_TRYOPEN,
    STATE_OPEN,
    STATE_CLOSED,
  ];

  static final $core.Map<$core.int, State> _byValue = $pb.ProtobufEnum.initByValue(values);
  static State valueOf($core.int value) => _byValue[value];

  const State._($core.int v, $core.String n) : super(v, n);
}

class Order extends $pb.ProtobufEnum {
  static const Order ORDER_NONE_UNSPECIFIED = Order._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'ORDER_NONE_UNSPECIFIED');
  static const Order ORDER_UNORDERED = Order._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'ORDER_UNORDERED');
  static const Order ORDER_ORDERED = Order._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'ORDER_ORDERED');

  static const $core.List<Order> values = <Order> [
    ORDER_NONE_UNSPECIFIED,
    ORDER_UNORDERED,
    ORDER_ORDERED,
  ];

  static final $core.Map<$core.int, Order> _byValue = $pb.ProtobufEnum.initByValue(values);
  static Order valueOf($core.int value) => _byValue[value];

  const Order._($core.int v, $core.String n) : super(v, n);
}

