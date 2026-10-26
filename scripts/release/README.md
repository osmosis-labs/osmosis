# Python release helpers

Install [uv](https://docs.astral.sh/uv/getting-started/installation/). These helpers
share one locked environment, replacing their individual `requirements.txt`
files. uv selects Python 3.10 from `.python-version` and downloads it if needed.
No manual virtual environment activation is necessary.

From the repository root:

```bash
uv sync --project scripts/release --locked

uv run --project scripts/release --locked python scripts/release/create_binaries_json/create_binaries_json.py --tag v32.0.0

# Offline smoke tests using a local HTTP server.
uv run --project scripts/release --locked python -m unittest discover -s scripts/release -p '*_test.py'
```

The binary-list wrapper also uses uv and can be called from any working directory:

```bash
bash scripts/release/create_binaries_json/create_all_binaries_json.sh
```

## Legacy chain-registry updater

`update_chain_registry/update_chain_registry.py` and its `run.sh` wrapper are
legacy helpers, **not a supported way to generate the current registry**. Their
source (`osmosis/main/chain.schema.json`) was deleted, and the current
[chain-registry entry](https://github.com/cosmos/chain-registry/blob/master/osmosis/chain.json)
uses a different codebase schema. Restoring this helper requires a separate
schema migration; changing the URL alone is insufficient.

The helper remains in the shared environment for maintenance and offline tests.
Errors and skipped updates exit nonzero and write diagnostics to stderr. The
wrapper validates JSON before replacing any output, so a failed command or
invalid JSON leaves an existing `chain.schema.json` untouched.

To update a dependency, run `uv add --project scripts/release 'package==version'`,
rerun the tests, and commit both `pyproject.toml` and `uv.lock`. CI uses `--locked`
to reject stale lockfiles.
