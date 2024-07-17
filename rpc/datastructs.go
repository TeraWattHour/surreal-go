package rpc

import (
	"encoding/json"
	"fmt"
	"time"
)

// Mirroring https://github.com/surrealdb/surrealdb.go implementation

type Incoming struct {
	ID     any             `json:"id" msgpack:"id"`
	Error  *Error          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

type Outgoing struct {
	ID     any    `json:"id"`
	Async  bool   `json:"async,omitempty"`
	Method string `json:"method,omitempty"`
	Params []any  `json:"params,omitempty"`
}

type LiveNotification struct {
	ID     string          `json:"id"`
	Action string          `json:"action"`
	Result json.RawMessage `json:"result"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (r *Error) Error() string {
	return fmt.Sprintf("#%d: %s", r.Code, r.Message)
}

type MappedBool bool

func (b *MappedBool) UnmarshalJSON(data []byte) error {
	*b = string(data) == `"OK"`
	return nil
}

type MappedDuration time.Duration

func (d *MappedDuration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = MappedDuration(duration)
	return nil
}

type RawResult []struct {
	OK     MappedBool `json:"status"`
	Time   MappedDuration
	Result json.RawMessage
}
