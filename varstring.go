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

// fixme
func (s *VarLenStringSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

// Bytes encodes the schema in a portable binary format
func (s *VarLenStringSchema) Bytes() []byte {

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
func (s *VarLenStringSchema) Encode(w io.Writer, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	v := reflect.ValueOf(i)

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer?
		if !v.IsValid() {

			fmt.Println("value encoded as a null...")

			// per the revised spec, 1 indicates null
			w.Write([]byte{1})
			return nil
		} else {
			// 0 indicates not null
			w.Write([]byte{0})
		}
	} else {
		// if nullable is false
		// but they are trying to encode a nil value.. then that is an error
		if !v.IsValid() {
			return fmt.Errorf("cannot enoded nil value when IsNullable is false")
		}
		// 0 indicates not null
		w.Write([]byte{0})
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
func (s *VarLenStringSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// first byte indicates whether value is null or not....
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(r, buf, 1)
	if err != nil {
		return err
	}
	valueIsNull := (buf[0] == 1)

	// if the data indicates this type is nullable, then the actual
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.IsNullable {
		if valueIsNull {
			if v.Kind() == reflect.Ptr {
				if v.CanSet() {
					v.Set(reflect.Zero(v.Type()))
					return nil
				} else {
					return fmt.Errorf("destination not settable")
				}
			} else {
				return fmt.Errorf("cannot decode null value to non pointer to (string) pointer type")
			}
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

	// Ensure v is settable
	// however: one important thing is to do make sure to process as many bytes as we are going to
	// read from r before we do this
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

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *VarLenStringSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)
}

func (s *VarLenStringSchema) Nullable() bool {
	return s.IsNullable
}

func (s *VarLenStringSchema) SetNullable(n bool) {
	s.IsNullable = n
}
