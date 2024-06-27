import subprocess
import re
import json
import time
import os
from concurrent.futures import ThreadPoolExecutor, as_completed

# The current flow of this script is as follows:
#
# We first have a list of pool IDs that have OSMO as one of the assets in the pool.
# This is to simplify the logic of determining what the non-OSMO asset is, and
# not having to worry about seeding wallets with with every possible asset.
# The idea is, each block, we swap over one more pool, until we have swapped over all pools.
# Ex. If we have 2 GAMM pools, 2 CL pools, and 1 CW pool:
# Block 1: Swap over GAMM pool 1
# Block 2: Swap over GAMM pool 1, GAMM pool 2
# Block 3: Swap over GAMM pool 1, GAMM pool 2, CL pool 1
# Block 4: Swap over GAMM pool 1, GAMM pool 2, CL pool 1, CL pool 2
# Block 5: Swap over GAMM pool 1, GAMM pool 2, CL pool 1, CL pool 2, CW pool 1
#
# The script creates and seeds as many wallets as their are pools. On the first run, it will seed all wallets.
# On subsequent runs, if the wallets exist and have sufficient balance, they will not be seeded for speed.
# The reason for using many wallets instead of incrementing the sequence number on a single wallet is because
# the single wallet must be submitted in a single threaded manner, in the order of the sequence number. With
# 2 second blocks, as we get more and more pools, the likelihood of going over the 2 second block time increases.
# With multiple wallets, we can submit transactions concurrently since there are no sequence number conflicts.
#
# Some notes:
#
# Its recommended to run this script on a machine with a good amount of CPU cores, as the script
# will submit transactions concurrently for each pool. The more cores, the faster the script will run. With a 16 vCPU machine,
# and 65 pool entries, only the last 3 blocks had a 2 block delay.
#
# You should use one of the lo-test1 - lo-test10 KEYRING_NAME keys since they are seeded with OSMO from in-place-testnet.
#
# The script works with the given five CW pools, but adding other CW pools might require some extra logic to handle their
# different query responses.
#
# Lastly, to run this script, you would do the following:
# 1. Download a mainnet snapshot
# 2. In the osmosis repo, run `make localnet-keys`
# 3. Run the in-place-testnet command (at the time of this writing, use the trigger upgrade flag as "v26" since this script is meant for a sdk v50 chain)
# 4. Hit the upgrade height and upgrade
# 5. Make SURE you `make install` the v26 (or higher) osmosisd binary, as that is what will be used to run this script
# 6. If you want to check on the txhashes that are output from the script, make sure to change your tx indexer in config.toml to "kv"
# 7. Run this script
#
# Some TODOs:
#
# 1. Delete all unsigned tx files and signed tx files in the event of an error
# 2. Figure out some better concurrency strategy for creating and signing the txs
# 3. Store all the tx hashes, and once complete, iterate over them all and ensure they are all successful and give a summary of the results

# Variables (change as needed)
KEYRING_NAME = "lo-test2"
KEYRING_BACKEND = "test"
BROADCAST_MODE = "async"
FEES = "100000uosmo"
GAS = 750000
CHAIN_ID = "localosmosis"
MIN_WALLET_BALANCE = 5000000
OSMO_GAMM_POOL_IDS = [1, 712, 704, 812, 678, 681, 796, 1057, 3, 9, 725, 832, 806, 840, 1241, 1687, 1632, 722, 584,560, 586, 5, 604, 497, 992, 799, 1244, 744, 1075, 1225] # 30 pools
OSMO_CL_POOL_IDS = [1252, 1135, 1093, 1134, 1090, 1133, 1248, 1323, 1094, 1095, 1263, 1590, 1096, 1265, 1098, 1097, 1092, 1464, 1400, 1388, 1104, 1325, 1281, 1114, 1066, 1215, 1449, 1077, 1399, 1770] # 30 pools
OSMO_CW_POOL_IDS = [1463, 1575, 1584, 1642, 1643] # 5 pools
# OSMO_GAMM_POOL_IDS = [1, 712, 704, 812, 678, 681, 796, 1057, 3, 9] # 10 pools
# OSMO_CL_POOL_IDS = [1252, 1135, 1093, 1134, 1090, 1133, 1248, 1323, 1094, 1095] # 10 pools
# OSMO_CW_POOL_IDS = [1463, 1575, 1584, 1642, 1643] # 5 pools

