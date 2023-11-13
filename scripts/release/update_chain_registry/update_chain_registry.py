# Usage: update_chain_registry.py [OPTIONS]

# Description:
# This script fetches the current chain-registry entry for Osmosis and returns an updated registry 
# with the new version and height information.
#
# The script fetches 
# - the checksums from the osmosis release page
# - package informations from the `go.mod` in osmosis repository 

# Options:
#   --upgrade_version VERSION  Required. The version tag for the upgrade (e.g., v19.0.0).
#   --upgrade_height HEIGHT   Required. The blockchain height at which the upgrade will happen (e.g., 11317300).
#   --debug                   Optional. Enable debug mode for more verbose output. Defaults to False.

# Example:
#   python update_chain_registry.py --upgrade_version v19.0.0 --upgrade_height 11317300

import requests
import json
import argparse
import sys

from utils.versions import compare_versions, same_major, validate_tag
from utils.go_mod import fetch_go_mod_from_tag, get_package_version

chain_json_url = "https://raw.githubusercontent.com/osmosis-labs/osmosis/main/chain.schema.json"
DEBUG = False

def fetch_data(url, url_type):
    try:
        response = requests.get(url)
        response.raise_for_status()
        if url_type == "json":
            return response.json()
        elif url_type == "schema":
            return json.loads(response.text)
        elif url_type == "text":
            return response.text
        else:
            raise ValueError("Invalid url_type. Use 'json' / 'schema' / 'text'")
    except requests.exceptions.RequestException as e:
        raise Exception(f"An error occurred while fetching data from {url}: {e}")
    except json.JSONDecodeError as e:
        raise Exception(f"Failed to parse JSON data from {url}: {e}")


def download_checksums(checksums_url):

    response = requests.get(checksums_url)
    if response.status_code != 200:
        raise ValueError(f"Failed to fetch sha256sum.txt. Status code: {response.status_code}")
    return response.text


def checksums_to_binaries_json(checksums):

    binaries = {}
    for line in checksums.splitlines():
        checksum, filename = line.split('  ')

        if not filename.endswith('.tar.gz') and filename.startswith('osmosisd'):
            try:
                _, tag, platform, arch = filename.split('-')
            except ValueError:
                print(f"Error: Expected binary name in the form: osmosisd-X.Y.Z-platform-architecture, but got {filename}")
                sys.exit(1)
            _, tag, platform, arch,  = filename.split('-')
            # exclude universal binaries and windows binaries
            if arch == 'all' or platform == 'windows':
                continue

            binaries[f"{platform}/{arch}"] = f"https://github.com/osmosis-labs/osmosis/releases/download/v{tag}/{filename}?checksum=sha256:{checksum}"

    return {
        "binaries": binaries
    }


def create_version_info(version, height):

    # Update packages versions
    go_mod = fetch_go_mod_from_tag(version)

    cosmos_sdk_version = get_package_version(go_mod, "github.com/cosmos/cosmos-sdk")
    cosmwasm_version = get_package_version(go_mod, "github.com/CosmWasm/wasmd")
    tendermint_version = get_package_version(go_mod, "github.com/cometbft/cometbft")
    ibc_go_version = get_package_version(go_mod, "github.com/cosmos/ibc-go/v7")

    if DEBUG:
        print(f"Cosmos SDK version  {cosmos_sdk_version}")
        print(f"CosmWasm version    {cosmwasm_version}")
        print(f"Tendermint version  {tendermint_version}")
        print(f"IBC Go version      {ibc_go_version}")

    version_info = {
        "name": version.split('.')[0],
        "tag" : version,
        "height": int(height),
        "recommended_version": version,
        "compatible_versions": [
            version
        ],
        "cosmos_sdk_version": cosmos_sdk_version,
        "consensus": {
            "type": "tendermint",
            "version": tendermint_version
        },
        "cosmwasm_version":cosmwasm_version,
        "cosmwasm_enabled": True,
        "ibc_go_version": ibc_go_version,
        "ics_enabled": [
            "ics20-1"
        ],
    }
    # Read binaries from the release sha256sum.txt
    checksums_url = f"https://github.com/osmosis-labs/osmosis/releases/download/{version}/sha256sum.txt"
    checksums = fetch_data(checksums_url, 'text')

    binaries_json = checksums_to_binaries_json(checksums)
    version_info['binaries'] = binaries_json['binaries']

    return version_info


