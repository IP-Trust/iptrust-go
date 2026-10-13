package iptrust

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production API endpoint.
	DefaultBaseURL = "https://api.iptrust.co"
	// DefaultTimeout is the per-request timeout applied when none is set.
	DefaultTimeout = 10 * time.Second
	// Version is the SDK version, sent in the User-Agent header.
	Version = "0.1.0"

	defaultUserAgent = "iptrust-go/" + Version
)

// Client talks to the IP Trust API. It is safe for concurrent use.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	userAgent  string
	headers    http.Header
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL. Useful for tests and proxies.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient sets the *http.Client used for requests. The default is
// http.DefaultClient. Request timeouts are enforced by the SDK via context,
// so the client's own Timeout may be left at zero.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithTimeout sets the per-request timeout. It is applied in addition to any
// deadline on the context passed to each call. Zero disables it.
// The default is DefaultTimeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithUserAgent replaces the default User-Agent header ("iptrust-go/<version>").
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithHeader adds a header to every API request. Calling it with the same
// key twice replaces the earlier value.
func WithHeader(key, value string) Option {
	return func(c *Client) { c.headers.Set(key, value) }
}

// NewClient returns a Client authenticated with apiKey. It returns
// ErrMissingAPIKey if apiKey is empty.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: http.DefaultClient,
		timeout:    DefaultTimeout,
		userAgent:  defaultUserAgent,
		headers:    http.Header{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// LookupIP returns intelligence for an IPv4 or IPv6 address.
//
// A non-2xx response is returned as an *Error: 400 for an invalid IP, 401
// for a bad key, 403 when your plan does not include the resource, 429 when
// the quota or rate limit is exceeded, and 500 for a server error.
func (c *Client) LookupIP(ctx context.Context, ip string) (IPResponse, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return IPResponse{}, errors.New("iptrust: an IP address is required")
	}
	var out IPResponse
	if err := c.do(ctx, http.MethodGet, "/ip/"+url.PathEscape(ip), nil, &out); err != nil {
		return IPResponse{}, err
	}
	return out, nil
}

// ListDatabases lists the databases available for download, including ones
// your plan does not include (see Database.Entitled).
func (c *Client) ListDatabases(ctx context.Context) ([]Database, error) {
	var out databaseListResponse
	if err := c.do(ctx, http.MethodGet, "/database", nil, &out); err != nil {
		return nil, err
	}
	return out.Databases, nil
}

// CreateDatabaseDownload generates a temporary signed URL for downloading a
// database file. The URL is valid for about an hour; check ExpiresAt.
//
// Returns an *Error with status 400 for an unknown type or format, 403 when
// your plan excludes the database, and 404 when it is not published in that
// format.
func (c *Client) CreateDatabaseDownload(ctx context.Context, typ DatabaseType, format DatabaseFormat) (DatabaseDownload, error) {
	var out DatabaseDownload
	body := downloadRequest{Type: typ, Format: format}
	if err := c.do(ctx, http.MethodPost, "/database/download", body, &out); err != nil {
		return DatabaseDownload{}, err
	}
	return out, nil
}

// DownloadDatabase requests a signed URL and fetches the file in one step.
// The caller must close the returned body. The client timeout is not
// applied to the file transfer, so pass a context with a deadline if you
// need one.
//
//	dl, body, err := client.DownloadDatabase(ctx, iptrust.DatabaseGeolocation, iptrust.FormatMMDB)
//	if err != nil {
//		return err
//	}
//	defer body.Close()
//	f, err := os.Create(dl.Filename)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	_, err = io.Copy(f, body)
func (c *Client) DownloadDatabase(ctx context.Context, typ DatabaseType, format DatabaseFormat) (DatabaseDownload, io.ReadCloser, error) {
	dl, err := c.CreateDatabaseDownload(ctx, typ, format)
	if err != nil {
		return DatabaseDownload{}, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dl.URL, nil)
	if err != nil {
		return DatabaseDownload{}, nil, fmt.Errorf("iptrust: build download request: %w", err)
	}
	// The signed URL carries its own auth; do not send the API key to it.
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return DatabaseDownload{}, nil, fmt.Errorf("iptrust: download database: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		return DatabaseDownload{}, nil, newErrorFromResponse(resp)
	}
	return dl, resp.Body, nil
}

// do performs an API request and decodes a JSON response into out.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("iptrust: encode request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("iptrust: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	for k, vs := range c.headers {
		req.Header[k] = append([]string(nil), vs...)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("iptrust: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return newErrorFromResponse(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("iptrust: decode response: %w", err)
	}
	return nil
}
