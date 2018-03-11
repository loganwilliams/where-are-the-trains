package pretty

import (
	"encoding/json"
	"bytes"
)

// Pretty-print JSON string.
func Json(in string) string {
    var out bytes.Buffer
    err := json.Indent(&out, []byte(in), "", "\t")
    if err != nil {
        return in
    }
    return out.String()
}