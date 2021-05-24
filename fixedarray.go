package schemer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type FixedLenArraySchema struct {
	IsNullable bool

	Length  int
	Element Schema
}

func (s *FixedLenArraySchema) IsValid() bool {
	return s.Length > 0
}

// Bytes encodes the schema in a portable binary format
func (s *FixedLenArraySchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = make([]byte, 1)

	schema[0] = 0b10010100 // bit pattern for fixed array schema

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	// encode array fixed length as a varint
	buf := make([]byte, binary.MaxVarintLen64)
	varIntByteLength := binary.PutVarint(buf, int64(s.Length))
	schema = append(schema, buf[0:varIntByteLength]...)

	// now encode the schema for the type of this array
	schema = append(schema, s.Element.Bytes()...)

	return schema

}

// if this function is called MarshalJSON it seems to be called
// recursively by the json library???
func (s *FixedLenArraySchema) DoMarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid floating point schema")
	}

	return json.Marshal(s)
}

// if this function is called UnmarshalJSON it seems to be called
// recursively by the json library???
func (s *FixedLenArraySchema) DoUnmarshalJSON(buf []byte) error {
	return json.Unmarshal(buf, s)
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *FixedLenArraySchema) Encode(w io.Writer, i interface{}) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid FixedLenArraySchema schema")
	}

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer
		if reflect.TypeOf(i).Kind() == reflect.Ptr ||
			reflect.TypeOf(i).Kind() == reflect.Interface &&
				reflect.ValueOf(i).IsNil() {
			// we encode a null value by writing a single non 0 byte
			w.Write([]byte{1})
			return nil
		} else {
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

	if k != reflect.Array {
		return fmt.Errorf("FixedLenArraySchema can only encode fixed length arrays")
	}

	if s.Length != v.Len() {
		return fmt.Errorf("source array size does not match schema size")
	}

	for i := 0; i < v.Len(); i++ {
		s.Element.Encode(w, v.Index(i).Interface())
	}

	return nil
}

func (s *FixedLenArraySchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot decode using invalid FixedLenArraySchema schema")
	}

	// if the schema indicates this type is nullable, then the actual floating point
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.IsNullable {
		buf := make([]byte, 1)
		_, err := io.ReadAtLeast(r, buf, 1)
		if err != nil {
			return err
		}
		if buf[0] != 0 {
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

	if k != reflect.Array {
		return fmt.Errorf("FixedLenArraySchema can only encode fixed length arrays")
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
func (s *FixedLenArraySchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)
}

func (s *FixedLenArraySchema) Nullable() bool {
	return s.IsNullable
}

func (s *FixedLenArraySchema) SetNullable(n bool) {
	s.IsNullable = n
}
