package iptrust

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// timestampLayouts are the formats accepted for API timestamps and dates.
var timestampLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// parseTimestamp parses an ISO 8601 timestamp or date. An empty string
// yields the zero time.
func parseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range timestampLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("iptrust: cannot parse timestamp %q", s)
}

// timestamp is a time.Time with lenient decoding: it accepts RFC 3339, a
// bare date, an empty string or null, the last three decoding to the zero
// time. It backs OptionalTime and is used in shadow structs during
// unmarshalling so required timestamps can be plain time.Time fields.
type timestamp time.Time

func (t *timestamp) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*t = timestamp{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := parseTimestamp(s)
	if err != nil {
		return err
	}
	*t = timestamp(parsed)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (p *Proxy) UnmarshalJSON(b []byte) error {
	type plain Proxy
	if err := json.Unmarshal(b, (*plain)(p)); err != nil {
		return err
	}
	p.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (v *VPN) UnmarshalJSON(b []byte) error {
	type plain VPN
	if err := json.Unmarshal(b, (*plain)(v)); err != nil {
		return err
	}
	v.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (a *ASN) UnmarshalJSON(b []byte) error {
	type plain ASN
	if err := json.Unmarshal(b, (*plain)(a)); err != nil {
		return err
	}
	a.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (l *Location) UnmarshalJSON(b []byte) error {
	type plain Location
	if err := json.Unmarshal(b, (*plain)(l)); err != nil {
		return err
	}
	l.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (c *Company) UnmarshalJSON(b []byte) error {
	type plain Company
	if err := json.Unmarshal(b, (*plain)(c)); err != nil {
		return err
	}
	c.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (k *KnownBot) UnmarshalJSON(b []byte) error {
	type plain KnownBot
	if err := json.Unmarshal(b, (*plain)(k)); err != nil {
		return err
	}
	k.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (h *Hosting) UnmarshalJSON(b []byte) error {
	type plain Hosting
	if err := json.Unmarshal(b, (*plain)(h)); err != nil {
		return err
	}
	h.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (a *AbuseContact) UnmarshalJSON(b []byte) error {
	type plain AbuseContact
	if err := json.Unmarshal(b, (*plain)(a)); err != nil {
		return err
	}
	a.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (t *TorExitNode) UnmarshalJSON(b []byte) error {
	type plain TorExitNode
	if err := json.Unmarshal(b, (*plain)(t)); err != nil {
		return err
	}
	t.Valid = true
	return nil
}

// UnmarshalJSON implements json.Unmarshaler and sets Valid.
func (p *PrivacyRelay) UnmarshalJSON(b []byte) error {
	type plain PrivacyRelay
	if err := json.Unmarshal(b, (*plain)(p)); err != nil {
		return err
	}
	p.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler. Sections with Valid set to false
// are omitted, matching the API's wire format.
func (r IPResponse) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"ip":       r.IP,
		"is_bogon": r.IsBogon,
	}
	type plainASN ASN
	type plainLocation Location
	type plainCompany Company
	type plainKnownBot KnownBot
	type plainHosting Hosting
	type plainAbuseContact AbuseContact
	type plainTorExitNode TorExitNode
	type plainPrivacyRelay PrivacyRelay
	type plainProxy Proxy
	type plainVPN VPN
	if r.ASN.Valid {
		out["asn"] = plainASN(r.ASN)
	}
	if r.Location.Valid {
		out["location"] = plainLocation(r.Location)
	}
	if r.Company.Valid {
		out["company"] = plainCompany(r.Company)
	}
	if r.KnownBot.Valid {
		out["known_bot"] = plainKnownBot(r.KnownBot)
	}
	if r.Hosting.Valid {
		out["hosting"] = plainHosting(r.Hosting)
	}
	if r.AbuseContact.Valid {
		out["abuse_contact"] = plainAbuseContact(r.AbuseContact)
	}
	if r.TorExitNode.Valid {
		out["tor_exit_node"] = plainTorExitNode(r.TorExitNode)
	}
	if r.PrivacyRelay.Valid {
		out["privacy_relay"] = plainPrivacyRelay(r.PrivacyRelay)
	}
	if r.Proxy.Valid {
		out["proxy"] = plainProxy(r.Proxy)
	}
	if r.VPN.Valid {
		out["vpn"] = plainVPN(r.VPN)
	}
	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *DatabaseFile) UnmarshalJSON(b []byte) error {
	type plain DatabaseFile
	aux := struct {
		*plain
		UpdatedAt timestamp `json:"updated_at"`
	}{plain: (*plain)(f)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	f.UpdatedAt = time.Time(aux.UpdatedAt)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Database) UnmarshalJSON(b []byte) error {
	type plain Database
	aux := struct {
		*plain
		UpdatedAt timestamp `json:"updated_at"`
	}{plain: (*plain)(d)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	d.UpdatedAt = time.Time(aux.UpdatedAt)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DatabaseDownload) UnmarshalJSON(b []byte) error {
	type plain DatabaseDownload
	aux := struct {
		*plain
		UpdatedAt timestamp `json:"updated_at"`
		ExpiresAt timestamp `json:"expires_at"`
	}{plain: (*plain)(d)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	d.UpdatedAt = time.Time(aux.UpdatedAt)
	d.ExpiresAt = time.Time(aux.ExpiresAt)
	return nil
}