# Constants (should not be changed)
GAMM_POOL_TYPE = "/osmosis.gamm.v1beta1.Pool"
CL_POOL_TYPE = "/osmosis.concentratedliquidity.v1beta1.Pool"
CW_POOL_TYPE = "/osmosis.cosmwasmpool.v1beta1.CosmWasmPool"
POOL_IDS = OSMO_GAMM_POOL_IDS + OSMO_CL_POOL_IDS + OSMO_CW_POOL_IDS
QUERY_MSG = '{"pool": {}}'
QUERY_MSG_2 = '{"get_total_pool_liquidity": {}}'
WALLETS = [f"wallet_{i}" for i in range(len(POOL_IDS))]
TX_FLAGS = (
    f"--keyring-backend={KEYRING_BACKEND} -b={BROADCAST_MODE} "
    f"--fees={FEES} --gas={GAS} -y -o=json"
)


def run_command(cmd):
    result = subprocess.run(cmd, capture_output=True, text=True)
    return result.stdout.strip(), result.stderr.strip()


def create_wallets():
    print("Using ", len(WALLETS), " wallets")
    wallets_to_seed = []  # Store wallets that need to be seeded after concurrent execution

    with ThreadPoolExecutor() as executor:
        futures = {executor.submit(create_or_check_wallet, wallet): wallet for wallet in WALLETS}

        for future in as_completed(futures):
            wallet, needs_seed = future.result()
            if needs_seed:
                wallets_to_seed.append(wallet)

    # Seed wallets that need funding one per height
    if len(wallets_to_seed) > 0:
        for wallet in wallets_to_seed:
            print(f"Seeding wallet {wallet}")
            seed_wallet(wallet)

def create_or_check_wallet(wallet):
    # Check if the wallet already exists
    res_show, err_show = run_command([
        "osmosisd", "keys", "show", wallet, "--keyring-backend", KEYRING_BACKEND
    ])
    if "not a valid name or address" in err_show:
        # Wallet does not exist, create it
        run_command([
            "osmosisd", "keys", "add", wallet, "--keyring-backend", KEYRING_BACKEND
        ])
        res_show, err_show = run_command([
            "osmosisd", "keys", "show", wallet, "--keyring-backend", KEYRING_BACKEND
        ])

    # Extract the wallet address
    address_match = re.search(r'address:\s+(\S+)', res_show)
    if address_match:
        wallet_address = address_match.group(1)
    else:
        print(f"Failed to extract address for wallet {wallet}")
        return wallet, False

    # Check if the wallet has MIN_WALLET_BALANCE
    res_balance, err_balance = run_command([
        "osmosisd", "q", "bank", "balances", wallet_address, "--output", "json"
    ])
    if err_balance:
        print(f"Failed to query balance for wallet {wallet}: {err_balance}")
        return wallet, False
    balance_json = json.loads(res_balance)
    uosmo_balance = next(
        (int(balance["amount"]) for balance in balance_json.get("balances", [])
         if balance["denom"] == "uosmo"), 0)

    if uosmo_balance >= MIN_WALLET_BALANCE:
        # Wallet has sufficient balance, return False to indicate that it does not need to be seeded
        return wallet, False

    # Wallet needs to be funded, return True to indicate that it needs to be seeded
    return wallet, True

