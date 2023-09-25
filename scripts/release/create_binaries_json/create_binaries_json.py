"""
Usage:
This script generates a JSON object containing binary download URLs and their corresponding checksums 
for a given release tag of osmosis-labs/osmosis or from a provided checksum URL.
The binary JSON is compatible with cosmovisor and with the chain registry.

You can run this script with the following commands:

❯ python create_binaries_json.py --checksums_url https://github.com/osmosis-labs/osmosis/releases/download/v16.1.1/sha256sum.txt

Output:
{
    "binaries": {
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-linux-arm64?checksum=<checksum>",
    "darwin/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-darwin-arm64?checksum=<checksum>",
    "darwin/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-darwin-amd64?checksum=<checksum>,
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-linux-amd64?checksum=><checksum>"
    }
}

Expects a checksum in the form:

<CHECKSUM>  osmosisd-<VERSION>-<OS>-<ARCH>[.tar.gz]
<CHECKSUM>  osmosisd-<VERSION>-<OS>-<ARCH>[.tar.gz]
...

Example:

f838618633c1d42f593dc33d26b25842f5900961e987fc08570bb81a062e311d  osmosisd-16.1.1-linux-amd64
fa6699a763487fe6699c8720a2a9be4e26a4f45aafaec87aa0c3aced4cbdd155  osmosisd-16.1.1-linux-amd64.tar.gz

(From: https://github.com/osmosis-labs/osmosis/releases/download/v16.1.1/sha256sum.txt)

❯ python create_binaries_json.py --tag v16.1.1

Output:
{
    "binaries": {
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-linux-arm64?checksum=<checksum>",
    "darwin/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-darwin-arm64?checksum=<checksum>",
    "darwin/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-darwin-amd64?checksum=<checksum>",
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/16.1.1/osmosisd-16.1.1-linux-amd64?checksum=><checksum>"
    }
}

Expect a checksum to be present at: 
https://github.com/osmosis-labs/osmosis/releases/download/<TAG>/sha256sum.txt
"""

import requests
import json
import argparse
import re
import sys

def validate_tag(tag):
    pattern = '^v[0-9]+.[0-9]+.[0-9]+$'
    return bool(re.match(pattern, tag))

def download_checksums(checksums_url):

    response = requests.get(checksums_url)
    if response.status_code != 200:
        raise ValueError(f"Failed to fetch sha256sum.txt. Status code: {response.status_code}")
    return response.text

def checksums_to_binaries_json(checksums):

    binaries = {}
    
    # Parse the content and create the binaries dictionary 
    for line in checksums.splitlines():
        checksum, filename = line.split('  ')

        # exclude tar.gz files
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

    binaries_json = {
        "binaries": binaries
    }

    return json.dumps(binaries_json, indent=2)

def main():

    parser = argparse.ArgumentParser(description="Create binaries json")
    parser.add_argument('--tag', metavar='tag', type=str, help='the tag to use (e.g v16.1.1)')
    parser.add_argument('--checksums_url', metavar='checksums_url', type=str, help='URL to the checksum')

    args = parser.parse_args()
    
    # Validate the tag format
    if args.tag and not validate_tag(args.tag):
        print("Error: The provided tag does not follow the 'vX.Y.Z' format.")
        sys.exit(1)

    # Ensure that only one of --tag or --checksums_url is specified
    if not bool(args.tag) ^ bool(args.checksums_url):
        parser.error("Only one of tag or --checksums_url must be specified")
        sys.exit(1)

    checksums_url = args.checksums_url if args.checksums_url else f"https://github.com/osmosis-labs/osmosis/releases/download/{args.tag}/sha256sum.txt"
    checksums = download_checksums(checksums_url)
    binaries_json = checksums_to_binaries_json(checksums)
    print(binaries_json)

if __name__ == "__main__":
    main()
