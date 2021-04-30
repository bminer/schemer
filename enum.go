package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type EnumSchema struct {
	WeakDecoding bool
}

func (s EnumSchema) IsValid() bool {
	return true
}

func (s EnumSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *EnumSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s EnumSchema) Encode(w io.Writer, v interface{}) error {

	value := reflect.ValueOf(v)
	t := value.Type()
	k := t.Kind()

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid EnumSchema schema")
	}

	switch k {

	default:
		return errors.New("can only encode boolean types when using BoolSchema")
	}

}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s BoolSchema) EnumSchema(r io.Reader, i interface{}) error {

	return nil
}