def seed_wallet(wallet):
    # Get the wallet address
    res, _ = run_command([
        "osmosisd", "keys", "show", wallet, "--keyring-backend", KEYRING_BACKEND
    ])
    address_match = re.search(r'address:\s+(\S+)', res)
    if address_match:
        wallet_address = address_match.group(1)
    else:
        print(f"Failed to extract address for wallet {wallet}")
        return

    # Send 50 OSMO to the wallet from the KEYRING_NAME wallet
    run_command([
        "osmosisd", "tx", "bank", "send", KEYRING_NAME, wallet_address,
        "50000000uosmo", f"--keyring-backend={KEYRING_BACKEND}",
        f"--chain-id={CHAIN_ID}", "-y", "--fees=1000000uosmo",
        "--gas=200000"
    ])
    wait_for_next_block()


def retrieve_status():
    status, _ = run_command(["osmosisd", "status"])
    status_json = json.loads(status)
    latest_block_height = status_json.get(
        "sync_info", {}).get("latest_block_height", 0)
    return int(latest_block_height)


def get_account_info(wallet):
    account_address, _ = run_command([
        "osmosisd", "keys", "show", wallet, "-a", "--keyring-backend",
        KEYRING_BACKEND
    ])
    account_info, _ = run_command([
        "osmosisd", "query", "auth", "account", account_address, "--output",
        "json"
    ])
    account_info_json = json.loads(account_info)
    sequence = int(
        account_info_json.get("account", {}).get("value",
                                                 {}).get("sequence", 0))
    account_number = int(
        account_info_json.get("account", {}).get("value",
                                                 {}).get("account_number", 0))
    return account_number, sequence


def wait_for_next_block():
    initial_height = retrieve_status()
    target_height = initial_height + 1
    current_height = initial_height
    while current_height < target_height:
        current_height = retrieve_status()
        time.sleep(0.05)

    print(f"Block height: {current_height}")


def get_pool(pool_id):
    response, _ = run_command(
        ["osmosisd", "q", "poolmanager", "pool", str(pool_id), "--output", "json"])
    response_json = json.loads(response)
    pool_type = response_json.get("pool", {}).get("@type", "")
    return pool_type, response


def get_non_osmo_pool_asset(pool_id):
    pool_type, response = get_pool(pool_id)
    response_json = json.loads(response)

    if pool_type == GAMM_POOL_TYPE:
        denom = next((asset["token"]["denom"]
                      for asset in response_json["pool"]["pool_assets"]
                      if asset["token"]["denom"] != "uosmo"), None)
        return denom
    elif pool_type == CL_POOL_TYPE:
        token0 = response_json["pool"]["token0"]
        token1 = response_json["pool"]["token1"]
        return token0 if token0 != "uosmo" else token1 if token1 != "uosmo" else None
    elif pool_type == CW_POOL_TYPE:
        contract_address = response_json["pool"]["contract_address"]
        # print("contract_address", contract_address)
        cw_response, cw_err = run_command([
            "osmosisd", "query", "wasm", "contract-state", "smart",
            contract_address, QUERY_MSG, "-o", "json"
        ])
        # print("cw_resp", cw_response)
        if "Error parsing into" in cw_err:
            cw_response, _ = run_command([
                "osmosisd", "query", "wasm", "contract-state", "smart",
                contract_address, QUERY_MSG_2, "-o", "json"
            ])
            # print("cw_resp 2", cw_response)
        cw_response_json = json.loads(cw_response)
        denom = next(
            (asset["denom"]
             for asset in cw_response_json["data"]["total_pool_liquidity"]
             if asset["denom"] != "uosmo"), None)
        return denom
    else:
        return None


# Initialize a dictionary to store sequence numbers for each wallet
wallet_sequences = {}


def initialize_wallet_sequences():
    for wallet in WALLETS:
        _, sequence = get_account_info(wallet)
        wallet_sequences[wallet] = sequence


