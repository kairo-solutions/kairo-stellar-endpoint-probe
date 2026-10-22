# Stellar Endpoint Probe

`kairo-stellar-endpoint-probe` is a small Go CLI for checking that a Horizon endpoint
and a Soroban RPC endpoint are reachable, healthy, and configured for the same
Stellar network. It makes read-only requests and uses only the Go standard
library.

[**Read the documentation**](https://kairo-solutions.github.io/kairo-stellar-endpoint-probe/)

Part of the **Kairo Stellar toolkit**:

- [Soroban Event Snapshot](https://github.com/kairo-solutions/kairo-soroban-event-snapshot) exports contract events over a ledger range.
- [Stellar Transaction Inspector](https://github.com/kairo-solutions/kairo-stellar-tx-inspector) decodes transaction-envelope XDR.
- **Stellar Endpoint Probe** checks Horizon/Soroban health and network identity.

## Usage

```sh
go run ./cmd/kairo-stellar-endpoint-probe \
  --horizon https://horizon-testnet.stellar.org \
  --soroban-rpc https://soroban-testnet.stellar.org \
  --network testnet
```

The probe checks the Horizon root response, Soroban RPC `getHealth` and
`getNetwork` methods, and compares both reported network passphrases. Use
`--network public` for the public network, or `--network custom
--network-passphrase "..."` for a private network.

Use `--json` for machine-readable output and `--timeout` to set the timeout for
each request. The command exits with status `0` when all checks pass, `1` when
an endpoint check fails, and `2` for invalid arguments or configuration.
Endpoint query parameters are omitted from reports to avoid printing API tokens.

## Development

```sh
go test ./...
go run ./cmd/kairo-stellar-endpoint-probe --help
```

## Scope

This is a read-only connectivity and network-identity diagnostic. It does not
submit transactions, require private keys, or replace the official Stellar CLI.
