package iptrust

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const sampleIP = `{
  "ip": "9.9.9.9",
  "is_bogon": false,
  "asn": {
    "number": "AS19281", "name": "QUAD9-AS-1", "company": "Quad9", "country": "US",
    "route": "9.9.9.0/24", "rir": "ARIN", "type": "business", "risk": 0.1, "risk_label": "very_low"
  },
  "location": {
    "resolution": "city", "city": "Berkeley", "state": "California", "country": "United States",
    "latitude": 37.87, "longitude": -122.27,
    "area": {"latitude": 37.87, "longitude": -122.27, "radius_km": 20,
             "territories": [{"latitude": 64.2, "longitude": -149.5, "radius_km": 1000}]},
    "meta": {
      "country_iso2": "US", "country_iso3": "USA", "country_numeric_code": "840",
      "country_phone_code": "1", "country_capital": "Washington", "country_currency": "USD",
      "country_currency_name": "US Dollar", "country_native": "United States",
      "country_region": "Americas", "country_subregion": "Northern America",
      "country_emoji": "🇺🇸", "country_emoji_u": "U+1F1FA U+1F1F8"
    }
  },
  "vpn": {"detected": false},
  "proxy": {"detected": true, "type": "open", "last_detected": "2026-03-04"}
}`

type recorded struct {
	method string
	path   string
	rawURL string
	header http.Header
	body   []byte
}

// newServer returns a test server that records the last request and replies
// with the given status and body.
func newServer(t *testing.T, status int, body string) (*httptest.Server, *recorded) {
	t.Helper()
	rec := &recorded{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path
		rec.rawURL = r.URL.RequestURI()
		rec.header = r.Header.Clone()
		rec.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

func newTestClient(t *testing.T, srv *httptest.Server, opts ...Option) *Client {
	t.Helper()
	opts = append([]Option{WithBaseURL(srv.URL)}, opts...)
	c, err := NewClient("test-key", opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestNewClientRequiresKey(t *testing.T) {
	t.Setenv("IPTRUST_API_KEY", "env-key") // must be ignored
	if _, err := NewClient(""); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("got %v, want ErrMissingAPIKey", err)
	}
	if _, err := NewClient("   "); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("got %v, want ErrMissingAPIKey", err)
	}
}

func TestWithBaseURLStripsTrailingSlash(t *testing.T) {
	srv, rec := newServer(t, 200, sampleIP)
	c := newTestClient(t, srv, WithBaseURL(srv.URL+"///"))
	if _, err := c.LookupIP(context.Background(), "9.9.9.9"); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/ip/9.9.9.9" {
		t.Fatalf("path = %q", rec.path)
	}
}

func TestLookupIP(t *testing.T) {
	srv, rec := newServer(t, 200, sampleIP)
	c := newTestClient(t, srv)

	res, err := c.LookupIP(context.Background(), "9.9.9.9")
	if err != nil {
		t.Fatal(err)
	}

	if rec.method != http.MethodGet || rec.path != "/ip/9.9.9.9" {
		t.Fatalf("request = %s %s", rec.method, rec.path)
	}
	if got := rec.header.Get("X-API-Key"); got != "test-key" {
		t.Fatalf("X-API-Key = %q", got)
	}
	if got := rec.header.Get("User-Agent"); got != "iptrust-go/"+Version {
		t.Fatalf("User-Agent = %q", got)
	}
	if got := rec.header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q", got)
	}
	if len(rec.body) != 0 {
		t.Fatalf("unexpected body %q", rec.body)
	}

	if res.IP != "9.9.9.9" || res.IsBogon {
		t.Fatalf("top level = %+v", res)
	}
	if !res.ASN.Valid || res.ASN.Number != "AS19281" || res.ASN.Type != (Optional[OrganizationType]{Valid: true, Value: OrganizationBusiness}) {
		t.Fatalf("asn = %+v", res.ASN)
	}
	if res.ASN.Domain.Valid || res.ASN.Description.Valid || res.ASN.RiskTypes != nil {
		t.Fatalf("absent asn fields should be invalid: %+v", res.ASN)
	}
	if res.ASN.Risk != (OptionalFloat{Valid: true, Value: 0.1}) || res.ASN.RiskLabel != (Optional[RiskLabel]{Valid: true, Value: RiskVeryLow}) {
		t.Fatalf("risk = %+v", res.ASN)
	}
	if !res.Location.Valid || res.Location.Resolution != ResolutionCity || res.Location.Meta.CountryISO2 != "US" {
		t.Fatalf("location = %+v", res.Location)
	}
	if res.Location.Area.RadiusKM != 20 || len(res.Location.Area.Territories) != 1 {
		t.Fatalf("area = %+v", res.Location.Area)
	}
	if !res.VPN.Valid || res.VPN.Detected || res.VPN.LastDetected.Valid || res.VPN.Provider.Valid {
		t.Fatalf("vpn = %+v", res.VPN)
	}
	if !res.Proxy.Valid || !res.Proxy.Detected || res.Proxy.Type != (Optional[ProxyType]{Valid: true, Value: ProxyOpen}) {
		t.Fatalf("proxy = %+v", res.Proxy)
	}
	want := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
	if !res.Proxy.LastDetected.Valid || !res.Proxy.LastDetected.Value.Equal(want) {
		t.Fatalf("proxy.last_detected = %v", res.Proxy.LastDetected)
	}
	if res.Company.Valid || res.KnownBot.Valid || res.Hosting.Valid || res.AbuseContact.Valid ||
		res.TorExitNode.Valid || res.PrivacyRelay.Valid {
		t.Fatalf("absent sections should not be valid: %+v", res)
	}
}

