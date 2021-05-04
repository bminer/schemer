package schemer

import (
	"fmt"
	"io"
)

// https://golangbyexample.com/go-size-range-int-uint/
//const uintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

type Schema interface {
	// Encode uses the schema to write the encoded value of v to the output stream
	Encode(w io.Writer, i interface{}) error
	// Decode uses the schema to read the next encoded value from the input stream and store it in v
	Decode(r io.Reader, i interface{}) error
	// MarshalSchemer encodes the schema in a portable binary format
	// MarshalJSON returns the JSON encoding of the schema
	DoMarshalJSON() ([]byte, error)
	// UnmarshalJSON updates the schema by decoding the JSON-encoded schema in b
	DoUnmarshalJSON(b []byte) error
	// Nullable returns true if and only if the type is nullable
	/*
		Nullable() bool
		// SetNullable sets the nullable flag for the schema
		SetNullable(n bool)
	*/
}

// NewSchema decodes a schema stored in buf and returns an error if the schema is invalid
func NewSchema(buf []byte) (Schema, error) {

	var bit3IsSet bool

	var floatSchema FloatSchema
	var complexSchema ComplexSchema

	// decode floating point schema
	if buf[0]&80 == 80 {
		bit3IsSet = (buf[0] & 4) == 4
		if bit3IsSet {
			floatSchema.Bits = 64
		} else {
			floatSchema.Bits = 32
		}
		floatSchema.IsNullable = (buf[0]&1 == 1)

		return floatSchema, nil
	}

	// decode complex number
	if buf[0]&96 == 96 {
		bit3IsSet = (buf[0] & 4) == 4
		if bit3IsSet {
			complexSchema.Bits = 128
		} else {
			complexSchema.Bits = 64
		}
		complexSchema.IsNullable = (buf[0]&1 == 1)

		return complexSchema, nil

	}

	return nil, fmt.Errorf("ComplexSchema only supports encoding Complex64 and Complex128 values")
}
