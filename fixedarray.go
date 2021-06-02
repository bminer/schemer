package schemer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type FixedArraySchema struct {
	SchemaOptions

	Length  int
	Element Schema
}

func (s *FixedArraySchema) DefaultGOType() reflect.Type {
	return reflect.ArrayOf(s.Length, s.Element.DefaultGOType())
}

func (s *FixedArraySchema) Valid() bool {
	return s.Length >= 0
}

// Bytes encodes the schema in a portable binary format
func (s *FixedArraySchema) MarshalSchemer() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = []byte{0b00100101}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schema[0] |= 128
	}

	// encode array fixed length as a varint
	buf := make([]byte, binary.MaxVarintLen64)
	varIntByteLength := binary.PutVarint(buf, int64(s.Length))
	schema = append(schema, buf[0:varIntByteLength]...)

	// now encode the schema for the type of this array
	schema = append(schema, s.Element.MarshalSchemer()...)

	return schema

}

func (s *FixedArraySchema) MarshalJSON() ([]byte, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid FixedArraySchema")
	}

	tmpMap := make(map[string]interface{}, 3)
	tmpMap["type"] = "array"
	tmpMap["length"] = strconv.Itoa(s.Length)
	tmpMap["nullable"] = strconv.FormatBool(s.SchemaOptions.Nullable)

	// now encode the schema for the element
	elementJSON, err := s.Element.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var elementMap map[string]interface{}

	err = json.Unmarshal(elementJSON, &elementMap)
	if err != nil {
		return nil, err
	}

	tmpMap["element"] = elementMap

	return json.Marshal(tmpMap)
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *FixedArraySchema) Encode(w io.Writer, i interface{}) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot encode using invalid FixedArraySchema schema")
	}

	v := reflect.ValueOf(i)

	ok, err := PreEncode(s, w, &v)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Array {
		return fmt.Errorf("FixedArraySchema can only encode fixed length arrays")
	}

	if s.Length != v.Len() {
		return fmt.Errorf("source array size does not match schema size")
	}

	for i := 0; i < v.Len(); i++ {
		s.Element.Encode(w, v.Index(i).Interface())
	}

	return nil
}

func (s *FixedArraySchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot decode using invalid FixedArraySchema schema")
	}

	v, err := PreDecode(s, r, v)
	if err != nil {
		return err
	}
	// if PreDecode() returns a zero value for v, it means we are done decoding
	if !(v.IsValid()) {
		return nil
	}

	t := v.Type()
	k := t.Kind()

	if k == reflect.Interface {
		v.Set(reflect.New(s.DefaultGOType()))

		v = v.Elem().Elem()
		t = v.Type()
		k = t.Kind()
	}

	if k != reflect.Array {
		return fmt.Errorf("FixedArraySchema can only encode fixed length arrays")
	}

	if s.Length != v.Len() {
		return fmt.Errorf("source array size does not match schema size")
	}

	for i := 0; i < s.Length; i++ {
		err := s.Element.DecodeValue(r, v.Index(i))
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *FixedArraySchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)
}

func (s *FixedArraySchema) Nullable() bool {
	return s.SchemaOptions.Nullable
}

func (s *FixedArraySchema) SetNullable(n bool) {
	s.SchemaOptions.Nullable = n
}
