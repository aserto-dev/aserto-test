package printer

import (
	"encoding/json"
	"io"
)

type JSON struct {
	enc *json.Encoder
}

func NewJSON(w io.Writer) *JSON {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	enc.SetIndent("", "  ")
	return &JSON{
		enc: enc,
	}
}

func (p *JSON) Print(v interface{}) error {
	return p.enc.Encode(v)
}
