# Checks and output

The probe performs independent read-only checks:

| Check | Request | Success condition |
| --- | --- | --- |
| Horizon root | `GET /` | Successful HTTP response with a network passphrase |
| Soroban RPC health | JSON-RPC `getHealth` | Result reports `status: healthy` |
| Soroban RPC latest ledger | JSON-RPC `getLatestLedger` | Result includes a positive ledger sequence |
| Soroban RPC network | JSON-RPC `getNetwork` | Result includes a passphrase |
| Expected network | Local comparison | Each returned passphrase matches the selected network |
| Cross-service consistency | Local comparison | Horizon and Soroban RPC passphrases match |

The Horizon request and three Soroban RPC requests run concurrently. The
endpoint checks are reported separately, so an unavailable service does not
hide the status of the other one. A consistency check is marked `skipped` when
either service cannot provide a passphrase.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | All requested checks passed |
| `1` | At least one endpoint or network check failed |
| `2` | Invalid command-line arguments or configuration |

Use the JSON report's `ok` field for the overall result and inspect individual
`checks` to identify a failing service. Successful latest-ledger checks include
the `ledger_sequence` in JSON and the ledger number in the text report. Latency
is measured in milliseconds.
Endpoint query strings are omitted from report output.

## Operational notes

- Endpoints must be absolute HTTP or HTTPS URLs and must not embed user
  credentials.
- Each request has its own timeout.
- Soroban RPC replies must use JSON-RPC 2.0 and match the request ID.
- Use HTTPS for production endpoints.
- The probe does not treat endpoint reachability as transaction simulation or
  proof of ledger freshness.
