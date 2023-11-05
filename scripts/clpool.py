import subprocess
import re

# min tick is -108000000, max tick is 342000000
# this moves from start tick up towards end tick
# expected price of osmo/dai is ~5000000000000 which corresponds to tick 112000000
# 50x price increase is 250_000_000_000_000 which corresponds to tick 127500000
# 50x price decrease is 100_000_000_000 which corresponds to tick 99000000
# 7750 positions
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


def check_position_creation(output):
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
        if check_position_creation(output):
            print(
                f"Pool created successfully")
        else:
            print(
                f"Failed to create pool with error: {output}, {stderr}")

run_osmosis_cmd(denom0, denom1, tick_spacing, spread_factor)