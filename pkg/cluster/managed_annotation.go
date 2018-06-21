package cluster

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"

	"github.com/ksonnet/ksonnet/pkg/util/serial"
)

// managedAnnotation is the contents of of the ksonnet.io/managed annotation.
type managedAnnotation struct {
	Pristine string `json:"pristine,omitempty"`
}

// Marshal marshals this object to JSON.
func (mm *managedAnnotation) Marshal() ([]byte, error) {
	return json.Marshal(mm)
}

// EncodePristine encodes a pristine copy of the object.
func (mm *managedAnnotation) EncodePristine(m map[string]interface{}) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	actions := []serial.Action{
		func() error { return json.NewEncoder(gz).Encode(m) },
		gz.Flush,
		gz.Close,
	}

	if err := serial.RunActions(actions...); err != nil {
		return err
	}

	mm.Pristine = base64.StdEncoding.EncodeToString(buf.Bytes())
	return nil
}

// DecodePristine decodes a pristone copy of the object.
func (mm *managedAnnotation) DecodePristine() (map[string]interface{}, error) {
	b, err := base64.StdEncoding.DecodeString(mm.Pristine)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var m map[string]interface{}
	if err := json.NewDecoder(zr).Decode(&m); err != nil {
		return nil, err
	}

	return m, nil
}
