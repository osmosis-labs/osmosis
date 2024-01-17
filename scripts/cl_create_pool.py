# This script is used for creating a concentrated liquidity pool on the Osmosis testnet.
# With the parameters defined below.
import subprocess
import re

denom0 = "stake"
denom1 = "uosmo"
tick_spacing = "1000"
spread_factor = "0.001"

# env
wallet_name = "validator1"
chain_id = "testing"
node = "http://localhost:26657"
home = "/root/.osmosisd/validator1"


def parse_sequence_from_error(error_message):
    pattern = r"account sequence mismatch, expected (\d+), got (\d+)"
    match = re.search(pattern, error_message)
    if match:
        expected_sequence = int(match.group(1))
        current_sequence = int(match.group(2))
        return expected_sequence, current_sequence
    return None, None


def check_pool_creation(output):
    pattern = r"code: (\d+)"
    match = re.search(pattern, output)
    if match:
        code = int(match.group(1))
        return code == 0
    return False


def run_osmosis_cmd(denom0, denom1, tick_spacing, spread_factor):
        cmd = [
            "osmosisd",
            "tx",
            "concentratedliquidity",
            "create-pool",
            denom0,
            denom1,
            tick_spacing,
            spread_factor,
            "--from",
            wallet_name,
            "--chain-id",
            chain_id,
            "--keyring-backend",
            "test",
            "--fees",
            "15000000stake",
            "--gas",
            "8000000",
            "--node",
            node,
            "--home",
            home,
            "-y"
        ]
    
    
        result = subprocess.run(cmd, capture_output=True, text=True)
        output = result.stdout.strip()
        stderr = result.stderr.strip()
        if check_pool_creation(output):
            print(
                f"Pool created successfully")
        else:
            print(
                f"Failed to create pool with error: {output}, {stderr}")

run_osmosis_cmd(denom0, denom1, tick_spacing, spread_factor)