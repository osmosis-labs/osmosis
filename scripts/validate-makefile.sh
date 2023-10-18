#!/bin/bash

# Function to extract commands from Makefile
extract_commands() {
    grep -oP '^[a-zA-Z_-]+:' $1 | sed 's/://'
}

# Function to check if command is mentioned in any @echo line in the Makefile
is_documented() {
    local cmd=$1
    # This looks for any line starting with @echo and checks if it contains the command.
    grep '^	@echo' Makefile | grep -q "$cmd"
}

# Get the Makefile from the main branch
git fetch origin main
git show origin/main:Makefile > Makefile.main

# Extract commands from both Makefiles
current_cmds=$(extract_commands Makefile)
main_cmds=$(extract_commands Makefile.main)

# Check for commands that are in the current branch but not in main
new_cmds=$(comm -23 <(echo "$current_cmds" | sort) <(echo "$main_cmds" | sort))

# Check each new command to see if it's mentioned in any @echo line
error=0
for cmd in $new_cmds; do
    if ! is_documented $cmd; then
        echo "Error: New command '$cmd' added without being mentioned in any @echo line."
        error=1
    fi
done

# Clean up
rm Makefile.main

exit $error
