package schemer

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
)

// each custom type has a unique name an a unique UUID
const ipV4SchemaName string = "ipv4"
const ipV4SchemaUUID byte = 02

type ipv4Schema struct {
	SchemaOptions
}

// CustomSchema receivers --------------------------------------

func (s *ipv4Schema) Name() string {
	return ipV4SchemaName
}

func (s *ipv4Schema) UUID() byte {
	return ipV4SchemaUUID
}

func (s *ipv4Schema) UnMarshalJSON(buf []byte) (Schema, error) {
	fields := make(map[string]interface{})

	err := json.Unmarshal(buf, &fields)
	if err != nil {
		return nil, err
	}

	b, ok := fields["nullable"].(bool)
	if !ok {
		return nil, fmt.Errorf("missing nullable field in JSON data while decoding ipv4Schema")
	}

	s.SchemaOptions.nullable = b
	return s, nil
}

func (s *ipv4Schema) UnMarshalSchemer(buf []byte, byteIndex *int) (Schema, error) {

	s.SchemaOptions.nullable = (buf[*byteIndex]&SchemaNullBit == SchemaNullBit)

	// advance to the UUID
	*byteIndex++
	if buf[*byteIndex] != s.UUID() {
		return nil, fmt.Errorf("invalid call to ipSchema(), invalid UUID encountered in binary schema")
	}

	// advance past the UUID
	*byteIndex++

	return s, nil
}

// returns an ipv4Schema if the passed in reflect.type is a net.IP
func (s *ipv4Schema) RegisteredSchema(t reflect.Type) Schema {
	if t.Name() == "IP" && t.PkgPath() == "net" {
		return s
	}

	return nil
}

// Schema receivers --------------------------------------

func (s *ipv4Schema) GoType() reflect.Type {
	var t net.IP
	retval := reflect.TypeOf(t)

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}
	return retval
}

func (s *ipv4Schema) MarshalJSON() ([]byte, error) {

	return json.Marshal(map[string]interface{}{
		"type":       "custom",
		"customtype": ipV4SchemaName,
		"nullable":   s.Nullable(),
	})
}

// Bytes encodes the schema in a portable binary format
func (s *ipv4Schema) MarshalSchemer() []byte {

	const schemerDateSize byte = 1 + 1 // 1 byte for the schema + 1 bytes for the UUID

	// string schemas are 1 byte long
	var schema []byte = make([]byte, schemerDateSize)

	schema[0] = CustomSchemaMask

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= 0x80
	}

	schema[1] = ipV4SchemaUUID

	return schema
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *ipv4Schema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of he output stream
func (s *ipv4Schema) EncodeValue(w io.Writer, v reflect.Value) error {

	ok, err := PreEncode(s, w, &v)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Slice || t.Name() != "IP" || t.PkgPath() != "net" {
		return fmt.Errorf("ipSchema only supports encoding net.IP values")
	}

	var bytesToEncode [4]byte

	bytesToEncode[0] = byte(v.Index(0).Uint())
	bytesToEncode[1] = byte(v.Index(1).Uint())
	bytesToEncode[2] = byte(v.Index(2).Uint())
	bytesToEncode[3] = byte(v.Index(3).Uint())

	fixedArraySchema := SchemaOf(bytesToEncode)
	err = fixedArraySchema.Encode(w, bytesToEncode)

	if err != nil {
		return err
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *ipv4Schema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded valuethe input stream and store it in v
func (s *ipv4Schema) DecodeValue(r io.Reader, v reflect.Value) error {

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

	var decodedBytes [4]byte

	fixedArraySchema := SchemaOf(decodedBytes)
	err = fixedArraySchema.Decode(r, &decodedBytes)

	if err != nil {
		return err
	}

	if k == reflect.Slice && t.Name() == "IP" || t.PkgPath() == "net" {
		v.Set(reflect.ValueOf(net.IPv4(decodedBytes[0], decodedBytes[1], decodedBytes[2], decodedBytes[3])))
		return nil
	}

	if s.weakDecoding {
		if k == reflect.String {
			v.SetString(string(decodedBytes[0]) + "." + string(decodedBytes[1]) + "." + string(decodedBytes[2]) + "." + string(decodedBytes[3]))
			return nil
		}
	}

	return fmt.Errorf("invalid destination")

}
