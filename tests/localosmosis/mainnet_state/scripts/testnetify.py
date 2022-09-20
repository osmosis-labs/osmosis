import json
import argparse
from datetime import datetime
from dataclasses import dataclass

# Classes 
@dataclass
class Validator:
    moniker: str
    pubkey: str
    hex_address: str
    operator_address: str
    consensus_address: str

@dataclass
class Account:
    pubkey: str
    address: str

# Contants
BONDED_TOKENS_POOL_MODULE_ADDRESS = "osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09"
DISTRIBUTION_MODULE_ADDRESS = "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"
DISTRIBUTION_MODULE_OFFSET = 2

config = {
    "governance_voting_period": "180s",
    "epoch_duration": '21600s',
}

def replace(d, old_value, new_value):
    """
    Replace all the occurences of `old_value` with `new_value`
    in `d` dictionary
    """
    for k in d.keys():
        if isinstance(d[k], dict):
            replace(d[k], old_value, new_value)
        elif isinstance(d[k], list):
            for i in range(len(d[k])):
                if isinstance(d[k][i], dict) or isinstance(d[k][i], list):
                    replace(d[k][i], old_value, new_value)
                else:
                    if d[k][i] == old_value:
                        d[k][i] = new_value
        else:
            if d[k] == old_value:
                d[k] = new_value

def replace_validator(genesis, old_validator, new_validator):
    
    replace(genesis, old_validator.hex_address, new_validator.hex_address)
    replace(genesis, old_validator.consensus_address, new_validator.consensus_address)

    # replace(genesis, old_validator.pubkey, new_validator.pubkey)
    for validator in genesis["validators"]:
        if validator['name'] == old_validator.moniker:
            validator['pub_key']['value'] = new_validator.pubkey
        
    for validator in genesis['app_state']['staking']['validators']:
        if validator['description']['moniker'] == old_validator.moniker:
            validator['consensus_pubkey']['key'] = new_validator.pubkey

    # This creates problems
    # replace(genesis, old_validator.operator_address, new_validator.operator_address)
    
    # replacing operator_address in lockup > synthetic_locks
    for synthetic_lock in genesis['app_state']['lockup']['synthetic_locks']:
        # {
        #   "duration": "1209600s",
        #   "end_time": "0001-01-01T00:00:00Z",
        #   "synth_denom": "gamm/pool/704/superbonding/osmovaloper1pgl5usqpelz3a4c04g6t3yyvlrc9yseylmtcnt",
        #   "underlying_lock_id": "1367816"
        # }

        break # skip for now
        if synthetic_lock['synth_denom'].endswith(old_validator.operator_address):
            synthetic_lock['synth_denom'] = synthetic_lock['synth_denom'].replace(old_validator.operator_address, new_validator.operator_address)

    # Replacing operator_address in incentives > gauges
    for gauge in genesis['app_state']['incentives']['gauges']:
        # {
        #     "coins": [],
        #     "distribute_to": {
        #         "denom": "gamm/pool/3/superbonding/osmovaloper1969qjenvrjxrypynv36j9hjmlxrqgmy0xq3dg2",
        #         "duration": "1209600s",
        #         "lock_query_type": "ByDuration",
        #         "timestamp": "0001-01-01T00:00:00Z"
        #     },
        #     "distributed_coins": [],
        #     "filled_epochs": "0",
        #     "id": "29620",
        #     "is_perpetual": true,
        #     "num_epochs_paid_over": "1",
        #     "start_time": "2022-09-10T12:19:10.101498191Z"
        # }

        break # skip for now
        if gauge['distribute_to']['denom'].endswith(old_validator.operator_address):
            gauge['distribute_to']['denom'] = gauge['distribute_to']['denom'].replace(old_validator.operator_address, new_validator.operator_address)

def replace_account(genesis, old_account, new_account):

    replace(genesis, old_account.address, new_account.address)
    replace(genesis, old_account.pubkey, new_account.pubkey)

