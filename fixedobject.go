package schemer

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type ObjectField struct {
	StructFieldOptions

	Name   string
	Schema Schema
}

type FixedObjectSchema struct {
	IsNullable bool
	Fields     []ObjectField
}

func (s FixedObjectSchema) IsValid() bool {
	return true
}

func (s FixedObjectSchema) MarshalJSON() ([]byte, error) {

	type tmpFixedObjectSchema FixedObjectSchema

	return json.MarshalIndent(struct {
		tmpFixedObjectSchema
	}{
		tmpFixedObjectSchema: tmpFixedObjectSchema(s),
	}, "", "  ")

}

func (s FixedObjectSchema) DoMarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s FixedObjectSchema) DoUnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Bytes encodes the schema in a portable binary format
// FIXME
func (s FixedObjectSchema) Bytes() []byte {

	// fixedObject schemas are 1 byte long
	var schemaBytes []byte = make([]byte, 1)

	schemaBytes[0] = 0b10100100

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schemaBytes[0] |= 1
	}

	// fixme:
	// update this to write as a var len int
	tmp := byte(len(s.Fields))
	schemaBytes = append(schemaBytes, tmp)

	// also need to concatenate the schemas for all other fields
	for _, f := range s.Fields {
		schemaBytes = append(schemaBytes, f.Schema.Bytes()...)
	}

	return schemaBytes
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s FixedObjectSchema) Encode(w io.Writer, i interface{}) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid FixedLenArraySchema schema")
	}

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	v := reflect.ValueOf(i)
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer
		if (reflect.TypeOf(v).Kind() == reflect.Ptr ||
			reflect.TypeOf(v).Kind() == reflect.Interface) &&
			reflect.ValueOf(v).IsNil() {
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

	if k != reflect.Struct {
		return fmt.Errorf("fixedObjectSchema can only encode structs")
	}

	// loop through all the schemas, and encode each one
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		err := s.Fields[i].Schema.Encode(w, f.Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s FixedObjectSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid FixedIntSchema schema")
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
				v = v.Elem()
				if v.Kind() == reflect.Ptr {
					// special way to return a nil pointer
					v.Set(reflect.Zero(v.Type()))
				} else {
					return fmt.Errorf("cannot decode null value to non pointer to pointer type")
				}
			} else {
				return fmt.Errorf("cannot decode null value to non pointer to pointer type")
			}
			return nil
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

	if k != reflect.Struct {
		return fmt.Errorf("FixedObjectSchema can only decode to structures")
	}

	for i := 0; i < t.NumField(); i++ {

		// this is where we figure out where to put the
		// value we have just decoded...

		x := s.Fields[i].FieldAliases[0]
		//s.Fields[i].StructFieldOptions.FieldAliases

		err := s.Fields[i].Schema.DecodeValue(r, v.FieldByName(x))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s FixedObjectSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	return s.DecodeValue(r, reflect.ValueOf(i))

}

func (s FixedObjectSchema) Nullable() bool {
	return s.IsNullable
}

func (s *FixedObjectSchema) SetNullable(n bool) {
	s.IsNullable = n
}