def generate_and_sign_tx(pool_id, wallet, tx_number):
    non_osmo_denom = get_non_osmo_pool_asset(pool_id)
    account_number, sequence = get_account_info(wallet)

    # Use and increment the sequence number from the dictionary
    sequence = wallet_sequences[wallet]
    wallet_sequences[wallet] += 1

    # Generate the unsigned transaction
    unsigned_tx_cmd = [
        "osmosisd", "tx", "poolmanager", "swap-exact-amount-in", "100000uosmo",
        "1", "--swap-route-pool-ids",
        str(pool_id), "--swap-route-denoms", non_osmo_denom, *TX_FLAGS.split(),
        "-s",
        str(sequence), "--account-number",
        str(account_number), "--generate-only", "--offline", "--from", wallet
    ]
    unsigned_tx, _ = run_command(unsigned_tx_cmd)

    # Save the unsigned transaction to a file
    unsigned_tx_file = f"unsigned_tx_{pool_id}_{wallet}_{tx_number}.json"
    with open(unsigned_tx_file, "w") as f:
        f.write(unsigned_tx)

    # Sign the transaction
    signed_tx_file = f"signed_tx_{pool_id}_{wallet}_{tx_number}.json"
    sign_cmd = [
        "osmosisd", "tx", "sign", unsigned_tx_file, "--from", wallet,
        "--keyring-backend", KEYRING_BACKEND, "--output-document",
        signed_tx_file, "--account-number",
        str(account_number), "-s",
        str(sequence), "--chain-id", CHAIN_ID, "--offline"
    ]
    print(f"Signed tx file: {signed_tx_file}")
    run_command(sign_cmd)

    # Remove the unsigned transaction file
    if os.path.exists(unsigned_tx_file):
        os.remove(unsigned_tx_file)

    return signed_tx_file


def broadcast_tx(signed_tx_file):
    broadcast_cmd = [
        "osmosisd", "tx", "broadcast", signed_tx_file, "--output", "json"
    ]
    # print(broadcast_cmd)
    result, error = run_command(broadcast_cmd)
    return result, error, signed_tx_file

def generate_and_sign_tx_for_wallet(wallet, pool_id, tx_count):
    for tx_number in range(1, tx_count + 1):
        generate_and_sign_tx(pool_id, wallet, tx_number)

#
# Script execution starts here
#

# Create wallets and initialize sequences
create_wallets()
wait_for_next_block()
initialize_wallet_sequences()

# Generate and sign transactions
print(f"Generating and signing transactions prior to broadcasting...")
wallet_to_pool_tx_count = {
    f"wallet_{i}": (POOL_IDS[i], len(POOL_IDS) - i)
    for i in range(len(POOL_IDS))
}
with ThreadPoolExecutor() as executor:
    futures = {executor.submit(generate_and_sign_tx_for_wallet, wallet, pool_id, tx_count): (wallet, pool_id, tx_count)
               for wallet, (pool_id, tx_count) in wallet_to_pool_tx_count.items()}
    for future in as_completed(futures):
        future.result()

    print("Completed generating and signing transactions")

# Broadcast transactions in rounds (one round per block)
for round_number in range(1, len(POOL_IDS) + 1):
    txhashes = []
    files_to_remove = []

    print(f"Broadcasting transactions for round {round_number}")

    # Broadcast transactions concurrently
    with ThreadPoolExecutor() as executor:
        futures = []
        for wallet_index in range(round_number):
            wallet = f"wallet_{wallet_index}"
            pool_id = POOL_IDS[wallet_index]
            tx_number = round_number - wallet_index
            if tx_number > 0:
                signed_tx_file = f"signed_tx_{pool_id}_{wallet}_{tx_number}.json"
                futures.append(executor.submit(broadcast_tx, signed_tx_file))

        # Collect results
        for future in as_completed(futures):
            result, error, signed_tx_file = future.result()
            if '"code":0' in result:
                txhash = json.loads(result).get("txhash")
                txhashes.append(txhash)
            else:
                print(f"Failed: {result}, {error},{signed_tx_file}")
            files_to_remove.append(signed_tx_file)

    print(f"Round {round_number} txhashes: {txhashes}")
    print(f"Count: {len(txhashes)}")

    # Remove files after broadcasting
    for file in files_to_remove:
        if os.path.exists(file):
            os.remove(file)

    # Wait for the next block after all transactions in the round are done
    wait_for_next_block()
