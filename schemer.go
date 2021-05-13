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
	Decode(r io.Reader, i interface{}) error
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

func CreateFixedIntegerSchema(signed bool, bits int, isNullable bool) FixedIntSchema {

	var fixedIntSchema FixedIntSchema

	fixedIntSchema.Signed = signed
	fixedIntSchema.Bits = bits
	fixedIntSchema.IsNullable = isNullable

	return fixedIntSchema
}

func CreateComplexSchema(bits int) ComplexSchema {
	var complexSchema ComplexSchema

	complexSchema.Bits = bits

	return complexSchema
}

func CreateBooleanSchema() BoolSchema {
	var boolSchema BoolSchema

	return boolSchema
}

func CreateFloatSchema(bits int, isNullable bool) FloatSchema {

	var floatSchema FloatSchema

	floatSchema.Bits = bits
	floatSchema.IsNullable = isNullable

	return floatSchema
}

func CreateFixedLenStringSchema(IsNullable bool, FixedLength int) FixedLenStringSchema {

	var fixedLenStringSchema FixedLenStringSchema

	fixedLenStringSchema.IsNullable = IsNullable
	fixedLenStringSchema.FixedLength = FixedLength

	return fixedLenStringSchema

}

func CreateVarLenStringSchema(IsNullable bool) VarLenStringSchema {
	var varStringSchema VarLenStringSchema

	varStringSchema.IsNullable = IsNullable

	return varStringSchema
}

func CreateFixedArraySchema(IsNullable bool, FixedLength int) FixedLenArraySchema {

	var fixedLenArraySchema FixedLenArraySchema

	fixedLenArraySchema.IsNullable = IsNullable
	fixedLenArraySchema.Length = FixedLength

	return fixedLenArraySchema

}

func CreateVarArraySchema(IsNullable bool) VarArraySchema {

	var varArraySchema VarArraySchema

	varArraySchema.IsNullable = IsNullable

	return varArraySchema

}

func CreateFixedObjectSchema(IsNullable bool) FixedObjectSchema {

	var fixedObjectSchema FixedObjectSchema

	fixedObjectSchema.IsNullable = IsNullable

	return fixedObjectSchema
}

func CreateVarObjectSchema(IsNullable bool) VarObjectSchema {

	var varObjectSchema VarObjectSchema

	varObjectSchema.IsNullable = IsNullable

	return varObjectSchema
}

func SchemaOf(i interface{}) Schema {

	// spec says: "SchemaOf(nil) returns a Schema for an empty struct."
	if i == nil {
		return CreateFixedObjectSchema(true)
	}

	v := reflect.ValueOf(i)
	return SchemaForValue(v)
}

// SchemaOf generates a Schema from the concrete value stored in the interface i.
// SchemaOf(nil) returns a Schema for an empty struct.
func SchemaForValue(v reflect.Value) Schema {

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {

		if v.IsNil() {
			// maybe we need way to return an error here...
			/*
				if !v.CanSet() {
					return fmt.Errorf("decode destination is not settable")
				}
			*/
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	t := v.Type()
	k := t.Kind()

	switch k {

	case reflect.Map:
		varObjectSchema := CreateVarObjectSchema(true)

		for _, mapKey := range v.MapKeys() {
			mapValue := v.MapIndex(mapKey)

			varObjectSchema.Key = SchemaForValue(mapKey)
			varObjectSchema.Value = SchemaForValue(mapValue)

		}

		return varObjectSchema

	case reflect.Struct:
		fixedObjectSchema := CreateFixedObjectSchema(true)
		var of ObjectField

		for i := 0; i < t.NumField(); i++ {

			f := v.Field(i)

			of.Name = t.Field(i).Name
			of.Schema = SchemaForValue(f)

			fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of)

		}

		return fixedObjectSchema

	case reflect.Slice:
		tmp := CreateVarArraySchema(true)
		tmp.Element = SchemaForValue(v.Index(0))
		return tmp

	case reflect.Array:
		tmp := CreateFixedArraySchema(true, v.Len())
		tmp.Element = SchemaForValue(v.Index(0))
		return tmp

	case reflect.String:
		return CreateVarLenStringSchema(true)

	case reflect.Int:
		return CreateFixedIntegerSchema(true, uintSize, false)

	case reflect.Int8:
		return CreateFixedIntegerSchema(true, 8, false)

	case reflect.Int16:
		return CreateFixedIntegerSchema(true, 16, false)

	case reflect.Int32:
		return CreateFixedIntegerSchema(true, 32, false)

	case reflect.Int64:
		return CreateFixedIntegerSchema(true, 64, false)

	case reflect.Uint:
		return CreateFixedIntegerSchema(false, uintSize, false)

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
	var varArraySchema VarArraySchema
	var varObjectSchema VarObjectSchema
	var fixedObjectSchema FixedObjectSchema

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
		varArraySchema.IsNullable = (buf[0]&1 == 1)

		return varArraySchema, nil
	}

	// decode var object schema
	if buf[0]&252 == 160 {
		varObjectSchema.IsNullable = (buf[0]&1 == 1)

		return varObjectSchema, nil
	}

	// fixed object schema
	if buf[0]&252 == 164 {
		fixedObjectSchema.IsNullable = (buf[0]&1 == 1)

		return fixedObjectSchema, nil
	}

	//Variant

	//Schema

	//Custom Type

	return nil, fmt.Errorf("invalid binary schema encountered")
}
