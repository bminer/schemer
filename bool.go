package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type BoolSchema struct {
	SchemaOptions
}

func (s *BoolSchema) DefaultGOType() reflect.Type {
	var b bool
	return reflect.TypeOf(b)
}

func (s *BoolSchema) MarshalJSON() ([]byte, error) {

	tmpMap := make(map[string]interface{}, 2)
	tmpMap["type"] = "bool"
	tmpMap["nullable"] = strconv.FormatBool(s.SchemaOptions.Nullable)

	return json.Marshal(tmpMap)
}

// Bytes encodes the schema in a portable binary format
func (s *BoolSchema) MarshalSchemer() []byte {

	// bool schemas are 1 byte long
	var schema []byte = []byte{0b00011100}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schema[0] |= 128
	}

	return schema
}

func PreEncode(s Schema, w io.Writer, v *reflect.Value) (bool, error) {

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		*v = v.Elem()
	}

	if s.Nullable() {
		// did the caller pass in a nil value, or a null pointer?
		if !v.IsValid() {
			// per the revised spec, 1 indicates null
			w.Write([]byte{1})
			return false, nil
		} else {
			// 0 indicates not null
			w.Write([]byte{0})
		}
	} else {
		// if nullable is false
		// but they are trying to encode a nil value.. then that is an error
		if !v.IsValid() {
			return false, fmt.Errorf("cannot enoded nil value when IsNullable is false")
		}
	}

	return true, nil
}

func PreDecode(s Schema, r io.Reader, v reflect.Value) (reflect.Value, error) {
	// if t is a ptr or interface type, remove exactly ONE level of indirection
	if k := v.Kind(); !v.CanSet() && (k == reflect.Ptr || k == reflect.Interface) {
		v = v.Elem()
	}

	buf := make([]byte, 1)

	// if the data indicates this type is nullable, then the actual
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.Nullable() {

		// first byte indicates whether value is null or not...
		_, err := io.ReadAtLeast(r, buf, 1)
		if err != nil {
			return reflect.Value{}, err
		}
		valueIsNull := (buf[0] == 1)

		if valueIsNull {
			//fmt.Println("nullable", "and null", v.Kind(), v.CanSet())
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
				if v.CanSet() {
					// special way to set pointer to nil value
					v.Set(reflect.Zero(v.Type()))
					return reflect.Value{}, nil
				}
				return reflect.Value{}, fmt.Errorf("destination not settable")
			} else {
				return reflect.Value{}, fmt.Errorf("cannot decode null value to non pointer to pointer type")
			}
		}
	}

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		if v.IsNil() {
			if k == reflect.Interface {
				break
			}
			if !v.CanSet() {
				return reflect.Value{}, fmt.Errorf("decode destination is not settable")
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	return v, nil
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *BoolSchema) Encode(w io.Writer, i interface{}) error {

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
func (s *BoolSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)

}

func (s *BoolSchema) DecodeValue(r io.Reader, v reflect.Value) error {

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
		v.Set(reflect.New(s.DefaultGOType()))

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

func (s *BoolSchema) Nullable() bool {
	return s.SchemaOptions.Nullable
}

func (s *BoolSchema) SetNullable(n bool) {
	s.SchemaOptions.Nullable = n
}
