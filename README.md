# DoubleCloud Go SDK

[![GoDoc](https://godoc.org/github.com/doublecloud/go-sdk?status.svg)](https://godoc.org/github.com/doublecloud/go-sdk)

Go SDK for [DoubleCloud API](https://double.cloud/docs/en/public-api/).

## Installation

```bash
go get github.com/doublecloud/go-sdk
```

## Example usages

### Initializing SDK

```go
key, err := iamkey.ReadFromJSONFile(*flags.saPath)
if err != nil {
    panic(err)
}
creds, err := dc.ServiceAccountKey(key)
if err != nil {
    panic(err)
}
sdk, err := dc.Build(ctx, dc.Config{
    Credentials: creds,
})
```

### More examples

More examples can be found in [examples dir](examples).