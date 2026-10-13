# iptrust-go

Official Go SDK for the [IP Trust](https://iptrust.co) API. Look up geolocation, ASN, company, hosting, known bot, VPN, proxy, Tor and privacy relay data for any IPv4 or IPv6 address, and download the underlying databases. No dependencies outside the standard library.

Get a free API key from the [IP Trust dashboard](https://dashboard.iptrust.co/). IP Trust is free for up to 10,000 requests a month.

## Install

```sh
go get github.com/IP-Trust/iptrust-go
```

Requires Go 1.22 or later.

## Quick start

```go
import "github.com/IP-Trust/iptrust-go"

client, err := iptrust.NewClient("your-api-key")
if err != nil {
    log.Fatal(err)
}

res, err := client.LookupIP(ctx, "9.9.9.9")
if err != nil {
    log.Fatal(err)
}

fmt.Println(res.Location.Country) // "United States"
fmt.Println(res.ASN.Number)       // "AS19281"
fmt.Println(res.VPN.Detected)     // false
```

## IP lookup

`LookupIP` calls `GET /ip/{ip}` and returns an `IPResponse` by value. The response has the same structure on every plan. Each section (`ASN`, `Location`, `VPN`, and so on) is a plain struct, never a pointer, with a `Valid` field that is `true` only when the API returned that section. Sections your plan does not include are left at their zero value with `Valid == false`.

```go
res, err := client.LookupIP(ctx, "2001:4860:4860::8888")

if res.IsBogon { /* reserved or unroutable */ }

if res.ASN.Valid {
    fmt.Println(res.ASN.Number)        // always present in a valid section
    if res.ASN.Risk.Valid {            // OptionalFloat: not every AS has been scored
        fmt.Println(res.ASN.Risk.Value) // 0.0 to 1.0
    }
    if res.ASN.RiskLabel.Valid && res.ASN.RiskLabel.Value == iptrust.RiskHigh {
        // ...
    }
}
if res.Location.Valid {
    res.Location.Resolution            // iptrust.ResolutionCity, ResolutionState, ResolutionCountry
    res.Location.Area.RadiusKM         // expected area, handy for impossible-travel checks
}
if res.Proxy.Valid {
    if res.Proxy.Type.Valid && res.Proxy.Type.Value == iptrust.ProxyOpen {
        // ...
    }
    if res.Proxy.LastDetected.Valid {  // OptionalTime
        fmt.Println(res.Proxy.LastDetected.Value.Format(time.DateOnly))
    }
}
```

A few conventions apply throughout:

- **Sections** have a `Valid bool`. Reading fields of an invalid section is safe; they are just zero values.
- **Optional fields** inside a section are wrapped in `iptrust.Optional[T]`, a struct with `Valid` and `Value`. Named instances cover the common cases: `OptionalString`, `OptionalFloat`, and `Optional[OrganizationType]`, `Optional[RiskLabel]`, `Optional[BotType]`, `Optional[ProxyType]` for the enumerations. A JSON `null` or missing key leaves `Valid` false.
- **Optional timestamps** use `iptrust.OptionalTime`, which works the same way. It accepts full RFC 3339 timestamps and bare dates. Required timestamps on the database types are plain `time.Time`, zero if the API omitted them.
- **Optional lists** such as `ASN.RiskTypes` and `Location.Area.Territories` are ordinary slices that are `nil` when absent.
- **Enumerated fields** are typed strings with constants for the documented values. Unknown values pass through unchanged, so compare rather than assume.

Full field reference: https://iptrust.co/docs

## Database downloads

For plans that include database downloads.

```go
// What's available and what your plan covers
dbs, err := client.ListDatabases(ctx)
for _, db := range dbs {
    fmt.Println(db.Type, db.Entitled, len(db.Formats))
}

// Get a signed URL (valid about an hour, see ExpiresAt)
dl, err := client.CreateDatabaseDownload(ctx, iptrust.DatabaseGeolocation, iptrust.FormatMMDB)
fmt.Println(dl.URL, dl.Filename, dl.SizeBytes)

// Or fetch the file in one step and stream it to disk
dl, body, err := client.DownloadDatabase(ctx, iptrust.DatabaseGeolocation, iptrust.FormatMMDB)
if err != nil {
    log.Fatal(err)
}
defer body.Close()

f, _ := os.Create(dl.Filename)
defer f.Close()
io.Copy(f, body)
```

The signed URL carries its own authorisation. Anyone holding it can download the file until it expires, so treat it like a credential. The client timeout is not applied to the file transfer itself; use a context deadline if you need one.

## Errors

Any non-2xx response is returned as an `*iptrust.Error`:

```go
var apiErr *iptrust.Error
if errors.As(err, &apiErr) {
    apiErr.StatusCode // 400
    apiErr.Code       // e.g. "bad_request", "token_invalid", "request_limit_exceeded"
    apiErr.Detail     // human-readable message from the API
    apiErr.ErrorID    // set on 5xx responses, quote it to support
}
```

| Status | Meaning |
|--------|---------|
| 400 | Invalid IP, or unknown database type/format |
| 401 | Missing or invalid API key |
| 403 | Your plan does not include this resource |
| 404 | Database not published in that format |
| 429 | Quota or rate limit exceeded |
| 500 | Server error, retry or contact support |

Timeouts and cancellations surface as wrapped `context.DeadlineExceeded` or `context.Canceled`, so `errors.Is` works. Transport failures from the HTTP client are wrapped and returned as-is.

## Options

```go
client, err := iptrust.NewClient(apiKey,
    iptrust.WithBaseURL("https://api.iptrust.co"), // override for testing
    iptrust.WithTimeout(10*time.Second),           // per request, 0 to disable
    iptrust.WithHTTPClient(myHTTPClient),          // custom *http.Client
    iptrust.WithUserAgent("my-app/1.0"),           // default is "iptrust-go/<version>"
    iptrust.WithHeader("X-Request-Source", "web"), // extra header on every API request
)
```

Every method takes a `context.Context`, so per-request cancellation and deadlines work the usual way. The `*Client` is safe for concurrent use.

## License

MIT