func TestIPResponseRoundTrip(t *testing.T) {
	var res IPResponse
	if err := json.Unmarshal([]byte(sampleIP), &res); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"ip", "is_bogon", "asn", "location", "vpn", "proxy"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing %q in marshalled output", k)
		}
	}
	for _, k := range []string{"company", "known_bot", "hosting", "tor_exit_node"} {
		if _, ok := m[k]; ok {
			t.Errorf("invalid section %q should be omitted", k)
		}
	}

	var again IPResponse
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatal(err)
	}
	if !again.ASN.Valid || again.ASN.Risk != (OptionalFloat{Valid: true, Value: 0.1}) || again.Company.Valid || again.ASN.Domain.Valid {
		t.Fatalf("round trip = %+v", again)
	}
	if again.Proxy.LastDetected != res.Proxy.LastDetected {
		t.Fatalf("last_detected round trip = %v", again.Proxy.LastDetected)
	}
}

func TestOptionalString(t *testing.T) {
	var v struct {
		Name OptionalString `json:"name"`
	}
	for in, want := range map[string]OptionalString{
		`{"name":"x"}`:  {Valid: true, Value: "x"},
		`{"name":""}`:   {Valid: true, Value: ""},
		`{"name":null}`: {},
		`{}`:            {},
	} {
		v.Name = OptionalString{Valid: true, Value: "stale"}
		if err := json.Unmarshal([]byte(in), &v); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if in == `{}` {
			continue // missing keys leave the field untouched, as with any Go field
		}
		if v.Name != want {
			t.Errorf("%s: got %+v, want %+v", in, v.Name, want)
		}
	}
	if err := json.Unmarshal([]byte(`{"name":5}`), &v); err == nil {
		t.Error("expected error for non-string input")
	}
}

func TestOptionalFloat(t *testing.T) {
	cases := []struct {
		in   string
		want OptionalFloat
	}{
		{`{"risk": 0}`, OptionalFloat{Valid: true, Value: 0}},
		{`{"risk": 0.75}`, OptionalFloat{Valid: true, Value: 0.75}},
		{`{"risk": null}`, OptionalFloat{}},
		{`{}`, OptionalFloat{}},
	}
	for _, tc := range cases {
		var v struct {
			Risk OptionalFloat `json:"risk"`
		}
		if err := json.Unmarshal([]byte(tc.in), &v); err != nil {
			t.Fatalf("%s: %v", tc.in, err)
		}
		if v.Risk != tc.want {
			t.Errorf("%s: got %+v, want %+v", tc.in, v.Risk, tc.want)
		}
	}
	if err := json.Unmarshal([]byte(`"x"`), new(OptionalFloat)); err == nil {
		t.Error("expected error for non-numeric input")
	}
}

func TestTimestampParsing(t *testing.T) {
	cases := []struct {
		in   string
		want OptionalTime
	}{
		{`"2026-01-02T03:04:05Z"`, OptionalTime{Valid: true, Value: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}},
		{`"2026-01-02T03:04:05.123456Z"`, OptionalTime{Valid: true, Value: time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC)}},
		{`"2026-01-02T03:04:05+01:00"`, OptionalTime{Valid: true, Value: time.Date(2026, 1, 2, 2, 4, 5, 0, time.UTC)}},
		{`"2026-01-02T03:04:05"`, OptionalTime{Valid: true, Value: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}},
		{`"2026-01-02"`, OptionalTime{Valid: true, Value: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}},
		{`""`, OptionalTime{}},
		{`null`, OptionalTime{}},
	}
	for _, tc := range cases {
		var v VPN
		if err := json.Unmarshal([]byte(`{"detected":true,"last_detected":`+tc.in+`}`), &v); err != nil {
			t.Fatalf("%s: %v", tc.in, err)
		}
		if v.LastDetected.Valid != tc.want.Valid || !v.LastDetected.Value.Equal(tc.want.Value) {
			t.Errorf("%s: got %v, want %v", tc.in, v.LastDetected, tc.want)
		}
		if !v.Valid || !v.Detected {
			t.Errorf("%s: other fields lost: %+v", tc.in, v)
		}
	}

	var v VPN
	if err := json.Unmarshal([]byte(`{"detected":true,"last_detected":"yesterday"}`), &v); err == nil {
		t.Error("expected error for unparseable timestamp")
	}

	var dl DatabaseDownload
	if err := json.Unmarshal([]byte(`{"type":"asn","format":"csv","filename":"a","url":"u","size_bytes":1}`), &dl); err != nil {
		t.Fatal(err)
	}
	if !dl.UpdatedAt.IsZero() || !dl.ExpiresAt.IsZero() || dl.Filename != "a" || dl.SizeBytes != 1 {
		t.Fatalf("dl = %+v", dl)
	}
}

