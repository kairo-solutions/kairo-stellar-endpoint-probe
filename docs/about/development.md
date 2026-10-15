# Development

## Requirements

- Go 1.21 or newer

## Run checks

```sh
gofmt -w .
go test -race ./...
go vet ./...
go build ./...
```

The tests use local HTTP test servers and do not require live Stellar
endpoints.

## Contribution scope

Focused bug fixes and improvements to the probe's documented behavior are
welcome. Include tests for observable changes and keep endpoint checks
read-only.
