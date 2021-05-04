package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type BoolSchema struct {
	WeakDecoding bool
}

func (s BoolSchema) IsValid() bool {
	return true
}

func (s BoolSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *BoolSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Bytes encodes the schema in a portable binary format
func (s BoolSchema) Bytes() []byte {

	// floating point schemas are 1 byte long
	var schema []byte = make([]byte, 1)

	schema[0] = 0b011100

	// no options!

	return schema
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s BoolSchema) Encode(w io.Writer, v interface{}) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid BoolSchema schema")
	}

	value := reflect.ValueOf(v)
	t := value.Type()
	k := t.Kind()

	if k != reflect.Bool {
		return fmt.Errorf("BoolSchema only supports encoding boolean values")
	}

	var boolToEncode byte

	if value.Bool() {
		boolToEncode = 1
	} else {
		boolToEncode = 0
	}

	switch k {
	case reflect.Bool:

		n, err := w.Write([]byte{
			boolToEncode,
		})
		if err == nil && n != 1 {
			return errors.New("unexpected number of bytes written")
		}

	default:
		return errors.New("can only encode boolean types when using BoolSchema")
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s BoolSchema) Decode(r io.Reader, i interface{}) error {

	v := reflect.ValueOf(i)
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

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot decode using invalid ComplexNumber schema")
	}

	// take a look at the schema..

	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(r, buf, 1)
	if err != nil {
		return err
	}
	tmpByte := buf[0]
	decodedBool := tmpByte != 0

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
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode to int type")
		}
		if decodedBool {
			v.SetInt(1)
		} else {
			v.SetInt(0)
		}

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode to uint type")
		}
		if decodedBool {
			v.SetUint(1)
		} else {
			v.SetUint(0)
		}

	case reflect.Bool:
		v.SetBool(decodedBool)

	case reflect.String:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode to string")
		}
		if decodedBool {
			v.SetString("True")
		} else {
			v.SetString("False")
		}

	default:
		return fmt.Errorf("invalid destination %v", k)
	}

	return nil
}
