package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type VarLenArraySchema struct {
	IsNullable bool

	Element Schema
}

func (s VarLenArraySchema) IsValid() bool {
	return true
}

func (s VarLenArraySchema) DecodeValue(r io.Reader, v reflect.Value) error {
	return nil
}

// Bytes encodes the schema in a portable binary format
func (s VarLenArraySchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = make([]byte, 1)

	schema[0] = 0b10010000 // bit pattern for var array

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	return schema

}

// if this function is called MarshalJSON it seems to be called
// recursively by the json library???
func (s VarLenArraySchema) DoMarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid floating point schema")
	}

	return json.Marshal(s)
}

// if this function is called UnmarshalJSON it seems to be called
// recursively by the json library???
func (s VarLenArraySchema) DoUnmarshalJSON(buf []byte) error {
	return json.Unmarshal(buf, s)
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s VarLenArraySchema) Encode(w io.Writer, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid VarLenArraySchema schema")
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer?
		if reflect.TypeOf(i).Kind() == reflect.Ptr ||
			reflect.TypeOf(i).Kind() == reflect.Interface &&
				reflect.ValueOf(i).IsNil() {

			// per the spec, we encode a null value by writing out a single byte
			// with the high bit set
			w.Write([]byte{128})
			return nil
		} else {
			// TODO: update the SPEC (encoding-format.md)
			// 0 means not null (with actual encoded bytes to follow)
			w.Write([]byte{0})
		}
	} else {
		if i == nil {
			return fmt.Errorf("cannot enoded nil value when IsNullable is false")
		}
	}

	v := reflect.ValueOf(i)

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		i = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	if k != reflect.String {
		return fmt.Errorf("StringSchema only supports encoding string values")
	}

	var stringToEncode string = v.String()
	var stringLen int = len(stringToEncode)

	err := writeVarUint(w, uint64(stringLen))
	if err != nil {
		return errors.New("cannot encode var string length as var int")
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s VarLenArraySchema) Decode(r io.Reader, i interface{}) error {

	return nil
}
