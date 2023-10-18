#!/bin/bash

# Function to extract commands from Makefile
extract_commands() {
    grep -oP '^[a-zA-Z_-]+:' $1 | sed 's/://'
}

# Function to check if command is documented in help
is_documented() {
    local cmd=$1
    local help_output=$(make -s $2) # assuming that running 'make help' or 'make {specific-help-command}' returns the help text
    echo "$help_output" | grep -q "$cmd"
}

# Compare Makefile and Makefile.prev
current_cmds=$(extract_commands Makefile)
previous_cmds=$(extract_commands Makefile.prev)

# Check for new commands and if they are documented
error=0
for cmd in $current_cmds; do
    if ! grep -q "^$cmd$" <<< "$previous_cmds"; then
        # This is a new command, check if it's documented under the help command
        if ! is_documented $cmd "help"; then # replace "help" with the specific help command if necessary
            echo "Error: New command '$cmd' added without documentation under the help command."
            error=1
        fi
    fi
done

exit $error
