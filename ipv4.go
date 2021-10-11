package schemer

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
)

// each custom type has a unique name an a unique UUID
const ipV4SchemaUUID byte = 2

type ipv4Schema struct {
	SchemaOptions
}

type ipv4SchemaGenerator struct{}

func (sg ipv4SchemaGenerator) SchemaOfType(t reflect.Type) (Schema, error) {
	nullable := false

	// Dereference pointer / interface types
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()

		// If we encounter any pointers, then we know this type is nullable
		nullable = true
	}

	if t.Name() == "IP" && t.PkgPath() == "net" {
		s := &ipv4Schema{}
		s.SetNullable(nullable)
		return s, nil
	}

	return nil, nil
}

func (sg ipv4SchemaGenerator) DecodeSchema(r io.Reader) (Schema, error) {

	tmpBuf := make([]byte, 1)
	_, err := r.Read(tmpBuf)
	if err != nil {
		return nil, err
	}

	if tmpBuf[0]&CustomMask == CustomMask {
		if tmpBuf[0]&(ipV4SchemaUUID<<4) != (ipV4SchemaUUID << 4) {
			return nil, nil
		}
	} else {
		return nil, nil
	}

	nullable := (tmpBuf[0]&NullMask == NullMask)

	s := ipv4Schema{}
	s.SetNullable(nullable)
	return &s, nil

}

func (sg ipv4SchemaGenerator) DecodeSchemaJSON(r io.Reader) (Schema, error) {

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]interface{})

	err = json.Unmarshal(buf, &fields)
	if err != nil {
		return nil, err
	}

	// Parse `type`
	tmp, ok := fields["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid schema type")
	}
	typeStr := strings.ToLower(tmp)

	if typeStr != "ipv4" {
		return nil, nil
	}

	// Parse `nullable`
	nullable := false
	tmp1, found := fields["nullable"]
	if found {
		if b, ok := tmp1.(bool); ok {
			nullable = b
		} else {
			return nil, fmt.Errorf("nullable must be a boolean")
		}
	}

	s := ipv4Schema{}
	s.SetNullable(nullable)
	return &s, nil
}

// returns an ipv4Schema if the passed in reflect.type is a net.IP
func (s *ipv4Schema) ForType(t reflect.Type) Schema {
	if t.Name() == "IP" && t.PkgPath() == "net" {
		return s
	}

	return nil
}

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
		"type":     "ipv4",
		"nullable": s.Nullable(),
	})
}

// Bytes encodes the schema in a portable binary format
func (s *ipv4Schema) MarshalSchemer() ([]byte, error) {

	const schemerDateSize byte = 1

	// string schemas are 1 byte long
	var schema []byte = make([]byte, schemerDateSize)

	schema[0] |= CustomMask
	schema[0] |= (ipV4SchemaUUID << 4)

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= 0x80
	}

	return schema, nil
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *ipv4Schema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of he output stream
func (s *ipv4Schema) EncodeValue(w io.Writer, v reflect.Value) error {

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
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

	fixedArraySchema, err := SchemaOf(bytesToEncode)
	if err != nil {
		return err
	}

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

	done, err := PreDecode(r, &v, s.Nullable())
	if err != nil || done {
		return err
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

	fixedArraySchema, err := SchemaOf(decodedBytes)
	if err != nil {
		return err
	}
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
