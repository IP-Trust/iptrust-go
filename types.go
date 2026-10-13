package iptrust

import (
	"encoding/json"
	"time"
)

// OrganizationType classifies an AS or company.
type OrganizationType string

// Known OrganizationType values.
const (
	OrganizationISP        OrganizationType = "isp"
	OrganizationHosting    OrganizationType = "hosting"
	OrganizationBusiness   OrganizationType = "business"
	OrganizationGovernment OrganizationType = "government"
	OrganizationEducation  OrganizationType = "education"
)

// RiskLabel is the risk level derived from the numeric ASN.Risk score.
type RiskLabel string

// Known RiskLabel values.
const (
	RiskVeryLow  RiskLabel = "very_low"
	RiskLow      RiskLabel = "low"
	RiskModerate RiskLabel = "moderate"
	RiskHigh     RiskLabel = "high"
	RiskVeryHigh RiskLabel = "very_high"
)

// RiskType is a risk signal category attached to an AS.
type RiskType string

// Known RiskType values.
const (
	RiskTypeConfirmedBot  RiskType = "confirmed_bot"
	RiskTypeSuspectedBot  RiskType = "suspected_bot"
	RiskTypeTrafficVolume RiskType = "traffic_volume"
	RiskTypeRepeatedAbuse RiskType = "repeated_abuse"
	RiskTypeTargetedAbuse RiskType = "targeted_abuse"
)

// LocationResolution is the precision of the geolocation data.
type LocationResolution string

// Known LocationResolution values.
const (
	ResolutionCity    LocationResolution = "city"
	ResolutionState   LocationResolution = "state"
	ResolutionCountry LocationResolution = "country"
)

// BotType is the category of a detected known bot.
type BotType string

// Known BotType values.
const (
	BotTypeCrawler BotType = "crawler"
	BotTypeAI      BotType = "ai"
	BotTypeOther   BotType = "other"
)

// ProxyType is a proxy classification.
type ProxyType string

// Known ProxyType values.
const (
	ProxyResidential ProxyType = "residential"
	ProxyOpen        ProxyType = "open"
)

// ASN holds Autonomous System information (Identity package).
type ASN struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Number is the Autonomous System Number, e.g. "AS8075".
	Number string `json:"number"`
	// Name is the registered name of the AS as listed by the RIR.
	Name string `json:"name"`
	// Company is the legal organization name that owns the AS.
	Company string `json:"company"`
	// Country is the ISO 3166-1 alpha-2 code where the AS is registered.
	Country string `json:"country"`
	// Route is the most specific BGP route prefix containing this IP.
	Route string `json:"route"`
	// RIR is the Regional Internet Registry, e.g. "ARIN", "RIPE", "APNIC".
	RIR string `json:"rir"`
	// Domain is the primary domain associated with the AS owner, if known.
	Domain OptionalString `json:"domain,omitzero"`
	// Type is the AS classification, if known.
	Type Optional[OrganizationType] `json:"type,omitzero"`
	// Description is a human-readable description of the AS owner.
	Description OptionalString `json:"description,omitzero"`
	// Risk is a score from 0.0 (very low) to 1.0 (very high). Risk.Valid is
	// false when not enough traffic has been observed to score the AS.
	Risk OptionalFloat `json:"risk,omitzero"`
	// RiskLabel is the risk level label, present when Risk is.
	RiskLabel Optional[RiskLabel] `json:"risk_label,omitzero"`
	// RiskTypes lists the risk signal categories. It is nil when the API
	// reported none.
	RiskTypes []RiskType `json:"risk_types,omitempty"`
}

// Area is a circle described by its centre and radius.
type Area struct {
	// Latitude of the centre (WGS84).
	Latitude float64 `json:"latitude"`
	// Longitude of the centre (WGS84).
	Longitude float64 `json:"longitude"`
	// RadiusKM is the radius in kilometres.
	RadiusKM float64 `json:"radius_km"`
}

// LocationArea is the area the IP is expected to be within. It matches
// Location.Resolution and is intended for checks such as impossible travel
// detection.
type LocationArea struct {
	Area
	// Territories is only present when the resolution is country and the
	// country has territories outside the main area, such as Alaska and
	// Hawaii for the United States.
	// It is nil otherwise.
	Territories []Area `json:"territories,omitempty"`
}

