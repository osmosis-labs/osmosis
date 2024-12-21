import argparse
from string import Template
import argparse
import re
import sys

# USAGE:
#
# This script generates a Mainnet Upgrade Guide using a template. It replaces variables like current_version, upgrade_version,
# proposal_id, and upgrade_block based on the arguments provided.
# 
# Example:
# Run the script using the following command:
# python create_upgrade_guide.py --current_version=v18 --upgrade_version=v19 --proposal_id=606 --upgrade_block=11317300 --upgrade_tag=v19.0.0
#
# Arguments:
# --current_version    : The current version before upgrade (e.g., v18)
# --upgrade_version    : The version to upgrade to (e.g., v19)
# --proposal_id        : The proposal ID related to the upgrade
# --upgrade_block      : The block height at which the upgrade will occur
# --upgrade_tag        : The specific version tag for the upgrade (e.g., v19.0.0)
#
# This will read a template file and replace the variables in it to generate a complete Mainnet Upgrade Guide.


def validate_tag(tag):
    pattern = r'^v[0-9]+.[0-9]+.[0-9]+$'
    return bool(re.match(pattern, tag))


def validate_version(version):
    # Regex to match 'v' followed by a number
    pattern = r'^v\d+$'
    return bool(re.match(pattern, version))


def main():

    parser = argparse.ArgumentParser(description="Create upgrade guide from template")
    parser.add_argument('--current_version', '-c', metavar='current_version', type=str, required=True, help='Current version (e.g v1)')
    parser.add_argument('--upgrade_version', '-u', metavar='upgrade_version', type=str, required=True, help='Upgrade version (e.g v2)')
    parser.add_argument('--upgrade_tag', '-t', metavar='upgrade_tag', type=str, required=True, help='Upgrade tag (e.g v2.0.0)')
    parser.add_argument('--proposal_id', '-p', metavar='proposal_id', type=str, required=True, help='Proposal ID')
    parser.add_argument('--upgrade_block', '-b', metavar='upgrade_block', type=str, required=True, help='Upgrade block height')

    args = parser.parse_args()

    if not validate_version(args.current_version):
        print("Error: The provided current_version does not follow the 'vX' format.")
        sys.exit(1)
    
    if not validate_version(args.upgrade_version):
        print("Error: The provided upgrade_version does not follow the 'vX' format.")
        sys.exit(1)

    if not validate_tag(args.upgrade_tag):
        print("Error: The provided tag does not follow the 'vX.Y.Z' format.")
        sys.exit(1)

    # Read the template from an external file
    with open('UPGRADE_TEMPLATE.md', 'r') as f:
        markdown_template = f.read()

    # Initialize the template
    t = Template(markdown_template)

    # Substitute the variables
    # Use Template.safe_substitute() over Template.substitute()
    # This method won't throw an error for missing placeholders, making it suitable for partial replacements.
    filled_markdown = t.safe_substitute(
        CURRENT_VERSION=args.current_version,
        UPGRADE_VERSION=args.upgrade_version,
        UPGRADE_TAG=args.upgrade_tag,
        PROPOSAL_ID=args.proposal_id,
        UPGRADE_BLOCK=args.upgrade_block
    )

    print(filled_markdown)


if __name__ == "__main__":
    main()
