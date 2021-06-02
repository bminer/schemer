package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type VarArraySchema struct {
	SchemaOptions

	Element Schema
}

func (s *VarArraySchema) DefaultGOType() reflect.Type {
	return reflect.SliceOf(s.Element.DefaultGOType())
}

// Bytes encodes the schema in a portable binary format
func (s *VarArraySchema) MarshalSchemer() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = []byte{0b00100100}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schema[0] |= 128
	}

	schema = append(schema, s.Element.MarshalSchemer()...)

	return schema

}

func (s *VarArraySchema) MarshalJSON() ([]byte, error) {

	tmpMap := make(map[string]interface{}, 2)
	tmpMap["type"] = "array"
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
func (s *VarArraySchema) Encode(w io.Writer, i interface{}) error {

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

	if k != reflect.Slice {
		return fmt.Errorf("VarArraySchema can only encode slices")
	}

	err = writeVarUint(w, uint64(v.Len()))
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

func (s *VarArraySchema) DecodeValue(r io.Reader, v reflect.Value) error {

	ok, err := PreDecode(s, r, &v)
	if err != nil {
		return err
	}
	if !ok {
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

	if k != reflect.Slice {
		return fmt.Errorf("VarArraySchema can only decode to slices")
	}

	expectedLen, err := readVarUint(r)
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

func (s *VarArraySchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)
}

func (s *VarArraySchema) Nullable() bool {
	return s.SchemaOptions.Nullable
}

func (s *VarArraySchema) SetNullable(n bool) {
	s.SchemaOptions.Nullable = n
}
