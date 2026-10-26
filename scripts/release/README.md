# Python release helpers

Install [uv](https://docs.astral.sh/uv/getting-started/installation/). These helpers
share one locked environment, replacing their individual `requirements.txt`
files. uv selects Python 3.10 from `.python-version` and downloads it if needed.
No manual virtual environment activation is necessary.

From the repository root:

```bash
uv sync --project scripts/release --locked

uv run --project scripts/release --locked python scripts/release/create_binaries_json/create_binaries_json.py --tag v32.0.0

uv run --project scripts/release --locked python scripts/release/update_chain_registry/update_chain_registry.py --upgrade_version v32.0.0 --upgrade_height 12345678

# Offline smoke tests using a local HTTP server.
uv run --project scripts/release --locked python -m unittest discover -s scripts/release -p '*_test.py'
```

The shell wrappers also use uv and can be called from any working directory:

```bash
bash scripts/release/create_binaries_json/create_all_binaries_json.sh

# This overwrites chain.schema.json at the repository root after success.
bash scripts/release/update_chain_registry/run.sh v32.0.0 12345678
```

To update a dependency, run `uv add --project scripts/release 'package==version'`,
rerun the tests, and commit both `pyproject.toml` and `uv.lock`. CI uses `--locked`
to reject stale lockfiles.
