# Quick start

## Install

With Go 1.21 or newer:

```sh
go install github.com/kairo-solutions/kairo-stellar-endpoint-probe/cmd/kairo-stellar-endpoint-probe@latest
```

Or run from a checkout:

```sh
go run ./cmd/kairo-stellar-endpoint-probe \
  --horizon https://horizon-testnet.stellar.org \
  --soroban-rpc https://soroban-testnet.stellar.org
```

## Choose a network

Testnet is the default. Public network checks are explicit:

```sh
kairo-stellar-endpoint-probe \
  --horizon https://horizon.stellar.org \
  --soroban-rpc https://soroban-rpc.mainnet.stellar.gateway.fm \
  --network public
```

For a private network, set the expected passphrase:

```sh
kairo-stellar-endpoint-probe \
  --horizon https://horizon.example.org \
  --soroban-rpc https://rpc.example.org \
  --network custom \
  --network-passphrase "My Private Stellar Network"
```

The passphrase is compared with both endpoint responses; it is not used as a
credential.

## Machine-readable output

Use `--json` to send a structured report to a script:

```sh
kairo-stellar-endpoint-probe \
  --horizon https://horizon-testnet.stellar.org \
  --soroban-rpc https://soroban-testnet.stellar.org \
  --json
```

Set `--timeout 10s` to change the per-request timeout. The default is five
seconds.
