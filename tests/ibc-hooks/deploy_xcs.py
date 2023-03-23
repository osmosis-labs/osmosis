import argparse
import asyncio
import itertools
import shlex

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


ENV = "mainnet"
SWAPROUTER_CODE_ID = None
CROSSCHAIN_SWAPS_CODE_ID = None
match ENV:
    case "testnet":
        BASE_API = "https://api.testnet.osmosis.zone"
        osmosisd = Command(node="https://rpc-test.osmosis.zone:443", keyring_backend="test", chain_id="osmo-test-4")
        CHANNEL_PREFIX_MAP = ''
        #SWAPROUTER_CODE_ID = 6477
        #CROSSCHAIN_SWAPS_CODE_ID = 6478
    case "mainnet":
        SWAPROUTER_CODE_ID = 10
        CROSSCHAIN_SWAPS_CODE_ID = 31
        BASE_API = "https://api-osmosis.imperator.co"
        osmosisd = Command(node="https://rpc.osmosis.zone:443", keyring_backend="test", chain_id="osmosis-1")
        CHANNEL_PREFIX_MAP = '["cosmos","channel-0"],["juno","channel-42"],["axelar","channel-208"],["stars","channel-75"],["akash","channel-1"]'
    case _:
        raise ValueError("Invalid environment")

POOL_API_ENDPOINT = BASE_API + "/stream/pool/v1/all?min_liquidity=0&order_key=liquidity&order_by=desc&offset=0&limit=20"
GAS_ADJUSTMENT = "--gas auto --gas-prices 0.1uosmo --gas-adjustment 1.5 -y"


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
        for token0, token1 in itertools.combinations(pool["pool_tokens"], 2):
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
        print(result)
        raise Exception("Unexpected response from wasm store")
    return result['logs'][0]['events'][1]['attributes'][1]['value']


def get_address(response):
    result, err = response
    if not isinstance(result, dict) or result['code'] != 0 or result['logs'][0]['events'][0]['attributes'][0]['key'] != '_contract_address':
        print(result, err)
        raise Exception("Unexpected response from wasm instantiate")
    return result['logs'][0]['events'][0]['attributes'][0]['value']


async def setup_swaprouter(moniker, deployer, gov, pools, dry_run=False):
    if SWAPROUTER_CODE_ID:
        swaprouter_id = SWAPROUTER_CODE_ID
    else:
        swaprouter_id = get_code_id(
            await osmosisd(f"tx wasm store ./bytecode/swaprouter.wasm --from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run))

    # msg = '{"owner": "%s"}' % deployer
    # swaprouter_addr = get_address(await osmosisd(
    #     f'tx wasm instantiate {swaprouter_id}',
    #     msg,
    #     f'--from {moniker} --admin {gov} --label swaprouter', GAS_ADJUSTMENT, dry_run=dry_run))

    swaprouter_addr = "osmo1fy547nr4ewfc38z73ghr6x62p7eguuupm66xwk8v8rjnjyeyxdqs6gdqx7"

    swap_router_messages = build_swap_router_messages(pools)
    for message in swap_router_messages:
        if message['set_route']['pool_route'][0]['pool_id'] in ['1', '678', '712', '704', '833', '674', '907', '812']:
            continue
        result, err = await osmosisd(f"tx wasm execute {swaprouter_addr}", json.dumps(message),
                                     f"--from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run)
        await asyncio.sleep(1)
        if not isinstance(result, dict) or result['code'] != 0:
            print(message)
            print(result, err)
            if ENV == 'testnet' or True:
                print("!!!!!!!!!!!!!!!!!!!!!")
                continue
            raise Exception("Error setting up swaprouter")

    result, err = await osmosisd(f"tx wasm execute {swaprouter_addr}", '{"transfer_ownership": {"new_owner": "%s"}}' % gov,
                                 f"--from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run)
    if not isinstance(result, dict) or result['code'] != 0:
        print(result)
        raise Exception("Error transfering swaprouter ownership")
    return swaprouter_addr


async def setup_xcs(moniker, deployer, swaprouter_addr, gov, dry_run=False):
    if CROSSCHAIN_SWAPS_CODE_ID:
        xcs_id = CROSSCHAIN_SWAPS_CODE_ID
    else:
        xcs_id = get_code_id(
            await osmosisd(f"tx wasm store ./bytecode/crosschain_swaps.wasm --from {moniker}", GAS_ADJUSTMENT, dry_run=dry_run))

    msg = '{"governor": "%s", "swap_contract": "%s", "channels": [%s]}' % (deployer, swaprouter_addr, CHANNEL_PREFIX_MAP)
    xcs_addr = get_address(await osmosisd(
        f'tx wasm instantiate {xcs_id}',
        msg,
        f'--from {moniker} --admin {gov} --label crosschain_swaps', GAS_ADJUSTMENT, dry_run=dry_run))

    return xcs_addr


async def deploy(dry_run=False, moniker="deployer"):
    (deployer, _), gov, pools = await asyncio.gather(
        osmosisd(f"keys show {moniker} -a", node=False, chain_id=False, block=False, as_json=False, print_cmd=False),
        get_gov_addr(),
        get_pools()
    )
    # normalize the deployer address
    deployer = deployer.strip()

    # Store the contracts
    #swaprouter_addr = await setup_swaprouter(moniker, deployer, gov, pools, dry_run=dry_run)
    swaprouter_addr = 'osmo1fy547nr4ewfc38z73ghr6x62p7eguuupm66xwk8v8rjnjyeyxdqs6gdqx7'
    #channels = await get_channels_for_denoms(pools)
    await setup_xcs(moniker, deployer, swaprouter_addr, gov, dry_run=dry_run)



def main():
    # Create an argument parser
    parser = argparse.ArgumentParser(description='Run the deploy function.')

    # Add the dry_run argument with default value False and type bool
    parser.add_argument('--dry_run', default=False, type=bool, help='Set to True for a dry run, default is False.')
    # Add the deployer argument with default value "deployer" and type str
    parser.add_argument('--deployer', default="deployer", type=str,
                        help='The name of the deployer account, default is "deployer".')

    # Parse the command-line arguments
    args = parser.parse_args()

    # Call the deploy function with the dry_run argument
    asyncio.run(deploy(dry_run=args.dry_run, moniker=args.deployer))


if __name__ == "__main__":
    main()
