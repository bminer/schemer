package schemer

import (
	"encoding/json"
	"io"
)

type BoolSchema struct {
	WeakDecoding bool
}

func (s BoolSchema) IsValid() bool {
	return true
}

func (s BoolSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *BoolSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s BoolSchema) Encode(w io.Writer, v interface{}) error {
	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s BoolSchema) Decode(r io.Reader, i interface{}) error {

	return nil
}
