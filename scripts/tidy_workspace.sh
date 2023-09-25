#!/bin/bash

# Read each line from go.work and extract submodule path
awk '/^use / {print substr($0, 5)}' go.work | while read -r submodule; do
  # Change directory to the submodule
  if cd "$submodule"; then
    # Run go mod tidy
    go mod tidy
    if [ $? -eq 0 ]; then
      echo "Successfully ran go mod tidy in $submodule"
    else
      echo "Failed to run go mod tidy in $submodule"
    fi
    # Change back to the original directory
    cd - > /dev/null
  else
    echo "Failed to enter $submodule, skipping."
  fi
done