func TestLookupIPEscapesPath(t *testing.T) {
	srv, rec := newServer(t, 200, `{"ip":"2001:db8::1","is_bogon":false}`)
	c := newTestClient(t, srv)
	if _, err := c.LookupIP(context.Background(), " 2001:db8::1 "); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/ip/2001:db8::1" {
		t.Fatalf("path = %q", rec.path)
	}

	srv2, rec2 := newServer(t, 400, `{"error":{"error_code":"bad_request"}}`)
	c2 := newTestClient(t, srv2)
	_, err := c2.LookupIP(context.Background(), "../database?x=1")
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
		t.Fatalf("err = %v", err)
	}
	if rec2.rawURL != "/ip/..%2Fdatabase%3Fx=1" {
		t.Fatalf("raw URL = %q, input escaped the /ip/ route", rec2.rawURL)
	}
}

func TestLookupIPRejectsEmpty(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv)
	if _, err := c.LookupIP(context.Background(), "  "); err == nil {
		t.Fatal("expected error")
	}
	if called {
		t.Fatal("server should not be called")
	}
}

func TestAPIError(t *testing.T) {
	srv, _ := newServer(t, 401, `{"error":{"error_code":"token_invalid","detail":"invalid api key"}}`)
	c := newTestClient(t, srv)
	_, err := c.LookupIP(context.Background(), "9.9.9.9")

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T %v", err, err)
	}
	if apiErr.StatusCode != 401 || apiErr.Code != "token_invalid" || apiErr.Detail != "invalid api key" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
	if msg := err.Error(); !strings.Contains(msg, "401") || !strings.Contains(msg, "invalid api key") {
		t.Fatalf("message = %q", msg)
	}
}

func TestAPIErrorInternal(t *testing.T) {
	srv, _ := newServer(t, 500, `{"error":{"error_code":"internal_server_error","error_id":"abc-123","detail":"boom","data":{"x":1}}}`)
	c := newTestClient(t, srv)
	_, err := c.ListDatabases(context.Background())

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T %v", err, err)
	}
	if apiErr.StatusCode != 500 || apiErr.ErrorID != "abc-123" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
	var data map[string]int
	if err := json.Unmarshal(apiErr.Data, &data); err != nil || data["x"] != 1 {
		t.Fatalf("data = %s", apiErr.Data)
	}
}

