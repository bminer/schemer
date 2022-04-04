package schemer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type EnumSchema struct {
	SchemaOptions

	Values map[int]string
}

func (s *EnumSchema) GoType() reflect.Type {
	var t int
	retval := reflect.TypeOf(t)

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}
	return retval
}

func (s *EnumSchema) MarshalJSON() ([]byte, error) {

	tmpMap := make(map[string]interface{}, 3)
	tmpMap["type"] = "enum"
	tmpMap["nullable"] = s.Nullable()

	if len(s.Values) > 0 {
		tmpMap["values"] = s.Values
	}

	return json.Marshal(tmpMap)
}

// Bytes encodes the schema in a portable binary format
func (s EnumSchema) MarshalSchemer() ([]byte, error) {

	// fixed length schemas are 1 byte long total
	var schema []byte = []byte{EnumByte}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	// write all the enumerated values as part of the schema...
	var buf bytes.Buffer

	mapSchema := VarObjectSchema{
		Key:   &VarIntSchema{Signed: false},
		Value: &VarStringSchema{},
	}

	err := mapSchema.Encode(&buf, s.Values)
	if err != nil {
		return nil, err
	}

	schema = append(schema, buf.Bytes()...)

	return schema, nil

}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *EnumSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output streamtream
func (s *EnumSchema) EncodeValue(w io.Writer, v reflect.Value) error {
	varIntSchema := VarIntSchema{
		Signed:        false,
		SchemaOptions: SchemaOptions{nullable: s.Nullable()},
	}
	return varIntSchema.EncodeValue(w, v)
}

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *EnumSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *EnumSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// first we decode the actual encoded binary value

	varIntSchema := VarIntSchema{
		Signed:        false,
		SchemaOptions: SchemaOptions{nullable: s.Nullable()},
	}

	var decodedVal uint64
	var valPtr *uint64 = &decodedVal // we pass in a pointer to varIntSchema.Decode so we can potentially
	// get back a nil value [in case we are dealing with a nullable type]

	err := varIntSchema.Decode(r, &valPtr)
	if err != nil {
		return err
	}

	// now we check to see if varIntSchema.Decode returned us a nil value
	if valPtr == nil {
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			if v.CanSet() {
				v.Set(reflect.Zero(v.Type()))
				return nil
			}
			v = v.Elem()
			if v.CanSet() {
				v.Set(reflect.Zero(v.Type()))
				return nil
			}
			return fmt.Errorf("destination not settable")
		} else {
			return fmt.Errorf("cannot decode null value to non pointer to pointer type")
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

	if k == reflect.Interface {
		v.Set(reflect.New(s.GoType()))

		v = v.Elem().Elem()
		t = v.Type()
		k = t.Kind()
	}

	// check to see if the decoded value is in our map of enumerated values
	if s.Values != nil {
		if _, ok := s.Values[int(decodedVal)]; !ok {
			// however, maybe it makes sense to allow this scenario
			// when weak decoding is specified??
			if !s.WeakDecoding() {
				return fmt.Errorf("decoded enumerated value not in map")
			}
		}
	}

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
		intVal := int64(decodedVal)
		if uint64(intVal) != decodedVal || v.OverflowInt(intVal) {
			return fmt.Errorf("decoded value %d overflows destination %v", decodedVal, k)
		}
		v.SetInt(intVal)
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		if v.OverflowUint(decodedVal) {
			return fmt.Errorf("decoded value %d overflows destination %v", decodedVal, k)
		}
		v.SetUint(decodedVal)
	case reflect.String:
		if !s.WeakDecoding() {
			return fmt.Errorf("cannot decode enum to string without weak decoding enabled")
		}

		// if we have the map, return the string value of the constant
		if s.Values != nil {
			if _, ok := s.Values[int(decodedVal)]; ok {
				v.SetString(s.Values[int(decodedVal)])
				return nil
			}
		}

		// otherwise, just return a string version of the decoded integer value
		v.SetString(strconv.FormatUint(decodedVal, 10))
	default:
		return fmt.Errorf("decoded value %d incompatible with %v", decodedVal, k)
	}

	return nil
}
