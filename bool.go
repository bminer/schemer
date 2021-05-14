package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type BoolSchema struct {
	IsNullable   bool
	WeakDecoding bool
}

func (s BoolSchema) IsValid() bool {
	return true
}

// if this function is called MarshalJSON it seems to be called
// recursively by the json library???
func (s BoolSchema) DoMarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid floating point schema")
	}

	return json.Marshal(s)
}

// if this function is called UnmarshalJSON it seems to be called
// recursively by the json library???
func (s BoolSchema) DoUnmarshalJSON(buf []byte) error {
	return json.Unmarshal(buf, s)
}

// Bytes encodes the schema in a portable binary format
func (s BoolSchema) Bytes() []byte {

	// bool schemas are 1 byte long
	var schema []byte = make([]byte, 1)

	schema[0] = 0b01110000

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	return schema
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s BoolSchema) Encode(w io.Writer, i interface{}) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid BoolSchema schema")
	}

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer
		if reflect.TypeOf(i).Kind() == reflect.Ptr ||
			reflect.TypeOf(i).Kind() == reflect.Interface &&
				reflect.ValueOf(i).IsNil() {
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

	v := reflect.ValueOf(i)

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		i = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	if k != reflect.Bool {
		return fmt.Errorf("BoolSchema only supports encoding boolean values")
	}

	var boolToEncode byte

	if v.Bool() {
		// we are trying to encode a true value
		// (but we have to make sure that the most sig bit is not set, because
		boolToEncode = 254
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

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)

}

func (s BoolSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot decode using invalid BoolSchema schema")
	}

	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(r, buf, 1)
	if err != nil {
		return err
	}

	// if the schema indicates this type is nullable
	if s.IsNullable {
		// if value is null, set nil to top level
		if buf[0] != 0 {
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
				v = v.Elem()
				if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
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

	// Ensure v is settable
	if !v.CanSet() {
		return fmt.Errorf("decode destination is not settable")
	}

	decodedBool := buf[0] > 0

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
