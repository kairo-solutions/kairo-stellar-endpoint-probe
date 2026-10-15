# Stellar Endpoint Probe

Read-only diagnostics for the two Stellar services your application depends on:
Horizon and Soroban RPC.

<div class="docs-hero" markdown>

<div class="docs-hero__intro" markdown>

**STELLAR INFRASTRUCTURE**

## Check service health and network identity

Verify that endpoints respond, Soroban RPC is healthy, and both services are
connected to the network you expect.

[Get started](guides/quickstart.md){ .md-button .md-button--primary }
[Browse the Go API](reference/api.md){ .md-button }

</div>

<div class="docs-hero__example" markdown>

**A quick look**

```sh
go run ./cmd/stellar-endpoint-probe \
  --horizon https://horizon-testnet.stellar.org \
  --soroban-rpc https://soroban-testnet.stellar.org \
  --network testnet
```

</div>

</div>

## What it checks

<div class="docs-capabilities" markdown>

<div markdown>

### Horizon

Reads the Horizon root response and extracts the network passphrase.

</div>

<div markdown>

### Soroban RPC

Calls `getHealth` and `getNetwork` to confirm the RPC service is healthy and
identify its configured Stellar network.

</div>

<div markdown>

### Network consistency

Compares the Horizon and RPC passphrases with each other and with the expected
public, testnet, or custom network.

</div>

</div>

## Explore

<div class="docs-paths" markdown>

<div markdown>

### Use the probe

Install the command, choose a network, and interpret human-readable or JSON
results.

[Quick start](guides/quickstart.md) ·
[Checks and output](guides/checks.md)

</div>

<div markdown>

### Integrate or contribute

Use the Go package from automation, or run the tests and checks locally.

[Go API](reference/api.md) ·
[Development](about/development.md)

</div>

</div>

## Install

```sh
go install github.com/kairo-solutions/stellar-endpoint-probe/cmd/stellar-endpoint-probe@latest
```

The probe is read-only. It does not require keys, sign transactions, or submit
network operations.
