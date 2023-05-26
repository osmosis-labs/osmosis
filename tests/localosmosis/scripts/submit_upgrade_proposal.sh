#!/bin/sh
set -e

# Validator mnemonic of the validator that will make the proposal and vote on it
# it should have enough voting power to pass the proposal
VALIDATOR_MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"

OSMOSIS_HOME=$HOME/.osmosisd-local/
RPC_NODE=http://localhost:26657/

# Default upgrade version
UPGRADE_VERSION=${1:-"v15"}

# Paramters
KEY=val
PROPOSAL_DEPOSIT=1600000000uosmo
TX_FEES=1000uosmo

# Define ANSI escape sequences for colors
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get chain info
get_chain_info() {
    echo
    echo "${YELLOW}Getting chain info...${NC}"
    CHAIN_ID=$(curl -s localhost:26657/status | jq -r '.result.node_info.network')
    
    ABCI_INFO=$(curl -s localhost:26657/abci_info)
    CURRENT_HEIGHT=$(echo "$ABCI_INFO" | jq -r .result.response.last_block_height)
    UPGRADE_HEIGHT=$((CURRENT_HEIGHT + 50))
    UPGRADE_INFO=""

    echo "CHAIN_ID: $CHAIN_ID"
    echo "CURRENT_HEIGHT: $CURRENT_HEIGHT"
    echo "UPGRADE_HEIGHT: $UPGRADE_HEIGHT"
    echo "UPGRADE_VERSION: $UPGRADE_VERSION"
}

# Make proposal and get proposal ID
make_proposal() {
    echo
    echo "${YELLOW}Creating software-upgrade proposal...${NC}"
    OSMOSIS_CMD="osmosisd tx gov submit-proposal software-upgrade \
        $UPGRADE_VERSION \
        --title \"$UPGRADE_VERSION Upgrade\" \
        --description \"$UPGRADE_VERSION Upgrade\" \
        --upgrade-height $UPGRADE_HEIGHT \
        --upgrade-info \"$UPGRADE_INFO\" \
        --chain-id $CHAIN_ID \
        --deposit $PROPOSAL_DEPOSIT \
        --from $KEY \
        --fees $TX_FEES \
        --keyring-backend test \
        -b block \
        --node $RPC_NODE \
        --home $OSMOSIS_HOME \
        --yes \
        -o json"

    PROPOSAL_JSON=$(eval "$OSMOSIS_CMD")
    PROPOSAL_ID=$(echo "$PROPOSAL_JSON" | jq -r '.logs[0].events[] | select(.type == "submit_proposal") | .attributes[] | select(.key == "proposal_id") | .value')
}


# Query proposal
query_proposal() {
    osmosisd q gov proposal $PROPOSAL_ID \
        --node $RPC_NODE \
        -o json | jq
}

# Vote on proposal
vote_on_proposal() {
    echo
    echo "${YELLOW}Voting on proposal $PROPOSAL_ID...${NC}"
    OSMOSIS_CMD="osmosisd tx gov vote $PROPOSAL_ID yes \
        --from $KEY \
        --chain-id $CHAIN_ID \
        --fees $TX_FEES \
        --node $RPC_NODE \
        --home $OSMOSIS_HOME \
        --yes \
        --keyring-backend test \
        -o json"

    # Execute the command and capture the output
    VOTE_OUTPUT=$(eval "$OSMOSIS_CMD")
    echo $VOTE_OUTPUT | jq

}


# Main function
main() {
    get_chain_info
    make_proposal
    query_proposal
    vote_on_proposal
}

# Run main function
main
