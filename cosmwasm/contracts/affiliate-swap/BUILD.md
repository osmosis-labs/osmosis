Build locally:

1) Install Rust toolchain with wasm target:

```bash
rustup toolchain install stable
rustup target add wasm32-unknown-unknown
```

2) Run tests:

```bash
cargo test -p affiliate-swap
```

3) Optimize for upload:

```bash
make -C cosmwasm/contracts/affiliate-swap optimize
```