def create_parser():

    parser = argparse.ArgumentParser(
    formatter_class=argparse.RawDescriptionHelpFormatter,
    description='Create a testnet from a state export')

    parser.add_argument(
        '-c',
        '--chain-id',
        type = str,
        default="localosmosis",
        help='Chain ID for the testnet \nDefault: localosmosis\n'
    )

    parser.add_argument(
        '-i',
        '--input',
        type = str,
        default="genesis.json",
        dest='input_genesis',
        help='Path to input genesis'
    )

    parser.add_argument(
        '-o',
        '--output',
        type = str,
        default="testnet_genesis.json",
        dest='output_genesis',
        help='Path to output genesis'
    )
    
    parser.add_argument(
        '--validator-hex-address',
        type = str,
        help='Validator hex address to replace'
    )

    parser.add_argument(
        '--validator-operator-address',
        type = str,
        help='Validator operator address to replace'
    )

    parser.add_argument(
        '--validator-consensus-address',
        type = str,
        help='Validator consensus address to replace'
    )

    parser.add_argument(
        '--validator-pubkey',
        type = str,
        help='Validator pubkey to replace'
    )

    parser.add_argument(
        '-v',
        '--verbose',
        action='store_true',
        help='More verbose output'
    )

    parser.add_argument(
        '--pretty-output', 
        action='store_true',
        help='Properly indent output genesis (increases time and file size)'
    )

    return parser

