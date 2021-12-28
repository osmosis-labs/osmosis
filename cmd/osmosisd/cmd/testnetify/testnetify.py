import json
import subprocess
import re, shutil, tempfile
from datetime import date


#get values from your priv_validator_key.json to later switch with high power validator

#get bas64
result = subprocess.run(["osmosisd","tendermint","show-validator"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
base64 = result.stdout.strip()

#get validator cons pubkey
val_pubkey = base64[base64.find('key":') +6 :-2]

#osmosisd debug pubkey {base64} to get address
debug_pubkey = subprocess.run(["osmosisd","debug", "pubkey", base64], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

#hex address
address = debug_pubkey.stderr[9: debug_pubkey.stderr.find("\n")]

#feed hex address into osmosisd debug addr {address} to get bech32 validator address (osmovaloper)
bech32 = subprocess.run(["osmosisd","debug", "addr", address], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
#osmovaloper
bech32_val = bech32.stderr[bech32.stderr.find("Val: ") + 5: -1]

#pass osmovaloper address into osmosisd debug bech32-convert -p osmovalcons
bech32_convert = subprocess.run(["osmosisd","debug", "bech32-convert", bech32_val, "-p", "osmovalcons"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
#osmovalcons
final_address = bech32_convert.stderr[:bech32_convert.stderr.find("\n")]


#own opp address
#mnemonic: kitchen comic flower drip sick prize account cheese truth income weekend nominee segment punch call satisfy captain earth ethics wasp clump tunnel orchard advance
#exchange this value with own address or use above mnemonic for following address
op_address = "osmo1qye772qje88p7ggtzrvl9nxvty6dkuuskkg52l"

#own pub key
op_base64_pre = subprocess.run(["osmosisd","query", "auth", "account", op_address], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
op_pubkey = op_base64_pre.stdout[op_base64_pre.stdout.find("key: ")+5:op_base64_pre.stdout.find("sequence")-1]


def sed_inplace(filename, pattern, repl):
    '''
    Perform the pure-Python equivalent of in-place `sed` substitution: e.g.,
    `sed -i -e 's/'${pattern}'/'${repl}' "${filename}"`.
    '''
    # For efficiency, precompile the passed regular expression.
    pattern_compiled = re.compile(pattern)

    # For portability, NamedTemporaryFile() defaults to mode "w+b" (i.e., binary
    # writing with updating). This is usually a good thing. In this case,
    # however, binary writing imposes non-trivial encoding constraints trivially
    # resolved by switching to text writing. Let's do that.
    with tempfile.NamedTemporaryFile(mode='w', delete=False) as tmp_file:
        with open(filename) as src_file:
            for line in src_file:
                tmp_file.write(pattern_compiled.sub(repl, line))

    # Overwrite the original file with the munged temporary file in a
    # manner preserving file attributes (e.g., permissions).
    shutil.copystat(filename, tmp_file.name)
    shutil.move(tmp_file.name, filename)


#validator cons pubkey:     val_pubkey
#osmovalcons:               final_address
#validator hex address:     address
#osmovaloper:               bech32_val
#actual account:            op_address
#accounts pubkey:           op_pubkey



#replace validator cons pubkey (this did not work well due to random forward slashes, did other way later)
#sed_inplace("testnet_genesis.json", "b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU=", val_pubkey)

#replace validator osmovalcons
sed_inplace("testnet_genesis.json", "osmovalcons1z6skn9g6s7py0klztr7acutr3anqd52k9x5p70", final_address)

#replace validator hex address
sed_inplace("testnet_genesis.json", "16A169951A878247DBE258FDDC71638F6606D156", address)

#replace validator osmovaloper
sed_inplace("testnet_genesis.json", "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n", bech32_val)

#replace actual account
sed_inplace("testnet_genesis.json", "osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5", op_address)

#replace actual account pubkey
sed_inplace("testnet_genesis.json", "AqlNb1FM8veQrT4/apR5B3hww8VApc0LTtZnXhq7FqG0", op_pubkey)




#open genesis json file with read write priv, load json
test_gen = open("testnet_genesis.json", "r+")
read_test_gen = json.loads(test_gen.read())
#print(read_test_gen.keys())



#validator pubkey must be replaced from b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU= in two locations
#first under read_test_gen['app_state']['staking']['validators']
#second under read_test_gen['validators']
#i tried to do this using sed_inplace above but the multiple slashes broke it so did this instead

#first val list index
app_state_val_list = read_test_gen['app_state']['staking']['validators']
val_index = [i for i, elem in enumerate(app_state_val_list) if 'Sentinel' in elem['description']['moniker']][0]
#first val list update key
app_state_val_list[val_index]['consensus_pubkey']['key'] = val_pubkey
#also update delegator shares and tokens
app_state_val_list[val_index]['delegator_shares'] = str(int(float(app_state_val_list[val_index]['delegator_shares']) + 1000000000000000)) + ".000000000000000000"
app_state_val_list[val_index]['tokens'] = str(int(app_state_val_list[val_index]['tokens']) + 1000000000000000)


#second val list index
val_list_2 = read_test_gen['validators']
val_list_2_index = [i for i, elem in enumerate(val_list_2) if 'Sentinel' in elem['name']][0]
#second val list update key
val_list_2[val_list_2_index]['pub_key']['value'] = val_pubkey








#change self delegation amount on operator address

#first location
app_state_del_list = read_test_gen['app_state']['staking']['delegations']
del_index = [i for i, elem in enumerate(app_state_del_list) if op_address in elem['delegator_address']][0]
#first val list update share (add 1 BN)
current_share = int(float(app_state_del_list[del_index]['shares']))
new_share = str(current_share + 1000000000000000)+".000000000000000000"
app_state_del_list[del_index]['shares'] = new_share

#second location
app_state_dist_list = read_test_gen['app_state']['distribution']['delegator_starting_infos']
dist_index = [i for i, elem in enumerate(app_state_dist_list) if op_address in elem['delegator_address']][0]
#second val list update stake (add 1 BN)
current_stake = int(float(app_state_dist_list[dist_index]['starting_info']['stake']))
new_stake = str(current_stake + 1000000000000000)+".000000000000000000"
app_state_dist_list[dist_index]['starting_info']['stake'] = new_stake











#get index of val power
val_power_list = read_test_gen['validators']
val_power_index = [i for i, elem in enumerate(val_power_list) if 'Sentinel' in elem['name']][0]
#change val power (add 1 BN)
current_power = int(val_power_list[val_power_index]['power'])
new_power = str(current_power + 1000000000)
val_power_list[val_power_index]['power'] = new_power
#get index of val power in app state (osmovaloper) (bech32_val)
last_val_power_list = read_test_gen['app_state']['staking']['last_validator_powers']
last_val_power_index = [i for i, elem in enumerate(last_val_power_list) if bech32_val in elem['address']][0]
val_power = int(read_test_gen['app_state']['staking']['last_validator_powers'][last_val_power_index]['power'])
new_val_power = str(val_power + 1000000000)
read_test_gen['app_state']['staking']['last_validator_powers'][last_val_power_index]['power'] = new_val_power








#update last_total_power (last total bonded across all validators, add 1BN)
last_total_power = int(read_test_gen['app_state']['staking']['last_total_power'])
new_last_total_power = str(last_total_power + 1000000000)
read_test_gen['app_state']['staking']['last_total_power'] = new_last_total_power










#update operator address amount (add 1 BN)
#find wallet index in bank balance
bank_balance_list = read_test_gen['app_state']['bank']['balances']
op_amount_index = [i for i, elem in enumerate(bank_balance_list) if op_address in elem['address']][0]
#get uosmo index from wallet
op_wallet = read_test_gen['app_state']['bank']['balances'][op_amount_index]['coins']
op_uosmo_index = [i for i, elem in enumerate(op_wallet) if 'uosmo' in elem['denom']][0]
#update uosmo amount
op_uosmo = int(op_wallet[op_uosmo_index]['amount'])
new_op_uosmo = str(op_uosmo + 1000000000000000)
op_wallet[op_uosmo_index]['amount'] = new_op_uosmo







#update total OSMO supply (add 2 BN)
#supply list (ibc, ion, osmo)
supply = read_test_gen['app_state']['bank']['supply']

#get index of osmo supply
osmo_index = [i for i, elem in enumerate(supply) if 'uosmo' in elem['denom']][0]

#get osmo supply value
osmo_supply = supply[osmo_index]['amount']

#update osmo supply to new total osmo value (add 2 Billion OSMO)
osmo_supply_new = int(osmo_supply) + 2000000000000000
supply[osmo_index]['amount'] = str(osmo_supply_new)











#update bonded_tokens_pool module balance osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09 (add 1BN)
#get list of bank balances
bank_bal_list = read_test_gen['app_state']['bank']['balances']
#get index of module account
module_acct_index = [i for i, elem in enumerate(bank_bal_list) if 'osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09' in elem['address']][0]
#get current value
module_denom_list = bank_bal_list[module_acct_index]['coins']
osmo_bal_index = [i for i, elem in enumerate(module_denom_list) if 'uosmo' in elem['denom']][0]
osmo_bal = bank_bal_list[module_acct_index]['coins'][osmo_bal_index]['amount']
#increase by 1BN 
bank_bal_list[module_acct_index]['coins'][osmo_bal_index]['amount'] = int(osmo_bal) + 1000000000000000








#edit gov params
#change epoch duration to 3600s
epochs_list = read_test_gen['app_state']['epochs']['epochs'][0]
duration_current = epochs_list['duration']
epochs_list['duration'] = '3600s'

#change current_epoch_start_time
start_time_current = epochs_list['current_epoch_start_time']
today = date.today()
date_format = today.strftime("%Y-%m-%d")







#osmo supply: osmo_supply
#new osmo supply: new_osmo_supply





#go back to begining of file, write over with new values
test_gen.seek(0)
json.dump(read_test_gen, test_gen)

#delete remainder in case new data is shorter than old
test_gen.truncate()