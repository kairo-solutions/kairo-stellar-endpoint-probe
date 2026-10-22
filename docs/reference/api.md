# Go API reference

Import the package as `endpointprobe`:

```go
import endpointprobe "github.com/kairo-solutions/kairo-stellar-endpoint-probe"
```

## `Config`

```go
type Config struct {
    HorizonURL        string
    SorobanRPCURL     string
    Network           string
    NetworkPassphrase string
    Timeout           time.Duration
}
```

Both endpoint URLs and a positive timeout are required. Set
`NetworkPassphrase` to enable checks against an expected network.

## `Probe`

```go
func Probe(ctx context.Context, cfg Config) (Report, error)
```

Returns an error for invalid configuration. Service failures are represented in
the report, allowing callers to display all outcomes in one run.

## `Report` and `Check`

`Report` contains the requested network, an overall `OK` flag, and ordered
checks. Each `Check` contains a name, `ok`/`fail`/`skipped` status, sanitized
endpoint, latency in milliseconds, and an explanatory message. The Soroban RPC
latest-ledger check also sets `LedgerSequence` on success.

```go
report, err := endpointprobe.Probe(ctx, endpointprobe.Config{
    HorizonURL:        "https://horizon-testnet.stellar.org",
    SorobanRPCURL:     "https://soroban-testnet.stellar.org",
    Network:           "testnet",
    NetworkPassphrase: "Test SDF Network ; September 2015",
    Timeout:           5 * time.Second,
})
```

Reports can be encoded as JSON using the standard `encoding/json` package.