def main():

    parser = create_parser()
    args = parser.parse_args()

    new_validator = Validator(
        moniker = "val",
        pubkey = args.validator_pubkey,
        hex_address = args.validator_hex_address,
        operator_address = args.validator_operator_address,
        consensus_address = args.validator_consensus_address
    )

    old_validator = Validator(
        moniker = "Sentinel dVPN",
        pubkey = "b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU=",
        hex_address = "16A169951A878247DBE258FDDC71638F6606D156",
        operator_address = "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n",
        consensus_address = "osmovalcons1z6skn9g6s7py0klztr7acutr3anqd52k9x5p70"
    )

    new_account = Account(
        pubkey = "A2MR6q+pOpLtdxh0tHHe2JrEY2KOcvRogtLxHDHzJvOh",
        address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj"
    )

    old_account = Account(
        pubkey = "AqlNb1FM8veQrT4/apR5B3hww8VApc0LTtZnXhq7FqG0",
        address = "osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5"
    )

    print("📝 Opening {}... (it may take a while)".format(args.input_genesis))
    with open(args.input_genesis, 'r') as f:
        genesis = json.load(f)
    
    # Replace chain-id
    if args.verbose:
        print("🔗 Replacing chain-id {} with {}".format(genesis['chain_id'], args.chain_id))
    genesis['chain_id'] = args.chain_id

    # Update gov module
    if args.verbose:
        print("🗳️ Update gov module")
        print("\tSetting governance_voting_period to", config["governance_voting_period"])
    genesis['app_state']['gov']['voting_params']['voting_period'] = config["governance_voting_period"]

    # Update epochs module
    if args.verbose:
        print("⌛ Update epochs module")
        print("\tSetting epoch_duration to", config["epoch_duration"])
        print("\tResetting current_epoch_start_time")
    genesis['app_state']['epochs']['epochs'][0]['duration'] = config["epoch_duration"]
    genesis['app_state']['epochs']['epochs'][0]['current_epoch_start_time'] = datetime.now().isoformat() + 'Z'
    
    # Prune IBC
    if args.verbose:
        print("🕸 Pruning IBC module")

    genesis['app_state']["ibc"]["channel_genesis"]["ack_sequences"] = []
    genesis['app_state']["ibc"]["channel_genesis"]["acknowledgements"] = []
    genesis['app_state']["ibc"]["channel_genesis"]["channels"] = []
    genesis['app_state']["ibc"]["channel_genesis"]["commitments"] = []
    genesis['app_state']["ibc"]["channel_genesis"]["receipts"] = []
    genesis['app_state']["ibc"]["channel_genesis"]["recv_sequences"] = []
    genesis['app_state']["ibc"]["channel_genesis"]["send_sequences"] = []

    genesis['app_state']["ibc"]["client_genesis"]["clients"] = []
    genesis['app_state']["ibc"]["client_genesis"]["clients_consensus"] = []
    genesis['app_state']["ibc"]["client_genesis"]["clients_metadata"] = []

    # Impersonate validator
    if args.verbose:
        print("🚀 Replacing validator")
        # print("\t{:50} -> {}".format(old_validator.moniker, new_validator.moniker))
        print("\t{:50} -> {}".format(old_validator.pubkey, new_validator.pubkey))
        print("\t{:50} -> {}".format(old_validator.consensus_address, new_validator.consensus_address))
        print("\t{:50} -> {}".format(old_validator.operator_address, new_validator.operator_address))
        print("\t{:50} -> {}".format(old_validator.hex_address, new_validator.hex_address))

    replace_validator(genesis, old_validator, new_validator)

    # Impersonate account
    if args.verbose:
        print("🧪 Replacing account")
        print("\t{:50} -> {}".format(old_account.address, new_account.address))
        print("\t{:50} -> {}".format(old_account.pubkey, new_account.pubkey))
    
    replace_account(genesis, old_account, new_account)
        
    # Update staking module
    if args.verbose:
        print("🥩 Update staking module")

    # Replace validator pub key in genesis['app_state']['staking']['validators']
    for validator in genesis['app_state']['staking']['validators']:

        if validator['operator_address'] == new_validator.operator_address:
            
            # Update delegator shares
            validator['delegator_shares'] = str(int(float(validator['delegator_shares']) + 1000000000000000)) + ".000000000000000000"
            print("\tUpdate {} delegated shares to {}".format(validator['operator_address'], validator['delegator_shares']))

            # Update tokens
            validator['tokens'] = str(int(validator['tokens']) + 1000000000000000)
            print("\tUpdate {} tokens to {}".format(validator['operator_address'], validator['tokens']))
            break
    
    # Update self delegation on operator address
    for delegation in genesis['app_state']['staking']['delegations']:
        # {
        #   "delegator_address": "osmo1qqq9txnw4c77sdvzx0tkedsafl5s3vk7f6y6ca",
        #   "shares": "100000000.000000000000000000",
        #   "validator_address": "osmovaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y4"
        # }
        if delegation['delegator_address'] == new_account.address:

            delegation['validator_address'] = new_validator.operator_address
            delegation['shares'] = str(int(float(delegation['shares'])) + 1000000000000000) + ".000000000000000000"

            print("\tUpdate {} delegation shares to {} to {}".format(new_account.address, delegation['validator_address'], delegation['shares']))
            break

    # Update genesis['app_state']['distribution']['delegator_starting_infos'] on operator address
    for delegator_starting_info in genesis['app_state']['distribution']['delegator_starting_infos']:
        # {
        #   "delegator_address": "osmo10p27pypvmlp6njzy4vprpd0svtmqesascst59z",
        #   "starting_info": {
        #     "height": "3854640",
        #     "previous_period": "2",
        #     "stake": "10.000000000000000000"
        #   },
        #   "validator_address": "osmovaloper1qyksxgv03ngylanwzh2mx6k768frgjc0em6800"
        # }
        if delegator_starting_info['delegator_address'] == new_account.address:
            delegator_starting_info['starting_info']['stake'] = str(int(float(delegator_starting_info['starting_info']['stake']) + 1000000000000000))+".000000000000000000"
            print("\tUpdate {} stake to {}".format(delegator_starting_info['delegator_address'], delegator_starting_info['starting_info']['stake']))
            break

    if args.verbose:
        print("🔋 Update validator power")

    # Update power in genesis["validators"]
    for validator in genesis["validators"]:
        if validator['address'] == new_validator.hex_address:
            validator['power'] = str(int(validator['power']) + 1000000000)
            print("\tUpdate {} validator power to {}".format(validator['address'], validator['power']))
            break 
    
    for validator_power in genesis['app_state']['staking']['last_validator_powers']:
        # {
        #   "address": "osmovaloper19fec46jadzr4y6gd596kkf8sskm7eyq32kkwc7",
        #   "power": "504451"
        # }
        if validator_power['address'] == new_validator.operator_address:
            validator_power['power'] = str(int(validator_power['power']) + 1000000000)
            if args.verbose:
                print("\tUpdate {} last_validator_power to {}".format(new_validator.operator_address, validator_power['power']))
            break
    
    # Update total power
    genesis['app_state']['staking']['last_total_power'] = str(int(genesis['app_state']['staking']['last_total_power']) + 1000000000)
    if args.verbose:
        print("\tUpdate last_total_power to {}".format(genesis['app_state']['staking']['last_total_power']))

    # Update bank module
    if args.verbose:
        print("💵 Update bank module")

    for balance in genesis['app_state']['bank']['balances']:
        # {
        #   "address": "osmo1qqq8fdsrcsz4jnlvrcqa6ds2ruflgrgej32k90",
        #   "coins": [
        #     {
        #       "amount": "459000000",
        #       "denom": "uosmo"
        #     },
        #     {
        #       "amount": "43286104",
        #       "denom": "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC"
        #     }
        #   ]
        # }
        if balance['address'] == new_account.address:
            for coin in balance['coins']:
                if coin['denom'] == "uosmo":
                    print("\tUpdate {} uosmo balance from {} to {}".format(new_account.address, coin["amount"], str(int(coin["amount"]) + 1000000000000000)))
                    coin["amount"] = str(int(coin["amount"]) + 1000000000000000)
                    break
            break
    
    # Add 1 BN uosmo to bonded_tokens_pool module address
    for balance in genesis['app_state']['bank']['balances']:
        if balance['address'] == BONDED_TOKENS_POOL_MODULE_ADDRESS:
            # Find uosmo
            for coin in balance['coins']:
                if coin['denom'] == "uosmo":
                    print("\tUpdate {} (bonded_tokens_pool_module) uosmo balance from {} to {}".format(BONDED_TOKENS_POOL_MODULE_ADDRESS, coin["amount"], str(int(coin["amount"]) + 1000000000000000)))
                    coin["amount"] = str(int(coin["amount"]) + 1000000000000000)
                    break
            break
    
    # Distribution module fix
    for balance in genesis['app_state']['bank']['balances']:
        if balance['address'] == DISTRIBUTION_MODULE_ADDRESS:
            # Find uosmo
            for coin in balance['coins']:
                if coin['denom'] == "uosmo":
                    print("\tUpdate {} (distribution_module) uosmo balance from {} to {}".format(DISTRIBUTION_MODULE_ADDRESS, coin["amount"], str(int(coin["amount"]) + 1000000000000000)))
                    coin["amount"] = str(int(coin["amount"]) - DISTRIBUTION_MODULE_OFFSET)
                    break
            break

    # Update bank balance 
    for supply in genesis['app_state']['bank']['supply']:
        if supply["denom"] == "uosmo":
            print("\tUpdate total uosmo supply from {} to {}".format(supply["amount"], str(int(supply["amount"]) + 2000000000000000 - DISTRIBUTION_MODULE_OFFSET)))
            supply["amount"] = str(int(supply["amount"]) + 2000000000000000 - DISTRIBUTION_MODULE_OFFSET)
            break
    
    print("📝 Writing {}... (it may take a while)".format(args.output_genesis))
    with open(args.output_genesis, 'w') as f:
        if args.pretty_output:
            f.write(json.dumps(genesis, indent=2))
        else:
            f.write(json.dumps(genesis))

if __name__ == '__main__':
    main()