// LocationMeta is country metadata attached to a location.
type LocationMeta struct {
	// CountryISO2 is the ISO 3166-1 alpha-2 code.
	CountryISO2 string `json:"country_iso2"`
	// CountryISO3 is the ISO 3166-1 alpha-3 code.
	CountryISO3 string `json:"country_iso3"`
	// CountryNumericCode is the ISO 3166-1 numeric code.
	CountryNumericCode string `json:"country_numeric_code"`
	// CountryPhoneCode is the international dialling code.
	CountryPhoneCode string `json:"country_phone_code"`
	// CountryCapital is the capital city.
	CountryCapital string `json:"country_capital"`
	// CountryCurrency is the ISO 4217 currency code.
	CountryCurrency string `json:"country_currency"`
	// CountryCurrencyName is the full currency name.
	CountryCurrencyName string `json:"country_currency_name"`
	// CountryNative is the country name in the local language.
	CountryNative string `json:"country_native"`
	// CountryRegion is the UN geographic macro-region.
	CountryRegion string `json:"country_region"`
	// CountrySubregion is the UN geographic sub-region.
	CountrySubregion string `json:"country_subregion"`
	// CountryEmoji is the country flag emoji.
	CountryEmoji string `json:"country_emoji"`
	// CountryEmojiU is the country flag as Unicode code points.
	CountryEmojiU string `json:"country_emoji_u"`
}

// Location holds geolocation information (Geolocation package).
type Location struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Resolution is the precision of the location data.
	Resolution LocationResolution `json:"resolution"`
	// City is the city-level location. When Resolution is state or country,
	// this is a suitable default city for that state or country.
	City string `json:"city"`
	// State is the state, province, or top-level administrative region.
	State string `json:"state"`
	// Country is the full country name.
	Country string `json:"country"`
	// Latitude is the approximate latitude of the city (WGS84).
	Latitude float64 `json:"latitude"`
	// Longitude is the approximate longitude of the city (WGS84).
	Longitude float64 `json:"longitude"`
	// Area is the area the IP is expected to be within.
	Area LocationArea `json:"area"`
	// Meta is country metadata.
	Meta LocationMeta `json:"meta"`
}

// Company is the company operating at this IP (Identity package).
type Company struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Name is the legal name of the company.
	Name string `json:"name"`
	// Type is the company classification, if known.
	Type Optional[OrganizationType] `json:"type,omitzero"`
	// Domain is the primary website domain, if known.
	Domain OptionalString `json:"domain,omitzero"`
	// Description is a brief description of the company, if known.
	Description OptionalString `json:"description,omitzero"`
}

// KnownBot holds known bot and crawler detection (Identity package).
type KnownBot struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Detected reports whether a known bot or crawler was detected.
	Detected bool `json:"detected"`
	// BotType is the bot category, when detected.
	BotType Optional[BotType] `json:"bot_type,omitzero"`
	// Name identifies the bot, e.g. "google_bot", "gpt_bot", "anthropic".
	// The list grows over time; see https://iptrust.co/docs for current names.
	Name OptionalString `json:"name,omitzero"`
}

// Hosting holds hosting and cloud provider detection (Identity package).
type Hosting struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Detected reports whether the IP belongs to a known hosting provider.
	Detected bool `json:"detected"`
	// Domain is the website domain of the hosting provider, when detected.
	Domain OptionalString `json:"domain,omitzero"`
	// Name is the name of the hosting or cloud provider, when detected.
	Name OptionalString `json:"name,omitzero"`
}

// AbuseContact is the abuse contact for the network (Identity package).
type AbuseContact struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Address is the physical mailing address.
	Address OptionalString `json:"address,omitzero"`
	// Email is the address for reporting abuse.
	Email OptionalString `json:"email,omitzero"`
	// Name is the abuse contact person or team.
	Name OptionalString `json:"name,omitzero"`
}

// TorExitNode holds Tor exit node detection (Anonymization package).
type TorExitNode struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Detected reports whether this IP is a known Tor exit node.
	Detected bool `json:"detected"`
}

// PrivacyRelay holds commercial privacy relay detection (Anonymization package).
type PrivacyRelay struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Detected reports whether the IP is part of a commercial privacy relay.
	Detected bool `json:"detected"`
	// Provider identifies the relay, e.g. "icloud_private_relay".
	Provider OptionalString `json:"provider,omitzero"`
}

// Proxy holds proxy detection (Anonymization package).
type Proxy struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Detected reports whether this IP is a detected proxy server.
	Detected bool `json:"detected"`
	// Type is the proxy classification, when detected.
	Type Optional[ProxyType] `json:"type,omitzero"`
	// LastDetected is when proxy activity was last observed. Valid is false
	// when the API did not report a date.
	LastDetected OptionalTime `json:"last_detected,omitzero"`
}

