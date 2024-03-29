package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type VarArraySchema struct {
	SchemaOptions

	Element Schema
}

func (s *VarArraySchema) GoType() reflect.Type {
	retval := reflect.SliceOf(s.Element.GoType())

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}

	return retval
}

// Bytes encodes the schema in a portable binary format
func (s *VarArraySchema) MarshalSchemer() ([]byte, error) {

	// fixed length schemas are 1 byte long total
	var schema []byte = []byte{VarArrayByte}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	m := s.Element.(Marshaler)
	tmp, err := m.MarshalSchemer()
	if err != nil {
		return nil, err
	}

	schema = append(schema, tmp...)
	return schema, nil
}

func (s *VarArraySchema) MarshalJSON() ([]byte, error) {

	tmpMap := make(map[string]interface{}, 2)
	tmpMap["type"] = "array"
	tmpMap["nullable"] = s.Nullable()

	m := s.Element.(json.Marshaler)

	// now encode the schema for the element
	elementJSON, err := m.MarshalJSON()
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
func (s *VarArraySchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output stream
func (s *VarArraySchema) EncodeValue(w io.Writer, v reflect.Value) error {

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Slice {
		return fmt.Errorf("VarArraySchema can only encode slices")
	}

	err = WriteUvarint(w, uint64(v.Len()))
	if err != nil {
		return errors.New("cannot encode var string length as var int")
	}

	for i := 0; i < v.Len(); i++ {
		err := s.Element.Encode(w, v.Index(i).Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *VarArraySchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *VarArraySchema) DecodeValue(r io.Reader, v reflect.Value) error {

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

	if k != reflect.Slice {
		return fmt.Errorf("VarArraySchema can only decode to slices")
	}

	expectedLen, err := ReadUvarint(r)
	if err != nil {
		return err
	}

	if v.IsNil() {
		if !v.CanSet() {
			return errors.New("v not settable")
		}
		v.Set(reflect.MakeSlice(t, int(expectedLen), int(expectedLen)))
	}

	// else we have an existing slice
	// right now by default, we will just keep their entries
	// but we have to decide if this behavior is OK??

	for i := 0; i < v.Len(); i++ {
		err := s.Element.DecodeValue(r, v.Index(i))
		if err != nil {
			return err
		}
	}

	return nil
}
