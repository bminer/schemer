package schemer

import (
	"encoding/json"
	"io"
)

type StringSchema struct {
	WeakDecoding bool
}

func (s StringSchema) IsValid() bool {
	return true
}

func (s StringSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *StringSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s StringSchema) Encode(w io.Writer, v interface{}) error {
	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s StringSchema) Decode(r io.Reader, i interface{}) error {

	return nil
}
