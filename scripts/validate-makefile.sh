#!/bin/bash

# Function to extract commands from Makefile
extract_commands() {
    grep -oP '^[a-zA-Z_-]+:' $1 | sed 's/://'
    # grep -oP '^[a-zA-Z_-]+:' $1 | sed 's/://'
}

# Function to check if command has help
has_help() {
    local cmd=$1
    local help_cmd="${cmd}-help"
    grep -q "^[[:space:]]*$help_cmd" $2
}

# Compare Makefile and Makefile.prev
current_cmds=$(extract_commands Makefile)
previous_cmds=$(extract_commands Makefile.prev)

# Check for new commands and if they have corresponding help
error=0
for cmd in $current_cmds; do
    if ! grep -q "^$cmd$" <<< "$previous_cmds"; then
        # This is a new command, check for help
        if ! has_help $cmd Makefile; then
            echo "Error: New command '$cmd' added without corresponding help command."
            error=1
        fi
    fi
done

exit $error
