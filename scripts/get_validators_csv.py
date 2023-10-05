import subprocess, json, csv
from typing import Dict, Tuple, List
from collections import namedtuple

# Set to 0 for current height.
block_height = 11155925
pagination_limit = 1000
validator_count = 150
INCLUDE_JAILED = False
OSMO_CONSTANT = 1_000_000

Validator = namedtuple(
    "Validator", ["moniker", "operator_address", "tokens", "commission", "jailed"]
)

# returns all validators
def get_all_validators() -> List[Validator]:
    command = f"osmosisd q staking validators --output=json --limit={pagination_limit}"
    if block_height > 0:
        command += f" --height={block_height}"
    response = get_json_cli_response(command)

    # Extract desired fields from each entry
    data = []
    for validator in response["validators"]:
        val = validator_from_json(validator)
        data.append(val)

    data.sort(key=lambda x: float(x.tokens), reverse=True)
    all_validators = data
    return all_validators

def get_json_cli_response(command):
    # Execute the shell command and capture the response
    process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
    output, _ = process.communicate()
    # Parse the response as JSON
    return json.loads(output)

def validator_from_json(obj) -> Validator:
    return Validator(
        moniker=obj["description"]["moniker"],
        operator_address=obj["operator_address"],
        tokens=obj["tokens"],
        commission=obj["commission"]["commission_rates"]["rate"],
        jailed=obj["jailed"],
    )

validators = get_all_validators()

# Extract desired fields from each entry
data = []
for validator in validators:
    if not INCLUDE_JAILED and validator.jailed:
        continue
    moniker = validator.moniker
    operator_address = validator.operator_address
    tokens = float(validator.tokens)/OSMO_CONSTANT  
    data.append([moniker, operator_address, tokens])

data = data[:validator_count]
total_stake = sum([float(x[2]) for x in data])

# add new column for percent of total stake
for i in range(len(data)):
    data[i].append(float(data[i][2]) / total_stake)

# Export data to CSV
csv_filename = "validator_data.csv"
with open(csv_filename, "w", newline="") as csvfile:
    writer = csv.writer(csvfile)
    writer.writerow(["Moniker", "Operator Address", "Tokens", "Percent vp"])  # Header row
    writer.writerows(data)

print(f"CSV file '{csv_filename}' has been created successfully.")