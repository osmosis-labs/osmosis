import json

test_gen = open("testnet_genesis.json", "r+")
read_test_gen = json.loads(test_gen.read())

#remove everything from channel_genesis except next_channel_sequence
for elem in list(read_test_gen['app_state']['ibc']['channel_genesis']):
    if elem == 'next_channel_sequence':
        continue
    else:
        del read_test_gen['app_state']['ibc']['channel_genesis'][elem]

#remove everything from client_genesis except create_localhost, params, and next_client_sequence
for elem in list(read_test_gen['app_state']['ibc']['client_genesis']):
    if elem == 'create_localhost' or elem == 'params' or elem == 'next_client_sequence':
        continue
    else:
        del read_test_gen['app_state']['ibc']['client_genesis'][elem]

#remove everything from distribution except params, fee_pool, outstanding_rewards, and previous_proposer
for elem in list(read_test_gen['app_state']['distribution']):
    if elem == 'params' or elem == 'fee_pool' or elem == 'outstanding_rewards' or elem == 'previous_proposer':
        continue
    else:
        del read_test_gen['app_state']['distribution'][elem]


print("Please wait while file writes over itself, this may take 60 seconds or more")
#go back to begining of file, write over with new values
test_gen.seek(0)
json.dump(read_test_gen, test_gen)

#delete remainder in case new data is shorter than old
test_gen.truncate()