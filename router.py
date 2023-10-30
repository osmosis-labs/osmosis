import json
import requests
import itertools
import logging
import os

import networkx as nx

logging.basicConfig(level=os.environ.get('LOG_LEVEL', 'WARNING'))
logger = logging.getLogger(__name__)

POOLS_API = 'https://api-osmosis.imperator.co/stream/pool/v1/all?min_liquidity=0&order_key=liquidity&order_by=desc&offset=0&limit=5000'

POOLS = json.loads(requests.get(POOLS_API).text)['pools']


def get_prices(pool):
    return 1, 1


def generate_graph():
    graph = nx.DiGraph()

    for pool in POOLS:
        tokens = pool.get('pool_tokens', [])

        if isinstance(tokens, dict):
            try:
                tokens = [tokens['asset0'], tokens['asset1']]
            except KeyError:
                logger.warning(f'could not parse pool tokens: {tokens}')
                continue
        elif not isinstance(tokens, list) or len(tokens) != 2:
            logger.warning(f'Pool {pool["pool_id"]} has {len(tokens)} tokens. Only 2 supported')
            continue

        try:
            prices = get_prices(pool)
            edge1 = (tokens[0]['symbol'], tokens[1]['symbol'], prices[0])
            edge2 = (tokens[1]['symbol'], tokens[0]['symbol'], prices[1])

            # TODO: Deal with liquidity properly (especially for CL. Maybe multiple edges?)
            graph.add_weighted_edges_from([edge1, edge2], capacity=pool.get('liquidity', 0))

        except Exception as e:
            logger.warning(f'Pool {pool["pool_id"]} raised exception', str(e))

    return graph


def extract_path(flow_dict, source, target):
    paths = []
    visited = set()

    def dfs(node, current_path, current_flow):
        if node == target:
            paths.append(" -> ".join(current_path) + f" (Flow: {current_flow})")
            return
        visited.add(node)
        for neighbor, flow in flow_dict[node].items():
            if neighbor not in visited and flow > 0:
                dfs(neighbor, current_path + [f"{neighbor} ({flow})"], min(current_flow, flow))

    dfs(source, [source], float('inf'))
    return paths


if __name__ == '__main__':
    g = generate_graph()
    flow = nx.max_flow_min_cost(g, 'NOM', 'FLIX')
    print(extract_path(flow, 'NOM', 'FLIX'))
