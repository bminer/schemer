package schemer

import (
	"fmt"
	"io"
	"reflect"
)

// https://golangbyexample.com/go-size-range-int-uint/
const uintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

type Schema interface {
	// Encode uses the schema to write the encoded value of v to the output stream
	Encode(w io.Writer, i interface{}) error
	// Decode uses the schema to read the next encoded value from the input stream and store it in v
	//Decode(r io.Reader, i interface{}) error
	// MarshalSchemer encodes the schema in a portable binary format
	// MarshalJSON returns the JSON encoding of the schema
	DoMarshalJSON() ([]byte, error)
	// UnmarshalJSON updates the schema by decoding the JSON-encoded schema in b
	DoUnmarshalJSON(b []byte) error
	// Nullable returns true if and only if the type is nullable
	/*
		Nullable() bool
		// SetNullable sets the nullable flag for the schema
		SetNullable(n bool)
	*/

	DecodeValue(r io.Reader, v reflect.Value) error
}

// functions to create schema

func CreateFixedIntegerSchema(signed bool, bits int, isNullable bool) Schema {

	var floatSchema FloatSchema

	floatSchema.Bits = bits
	floatSchema.IsNullable = isNullable

	return floatSchema

}

func CreateComplexSchema(bits int) Schema {
	var complexSchema ComplexSchema

	complexSchema.Bits = bits

	return complexSchema
}

func CreateBooleanSchema() Schema {
	var boolSchema BoolSchema

	return boolSchema
}

func CreateFloatSchema(bits int, isNullable bool) Schema {

	var floatSchema FloatSchema

	floatSchema.Bits = bits
	floatSchema.IsNullable = isNullable

	return floatSchema
}

func CreateFixedLenStringSchema(IsNullable bool, FixedLength int) Schema {

	var fixedLenStringSchema FixedLenStringSchema

	fixedLenStringSchema.IsNullable = IsNullable
	fixedLenStringSchema.FixedLength = FixedLength

	return fixedLenStringSchema

}

func SchemaOfType(t reflect.Type) Schema {

	k := t.Kind()

	switch k {

	case reflect.Int:
		return CreateFixedIntegerSchema(true, uintSize/8, false)

	case reflect.Int8:
		return CreateFixedIntegerSchema(true, 8, false)

	case reflect.Int16:
		return CreateFixedIntegerSchema(true, 16, false)

	case reflect.Int32:
		return CreateFixedIntegerSchema(true, 32, false)

	case reflect.Int64:
		return CreateFixedIntegerSchema(true, 64, false)

	case reflect.Uint:
		return CreateFixedIntegerSchema(false, uintSize/8, false)

	case reflect.Uint8:
		return CreateFixedIntegerSchema(false, 8, false)

	case reflect.Uint16:
		return CreateFixedIntegerSchema(false, 16, false)

	case reflect.Uint32:
		return CreateFixedIntegerSchema(false, 32, false)

	case reflect.Uint64:
		return CreateFixedIntegerSchema(false, 64, false)

	case reflect.Complex64:
		return CreateComplexSchema(64)

	case reflect.Complex128:
		return CreateComplexSchema(128)

	case reflect.Bool:
		return CreateBooleanSchema()

	case reflect.Float32:
		return CreateFloatSchema(32, false)

	case reflect.Float64:
		return CreateFloatSchema(64, false)

	}

	return nil
}

// SchemaOf generates a Schema from the concrete value stored in the interface i.
// SchemaOf(nil) returns a Schema for an empty struct.
func SchemaOf(i interface{}) Schema {
	v := reflect.ValueOf(i)
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}
	t := v.Type()

	return SchemaOfType(t)
}

// NewSchema decodes a schema stored in buf and returns an error if the schema is invalid
func NewSchema(buf []byte) (Schema, error) {

	var bit3IsSet bool

	var fixedIntSchema FixedIntSchema
	var varIntSchema VarIntSchema
	var floatSchema FloatSchema
	var complexSchema ComplexSchema
	var boolSchema BoolSchema
	var fixedLenStringSchema FixedLenStringSchema
	var varLenStringSchema VarLenStringSchema
	var enumSchema EnumSchema
	var fixedLenArraySchema FixedLenArraySchema
	var varLenArraySchema VarLenArraySchema

	// decode fixed int schema
	// (bits 7 and 8 should be clear)
	if buf[0]&192 == 0 {
		fixedIntSchema.IsNullable = (buf[0]&1 == 1)
		fixedIntSchema.Signed = (buf[0] & 4) == 4
		fixedIntSchema.Bits = 8 << ((buf[0] & 56) >> 3)

		return fixedIntSchema, nil
	}

	// decode varint schema
	// (bits 7 should be set)
	if buf[0]&112 == 64 {
		varIntSchema.IsNullable = (buf[0]&1 == 1)
		varIntSchema.Signed = (buf[0] & 4) == 4

		return varIntSchema, nil
	}

	// decode floating point schema
	// (bits 5 and 7 should be set)
	if buf[0]&112 == 80 {
		bit3IsSet = (buf[0] & 4) == 4
		if bit3IsSet {
			floatSchema.Bits = 64
		} else {
			floatSchema.Bits = 32
		}
		floatSchema.IsNullable = (buf[0]&1 == 1)

		return floatSchema, nil
	}

	// decode complex number
	// (bits 6 and 7 should be set)
	if buf[0]&112 == 96 {
		bit3IsSet = (buf[0] & 4) == 4
		if bit3IsSet {
			complexSchema.Bits = 128
		} else {
			complexSchema.Bits = 64
		}
		complexSchema.IsNullable = (buf[0]&1 == 1)

		return complexSchema, nil
	}

	// decode boolean
	// (bits 5,6,7 are all set)
	if buf[0]&116 == 112 {
		boolSchema.IsNullable = (buf[0]&1 == 1)
		return boolSchema, nil
	}

	// decode enum
	// (bits 3,5,6,7 should all be set)
	if buf[0]&116 == 116 {
		enumSchema.IsNullable = (buf[0]&1 == 1)

		return enumSchema, nil
	}

	// decode fixed len string
	// (bits 8 and 3 should be set)
	if buf[0]&252 == 132 {
		fixedLenStringSchema.IsNullable = (buf[0]&1 == 1)
		//fixedLenStringSchema.FixedLength = ???
		//the binary schema does not encode the length

		return fixedLenStringSchema, nil
	}

	// decode var len string
	// (bits 8 should be set, bit 3 should be clear)
	if buf[0]&252 == 128 {
		varLenStringSchema.IsNullable = (buf[0]&1 == 1)

		return varLenStringSchema, nil
	}

	// decode fixed array schema
	// (bits 3, 5, 8)
	if buf[0]&252 == 148 {
		fixedLenArraySchema.IsNullable = (buf[0]&1 == 1)

		return fixedLenArraySchema, nil
	}

	// decode var array schema
	// (bits 3, 5, 8)
	if buf[0]&252 == 144 {
		varLenArraySchema.IsNullable = (buf[0]&1 == 1)

		return varLenArraySchema, nil
	}

	//Object

	//Variant

	//Schema

	//Custom Type

	return nil, fmt.Errorf("invalid binary schema encountered")
}
