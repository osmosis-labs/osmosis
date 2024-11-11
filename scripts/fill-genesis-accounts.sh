#!/bin/bash

# Check if the filename was provided
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <file>"
  exit 1
fi

# Assign the file argument to a variable
file="$1"

# Check if the file exists
if [ ! -f "$file" ]; then
  echo "File not found: $file"
  exit 1
fi

# Read each line from the file and pass it to symphonyd
while IFS= read -r line; do
  symphonyd add-genesis-account "$line" 10000000note
done < "$file"
Usage