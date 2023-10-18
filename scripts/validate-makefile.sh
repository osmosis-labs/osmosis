#!/bin/bash

# Function to extract commands from Makefile
extract_commands() {
    grep -oP '^[a-zA-Z_-]+:' $1 | sed 's/://'
}

# Function to check if command is documented in help
is_documented() {
    local cmd=$1
    local help_cmd="$2"
    local help_output=$(make -s $help_cmd) # This captures the output of the help command
    echo "$help_output" | grep -q "$cmd"
}

# Get the Makefile from the main branch
git fetch origin main
git show origin/main:Makefile > Makefile.main

# Extract commands from both Makefiles
current_cmds=$(extract_commands Makefile)
main_cmds=$(extract_commands Makefile.main)

# Check for commands that are in the current branch but not in main
new_cmds=$(comm -23 <(echo "$current_cmds" | sort) <(echo "$main_cmds" | sort))

# Check each new command to see if it's documented
error=0
for cmd in $new_cmds; do
    if ! is_documented $cmd "help"; then # replace "help" with your actual help command if it's different
        echo "Error: New command '$cmd' added without documentation under the help command."
        error=1
    fi
done

# Clean up
rm Makefile.main

exit $error
