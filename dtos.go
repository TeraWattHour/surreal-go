package surreal

import "encoding/json"

type LiveDiff struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}
