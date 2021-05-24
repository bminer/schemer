package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type VarObjectSchema struct {
	IsNullable bool

	Key   Schema
	Value Schema
}

func (s *VarObjectSchema) IsValid() bool {
	return true
}

func (s *VarObjectSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *VarObjectSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Bytes encodes the schema in a portable binary format
// FIXME
func (s *VarObjectSchema) Bytes() []byte {

	// string schemas are 1 byte long
	var schema []byte = make([]byte, 1)

	schema[0] = 0b10100000

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	// bit 3 is clear from above, indicating this is a var length string

	schema = append(schema, s.Key.Bytes()...)
	schema = append(schema, s.Value.Bytes()...)

	return schema
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *VarObjectSchema) Encode(w io.Writer, i interface{}) error {

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
		// 0 indicates not null
		w.Write([]byte{0})
	}

	v := reflect.ValueOf(i)
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	if k != reflect.Map {
		return fmt.Errorf("varObjectSchema can only encode maps")
	}

	err := writeVarUint(w, uint64(v.Len()))
	if err != nil {
		return errors.New("cannot encode var string length as var int")
	}

	for _, mapKey := range v.MapKeys() {
		mapValue := v.MapIndex(mapKey)

		err := s.Key.Encode(w, mapKey.Interface()) // encode key
		if err != nil {
			return err
		}
		err = s.Value.Encode(w, mapValue.Interface()) // encode value
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *VarObjectSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot decode using invalid DecodeValue schema")
	}

	// first byte indicates whether value is null or not...
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(r, buf, 1)
	if err != nil {
		return err
	}
	valueIsNull := (buf[0] == 1)

	// if the data indicates this type is nullable, then the actual
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.IsNullable {
		if valueIsNull {
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Map {
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

	if k != reflect.Map {
		return fmt.Errorf("VarObjectSchema can only decode to maps")
	}

	// we wrote the number of entries in the map as a var int
	// when we did the encoding
	expectedNumEntries, err := readVarUint(r)
	if err != nil {
		return err
	}

	if v.IsNil() {
		if !v.CanSet() {
			return errors.New("v not settable")
		}
		var mapType = reflect.MapOf(t.Key(), t.Elem())
		v.Set(reflect.MakeMap(mapType))
	} else {
		// we have an existing map
		// right now by default, we will just keep their entries
		// but we have to decide if this behavior is OK
	}

	for i := 0; i < int(expectedNumEntries); i++ {

		key := reflect.New(t.Key())
		val := reflect.New(t.Elem())

		err = s.Key.DecodeValue(r, key) // decode key
		if err != nil {
			return err
		}
		err = s.Value.DecodeValue(r, val) // decode value
		if err != nil {
			return err
		}

		v.SetMapIndex(reflect.Indirect(key), reflect.Indirect(val))
	}

	return nil
}

func (s *VarObjectSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	return s.DecodeValue(r, reflect.ValueOf(i))
}

func (s *VarObjectSchema) Nullable() bool {
	return s.IsNullable
}

func (s *VarObjectSchema) SetNullable(n bool) {
	s.IsNullable = n
}
