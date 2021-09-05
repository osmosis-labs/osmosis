import 'package:cosmos_ui_components/components/template/cosmos_password_field.dart';
import 'package:cosmos_ui_components/cosmos_ui_components.dart';
import 'package:flutter/material.dart';

class PasswordSetupSheet extends StatefulWidget {
  final void Function(String) submitClicked;

  const PasswordSetupSheet({
    Key? key,
    required this.submitClicked,
  }) : super(key: key);

  @override
  _PasswordSetupSheetState createState() => _PasswordSetupSheetState();
}

class _PasswordSetupSheetState extends State<PasswordSetupSheet> {
  String? password;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: CosmosAppTheme.spacingM),
          child: CosmosPasswordField(
            onPasswordUpdated: (value) => setState(() => password = value),
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: password == null ? null : () => widget.submitClicked(password!),
        child: const Icon(Icons.arrow_forward),
      ),
    );
  }
}
