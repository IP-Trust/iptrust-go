// Package iptrust is the official Go client for the IP Trust API
// (https://iptrust.co): IP geolocation, ASN, company, hosting, known bot,
// VPN, proxy, Tor and privacy relay detection, plus database downloads.
//
// Create a client with your API key and look up an address:
//
//	client, err := iptrust.NewClient("your-api-key")
//	if err != nil {
//		log.Fatal(err)
//	}
//	res, err := client.LookupIP(ctx, "9.9.9.9")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.Location.Valid {
//		fmt.Println(res.Location.Country)
//	}
//
// The lookup response has one shape for every plan. Each section is a plain
// struct with a Valid field that is false when your plan does not include
// it, so check Valid before reading its fields. Optional values inside a
// section use the Optional wrappers (OptionalString, OptionalFloat,
// OptionalTime and Optional[T] for enumerations) with the same Valid
// convention. Required timestamps are plain time.Time values.
//
// Any non-2xx response is returned as an *Error, which carries the HTTP
// status and the machine-readable error code from the API:
//
//	var apiErr *iptrust.Error
//	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests {
//		// back off
//	}
package iptrust