// VPN holds VPN detection (Anonymization package).
type VPN struct {
	// Valid is true when this section was present in the response. It is
	// false when your plan does not include it.
	Valid bool `json:"-"`
	// Detected reports whether this IP is part of a VPN service.
	Detected bool `json:"detected"`
	// Provider identifies the VPN provider, e.g. "nord_vpn".
	Provider OptionalString `json:"provider,omitzero"`
	// LastDetected is when VPN activity was last observed. Valid is false
	// when the API did not report a date.
	LastDetected OptionalTime `json:"last_detected,omitzero"`
}

// IPResponse is the response from GET /ip/{ip}.
//
// The shape is the same for every plan. Sections not included in your plan
// are left at their zero value with Valid set to false.
type IPResponse struct {
	// IP is the queried address, echoed back.
	IP string `json:"ip"`
	// IsBogon reports whether the IP is in a bogon (unroutable or reserved) range.
	IsBogon      bool         `json:"is_bogon"`
	ASN          ASN          `json:"asn"`
	Location     Location     `json:"location"`
	Company      Company      `json:"company"`
	KnownBot     KnownBot     `json:"known_bot"`
	Hosting      Hosting      `json:"hosting"`
	AbuseContact AbuseContact `json:"abuse_contact"`
	TorExitNode  TorExitNode  `json:"tor_exit_node"`
	PrivacyRelay PrivacyRelay `json:"privacy_relay"`
	Proxy        Proxy        `json:"proxy"`
	VPN          VPN          `json:"vpn"`
}

// DatabaseType identifies a downloadable database.
type DatabaseType string

// Known DatabaseType values.
const (
	DatabaseASN           DatabaseType = "asn"
	DatabaseCompany       DatabaseType = "company"
	DatabaseGeolocation   DatabaseType = "geolocation"
	DatabaseAbuseContact  DatabaseType = "abusecontact"
	DatabaseKnownBots     DatabaseType = "knownbots"
	DatabaseHosting       DatabaseType = "hosting"
	DatabaseAnonymization DatabaseType = "anonymization"
)

// DatabaseFormat is the file format of a downloadable database.
type DatabaseFormat string

// Known DatabaseFormat values.
const (
	// FormatMMDB is a MaxMind DB file.
	FormatMMDB DatabaseFormat = "mmdb"
	// FormatCSV is a gzipped CSV export.
	FormatCSV DatabaseFormat = "csv"
)

// DatabaseFile describes one published format of a database.
type DatabaseFile struct {
	Format DatabaseFormat `json:"format"`
	// SizeBytes is the file size as published. For CSV this is the
	// compressed size.
	SizeBytes int64 `json:"size_bytes"`
	// UpdatedAt is when this file was published (UTC). It is the zero time
	// if the API omitted it.
	UpdatedAt time.Time `json:"updated_at"`
}

// Database is one entry from GET /database.
type Database struct {
	Type DatabaseType `json:"type"`
	// Entitled reports whether your plan includes downloads of this database.
	Entitled bool `json:"entitled"`
	// UpdatedAt is when the most recently refreshed file was published (UTC).
	// It is the zero time if the API omitted it.
	UpdatedAt time.Time `json:"updated_at"`
	// Formats lists the published formats.
	Formats []DatabaseFile `json:"formats"`
}

// DatabaseDownload is the response from POST /database/download.
type DatabaseDownload struct {
	Type   DatabaseType   `json:"type"`
	Format DatabaseFormat `json:"format"`
	// Filename is a suggested filename for the downloaded file.
	Filename string `json:"filename"`
	// URL is the signed download URL. It needs no API key, and anyone
	// holding it can download the file until it expires, so treat it as a
	// credential.
	URL string `json:"url"`
	// SizeBytes is the file size in bytes.
	SizeBytes int64 `json:"size_bytes"`
	// UpdatedAt is when the file was published (UTC). It is the zero time if
	// the API omitted it.
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt is when the signed URL stops working (UTC). It is the zero
	// time if the API omitted it.
	ExpiresAt time.Time `json:"expires_at"`
}

type databaseListResponse struct {
	Databases []Database `json:"databases"`
}

type downloadRequest struct {
	Type   DatabaseType   `json:"type"`
	Format DatabaseFormat `json:"format"`
}

// apiErrorBody is the error body returned by every endpoint on failure.
type apiErrorBody struct {
	Error struct {
		ErrorCode string          `json:"error_code"`
		ErrorID   string          `json:"error_id"`
		Detail    string          `json:"detail"`
		Data      json.RawMessage `json:"data"`
	} `json:"error"`
}
