package schemer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
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

func (s *FixedObjectSchema) GoType() reflect.Type {
	return reflect.TypeOf(struct{}{})
}

func (s *FixedObjectSchema) MarshalJSON() ([]byte, error) {
	tmpMap := make(map[string]interface{}, 4)
	tmpMap["type"] = "object"
	tmpMap["nullable"] = s.SchemaOptions.Nullable

	var fieldMap []map[string]interface{}

	for i := range s.Fields {
		fields := make(map[string]interface{})
		fields["name"] = s.Fields[i].Aliases[0]

		b, err := s.Fields[i].Schema.MarshalJSON()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &fields)
		if err != nil {
			return nil, err
		}
		fieldMap = append(fieldMap, fields)
	}

	tmpMap["fields"] = fieldMap
	return json.MarshalIndent(tmpMap, "", "   ")
}

// Bytes encodes the schema in a portable binary format
func (s *FixedObjectSchema) MarshalSchemer() []byte {

	// fixedObject schemas are 1 byte long
	var schemaBytes []byte = []byte{FixedObjectSchemaBinaryFormat}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schemaBytes[0] |= 0x80
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

		schemaBytes = append(schemaBytes, f.Schema.MarshalSchemer()...)
	}
	return schemaBytes
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *FixedObjectSchema) Encode(w io.Writer, i interface{}) error {

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

var cacheMap map[string]string

// s is a source alias we are tryig to figure out where to put
// v will be a struct...
func (s *FixedObjectSchema) findDestinationField(sourceFieldAlias string, v reflect.Value) string {

	if cacheMap == nil {
		cacheMap = make(map[string]string)
	}
	r, ok := cacheMap[sourceFieldAlias]
	if ok {
		return r //CacheMap[sourceFieldAlias]
	}

	// see if there is a place in the destination struct that matches the alias passed in...
	for i := 0; i < v.Type().NumField(); i++ {
		fieldName := v.Type().Field(i).Name

		if sourceFieldAlias == fieldName {
			cacheMap[sourceFieldAlias] = sourceFieldAlias
			return sourceFieldAlias
		}

		// parse the tags on this field, to see if any aliases are present...
		tagOpts, _ := ParseStructTag(v.
			Type().Field(i).Tag.Get(SchemerTagName))

		// if any of the aliases on this destination field match sourceFieldAlias, then we have a match!
		for j := 0; j < len(tagOpts.FieldAliases); j++ {
			if sourceFieldAlias == tagOpts.FieldAliases[j] {
				cacheMap[sourceFieldAlias] = fieldName
				return fieldName
			}

		}

	}
	return ""
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *FixedObjectSchema) DecodeValue(r io.Reader, v reflect.Value) error {

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
		v.Set(reflect.New(s.GoType()))

		v = v.Elem().Elem()
		t = v.Type()
		k = t.Kind()
	}

	if k != reflect.Struct {
		return fmt.Errorf("FixedObjectSchema can only decode to structures")
	}

	var stringToMatch string

	// loop through all the potential source fields
	// and see if there is anywhere we can put them
	for i := 0; i < len(s.Fields); i++ {
		found := false

		for j := 0; j < len(s.Fields[i].Aliases); j++ {

			stringToMatch = s.Fields[i].Aliases[j]
			structFieldToPopulate := s.findDestinationField(stringToMatch, v)

			if structFieldToPopulate != "" {
				found = true
				err := s.Fields[i].Schema.DecodeValue(r, v.FieldByName(structFieldToPopulate))

				if err != nil {
					return err
				}
			}

		}

		if !found {
			// otherwise, there is just an extra field from the source struct that we cannot match
			// in the destination struct
			// since there is no where to put the field, we just need to skip it
			// (but we still need to call DecodeValue here to process the bytes of the encoded data!)
			var ignoreMe interface{}
			err := s.Fields[i].Schema.Decode(r, &ignoreMe)
			if err != nil {
				return err
			}
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