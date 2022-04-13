import json
import subprocess
import re, shutil, tempfile
from datetime import datetime


#get values from your priv_validator_key.json to later switch with high power validator

daemon_name = "osmosisd"

#get bas64
result = subprocess.run([daemon_name,"tendermint","show-validator"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
base64 = result.stdout.strip()
##base64 = '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"3QVAkiUIkKR3B6kkbd+QqzWDdcExoggbZV5fwH4jKDs="}'

#get validator cons pubkey
val_pubkey = base64[base64.find('key":') +6 :-2]
##val_pubkey = "3QVAkiUIkKR3B6kkbd+QqzWDdcExoggbZV5fwH4jKDs="

#osmosisd debug pubkey {base64} to get address
debug_pubkey = subprocess.run([daemon_name,"debug", "pubkey", base64], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

#address
address = debug_pubkey.stderr[9: debug_pubkey.stderr.find("\n")]
##based on show-valdiator
##address = "214D831D6F49A75F9104BDC3F2E12A6CC1FC5669"

#feed address into osmosisd debug addr {address} to get bech32 validator address (osmovaloper)
bech32 = subprocess.run([daemon_name,"debug", "addr", address], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
#osmovaloper
bech32_val = bech32.stderr[bech32.stderr.find("Val: ") + 5: -1]
##operator address
##bech32_val = "osmovaloper1y9xcx8t0fxn4lygyhhpl9cf2dnqlc4nf4pymm4"

#pass osmovaloper address into osmosisd debug bech32-convert -p osmovalcons
bech32_convert = subprocess.run([daemon_name,"debug", "bech32-convert", bech32_val, "-p", "osmovalcons"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
#osmovalcons
final_address = bech32_convert.stderr[:bech32_convert.stderr.find("\n")]
##osmovalcons is taken from show-validator
##final_address = "osmovalcons1y9xcx8t0fxn4lygyhhpl9cf2dnqlc4nfpjh8h5"

#own opp address
#exchange the op_address and op_pubkey with own address and pubkey or use above mnemonic for following address
#bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort
#CAN MODIFY
op_address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj"

#own pub key
#op_base64_pre = subprocess.run(["osmosisd","query", "auth", "account", op_address], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
#op_pubkey = op_base64_pre.stdout[op_base64_pre.stdout.find("key: ")+5:op_base64_pre.stdout.find("sequence")-1]
#CAN MODIFY
op_pubkey = "A2MR6q+pOpLtdxh0tHHe2JrEY2KOcvRogtLxHDHzJvOh"

#feed address into osmosisd debug addr {address} to get bech32 validator op address (osmovaloper)
bech32_op = subprocess.run([daemon_name,"debug", "addr", op_address], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
#osmovaloper
bech32_valoper = bech32_op.stderr[bech32_op.stderr.find("Val: ") + 5: -1]
# osmovaloper12smx2wdlyttvyzvzg54y2vnqwq2qjatex7kgq4

def sed_inplace(filename, pattern, repl):
    '''
    Perform the pure-Python equivalent of in-place `sed` substitution: e.g.,
    `sed -i -e 's/'${pattern}'/'${repl}' "${filename}"`.
    '''
    # For efficiency, precompile the passed regular expression.
    pattern_compiled = re.compile(pattern)

    # For portability, NamedTemporaryFile() defaults to mode "w+b" (i.e., binary
    # writing with updating)
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
#osmovaloper:               bech32_valoper
#actual account:            op_address
#accounts pubkey:           op_pubkey



#replace validator cons pubkey (this did not work well due to random forward slashes, did other way later)
#sed_inplace("testnet_genesis.json", "b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU=", val_pubkey)

#replace validator osmovalcons
print("Replacing osmovalcons1z6skn9g6s7py0klztr7acutr3anqd52k9x5p70 with " + final_address)
sed_inplace("testnet_genesis.json", "osmovalcons1z6skn9g6s7py0klztr7acutr3anqd52k9x5p70", final_address)

#replace validator hex address
print("Replacing 16A169951A878247DBE258FDDC71638F6606D156 with " + address)
sed_inplace("testnet_genesis.json", "16A169951A878247DBE258FDDC71638F6606D156", address)

#replace validator osmovaloper
print("Replacing osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n with " + bech32_valoper)
sed_inplace("testnet_genesis.json", "osmovaloper1cyw4vw20el8e7ez8080md0r8psg25n0cq98a9n", bech32_valoper)

#replace actual account
print("Replacing osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5 with " + op_address)
sed_inplace("testnet_genesis.json", "osmo1cyw4vw20el8e7ez8080md0r8psg25n0c6j07j5", op_address)

#replace actual account pubkey
print("Replacing AqlNb1FM8veQrT4/apR5B3hww8VApc0LTtZnXhq7FqG0 with " + op_pubkey)
sed_inplace("testnet_genesis.json", "AqlNb1FM8veQrT4/apR5B3hww8VApc0LTtZnXhq7FqG0", op_pubkey)



#open genesis json file with read write priv, load json
test_gen = open("testnet_genesis.json", "r+")
read_test_gen = json.loads(test_gen.read())
#print(read_test_gen.keys())


#change chain-id
print("Current chain-id is " + read_test_gen['chain_id'])
#CAN MODIFY
new_chain_id = "osmo-test-2"
read_test_gen['chain_id'] = new_chain_id
print("New chain-id is " + read_test_gen['chain_id'])


#validator pubkey must be replaced from b77zCh/VsRgVvfGXuW4dB+Dhg4PrMWWBC5G2K/qFgiU= in two locations
#first under read_test_gen['app_state']['staking']['validators']
#second under read_test_gen['validators']
#i tried to do this using sed_inplace above but the multiple slashes broke it so did this instead

#first val list index
app_state_val_list = read_test_gen['app_state']['staking']['validators']
val_index = [i for i, elem in enumerate(app_state_val_list) if 'Sentinel' in elem['description']['moniker']][0]
#first val list update key
#based on val
app_state_val_list[val_index]['consensus_pubkey']['key'] = val_pubkey
#also update delegator shares and tokens
current_del_share = str(app_state_val_list[val_index]['delegator_shares'])
print("Current delegator shares is " + current_del_share)
app_state_val_list[val_index]['delegator_shares'] = str(int(float(app_state_val_list[val_index]['delegator_shares']) + 1000000000000000)) + ".000000000000000000"
print("New delegator shares is " + app_state_val_list[val_index]['delegator_shares'])
print("Current delegator tokens is " + app_state_val_list[val_index]['tokens'])
app_state_val_list[val_index]['tokens'] = str(int(app_state_val_list[val_index]['tokens']) + 1000000000000000)
print("New delegator tokens is " + app_state_val_list[val_index]['tokens'])

#second val list index
val_list_2 = read_test_gen['validators']
val_list_2_index = [i for i, elem in enumerate(val_list_2) if 'Sentinel' in elem['name']][0]
#second val list update key
#based on val
val_list_2[val_list_2_index]['pub_key']['value'] = val_pubkey


#distribution module fix
dist_address = "osmo1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80yhvld"
app_state_balances_list = read_test_gen['app_state']['bank']['balances']
dist_index = [i for i, elem in enumerate(app_state_balances_list) if dist_address in elem['address']][0]
dist_all = app_state_balances_list[dist_index]['coins']
osmo_index = [i for i, elem in enumerate(dist_all) if 'uosmo' in elem['denom']][0]
current_dist_osmo_bal = dist_all[osmo_index]['amount']
dist_offset_amt = 2
print("Current distribution account uosmo balance is " + current_dist_osmo_bal)
new_dist_osmo_bal = str(int(current_dist_osmo_bal) - dist_offset_amt)
print("New distribution account uosmo balance is " + new_dist_osmo_bal)
dist_all[osmo_index]['amount'] = new_dist_osmo_bal


#change self delegation amount on operator address

#first location
app_state_del_list = read_test_gen['app_state']['staking']['delegations']
del_index = [i for i, elem in enumerate(app_state_del_list) if op_address in elem['delegator_address']][0]
#first val list update share (add 1 BN)
current_share = app_state_del_list[del_index]['shares']
print("Current self delegation is " + str(current_share))
new_share = str(int(float(current_share)) + 1000000000000000)+".000000000000000000"
print("New self delegation is " + new_share)
app_state_del_list[del_index]['shares'] = new_share

#second location
app_state_dist_list = read_test_gen['app_state']['distribution']['delegator_starting_infos']
dist_index = [i for i, elem in enumerate(app_state_dist_list) if op_address in elem['delegator_address']][0]
#second val list update stake (add 1 BN)
current_stake = app_state_dist_list[dist_index]['starting_info']['stake']
print("Current stake is " + str(current_stake))
new_stake = str(int(float(current_stake)) + 1000000000000000)+".000000000000000000"
print("New stake is " + new_stake)
app_state_dist_list[dist_index]['starting_info']['stake'] = new_stake


#get index of val power
val_power_list = read_test_gen['validators']
val_power_index = [i for i, elem in enumerate(val_power_list) if 'Sentinel' in elem['name']][0]
#change val power (add 1 BN)
current_power = int(val_power_list[val_power_index]['power'])
print("Current validator power is " + str(current_power))
new_power = str(current_power + 1000000000)
print("New validator power is " + new_power)
val_power_list[val_power_index]['power'] = new_power
#get index of val power in app state (osmovaloper) (bech32_valoper)
last_val_power_list = read_test_gen['app_state']['staking']['last_validator_powers']
last_val_power_index = [i for i, elem in enumerate(last_val_power_list) if bech32_valoper in elem['address']][0]
val_power = int(read_test_gen['app_state']['staking']['last_validator_powers'][last_val_power_index]['power'])
print("Current validator power in second location is " + str(val_power))
new_val_power = str(val_power + 1000000000)
print("New validator power in second location is " + new_val_power)
read_test_gen['app_state']['staking']['last_validator_powers'][last_val_power_index]['power'] = new_val_power


#update last_total_power (last total bonded across all validators, add 1BN)
last_total_power = int(read_test_gen['app_state']['staking']['last_total_power'])
print("Current last total power is " + str(last_total_power))
new_last_total_power = str(last_total_power + 1000000000)
print("New last total power is " + new_last_total_power)
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
print("Current operator address uosmo balance is " + str(op_uosmo))
new_op_uosmo = str(op_uosmo + 1000000000000000)
print("New operator address uosmo balance is " + new_op_uosmo)
op_wallet[op_uosmo_index]['amount'] = new_op_uosmo


#update total OSMO supply (add 2 BN)
#supply list (ibc, ion, osmo)
supply = read_test_gen['app_state']['bank']['supply']

#get index of osmo supply
osmo_index = [i for i, elem in enumerate(supply) if 'uosmo' in elem['denom']][0]

#get osmo supply value
osmo_supply = supply[osmo_index]['amount']
print("Current OSMO supply is " + osmo_supply)

#update osmo supply to new total osmo value (add 2 Billion OSMO)
#subtract by however much module account is subtracted by
osmo_supply_new = int(osmo_supply) + 2000000000000000 - dist_offset_amt
print("New OSMO supply is " + str(osmo_supply_new))
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
print("Current bonded tokens pool module account balance is " + osmo_bal)
#increase by 1BN
new_osmo_bal = int(osmo_bal) + 1000000000000000
print("New bonded tokens pool module account balance is " + str(new_osmo_bal))
bank_bal_list[module_acct_index]['coins'][osmo_bal_index]['amount'] = str(new_osmo_bal)


#edit epoch params
#change epoch duration to 21600s
epochs_list = read_test_gen['app_state']['epochs']['epochs'][0]
duration_current = epochs_list['duration']
print("Current epoch duration is " + duration_current)
#21600s for 6 hour epoch
#CAN MODIFY
new_duration = '21600s'
print("New epoch duration is " + new_duration)
epochs_list['duration'] = new_duration

#change current_epoch_start_time
start_time_current = epochs_list['current_epoch_start_time']
print("Current epoch start time is " + start_time_current)
#today = date.today()
now = datetime.now()
#date_format = now.strftime("%Y-%m-%d")
date_format = now.strftime("%Y-%m-%d"+"T"+"%H:%M:")
start_time_current_list = list(start_time_current)
start_time_current_list[:17] = date_format
start_time_new = ''.join(start_time_current_list)
epochs_list['current_epoch_start_time'] = start_time_new
print("New epoch start time is " + start_time_new)


#edit gov params
#change VotingPeriod
current_voting_period = read_test_gen['app_state']['gov']['voting_params']['voting_period']
print("Current voting period is " + current_voting_period)
#180s for 3 minute voting period
#CAN MODIFY
new_voting_period = "180s"
print("New voting period is " + new_voting_period)
read_test_gen['app_state']['gov']['voting_params']['voting_period'] = new_voting_period


print("Please wait while file writes over itself, this may take 60 seconds or more")
#go back to begining of file, write over with new values
test_gen.seek(0)
json.dump(read_test_gen, test_gen)

#delete remainder in case new data is shorter than old
test_gen.truncate()
