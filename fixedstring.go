package schemer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type FixedStringSchema struct {
	SchemaOptions
	Length int
}

func (s *FixedStringSchema) GoType() reflect.Type {
	var t string
	retval := reflect.TypeOf(t)

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}

	return retval
}

func (s *FixedStringSchema) Valid() bool {
	return s.Length > 0
}

func (s *FixedStringSchema) MarshalJSON() ([]byte, error) {

	if !s.Valid() {
		return nil, fmt.Errorf("invalid FixedStringSchema")
	}

	return json.Marshal(map[string]interface{}{
		"type":     "string",
		"length":   s.Length,
		"nullable": s.Nullable(),
	})

}

// Bytes encodes the schema in a portable binary format
func (s *FixedStringSchema) MarshalSchemer() ([]byte, error) {

	// string schemas are 1 byte long
	var schema []byte = []byte{FixedStringByte}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	// set bit 1, which indicates this is fixed len string
	schema[0] |= 1

	// encode fixed length as a varint
	buf := make([]byte, binary.MaxVarintLen64)
	varIntByteLength := binary.PutVarint(buf, int64(s.Length))

	schema = append(schema, buf[0:varIntByteLength]...)

	return schema, nil
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *FixedStringSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output stream
func (s *FixedStringSchema) EncodeValue(w io.Writer, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot encode using invalid FixedStringSchema")
	}

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.String {
		return fmt.Errorf("FixedStringSchema only supports encoding string values")
	}

	var stringToEncode string = v.String()

	// if we are encoding a fixed len string, we just need to pad it

	formatString := "%-" + strconv.Itoa(s.Length) + "v"
	stringToEncode = fmt.Sprintf(formatString, stringToEncode)

	_, err = w.Write([]byte(stringToEncode))
	if err != nil {
		return err
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *FixedStringSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *FixedStringSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot decode using invalid FixedStringSchema")
	}

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

	var decodedString string

	buf := make([]byte, s.Length)
	_, err = io.ReadAtLeast(r, buf, s.Length)
	if err != nil {
		return err
	}

	// when we return as a string, we will return it with the padding intact
	decodedString = string(buf)

	// but for conversions, having a trimmed up string will make things easier
	trimString := strings.Trim(decodedString, " ")

	// Ensure v is settable
	if !v.CanSet() {
		return fmt.Errorf("decode destination is not settable")
	}

	// take a look at the destination
	// bools can be decoded to integer types, bools, and strings
	switch k {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		i, err := strconv.Atoi(trimString)
		if err != nil {
			return err
		}
		if v.OverflowInt(int64(i)) {
			return fmt.Errorf("decoded float overlows destination")
		}
		v.SetInt(int64(i))
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		i, err := strconv.Atoi(trimString)
		if err != nil {
			return err
		}
		if v.OverflowUint(uint64(i)) {
			return fmt.Errorf("decoded float overlows destination")
		}
		v.SetUint(uint64(i))
	case reflect.Float32:
		vFloat, err := strconv.ParseFloat(trimString, 32)
		if err != nil {
			return err
		}
		if v.OverflowFloat(vFloat) {
			return fmt.Errorf("decoded float overlows destination")
		}
		v.SetFloat(vFloat)
	case reflect.Float64:
		vFloat, err := strconv.ParseFloat(trimString, 64)
		if err != nil {
			return err
		}
		if v.OverflowFloat(vFloat) {
			return fmt.Errorf("decoded float overlows destination")
		}
		v.SetFloat(vFloat)
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		// see if we can put it into the destination??
		return fmt.Errorf("not implemented")
	case reflect.Bool:
		if !s.WeakDecoding() {
			return fmt.Errorf("cannot decode int to bool without weak decoding")
		}
		if trimString == "1" || trimString == "t" || trimString == "T" || trimString == "TRUE" || trimString == "true" || trimString == "True" {
			v.SetBool(true)
			return nil
		}
		if trimString == "0" || trimString == "f" || trimString == "F" || trimString == "FALSE" || trimString == "false" || trimString == "False" {
			v.SetBool(false)
			return nil
		}

		return fmt.Errorf("cannot decode string value %s into bool value", trimString)
	case reflect.String:
		v.SetString(decodedString)
	default:
		return fmt.Errorf("invalid destination %v", k)
	}

	return nil
}
