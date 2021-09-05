import 'package:equatable/equatable.dart';

class Denom extends Equatable {
  final String text;

  const Denom(
    this.text,
  );

  @override
  List<Object> get props => [
        text,
      ];

  @override
  String toString() => text;
}
