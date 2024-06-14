package model

import (
	"encoding/json"
	"errors"
	"time"
)

type LotusRequest struct {
	Jsonrpc string     `json:"jsonrpc"`
	Method  string     `json:"method"`
	Params  rawMessage `json:"params"`
	ID      int        `json:"id"`
}

type rawMessage []byte

// MarshalJSON returns m as the JSON encoding of m.
func (m rawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *rawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

// Response defines a JSON RPC response from the spec
// http://www.jsonrpc.org/specification#response_object
type LotusResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	ID      interface{}     `json:"id"`
	Error   *LotusRespError `json:"error,omitempty"`
}

type LotusRespError struct {
	Code    errorCode       `json:"code"`
	Message string          `json:"message"`
	Meta    json.RawMessage `json:"meta,omitempty"`
}
type errorCode int

type Location struct {
	ID          int64     `db:"id" json:"id"`
	Ip          string    `db:"ip" json:"ip"`
	Continent   string    `db:"continent" json:"continent"`
	Province    string    `db:"province" json:"province"`
	City        string    `db:"city" json:"city"`
	Country     string    `db:"country" json:"country"`
	Latitude    string    `db:"latitude" json:"latitude"`
	Longitude   string    `db:"longitude" json:"longitude"`
	AreaCode    string    `db:"area_code" json:"area_code"`
	Isp         string    `db:"isp" json:"isp"`
	ZipCode     string    `db:"zip_code" json:"zip_code"`
	Elevation   string    `db:"elevation" json:"elevation"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	CountryCode string    `db:"-" json:"country_code"`
}
