package schemer

import (
	"encoding/json"
	"io"
)

type ObjectSchema struct {
	WeakDecoding bool
}

func (s ObjectSchema) IsValid() bool {
	return true
}

func (s ObjectSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *ObjectSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s ObjectSchema) Encode(w io.Writer, v interface{}) error {
	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s ObjectSchema) Decode(r io.Reader, i interface{}) error {

	return nil
}
