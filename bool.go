package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// BoolSchema is a Schema for encoding and decoding boolean values
type BoolSchema struct {
	SchemaOptions
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *BoolSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output
// stream
func (s *BoolSchema) EncodeValue(w io.Writer, v reflect.Value) error {

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Bool {
		return fmt.Errorf("BoolSchema only supports encoding boolean values")
	}

	var boolToEncode byte

	if v.Bool() {
		// we are trying to encode a true value
		boolToEncode = 1
	}

	switch k {
	case reflect.Bool:
		n, err := w.Write([]byte{boolToEncode})
		if err == nil && n != 1 {
			return errors.New("unexpected number of bytes written")
		}

	default:
		return errors.New("can only encode boolean types when using BoolSchema")
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input
// stream and stores it in i
func (s *BoolSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input
// stream and stores it in v
func (s *BoolSchema) DecodeValue(r io.Reader, v reflect.Value) error {

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

	buf := make([]byte, 1)

	_, err = io.ReadAtLeast(r, buf, 1)
	if err != nil {
		return err
	}

	decodedBool := buf[0] > 0

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
		if !s.WeakDecoding() {
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
		if !s.WeakDecoding() {
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
		if !s.WeakDecoding() {
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

// GoType returns the default Go type that represents the schema
func (s *BoolSchema) GoType() reflect.Type {
	var b bool
	retval := reflect.TypeOf(b)

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}
	return retval
}

// MarshalJSON encodes the schema in a JSON format
func (s *BoolSchema) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":     "bool",
		"nullable": s.Nullable(),
	})
}

// MarshalSchemer encodes the schema in a portable binary format
func (s *BoolSchema) MarshalSchemer() ([]byte, error) {
	// bool schemas are 1 byte long
	var schema []byte = []byte{BoolByte}

	// The most significant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	return schema, nil
}
