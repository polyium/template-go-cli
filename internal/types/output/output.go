package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/spf13/pflag"
)

// Type string that implements Cobra's [cobra.Type] interface for valid string enumeration values.
type Type string

// Runtime conformator to ensure implementation satisfies the interface.
var _ pflag.Value = (*Type)(nil)

const (
	JSON Type = "json"
	YAML Type = "yaml"
)

// String is used both by fmt.Print and by Cobra in help text.
func (o *Type) String() string {
	return string(*o)
}

// Set must have pointer receiver so it doesn't change the value of a copy.
func (o *Type) Set(v string) error {
	switch v {
	case "json", "yaml":
		*o = Type(v)
		return nil
	default:
		return errors.New("must be one of \"json\" or \"yaml\"")
	}
}

// Type is only used in help text.
func (o *Type) Type() string {
	return "(yaml|json)"
}

// Write serializes the provided datum into the specified format (JSON or YAML) and writes it to the given [io.Writer].
// Returns an error if encoding fails or encounters an issue during writing.
func Write(format Type, datum interface{}) (*bytes.Buffer, error) {
	var writer bytes.Buffer

	switch format {
	case JSON:
		encoder := json.NewEncoder(&writer)
		encoder.SetIndent("", "    ")

		if e := encoder.Encode(datum); e != nil {
			return nil, fmt.Errorf("failed to encode json: %w", e)
		}
	case YAML:
		if e := yaml.NewEncoder(&writer, yaml.Indent(4)).Encode(datum); e != nil {
			return nil, fmt.Errorf("failed to encode yaml: %w", e)
		}
	}

	return &writer, nil
}
