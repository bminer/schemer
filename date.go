package schemer

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"
)

// each custom type has a unique name an a unique UUID
const dateSchemaName string = "date"
const dateSchemaUUID byte = 01 // each custom type has a unique id

type dateSchema struct {
	SchemaOptions
}

// CustomSchema receivers --------------------------------------

func (s *dateSchema) Name() string {
	return dateSchemaName
}

func (s *dateSchema) UUID() byte {
	return dateSchemaUUID
}

func (s *dateSchema) Unmarshaljson(buf []byte) (Schema, error) {
	fields := make(map[string]interface{})

	err := json.Unmarshal(buf, &fields)
	if err != nil {
		return nil, err
	}

	b, ok := fields["nullable"].(bool)
	if !ok {
		return nil, fmt.Errorf("missing nullable field in JSON data while decoding dateSchema")
	}

	s.SchemaOptions.nullable = b
	return s, nil
}

func (s *dateSchema) UnMarshalSchemer(buf []byte, byteIndex *int) (Schema, error) {

	s.SchemaOptions.nullable = (buf[*byteIndex]&SchemaNullBit == SchemaNullBit)

	// advance to the UUID
	*byteIndex++
	if buf[*byteIndex] != s.UUID() {
		return nil, fmt.Errorf("invalid call to dateSchema(), invalid UUID encountered in binary schema")
	}

	// advance past the UUID
	*byteIndex++

	return s, nil
}

// loop through all registered GO types
func (s *dateSchema) ForType(t reflect.Type) Schema {
	if t.Name() == "Time" && t.PkgPath() == "time" {
		return s
	}
	return nil
}

// Schema receivers --------------------------------------

func (s *dateSchema) GoType() reflect.Type {
	var t time.Time
	retval := reflect.TypeOf(t)

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}
	return retval
}

func (s *dateSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(map[string]interface{}{
		"type":       "custom",
		"customtype": dateSchemaName,
		"nullable":   s.Nullable(),
	})
}

// Bytes encodes the schema in a portable binary format
func (s *dateSchema) MarshalSchemer() []byte {

	const schemerDateSize byte = 1 + 1 // 1 byte for the schema + 1 bytes for the UUID

	// string schemas are 1 byte long
	var schema []byte = make([]byte, schemerDateSize)

	schema[0] = CustomSchemaMask

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= 0x80
	}

	schema[1] = dateSchemaUUID

	return schema
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *dateSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of he output stream
func (s *dateSchema) EncodeValue(w io.Writer, v reflect.Value) error {

	ok, err := PreEncode(s, w, &v)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Struct || t.Name() != "Time" || t.PkgPath() != "time" {
		return fmt.Errorf("dateSchema only supports encoding time.Time values")
	}

	// call method UnixNano() on v, which is guarenteed to be a time.Time() due to the above check.
	// The method returns: "The number of nanoseconds elapsed since January 1, 1970 UTC"
	// because w are calling the method using reflection, we get back a slice of reflect.Values
	retValSlice := v.MethodByName("UnixNano").Call(nil)
	milisecondsToEncode := retValSlice[0].Int() / 1000000

	varIntSchema := SchemaOf(milisecondsToEncode)
	err = varIntSchema.Encode(w, milisecondsToEncode)

	if err != nil {
		return err
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *dateSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded valuethe input stream and store it in v
func (s *dateSchema) DecodeValue(r io.Reader, v reflect.Value) error {

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

	// Ensure v is settable
	if !v.CanSet() {
		return fmt.Errorf("decode destination is not settable")
	}

	var decodedMilliseconds int64

	varIntSchema := SchemaOf(decodedMilliseconds)
	err = varIntSchema.Decode(r, &decodedMilliseconds)

	if err != nil {
		return err
	}

	nanoSeconds := decodedMilliseconds * 1000000

	if k == reflect.Struct && t.Name() == "Time" || t.PkgPath() == "time" {
		v.Set(reflect.ValueOf(time.Unix(0, nanoSeconds)))
		return nil
	}

	// maybe it makes sense to just return the raw nanoseconds if they are trying
	// to decode to an integer
	if k == reflect.Int64 {
		v.SetInt(nanoSeconds)
		return nil
	}

	if s.weakDecoding {
		if k == reflect.String {
			v.SetString(time.Unix(0, nanoSeconds).String())
			return nil
		}
	}

	return fmt.Errorf("invalid destination")
}
