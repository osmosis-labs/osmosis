import argparse
import asyncio
import glob
import hashlib
import itertools
import os
from collections import defaultdict

import httpx
import json
from asyncio import subprocess


class Command(object):
    def __init__(self, command="osmosisd", node=None, keyring_backend=None, chain_id=None):
        self.command = command
        self.node = node or "http://localhost:26657"
        self.keyring_backend = keyring_backend or "test"
        self.chain_id = chain_id or "osmosis-1"

    def parse_arg(self, arg):
        if '"' in arg or "'" in arg:
            return [arg]
        return arg.split(" ")

    async def query(self, *args, as_json=True, print_cmd=True):
        args = ["query"] + list(args)
        return await self(*args, block=False, keyring=False, as_json=as_json, print_cmd=print_cmd)

    async def __call__(self, *args, chain_id=True, node=True, block=True, keyring=True, as_json=True, dry_run=False,
                       print_cmd=True):
        if len(args) == 0:
            raise ValueError("No arguments passed to command")

        cmd = ["osmosisd"]
        if chain_id:
            cmd += ["--chain-id", self.chain_id]
        if node:
            cmd += ["--node", self.node]
        if keyring:
            cmd += ["--keyring-backend", self.keyring_backend]
        if block:
            cmd += ["--broadcast-mode", "block"]
        if as_json:
            cmd += ["--output", "json"]
        cmd += itertools.chain.from_iterable(self.parse_arg(a) for a in args)

        if print_cmd:
            print(" ".join(cmd))
        if dry_run:
            return None, None
        proc = await subprocess.create_subprocess_exec(*cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        stdout, stderr = await proc.communicate()
        stdout, stderr = stdout.decode("utf-8"), stderr.decode("utf-8")
        if as_json:
            try:
                return json.loads(stdout), stderr
            except json.decoder.JSONDecodeError:
                pass

        return stdout, stderr


SWAPROUTER_CODE_ID = None
CROSSCHAIN_SWAPS_CODE_ID = None
SWAPROUTER_PATH = './bytecode/swaprouter.wasm'
REGISTRY_PATH = './bytecode/crosschain_registry.wasm'
CROSSCHAIN_SWAPS_PATH = './bytecode/crosschain_swaps.wasm'

ENV = "testnet"
match ENV:
    case "testnet":
        BASE_API = "https://api.testnet.osmosis.zone"
        osmosisd = Command(node="https://rpc.testnet.osmosis.zone:443", keyring_backend="test", chain_id="osmo-test-5")
        CHANNEL_PREFIX_MAP = []
        CHAIN_REGISTRY_PATH = "~/devel/chain-registry/testnets"
        # SWAPROUTER_CODE_ID = 6477
        # CROSSCHAIN_SWAPS_CODE_ID = 6478
    case "edgenet":
        BASE_API = "https://api-osmosis.imperator.co"
        osmosisd = Command(node="https://rpc.edgenet.osmosis.zone:443", keyring_backend="test", chain_id="edgenet")
        CHAIN_REGISTRY_PATH = "~/devel/chain-registry"
    case "mainnet":
        SWAPROUTER_CODE_ID = 10
        CROSSCHAIN_SWAPS_CODE_ID = 31
        BASE_API = "https://api-osmosis.imperator.co"
        osmosisd = Command(node="https://rpc.osmosis.zone:443", keyring_backend="test", chain_id="osmosis-1")
        CHAIN_REGISTRY_PATH = "~/devel/chain-registry"
    case _:
        raise ValueError("Invalid environment")

POOL_API_ENDPOINT = BASE_API + "/stream/pool/v1/all?min_liquidity=0&order_key=liquidity&order_by=desc&offset=0&limit=20"
GAS_ADJUSTMENT = "--gas auto --gas-prices 0.1uosmo --gas-adjustment 1.5 -y"
CHAIN_REGISTRY_PATH = os.path.expanduser(CHAIN_REGISTRY_PATH)


async def get_gov_addr():
    module_accounts, err = await osmosisd.query("auth module-accounts", print_cmd=False)
    gov_account = next(i for i in module_accounts["accounts"] if i["name"] == "gov")
    gov = gov_account["base_account"]["address"]
    return gov


async def get_pools():
    async with httpx.AsyncClient() as client:
        response = await client.get(POOL_API_ENDPOINT)
    pools = response.json()
    return pools['pools']


def build_swap_router_messages(pools):
    messages = []
    for pool in pools:
        pool_id = pool["pool_id"]
        if isinstance(pool["pool_tokens"], dict):
            assets = pool["pool_tokens"].values()
        else:
            assets = pool["pool_tokens"]

        for token0, token1 in itertools.combinations(assets, 2):
            messages.append(
                {"set_route": {
                    "input_denom": token0["denom"],
                    "output_denom": token1["denom"],
                    "pool_route": [{"pool_id": str(pool_id), "token_out_denom": token1["denom"]}]
                }}
            )
            messages.append(
                {"set_route": {
                    "input_denom": token1["denom"],
                    "output_denom": token0["denom"],
                    "pool_route": [{"pool_id": str(pool_id), "token_out_denom": token0["denom"]}]
                }}
            )
    return messages


async def get_channels_for_denoms(pools):
    denoms = {i['denom'] for i in itertools.chain(*[i['pool_tokens'] for i in pools])}
    channels = {}

    async def fetch_trace(denom):
        trace, err = await osmosisd.query("ibc-transfer denom-trace", denom, print_cmd=False)
        if err:
            print(err)
            return None, denom
        return trace, denom

    for trace, denom in await asyncio.gather(*[fetch_trace(denom) for denom in denoms if denom.startswith("ibc/")]):
        if trace:
            channels[denom] = trace["denom_trace"]["path"].split("/")[1]
    return channels


def get_code_id(response):
    result, err = response
    if not isinstance(result, dict) or result['code'] != 0 or result['logs'][0]['events'][1]['attributes'][1]['key'] != 'code_id':
        print(result, err)
        raise Exception("Unexpected response from wasm store")
    return result['logs'][0]['events'][1]['attributes'][1]['value']


def get_address(response):
    result, err = response
    if not isinstance(result, dict) or result['code'] != 0 or result['logs'][0]['events'][0]['attributes'][0][
        'key'] != '_contract_address':
        print(result, err)
        raise Exception("Unexpected response from wasm instantiate")
    return result['logs'][0]['events'][0]['attributes'][0]['value']


def generate_channels():
    file_paths = glob.glob(f'{CHAIN_REGISTRY_PATH}/_IBC/*')

    for file_path in file_paths:
        with open(file_path, 'r') as file:
            content = json.load(file)
            chain_1_name = content['chain_1']['chain_name']
            chain_2_name = content['chain_2']['chain_name']
            channels = content['channels']

            preferred_channel = None
            for channel in channels:
                if channel.get('tags', {}).get('status') != 'live':
                    continue

                if channel['chain_1'].get('port_id') != 'transfer' or channel['chain_2'].get('port_id') != 'transfer':
                    continue

                if channel.get('tags', {}).get('preferred') == True:
                    preferred_channel = channel
                    break

            if preferred_channel is not None:
                yield chain_1_name, chain_2_name, preferred_channel


def get_channel_links():
    result = {}

    for chain_1_name, chain_2_name, channel in generate_channels():
        channel_1_id = channel['chain_1']['channel_id']
        channel_2_id = channel['chain_2']['channel_id']

        result[(chain_1_name, chain_2_name)] = channel_1_id
        result[(chain_2_name, chain_1_name)] = channel_2_id

    return result


def get_bech32_prefixes():
    result_list = []

    for chain_dir in os.listdir(CHAIN_REGISTRY_PATH):
        chain_path = os.path.join(CHAIN_REGISTRY_PATH, chain_dir, "chain.json")

        if os.path.isfile(chain_path):
            with open(chain_path, 'r') as file:
                content = json.load(file)
                chain_name = content.get('chain_name')
                bech32_prefix = content.get('bech32_prefix')

                if chain_name and bech32_prefix:
                    result_list.append(
                        {"operation": "set", "chain_name": chain_name, "prefix": bech32_prefix}
                    )

    return result_list


def get_denom_aliases():
    result = {}

    for chain_1_name, chain_2_name, channel in generate_channels():
        if "osmosis" not in chain_1_name.lower() and "osmosis" not in chain_2_name.lower():
            continue
        osmosis_chain, other_chain = ("chain_1", chain_2_name) if 'osmosis' in chain_1_name.lower() else ("chain_2", chain_1_name)
        osmosis_channel_id = channel[osmosis_chain]['channel_id']
        osmosis_port_id = channel[osmosis_chain]['port_id']

        channel_info = f"{osmosis_port_id}/{osmosis_channel_id}"

        # Now handle the asset list for each other_chain
        assetlist_path = f"{CHAIN_REGISTRY_PATH}/{other_chain}/assetlist.json"

        if not os.path.exists(assetlist_path):
            continue

        with open(assetlist_path, 'r') as file:
            asset_content = json.load(file)
            for asset in asset_content['assets']:
                if asset.get("type_asset"):
                    continue
                base_denom = asset['base']
                # Skip non-alphanumeric aliases as contract doesn't support them. Consider supporting this in the future
                if not base_denom.isalnum():
                    continue

                ibc_hash = hashlib.sha256(f'{channel_info}/{base_denom}'.encode('utf-8')).hexdigest().upper()
                result[base_denom] = f'ibc/{ibc_hash}'

    return result


async def setup_registry(moniker, owner, gov, pools, dry_run=False):
    registry_id = get_code_id(
        await osmosisd(f"tx wasm store {REGISTRY_PATH} --from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run))
    msg = '{"owner": "%s"}' % owner
    registry_addr = get_address(await osmosisd(
        f'tx wasm instantiate {registry_id}',
        msg,
        f'--from {moniker} --admin {gov} --label registry', GAS_ADJUSTMENT, dry_run=dry_run))

    # Set channels
    msg = json.dumps({
        "modify_chain_channel_links": {"operations": [{
            "operation": "set",
            "source_chain": src,
            "destination_chain": dst,
            "channel_id": channel
        } for ((src, dst), channel) in get_channel_links().items()]}
    })
    result, err = await osmosisd(f"tx wasm execute {registry_addr}", msg,
                                 f"--from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run)

    if not result or not result.get('txhash'):
        print(err)
        raise Exception("Failed to set channels")
    else:
        print(result.get('txhash'))

    # Set Bech32 prefixes
    msg = json.dumps({
        "modify_bech32_prefixes": {"operations": get_bech32_prefixes()}
    })
    result, err = await osmosisd(f"tx wasm execute {registry_addr}", msg,
                                 f"--from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run)

    if not result or not result.get('txhash'):
        print(err)
        raise Exception("Failed to set bech32 prefixes")
    else:
        print(result.get('txhash'))

    # set denom aliases
    msg = json.dumps({
        "modify_denom_alias": {"operations": [{
            "operation": "set",
            "full_denom_path": full_denom_path,
            "alias": alias
        } for alias, full_denom_path in get_denom_aliases().items()]}
    })
    result, err = await osmosisd(f"tx wasm execute {registry_addr}", msg,
                                 f"--from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run)
    if not result or not result.get('txhash'):
        print(err)
        raise Exception("Failed to set denom aliases")
    else:
        print(result.get('txhash'))




    return registry_addr


async def setup_swaprouter(moniker, owner, gov, pools, dry_run=False):
    if SWAPROUTER_CODE_ID:
        swaprouter_id = SWAPROUTER_CODE_ID
    else:
        swaprouter_id = get_code_id(
            await osmosisd(f"tx wasm store {SWAPROUTER_PATH} --from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run))

    msg = '{"owner": "%s"}' % owner
    swaprouter_addr = get_address(await osmosisd(
        f'tx wasm instantiate {swaprouter_id}',
        msg,
        f'--from {moniker} --admin {gov} --label swaprouter', GAS_ADJUSTMENT, dry_run=dry_run))

    swap_router_messages = build_swap_router_messages(pools)
    for message in swap_router_messages:
        result, err = await osmosisd(f"tx wasm execute {swaprouter_addr}", json.dumps(message),
                                     f"--from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run)
        await asyncio.sleep(1)
        if not isinstance(result, dict) or result['code'] != 0:
            print(message)
            print(result, err)
            if ENV != 'mainnet':
                print("!!!!!!!!!!!!!!!!!!!!!")
                continue
            raise Exception("Error setting up swaprouter")

    # result, err = await osmosisd(f"tx wasm execute {swaprouter_addr}", '{"transfer_ownership": {"new_owner": "%s"}}' % gov,
    #                              f"--from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run)
    # if not isinstance(result, dict) or result['code'] != 0:
    #     print(result)
    #     raise Exception("Error transfering swaprouter ownership")
    return swaprouter_addr


async def setup_xcs(moniker, governor, swaprouter_addr, registry_addr, gov, dry_run=False):
    if CROSSCHAIN_SWAPS_CODE_ID:
        xcs_id = CROSSCHAIN_SWAPS_CODE_ID
    else:
        xcs_id = get_code_id(
            await osmosisd(f"tx wasm store {CROSSCHAIN_SWAPS_PATH} --from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run))

    msg = '{"governor": "%s", "swap_contract": "%s", "registry_contract": "%s"}' % (
        governor, swaprouter_addr, registry_addr)
    xcs_addr = get_address(await osmosisd(
        f'tx wasm instantiate {xcs_id}',
        msg,
        f'--from {moniker} --admin {gov} --label crosschain_swaps', GAS_ADJUSTMENT, dry_run=dry_run))

    return xcs_addr


async def deploy(dry_run=False, moniker="deployer", owner=None):
    (deployer, _), gov, pools = await asyncio.gather(
        osmosisd(f"keys show {moniker} -a", node=False, chain_id=False, block=False, as_json=False, print_cmd=False),
        get_gov_addr(),
        get_pools()
    )
    # normalize the deployer address
    deployer = deployer.strip()

    owner = owner or deployer
    if not owner:
        raise ValueError("owner is required")
    else:
        print(f'Owner set to {owner}')

    print(pools)
    registry_addr = await setup_registry(moniker, deployer, gov, pools, dry_run=dry_run)

    # Store the contracts
    swaprouter_addr = await setup_swaprouter(moniker, deployer, gov, pools, dry_run=dry_run)

    await setup_xcs(moniker, deployer, swaprouter_addr, registry_addr, gov, dry_run=dry_run)


def main():
    # Create an argument parser
    parser = argparse.ArgumentParser(description='Run the deploy function.')

    # Add the dry_run argument with default value False and type bool
    parser.add_argument('--dry_run', default=False, type=bool, help='Set to True for a dry run, default is False.')
    # Add the deployer argument with default value "deployer" and type str
    parser.add_argument('--deployer', default="deployer", type=str,
                        help='The name of the deployer account, default is "deployer".')
    # Add the deployer argument with default value "deployer" and type str
    parser.add_argument('--owner', default=None, type=str,
                        help='The owner account. Will default to the deployer if omitted')

    # Parse the command-line arguments
    args = parser.parse_args()

    # Call the deploy function with the dry_run argument
    asyncio.run(deploy(dry_run=args.dry_run, moniker=args.deployer, owner=args.owner))


if __name__ == "__main__":
    main()
