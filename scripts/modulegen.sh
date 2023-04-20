#!/usr/bin/env bash

read -p "Enter module name (lowercase letters only):" module_name

# Validate that the new module name is only lowercase letters a-z
if [[ ! "$module_name" =~ ^[a-z]+$ ]]; then
  echo "Error: The new module name can only contain lowercase letters a-z."
  exit 1
fi

go run cmd/modulegen/main.go -module_name $module_name
make proto-gen