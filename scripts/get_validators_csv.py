import subprocess, json, csv

pagination_limit = 1000
num_validators = 150
command = "osmosisd q staking validators --output=json --limit=1000"

# Execute the shell command and capture the response
process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
output, _ = process.communicate()

# Parse the response as JSON
response = json.loads(output)

# Extract desired fields from each entry
data = []
for validator in response["validators"]:
    moniker = validator["description"]["moniker"]
    operator_address = validator["operator_address"]
    tokens = validator["tokens"]
    data.append([moniker, operator_address, tokens])

data.sort(key=lambda x: float(x[2]), reverse=True)
data = data[:num_validators]
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