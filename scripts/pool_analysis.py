import json

filename = "balance_derived.json"

f = open(filename, 'r+')
obj = json.loads(f.read())

accs = obj['accounts']
res = {}
for k in accs.keys():
    if len(accs[k]['bonded_by_select_pools']) > 0:
        res[accs[k]['address']] = accs[k]['bonded_by_select_pools']

pool_id = '560'
bucket_denom = "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC"
bucket_exp = 2
buckets = {}
for i in range(50):
    buckets[i] = 0

for k in res.keys():
    coins = res[k][pool_id]
    for c in coins:
        if c['denom'] != bucket_denom:
            continue
        amt = int(c['amount'])
        binSize = len(bin(amt)[2:])
        buckets[binSize] += 1

print(buckets)
for i in range(50):
    print("number_of_LPs", )