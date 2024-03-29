package schemer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type FixedArraySchema struct {
	SchemaOptions

	Length  int
	Element Schema
}

func (s *FixedArraySchema) GoType() reflect.Type {
	retval := reflect.ArrayOf(s.Length, s.Element.GoType())

	if s.Nullable() {
		return reflect.PtrTo(retval)
	}

	return retval
}

func (s *FixedArraySchema) Valid() bool {
	return s.Length >= 0
}

// Bytes encodes the schema in a portable binary format
func (s *FixedArraySchema) MarshalSchemer() ([]byte, error) {

	// fixed length schemas are 1 byte long total
	var schema []byte = []byte{FixedArrayByte}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	// encode array fixed length as a varint
	buf := make([]byte, binary.MaxVarintLen64)
	varIntByteLength := binary.PutVarint(buf, int64(s.Length))
	schema = append(schema, buf[0:varIntByteLength]...)

	// now encode the schema for the type of this array

	m, ok := s.Element.(Marshaler)
	if !ok {
		return nil, fmt.Errorf("element does not implement MarshalSchemer")
	}

	elementData, err := m.MarshalSchemer()
	if err != nil {
		return nil, fmt.Errorf("element does not implement MarshalSchemer")
	}

	schema = append(schema, elementData...)
	return schema, nil

}

func (s *FixedArraySchema) MarshalJSON() ([]byte, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid FixedArraySchema")
	}

	tmpMap := make(map[string]interface{}, 3)
	tmpMap["type"] = "array"
	tmpMap["length"] = s.Length
	tmpMap["nullable"] = s.Nullable()

	// now encode the schema for the element

	t, ok := s.Element.(json.Marshaler)
	if !ok {
		return nil, fmt.Errorf("element does not implement MarshalJSON")
	}

	elementJSON, err := t.MarshalJSON()
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

// Encode uses the schema to write the encoded value of i to the output stream
func (s *FixedArraySchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output stream
func (s *FixedArraySchema) EncodeValue(w io.Writer, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot encode using invalid FixedArraySchema schema")
	}

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
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

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *FixedArraySchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *FixedArraySchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot decode using invalid FixedArraySchema schema")
	}

	done, err := PreDecode(r, &v, s.Nullable())
	if err != nil || done {
		return err
	}

	t := v.Type()
	k := t.Kind()

	if k == reflect.Interface {
		v.Set(reflect.New(s.GoType()))

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