func TestAPIErrorNonJSON(t *testing.T) {
	srv, _ := newServer(t, 502, "<html>Bad Gateway</html>")
	c := newTestClient(t, srv)
	_, err := c.LookupIP(context.Background(), "9.9.9.9")

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T %v", err, err)
	}
	if apiErr.StatusCode != 502 || apiErr.Code != "" || apiErr.Body != "<html>Bad Gateway</html>" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
	if !strings.Contains(err.Error(), "Bad Gateway") {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestInvalidJSONResponse(t *testing.T) {
	srv, _ := newServer(t, 200, "not json")
	c := newTestClient(t, srv)
	_, err := c.LookupIP(context.Background(), "9.9.9.9")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *Error
	if errors.As(err, &apiErr) {
		t.Fatalf("decode failure should not be an *Error: %v", err)
	}
}

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, WithTimeout(20*time.Millisecond))

	_, err := c.LookupIP(context.Background(), "9.9.9.9")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, WithTimeout(0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.LookupIP(ctx, "9.9.9.9")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestCustomHeadersAndUserAgent(t *testing.T) {
	srv, rec := newServer(t, 200, sampleIP)
	c := newTestClient(t, srv,
		WithUserAgent("my-app/1.0"),
		WithHeader("X-Extra", "yes"),
		WithHeader("X-API-Key", "must-not-override"),
	)
	if _, err := c.LookupIP(context.Background(), "9.9.9.9"); err != nil {
		t.Fatal(err)
	}
	if got := rec.header.Get("User-Agent"); got != "my-app/1.0" {
		t.Fatalf("User-Agent = %q", got)
	}
	if got := rec.header.Get("X-Extra"); got != "yes" {
		t.Fatalf("X-Extra = %q", got)
	}
	if got := rec.header.Get("X-API-Key"); got != "test-key" {
		t.Fatalf("X-API-Key = %q, extra headers must not override the key", got)
	}
}

func TestListDatabases(t *testing.T) {
	srv, rec := newServer(t, 200, `{"databases":[
		{"type":"geolocation","entitled":true,"updated_at":"2026-01-02T03:04:05Z",
		 "formats":[{"format":"mmdb","size_bytes":123,"updated_at":"2026-01-02T03:04:05Z"}]},
		{"type":"asn","entitled":false,"updated_at":"2026-01-01T00:00:00Z","formats":[]}
	]}`)
	c := newTestClient(t, srv)

	dbs, err := c.ListDatabases(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rec.method != http.MethodGet || rec.path != "/database" {
		t.Fatalf("request = %s %s", rec.method, rec.path)
	}
	if len(dbs) != 2 || dbs[0].Type != DatabaseGeolocation || !dbs[0].Entitled {
		t.Fatalf("dbs = %+v", dbs)
	}
	if len(dbs[0].Formats) != 1 || dbs[0].Formats[0].Format != FormatMMDB || dbs[0].Formats[0].SizeBytes != 123 {
		t.Fatalf("formats = %+v", dbs[0].Formats)
	}
	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if !dbs[0].UpdatedAt.Equal(want) {
		t.Fatalf("updated_at = %v", dbs[0].UpdatedAt)
	}
}

func TestCreateDatabaseDownload(t *testing.T) {
	srv, rec := newServer(t, 200, `{"type":"geolocation","format":"mmdb","filename":"geolocation.mmdb",
		"url":"https://files.example/x?sig=1","size_bytes":42,
		"updated_at":"2026-01-02T03:04:05Z","expires_at":"2026-01-02T04:04:05Z"}`)
	c := newTestClient(t, srv)

	dl, err := c.CreateDatabaseDownload(context.Background(), DatabaseGeolocation, FormatMMDB)
	if err != nil {
		t.Fatal(err)
	}
	if rec.method != http.MethodPost || rec.path != "/database/download" {
		t.Fatalf("request = %s %s", rec.method, rec.path)
	}
	if got := rec.header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["type"] != "geolocation" || body["format"] != "mmdb" {
		t.Fatalf("body = %s", rec.body)
	}
	if dl.Filename != "geolocation.mmdb" || dl.SizeBytes != 42 || dl.URL == "" {
		t.Fatalf("dl = %+v", dl)
	}
	if dl.ExpiresAt.Sub(dl.UpdatedAt) != time.Hour {
		t.Fatalf("expires_at = %v", dl.ExpiresAt)
	}
}

func TestDownloadDatabase(t *testing.T) {
	var fileHeader http.Header
	files := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileHeader = r.Header.Clone()
		_, _ = io.WriteString(w, "file-bytes")
	}))
	t.Cleanup(files.Close)

	srv, _ := newServer(t, 200, `{"type":"asn","format":"csv","filename":"asn.csv.gz",
		"url":"`+files.URL+`/asn.csv.gz?sig=1","size_bytes":10,
		"updated_at":"2026-01-02T03:04:05Z","expires_at":"2026-01-02T04:04:05Z"}`)
	c := newTestClient(t, srv)

	dl, body, err := c.DownloadDatabase(context.Background(), DatabaseASN, FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "file-bytes" || dl.Filename != "asn.csv.gz" {
		t.Fatalf("data = %q, dl = %+v", data, dl)
	}
	if fileHeader.Get("X-API-Key") != "" {
		t.Fatal("API key must not be sent to the signed URL")
	}
	if fileHeader.Get("User-Agent") != "iptrust-go/"+Version {
		t.Fatalf("User-Agent = %q", fileHeader.Get("User-Agent"))
	}
}

func TestDownloadDatabaseFailedFetch(t *testing.T) {
	files := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "expired", http.StatusForbidden)
	}))
	t.Cleanup(files.Close)

	srv, _ := newServer(t, 200, `{"type":"asn","format":"csv","filename":"asn.csv.gz",
		"url":"`+files.URL+`/x","size_bytes":10,
		"updated_at":"2026-01-02T03:04:05Z","expires_at":"2026-01-02T04:04:05Z"}`)
	c := newTestClient(t, srv)

	_, body, err := c.DownloadDatabase(context.Background(), DatabaseASN, FormatCSV)
	if body != nil {
		t.Fatal("body should be nil on error")
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 403 || apiErr.Body != "expired" {
		t.Fatalf("err = %v", err)
	}
}
