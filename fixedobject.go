package schemer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type ObjectField struct {
	Aliases []string
	Schema  Schema
}

type FixedObjectSchema struct {
	SchemaOptions
	Fields []ObjectField
}

// fixme
func (s *FixedObjectSchema) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// Bytes encodes the schema in a portable binary format
func (s *FixedObjectSchema) Bytes() []byte {

	// fixedObject schemas are 1 byte long
	var schemaBytes []byte = []byte{0b00101001}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schemaBytes[0] |= 128
	}

	// encode total number of fields as a varint
	buf := make([]byte, binary.MaxVarintLen64)
	varIntByteLength := binary.PutVarint(buf, int64(len(s.Fields)))

	schemaBytes = append(schemaBytes, buf[0:varIntByteLength]...)

	// also need to concatenate the schemas for all other fields
	for _, f := range s.Fields {

		// first encode the number of aliases for this field
		// (which will always be at least one... meaning the name of the field from the source struct at least!)
		buf := make([]byte, binary.MaxVarintLen64)
		varIntByteLength := binary.PutVarint(buf, int64(len(f.Aliases)))

		schemaBytes = append(schemaBytes, buf[0:varIntByteLength]...)

		// now write each field alias
		for i := 0; i < len(f.Aliases); i++ {
			s := f.Aliases[i]
			varLenStringSchema := SchemaOf(s)
			var buf bytes.Buffer
			varLenStringSchema.Encode(&buf, s)
			schemaBytes = append(schemaBytes, buf.Bytes()...)
		}

		schemaBytes = append(schemaBytes, f.Schema.Bytes()...)
	}

	return schemaBytes
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *FixedObjectSchema) Encode(w io.Writer, i interface{}) error {

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

	if k != reflect.Struct {
		return fmt.Errorf("fixedObjectSchema can only encode structs")
	}
	// loop through all the schemas in this object
	// and encode each field
	for i := 0; i < len(s.Fields); i++ {

		f := v.Field(i)
		err := s.Fields[i].Schema.Encode(w, f.Interface())
		if err != nil {
			return err
		}

	}

	return nil
}

// s is a source alias we are tryig to figure out where to put
// v will be a struct...
func (s *FixedObjectSchema) findDestinationField(sourceFieldAlias string, v reflect.Value) string {

	// see if there is a place in the destination struct that matches the alias passed in...
	for i := 0; i < v.Type().NumField(); i++ {
		fieldName := v.Type().Field(i).Name

		if sourceFieldAlias == fieldName {
			return sourceFieldAlias
		}

		// parse the tags on this field, to see if any aliases are present...
		schemerTagOptions := SchemerTagOptions{}
		parseStructTag(v.
			Type().Field(i).Tag.Get(SchemaTagName), &schemerTagOptions)

		// if any of the aliases on this destination field match sourceFieldAlias, then we have a match!
		for j := 0; j < len(schemerTagOptions.FieldAliases); j++ {
			if sourceFieldAlias == schemerTagOptions.FieldAliases[j] {
				return fieldName
			}

		}

	}

	return ""

}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *FixedObjectSchema) DecodeValue(r io.Reader, v reflect.Value) error {

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

	if k != reflect.Struct {
		return fmt.Errorf("FixedObjectSchema can only decode to structures")
	}

	// loop through all the potential source fields
	// and see if there is anywhere we can put them
	for i := 0; i < len(s.Fields); i++ {
		Foundmatch := false

		for j := 0; j < len(s.Fields[i].Aliases); j++ {

			stringToMatch := s.Fields[i].Aliases[j]
			structFieldToPopulate := s.findDestinationField(stringToMatch, v)

			if structFieldToPopulate != "" {
				Foundmatch = true
				err := s.Fields[i].Schema.DecodeValue(r, v.FieldByName(structFieldToPopulate))
				if err != nil {
					return err
				}
			}

		}

		if !Foundmatch {
			// otherwise, there is just an extra field from the source struct that we cannot match
			// in the destination struct
			// since there is no where to put the field, we just need to skip it
			// (but we still need to call DecodeValue here to process the bytes of the encoded data!)
			var ignoreMe reflect.Value = reflect.Value{}
			s.Fields[i].Schema.Decode(r, &ignoreMe)
			// ignore error, we just needed to process the bytes in R
		}

	}

	return nil
}

func (s *FixedObjectSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	return s.DecodeValue(r, reflect.ValueOf(i))

}

func (s *FixedObjectSchema) Nullable() bool {
	return s.SchemaOptions.Nullable
}

func (s *FixedObjectSchema) SetNullable(n bool) {
	s.SchemaOptions.Nullable = n
}