def update_codebase(codebase, version, height):

    global DEBUG

    curr_recommended_version = codebase["recommended_version"]
    curr_compatible_versions = codebase["compatible_versions"]

    if DEBUG:
        print(f"New version: {version}")
        print(f"Current version: {curr_recommended_version}")
        print(f"Current compatible versions: {curr_compatible_versions}")

    if compare_versions(version, curr_recommended_version) < 0:
        if DEBUG:
            # I should still handle this case as it could be
            # a minor release for a previous major verson
            print("New version is older than recommended version")
            print("Skipping update")
        return
    elif compare_versions(version, curr_recommended_version) == 0:
        if DEBUG:
            print("New version is equal than recommended version")
            print("Skipping update")
        return
    else:

        if same_major(version, curr_recommended_version):
            if DEBUG:
                print(f"Minor release from {version} to {curr_recommended_version}")

            version_info = create_version_info(version, height)
            version_info["compatible_versions"].extend(codebase["compatible_versions"])

            # Replace minor version
            for idx, codebase_version in enumerate(codebase["versions"]):
                if codebase_version['name'] == version.split('.')[0]: # same major version
                    codebase["versions"][idx] = version_info
                    break

            # Since it's a latest release, I have also to update the top info
            codebase["recommended_version"] = version_info["recommended_version"]
            codebase["compatible_versions"] = version_info["compatible_versions"]
            codebase["binaries"] = version_info["binaries"]
            codebase["cosmos_sdk_version"] = version_info["cosmos_sdk_version"]
            codebase["consensus"] = version_info["consensus"]
            codebase["cosmwasm_version"] = version_info["cosmwasm_version"]
            codebase["ibc_go_version"] = version_info["ibc_go_version"]

        else:
            if DEBUG:
                print(f"Major release from {version} to {curr_recommended_version}")

            version_info = create_version_info(version, height)

            # Add new version to versions list
            codebase["versions"].append(version_info)
            
            # Update top level info
            codebase["recommended_version"] = version_info["recommended_version"]
            codebase["compatible_versions"] = version_info["compatible_versions"]
            codebase["binaries"] = version_info["binaries"]
            codebase["cosmos_sdk_version"] = version_info["cosmos_sdk_version"]
            codebase["consensus"] = version_info["consensus"]
            codebase["cosmwasm_version"] = version_info["cosmwasm_version"]
            codebase["ibc_go_version"] = version_info["ibc_go_version"]

            # Update previous release "next_version_name" to current one
            codebase["versions"][-2]["next_version_name"] = version_info["name"]

    return codebase

def main():

    global DEBUG

    parser = argparse.ArgumentParser(description="Create binaries json")
    parser.add_argument('--upgrade_version', required=True, type=str, help='The upgrade tag to use (e.g v19.0.0)')
    # TODO: Upgrade height is required only with new major release
    parser.add_argument('--upgrade_height', required=True, type=int, help='The height of the upgrade (e.g. 10000000)')
    parser.add_argument('--debug', action='store_true', default=False)

    args = parser.parse_args()

    # Validate the tag format
    if args.upgrade_version and not validate_tag(args.upgrade_version):
        print("Error: The provided version does not follow the 'vX.Y.Z' format.")
        sys.exit(1)

    DEBUG = args.debug

    try:
        if DEBUG:
            print(f"Fetching chain json from {chain_json_url}...")
        chain_json = fetch_data(chain_json_url, "json")

        updated_codebase = update_codebase(chain_json["codebase"], args.upgrade_version, args.upgrade_height)

        if updated_codebase:
            chain_json["codebase"] = updated_codebase
            print(json.dumps(chain_json, indent=2))
        else:
            print("Couldn't update codebase")

    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == "__main__":
    main()
