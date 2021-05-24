package schemer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type ObjectField struct {
	StructFieldOptions
	Schema Schema
}

type FixedObjectSchema struct {
	IsNullable bool
	Fields     []ObjectField
}

func (s *FixedObjectSchema) IsValid() bool {
	return true
}

func (s *FixedObjectSchema) MarshalJSON() ([]byte, error) {

	mapD := map[string]int{"apple": 5, "lettuce": 7}
	return json.Marshal(mapD)

}

func (s *FixedObjectSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Bytes encodes the schema in a portable binary format
// FIXME
func (s *FixedObjectSchema) Bytes() []byte {

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

		// first encode the number of aliases for this field
		// (which will always be at least one... meaning the name of the field from the source struct at least!)
		numAlias := byte(len(f.StructFieldOptions.FieldAliases))
		schemaBytes = append(schemaBytes, numAlias)

		// now write each field alias
		for i := 0; i < len(f.StructFieldOptions.FieldAliases); i++ {
			s := f.FieldAliases[i]
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

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid FixedObjectSchema schema")
	}

	v := reflect.ValueOf(i)

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}

	if s.IsNullable {
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
		// 0 indicates not null
		w.Write([]byte{0})
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Struct {
		return fmt.Errorf("fixedObjectSchema can only encode structs")
	}
	// loop through all the schemas in this object
	// and encode each field
	for i := 0; i < len(s.Fields); i++ {

		if !s.Fields[i].StructFieldOptions.ShouldSkip {
			f := v.Field(i)
			err := s.Fields[i].Schema.Encode(w, f.Interface())
			if err != nil {
				return err
			}
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
		structFieldOptions := StructFieldOptions{}
		parseStructTag(v.
			Type().Field(i).Tag.Get(SchemaTagName), &structFieldOptions)

		// if any of the aliases on this destination field match sourceFieldAlias, then we have a match!
		for j := 0; j < len(structFieldOptions.FieldAliases); j++ {
			if sourceFieldAlias == structFieldOptions.FieldAliases[j] {
				return fieldName
			}

		}

	}

	return ""

}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *FixedObjectSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot decode using invalid FixedObjectSchema schema")
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

		for j := 0; j < len(s.Fields[i].FieldAliases); j++ {

			stringToMatch := s.Fields[i].FieldAliases[j]
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
	return s.IsNullable
}

func (s *FixedObjectSchema) SetNullable(n bool) {
	s.IsNullable = n
}
