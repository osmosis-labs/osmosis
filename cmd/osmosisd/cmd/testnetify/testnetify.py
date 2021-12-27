import json
import subprocess


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




#get own address and replace with Sentinel addressed used for self delelgation



#open genesis json file with read write priv, load json
test_gen = open("testnet_genesis.json", "r+")
read_test_gen = json.loads(test_gen.read())

#print(read_test_gen.keys())

#supply list (ibc, ion, osmo)
supply = read_test_gen['app_state']['bank']['supply']

#iterate over list and find uosmo line, get total osmo supply and find its index 
for i in supply:
    if i['denom'] == 'uosmo':
        osmo_index = supply.index(i)
        osmo_supply = supply[osmo_index]['amount']
        break

print(supply[osmo_index]['amount'])

#update osmo supply to new total osmo value (add 2 Billion OSMO)
osmo_supply_new = int(osmo_supply) + 2000000000000000
supply[osmo_index]['amount'] = str(osmo_supply_new)


print(supply[osmo_index]['amount'])





#go back to begining of file, write over with new values
test_gen.seek(0)
json.dump(read_test_gen, test_gen)

#delete remainder in case new data is shorter than old
test_gen.truncate()