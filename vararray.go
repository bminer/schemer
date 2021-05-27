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

// Bytes encodes the schema in a portable binary format
func (s *VarArraySchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = []byte{0b00100100}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schema[0] |= 128
	}

	schema = append(schema, s.Element.Bytes()...)

	return schema

}

func (s *VarArraySchema) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *VarArraySchema) Encode(w io.Writer, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	v := reflect.ValueOf(i)

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}

	if s.SchemaOptions.Nullable {
		// did the caller pass in a nil value, or a null pointer?
		if !v.IsValid() {

			fmt.Println("value encoded as a null...")

			// per the revised spec, 1 indicates null
			w.Write([]byte{1})
			return nil
		} else {
			// 0 indicates not null
			w.Write([]byte{0})
		}
	} else {
		// if nullable is false
		// but they are trying to encode a nil value.. then that is an error
		if !v.IsValid() {
			return fmt.Errorf("cannot enoded nil value when IsNullable is false")
		}
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Slice {
		return fmt.Errorf("VarArraySchema can only encode slices")
	}

	err := writeVarUint(w, uint64(v.Len()))
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

	// if the data indicates this type is nullable, then the actual
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.SchemaOptions.Nullable {

		// first byte indicates whether value is null or not...
		buf := make([]byte, 1)
		_, err := io.ReadAtLeast(r, buf, 1)
		if err != nil {
			return err
		}
		valueIsNull := (buf[0] == 1)

		if valueIsNull {
			if v.Kind() == reflect.Ptr {
				if v.CanSet() {
					v.Set(reflect.Zero(v.Type()))
					return nil
				}
				v = v.Elem()
				if v.CanSet() {
					v.Set(reflect.Zero(v.Type()))
					return nil
				}
				return fmt.Errorf("destination not settable")
			} else {
				return fmt.Errorf("cannot decode null value to non pointer to pointer type")
			}
		}
	}

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {

		if v.IsNil() {
			if !v.CanSet() {
				return fmt.Errorf("decode destination is not settable")
			}
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

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
	} else {
		// we have an existing slice
		// right now by default, we will just keep their entries
		// but we have to decide if this behavior is OK??
	}

	/*
		if int(expectedLen) != v.Len() {
			return fmt.Errorf("encoded length does not match destination slice size")
		}
	*/

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
