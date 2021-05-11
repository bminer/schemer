package schemer

import (
	"encoding/json"
	"fmt"
	"io"
)

type FixedLenArraySchema struct {
	IsNullable bool

	Length  int
	Element Schema
}

func (s FixedLenArraySchema) IsValid() bool {
	return s.Length > 0
}

// Bytes encodes the schema in a portable binary format
func (s FixedLenArraySchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = make([]byte, 1)

	schema[0] = 0b10010100 // bit pattern for fixed array schema

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	return schema

}

// if this function is called MarshalJSON it seems to be called
// recursively by the json library???
func (s FixedLenArraySchema) DoMarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid floating point schema")
	}

	return json.Marshal(s)
}

// if this function is called UnmarshalJSON it seems to be called
// recursively by the json library???
func (s FixedLenArraySchema) DoUnmarshalJSON(buf []byte) error {
	return json.Unmarshal(buf, s)
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s FixedLenArraySchema) Encode(w io.Writer, i interface{}) error {

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s FixedLenArraySchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	return nil
}
