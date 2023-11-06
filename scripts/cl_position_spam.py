# This script is used to spam the Osmosis testnet with concentrated liquidity positions.
# It reuses sequence numbers, allowing for submitting multiple transactions within the same block
# from multiple accounts
import subprocess
import re

# min tick is -108000000, max tick is 342000000
# this moves from start tick up towards end tick
# expected price of osmo/dai is ~5000000000000 which corresponds to tick 112000000
# 50x price increase is 250_000_000_000_000 which corresponds to tick 127500000
# 50x price decrease is 100_000_000_000 which corresponds to tick 99000000
# 7750 positions
start_tick = 112000000
end_tick = 127500000
tick_width = 1000
tick_gap = 1000
sequence = 0

# env
wallet_name = "validator1"
position_max_coins = "1000uosmo,100000000stake"
chain_id = "testing"
node = "http://localhost:26657"
home = "/root/.osmosisd/validator1"
retry_limit = 50
pool_id = "1"


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


def write_to_file(start_tick_str, end_tick_str):
    with open("last_ticks.txt", "w") as file:
        file.write(f"Last start_tick_str: {start_tick_str}\n")
        file.write(f"Last end_tick_str: {end_tick_str}\n")


def run_osmosis_cmd(start_tick, end_tick, tick_width, tick_gap, sequence, retry_limit):
    last_start_tick_str = None
    last_end_tick_str = None

    for i, tick in enumerate(range(start_tick, end_tick, tick_width + tick_gap)):
        start_tick = tick
        end_tick = tick + tick_width
        
        if start_tick < 0:
            start_tick_str = f"[{start_tick}]"
        else:
            start_tick_str = str(start_tick)
        
        if end_tick < 0:
            end_tick_str = f"[{end_tick}]"
        else:
            end_tick_str = str(end_tick)
        
        cmd = [
            "osmosisd",
            "tx",
            "concentratedliquidity",
            "create-position",
            pool_id,
            start_tick_str,
            end_tick_str,
            position_max_coins,
            "0",
            "0",
            "--from",
            wallet_name,
            "--chain-id",
            chain_id,
            "--keyring-backend",
            "test",
            "--fees",
            "15000000stake",
            "--gas",
            "25000000",
            "--node",
            node,
            "--home",
            home,
            "-s",
            str(sequence),
            "-y"
        ]
    
    
        retry_count = 0
        while retry_count <= retry_limit:
            try:
                result = subprocess.run(cmd, capture_output=True, text=True)
                output = result.stdout.strip()
                stderr = result.stderr.strip()
                if check_position_creation(output):
                    print(
                        f"Position created from {start_tick_str} to {end_tick_str} successfully.")
                    sequence += 1  # Increase sequence number by 1
                    break
                else:
                    expected_sequence, current_sequence = parse_sequence_from_error(
                        output)
                    if expected_sequence is not None and current_sequence is not None:
                        if current_sequence < expected_sequence:
                            sequence = expected_sequence  # Set sequence back to expected value
                            break
                    print(
                        f"Failed to create position from {start_tick_str} to {end_tick_str} with error: {output}, {stderr}")
            except subprocess.CalledProcessError as e:
                print(
                    f"Failed to create position from {start_tick_str} to {end_tick_str} with error 2: {e.output.decode('utf-8')}, {e.stderr.decode('utf-8')}")
            
            retry_count += 1
        
        last_start_tick_str = start_tick_str
        last_end_tick_str = end_tick_str
        
        if retry_count > retry_limit:
            write_to_file(last_start_tick_str, last_end_tick_str)

run_osmosis_cmd(start_tick, end_tick, tick_width,
                tick_gap, sequence, retry_limit)