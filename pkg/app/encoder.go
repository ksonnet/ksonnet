package app

import (
	"io"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// Encoder writes items to a serialized form.
type Encoder interface {
	// Encode writes an item to a stream. Implementations may return errors
	// if the data to be encoded is invalid.
	Encode(i interface{}, w io.Writer) error
}

var (
	defaultYAMLEncoder = &YAMLEncoder{}
)

// YAMLEncoder write items to a serialized form in YAML format.
type YAMLEncoder struct{}

// Encode encodes data in yaml format.
func (e *YAMLEncoder) Encode(i interface{}, w io.Writer) error {
	b, err := yaml.Marshal(i)
	if err != nil {
		return errors.Wrap(err, "encoding data")
	}

	_, err = w.Write(b)
	return err
}
