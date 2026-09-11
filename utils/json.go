package utils

import (
	"bytes"
	"encoding/json"
	"io"
)

func JsonDecode(body io.Reader, v any) error {
	if v == nil {
		return nil
	}
	return json.NewDecoder(body).Decode(v)
}

func JsonEncode(v any) (io.Reader, error) {
	if v == nil {
		return nil, nil
	}
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
