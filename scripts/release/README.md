# Python release helpers

Install [uv](https://docs.astral.sh/uv/getting-started/installation/). The
binary-JSON helper uses a locked environment, replacing its `requirements.txt`
file. uv selects Python 3.10 from `.python-version` and downloads it if needed.
No manual virtual environment activation is necessary.

From the repository root:

```bash
uv sync --project scripts/release --locked

uv run --project scripts/release --locked python scripts/release/create_binaries_json/create_binaries_json.py --tag v32.0.0

# Offline smoke tests using a local HTTP server.
uv run --project scripts/release --locked python -m unittest discover -s scripts/release -p '*_test.py'
```

`create_upgrade_guide/create_upgrade_guide.py` is a separate manual helper using
only Python's standard library. Run it from its own directory so it can find
`UPGRADE_TEMPLATE.md`; see its usage comments for arguments.

The obsolete chain-registry updater and the batch binary-list wrapper (whose
hard-coded releases stopped at v25) have been removed. Make registry changes in
[cosmos/chain-registry](https://github.com/cosmos/chain-registry/tree/master/osmosis).
Use the single-release binary-JSON command above for a specific tag.

To update a dependency, run `uv add --project scripts/release 'package==version'`,
rerun the tests, and commit both `pyproject.toml` and `uv.lock`. The `--locked`
flag rejects stale lockfiles.
