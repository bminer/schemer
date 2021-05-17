package schemer

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type EnumSchema struct {
	IsNullable   bool
	WeakDecoding bool

	Values map[int]string
}

func (s EnumSchema) IsValid() bool {
	return true
}

func (s EnumSchema) DoMarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s EnumSchema) DoUnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Bytes encodes the schema in a portable binary format
func (s EnumSchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = make([]byte, 1)

	schema[0] = 0b01110100 // bit pattern for enum

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	return schema

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s EnumSchema) Encode(w io.Writer, v interface{}) error {

	varIntSchema := VarIntSchema{Signed: true, IsNullable: s.IsNullable}

	if v == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	return varIntSchema.Encode(w, v)

}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s EnumSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)
}

func (s EnumSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// first we decode the actual encoded binary value

	varIntSchema := VarIntSchema{Signed: true, IsNullable: s.IsNullable}

	var decodedVarInt int64
	var intPtr *int64 = &decodedVarInt // we pass in a pointer to varIntSchema.Decode so we can potentially
	// get back a nil value [in case we are dealing with a nullable type]

	err := varIntSchema.Decode(r, &intPtr)
	if err != nil {
		return err
	}

	// now we check to see if varIntSchema.Decode returned us a nil value
	if intPtr == nil {
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
			if v.Kind() == reflect.Ptr {
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

	// check to see if the decoded value is in our map of enumerated values
	if s.Values != nil {
		if _, ok := s.Values[int(decodedVarInt)]; !ok {
			// however, maybe it makes sense to allow this scenario
			// when weak decoding is specified??
			if !s.WeakDecoding {
				return fmt.Errorf("decoded enumerated value not in map")
			}
		}
	}

	// if we are not dealing with a nil value
	// then we have to determine what to do with the value, based on where we are trying to decode it to

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

	// Write to destination
	// per the spec, we can decode enums to ints, enums, or strings
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
		if v.OverflowInt(decodedVarInt) {
			return fmt.Errorf("decoded value %d overflows destination %v", decodedVarInt, k)
		}
		v.SetInt(int64(decodedVarInt))
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		if decodedVarInt < 0 {
			return fmt.Errorf("decoded value %d incompatible with %v", decodedVarInt, k)
		}
		uintVal := uint64(decodedVarInt)
		if v.OverflowUint(uintVal) {
			return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
		}
		v.SetUint(uintVal)
	case reflect.String:
		if !s.WeakDecoding {
			return fmt.Errorf("cannot decode enum to string without weak decoding enabled")
		}

		// if we have the map, return the string value of the constant
		if s.Values != nil {
			if _, ok := s.Values[int(decodedVarInt)]; ok {
				v.SetString(s.Values[int(decodedVarInt)])
				return nil
			}
		}

		// otherwise, just return a string version of the decoded integer value
		v.SetString(strconv.FormatInt(decodedVarInt, 10))
	default:
		return fmt.Errorf("decoded value %d incompatible with %v", decodedVarInt, k)
	}

	return nil
}
