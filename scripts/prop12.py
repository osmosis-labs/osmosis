#!/bin/python3

import json
import functools
import sys

fname = sys.argv[1]
f = open(fname)
data = json.loads(f.read())

locks = data["app_state"]["lockup"]["locks"]

poolincentives = data["app_state"]["poolincentives"]
gauge_weights = {x["gauge_id"] : int(x["weight"]) for x in poolincentives["distr_info"]["records"]}
total_weight = int(poolincentives["distr_info"]["total_weight"])
osmo_issued = (300000000*1000000*0.45)/365
osmo_per_gauge = {gid : (osmo_issued*gauge_weights[gid]/total_weight) for gid in gauge_weights.keys()}

incentives = data["app_state"]["incentives"]
gauge_ids = {(x["distribute_to"]["denom"], x["distribute_to"]["duration"]) : x["id"] for x in incentives["gauges"] if len(x["coins"])>0 and x["coins"][0]["denom"]=="uosmo"}

accum = {}
for l in locks:
  amount = int(l["coins"][0]["amount"])
  denom = l["coins"][0]["denom"]
  duration = float(l["duration"][:-1])
  addr = l["owner"]

  if duration >= 86400:
    gid = gauge_ids[(denom, "86400s")]
    cur = accum.get(gid, {})
    cur[addr] = cur.get(addr, 0)+amount
    accum[gid] = cur

  if duration >= 604800:
    gid = gauge_ids[(denom, "604800s")]
    cur = accum.get(gid, {})
    cur[addr] = cur.get(addr, 0)+amount
    accum[gid] = cur
  
  if duration >= 1209600:
    gid = gauge_ids[(denom, "1209600s")]
    cur = accum.get(gid, {})
    cur[addr] = cur.get(addr, 0)+amount
    accum[gid] = cur

gauge_totals = {gid : sum(accum[gid].values()) for gid in accum.keys()}
mint_by_gauge_addr = {gid : {addr : osmo_per_gauge.get(gid, 0)*accum[gid][addr]/gauge_totals[gid] for addr in accum[gid].keys()} for gid in accum.keys()}
mint_by_addr = functools.reduce(lambda a,b: {addr: a[addr]+b[addr] if addr in a and addr in b else a.get(addr, b.get(addr)) for addr in a.keys() | b.keys()}, mint_by_gauge_addr.values())

print("\n".join([addr+", "+str(int(mint_by_addr[addr])) for addr in mint_by_addr.keys()]))