package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type VarLenStringSchema struct {
	IsNullable   bool
	WeakDecoding bool
}

func (s VarLenStringSchema) IsValid() bool {

	return true
}

// fixme
func (s VarLenStringSchema) DoMarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

// fixme
func (s VarLenStringSchema) DoUnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Bytes encodes the schema in a portable binary format
func (s VarLenStringSchema) Bytes() []byte {

	// string schemas are 1 byte long
	var schema []byte = make([]byte, 1)

	schema[0] = 0b10000000

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	// bit 3 is clear from above, indicating this is a var length string

	return schema
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s VarLenStringSchema) Encode(w io.Writer, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid StringSchema schema")
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer?
		if i == nil ||
			(reflect.TypeOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil()) {

			// per the spec, we encode a null value by writing out a single byte
			// with the high bit set
			w.Write([]byte{128})
			return nil
		} else {
			// TODO: update the SPEC (encoding-format.md)
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
		i = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	if k != reflect.String {
		return fmt.Errorf("StringSchema only supports encoding string values")
	}

	var stringToEncode string = v.String()
	var stringLen int = len(stringToEncode)

	err := writeVarUint(w, uint64(stringLen))
	if err != nil {
		return errors.New("cannot encode var string length as var int")
	}

	n, err := w.Write([]byte(stringToEncode))
	if err == nil && n != stringLen {
		return errors.New("unexpected number of bytes written")
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s VarLenStringSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot decode using invalid VarLenStringSchema schema")
	}

	// TODO: update the spec to reflect this change
	// data is proceeded by one byte which tells us if the data is null or not
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(r, buf, 1)
	if err != nil {
		return err
	}
	tmpByte := buf[0]

	// if the data indicates this type is nullable, then the actual floating point
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.IsNullable {
		if tmpByte == 1 {
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
				if v.Kind() == reflect.Ptr {
					// special way to return a nil pointer
					v.Set(reflect.Zero(v.Type()))
				} else {
					return fmt.Errorf("cannot decode null value to non pointer to (string) pointer type")
				}
			} else {
				return fmt.Errorf("cannot decode null value to non pointer to (string) pointer type")
			}
			return nil
		}
	}

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	// Ensure v is settable
	if !v.CanSet() {
		return fmt.Errorf("decode destination is not settable")
	}

	expectedLen, err := readVarUint(r)
	if err != nil {
		return err
	}

	buf = make([]byte, int(expectedLen))
	_, err = io.ReadAtLeast(r, buf, int(expectedLen))
	if err != nil {
		return err
	}

	// when we return as a string, we will return it with the padding intact
	decodedString := string(buf)

	// but for conversions, having a trimmed up string will make things easier
	trimString := strings.Trim(decodedString, " ")

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
		if !s.WeakDecoding {
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
