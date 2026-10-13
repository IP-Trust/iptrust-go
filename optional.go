package iptrust

import (
	"bytes"
	"encoding/json"
	"time"
)

// Optional is a value that may be absent from an API response. Valid is
// false when the field was missing or null, in which case Value is the zero
// value of T.
//
// Check Valid before reading Value. Named instances exist for the common
// cases: OptionalString, OptionalFloat and the enumerations such as
// Optional[OrganizationType].
type Optional[T any] struct {
	Valid bool
	Value T
}

// MarshalJSON implements json.Marshaler. An invalid value encodes as null.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(o.Value)
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		*o = Optional[T]{}
		return nil
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*o = Optional[T]{Valid: true, Value: v}
	return nil
}

// OptionalString is a string that may be absent from an API response.
type OptionalString = Optional[string]

// OptionalFloat is a float64 that may be absent from an API response.
type OptionalFloat = Optional[float64]

// OptionalTime is a time.Time that may be absent from an API response.
// Valid is false when the field was missing, null or empty. When decoding,
// it accepts RFC 3339 timestamps and bare dates (2006-01-02).
type OptionalTime struct {
	Valid bool
	Value time.Time
}

// MarshalJSON implements json.Marshaler. An invalid value encodes as null.
func (o OptionalTime) MarshalJSON() ([]byte, error) {
	if !o.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(o.Value)
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *OptionalTime) UnmarshalJSON(b []byte) error {
	var ts timestamp
	if err := ts.UnmarshalJSON(b); err != nil {
		return err
	}
	t := time.Time(ts)
	if t.IsZero() {
		*o = OptionalTime{}
		return nil
	}
	*o = OptionalTime{Valid: true, Value: t}
	return nil
}
