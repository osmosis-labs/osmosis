# Concentrated Liquidity Go Client

This Go-client allows connecting to an Osmosis chain via Ignite CLI and
setting up a concentrated liquidity pool with positions.

## General Setup FAQ

- Update constants at the top of the file accordingly.
   * Make sure keyring is set up.
   * Client home is pointing to the right place.

## LocalOsmosis Setup

Make sure that you run `localosmosis` in the background and have keys
added to your keyring with:
```bash
make localnet-keys
```

See `tests/localosmosis` for more info.
