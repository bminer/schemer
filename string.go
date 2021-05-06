package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type StringSchema struct {
	IsNullable    bool
	IsFixedLength bool
	FixedLength   int
	WeakDecoding  bool
}

func (s StringSchema) IsValid() bool {

	// just if we are using fixed length encoding, make sure
	// they specified a fixed length...
	if s.IsFixedLength {
		return s.FixedLength > 0
	}

	return true
}

// fixme
func (s StringSchema) DoMarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

// fixme
func (s StringSchema) DoUnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Bytes encodes the schema in a portable binary format
func (s StringSchema) Bytes() []byte {

	// string schemas are 1 byte long
	var schema []byte = make([]byte, 1)

	schema[0] = 0b10000000

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	// bit number 3 indicates if the string is fixed length
	if s.IsFixedLength {
		schema[0] |= 4
	}

	return schema
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s StringSchema) Encode(w io.Writer, i interface{}) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid StringSchema schema")
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer?
		if i == nil ||
			(reflect.TypeOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil()) {

			if s.IsFixedLength {
				// per the spec, we encode a null value by writing a non 0 value...
				w.Write([]byte{1})
				return nil
			} else {
				// TODO: fixme!
				return fmt.Errorf("not implemented")
			}
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

	// if we are encoding a fixed len string, we just need to pad it
	if s.IsFixedLength {
		formatString := "%-" + strconv.Itoa(s.FixedLength) + "v"
		stringToEncode = fmt.Sprintf(formatString, stringToEncode)

		n, err := w.Write([]byte(stringToEncode))
		if err == nil && n != s.FixedLength {
			return errors.New("unexpected number of bytes written")
		}

	} else {
		// FIXME:
		// not implemented
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s StringSchema) Decode(r io.Reader, i interface{}) error {

	/*
		v := reflect.ValueOf(i)

		// just double check the schema they are using
		if !s.IsValid() {
			return fmt.Errorf("cannot decode using invalid StringSchema schema")
		}

		// data is proceeded by one byte, which means
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
						return fmt.Errorf("cannot decode null value to non pointer to (bool) pointer type")
					}
				} else {
					return fmt.Errorf("cannot decode null value to non pointer to (bool) pointer type")
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



		// take a look at the destination
		// bools can be decoded to integer types, bools, and strings
		switch k {

		case reflect.String:
			//

		default:
			return fmt.Errorf("invalid destination %v", k)
		}
	*/

	return nil
}
