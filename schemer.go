package schemer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// SchemaTagName represents the tag prefix that the schemer library uses on struct tags
var SchemaTagName string = "schemer"

// StructFieldOptions represents information that can be read from struct field tags
type StructFieldOptions struct {
	FieldAliases []string

	Nullable     bool
	NotNull      bool // this is a separate field because it is an override
	WeakDecoding bool
	ShouldSkip   bool
}

type SchemaOptions struct {
	WeakDecoding bool
	Nullable     bool
}

type Schema interface {
	// Encode uses the schema to write the encoded value of v to the output stream
	Encode(w io.Writer, i interface{}) error

	// Decode uses the schema to read the next encoded value from the input stream and store it in v
	Decode(r io.Reader, i interface{}) error

	// MarshalJSON returns the JSON encoding of the schema
	MarshalJSON() ([]byte, error)

	// Nullable returns true if and only if the type is nullable
	Nullable() bool

	//SetNullable sets the nullable flag for the schema
	SetNullable(n bool)

	DecodeValue(r io.Reader, v reflect.Value) error

	// Bytes encodes the schema in a portable binary format
	Bytes() []byte
}

var byteIndex int = 0

func SchemaOf(i interface{}) Schema {

	// spec says: "SchemaOf(nil) returns a Schema for an empty struct."
	if i == nil {
		var fixedObjectSchema *FixedObjectSchema = &(FixedObjectSchema{})
		fixedObjectSchema.IsNullable = true
		return fixedObjectSchema
	}

	t := reflect.TypeOf(i)

	// if t is a ptr or interface type, remove exactly ONE level of indirection
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return SchemaOfType(t)
}

// parseStructTag tags a tagname as a string, parses it, and populates FieldOptions
// the format of the tag must be:
// tag := (alias)?("," option)*
// alias := identifier
//			"["identifier(","identifier)*"]"
//	option := "weak", "null", "not null"
func parseStructTag(tagStr string, structFieldOptions *StructFieldOptions) error {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error parsing struct tag:", err)
		}
	}()

	tagStr = strings.Trim(tagStr, " ")
	if len([]rune(tagStr)) == 0 {
		return nil
	}

	// special case meaning to skip this field
	if tagStr == "-" {
		structFieldOptions.ShouldSkip = true
		return nil
	}

	// if first part has a "]", then extract everything up to there
	// otherwise, extract everything up to the first comma

	var i int
	var aliasStr string
	var optionStr string

	// if the alias portion of the string contains [], then we want to grab everything up
	// to the ] and call that our aliasStr
	if strings.Contains(tagStr, "]") {

		i = strings.Index(tagStr, "]")
		aliasStr = tagStr[0 : i+1]
		tagStr = tagStr[i+1:] // eat off what we just processed
		tagStr = strings.Trim(tagStr, " ")

		if len([]rune(tagStr)) > 0 {

			if !strings.Contains(tagStr, ",") {
				return fmt.Errorf("missing comma after field alias")
			} else {
				// our options are just whatever is left after the comma
				optionStr = tagStr[strings.Index(tagStr, ",")+1:]
				optionStr = strings.Trim(optionStr, " ")
			}

		} else {
			optionStr = ""
		}
	} else {
		i = strings.Index(tagStr, ",")

		if i > 0 {

			// alias string is everything up to the comma
			aliasStr = tagStr[0:i]
			aliasStr = strings.Trim(aliasStr, " ")
			tagStr = tagStr[i+1:] // eat off what we just processed
			tagStr = strings.Trim(tagStr, " ")

			if len([]rune(tagStr)) > 0 {
				// options are everything after the comma
				optionStr = tagStr[strings.Index(tagStr, ",")+1:]
				optionStr = strings.Trim(optionStr, " ")
			} else {
				optionStr = ""
			}
		} else {
			aliasStr = strings.Trim(tagStr, " ")
			optionStr = ""
		}

	}

	// parse aliasStr, and put each field into .FieldAliases
	x := strings.Replace(aliasStr, "[", "", -1)
	y := strings.Replace(x, "]", "", -1)
	structFieldOptions.FieldAliases = strings.Split(y, ",")
	for i, f := range structFieldOptions.FieldAliases {
		structFieldOptions.FieldAliases[i] = strings.Trim(f, " ")
	}

	// parse options, and string and put each option into correct field, such as .Nullable
	structFieldOptions.Nullable = strings.Contains(strings.ToUpper(optionStr), "NUL")
	structFieldOptions.NotNull = strings.Contains(strings.ToUpper(optionStr), "!NUL")
	structFieldOptions.WeakDecoding = strings.Contains(strings.ToUpper(optionStr), "WEAK")

	return nil

}

// SchemaOfType returns a schema for the specified reflection type
func SchemaOfType(t reflect.Type) Schema {

	var shouldBeNullable = false

	// Dereference pointer / interface types
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()

		// if we encounter any pointers, then we know it makes sense
		// to allow this type to be nullable
		shouldBeNullable = true
	}

	k := t.Kind()

	switch k {
	case reflect.Map:
		var varObjectSchema *VarObjectSchema = &(VarObjectSchema{})

		varObjectSchema.IsNullable = shouldBeNullable
		varObjectSchema.Key = SchemaOfType(t.Key())
		varObjectSchema.Value = SchemaOfType(t.Elem())

		return varObjectSchema
	case reflect.Struct:
		var fixedObjectSchema *FixedObjectSchema = &(FixedObjectSchema{})
		fixedObjectSchema.IsNullable = shouldBeNullable

		var of ObjectField

		for i := 0; i < t.NumField(); i++ {
			of.Schema = SchemaOfType(t.Field(i).Type)

			structFieldOptions := StructFieldOptions{}

			parseStructTag(t.Field(i).Tag.Get(SchemaTagName), &structFieldOptions)
			// ignore result here... if an error parsing the tags occured, just don't worry about it
			// (in case they are in the wrong format or something)

			// if no tags aliases exist, then use the actual field name from the struct
			if len(structFieldOptions.FieldAliases) == 0 {
				structFieldOptions.FieldAliases = make([]string, 1)
				structFieldOptions.FieldAliases[0] = t.Field(i).Name
			}

			of.StructFieldOptions = structFieldOptions

			// check if this field is not exported (by looking at PkgPath)
			// or if the schemer tags on the field say that we should skip it...
			if len(t.Field(i).PkgPath) == 0 && !of.StructFieldOptions.ShouldSkip {
				fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of)
			} else {
				fmt.Println("skipped field:", t.Field(i).Name)
			}
		}
		return fixedObjectSchema
	case reflect.Slice:
		var varArraySchema *VarArraySchema = &(VarArraySchema{})
		varArraySchema.IsNullable = shouldBeNullable
		varArraySchema.Element = SchemaOfType(t.Elem())
		return varArraySchema
	case reflect.Array:
		var fixedLenArraySchema *FixedLenArraySchema = &(FixedLenArraySchema{})

		fixedLenArraySchema.IsNullable = shouldBeNullable
		fixedLenArraySchema.Length = t.Len()

		fixedLenArraySchema.Element = SchemaOfType(t.Elem())
		return fixedLenArraySchema
	case reflect.String:
		var varStringSchema *VarLenStringSchema = &(VarLenStringSchema{})

		varStringSchema.IsNullable = shouldBeNullable

		return varStringSchema
	case reflect.Int:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema

	case reflect.Int8:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Int16:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Int32:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Int64:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint8:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint16:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint32:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint64:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.IsNullable = shouldBeNullable

		return varIntSchema
	case reflect.Complex64:
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		complexSchema.Bits = 64
		complexSchema.SchemaOptions.Nullable = shouldBeNullable

		return complexSchema
	case reflect.Complex128:
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		complexSchema.SchemaOptions.Nullable = shouldBeNullable
		complexSchema.Bits = 128

		return complexSchema
	case reflect.Bool:
		var boolSchema *BoolSchema = &(BoolSchema{})

		boolSchema.SchemaOptions.Nullable = shouldBeNullable

		return boolSchema
	case reflect.Float32:
		var floatSchema *FloatSchema = &(FloatSchema{})

		floatSchema.Bits = 32
		floatSchema.IsNullable = shouldBeNullable

		return floatSchema
	case reflect.Float64:
		var floatSchema *FloatSchema = &(FloatSchema{})

		floatSchema.Bits = 64
		floatSchema.IsNullable = shouldBeNullable

		return floatSchema
	}

	return nil

}

// DecodeSchema takes a buffer of bytes representing a binary schema, and returns a Schema (by
// calling the routine decodeSchemaInternal.) After the schema is decoded, byteIndex is
// reset, so that we can process another Schema
func DecodeSchema(buf []byte) (Schema, error) {

	byteIndex = 0 // always start at the beginning of the buffer

	// and then decodeSchemaInternal() will advance byteIndex as it goes
	tmp, err := decodeSchemaInternal(buf)

	return tmp, err
}

// DecodeJSONSchema takes a buffer of JSON data and parses it to create a schema
func DecodeJSONSchema(buf []byte) (Schema, error) {
	return nil, nil
}

// decodeSchemaInternal processes buf[] to actually decode the binary schema.
// As each byte is processed, this routine advances byteIndex, which indicates
// how far into the buffer we have processed already.
// Note that any recursive calls inside of this routine should call decodeSchemaInternal()
// not DecodeSchema
func decodeSchemaInternal(buf []byte) (Schema, error) {

	var bit3IsSet bool
	var err error

	// decode fixed int schema
	// (bits 7 and 8 should be clear)
	if buf[byteIndex]&192 == 0 {
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.IsNullable = (buf[byteIndex]&1 == 1)
		fixedIntSchema.Signed = (buf[byteIndex] & 4) == 4
		fixedIntSchema.Bits = 8 << ((buf[byteIndex] & 56) >> 3)

		byteIndex++

		return fixedIntSchema, nil
	}

	// decode varint schema
	// (bits 7 should be set)
	if buf[byteIndex]&240 == 64 {
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.IsNullable = (buf[byteIndex]&1 == 1)
		varIntSchema.Signed = (buf[byteIndex] & 4) == 4

		byteIndex++

		return varIntSchema, nil
	}

	// decode floating point schema
	// (bits 5 and 7 should be set)
	if buf[byteIndex]&112 == 80 {
		var floatSchema *FloatSchema = &(FloatSchema{})

		bit3IsSet = (buf[byteIndex] & 4) == 4
		if bit3IsSet {
			floatSchema.Bits = 64
		} else {
			floatSchema.Bits = 32
		}
		floatSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return floatSchema, nil
	}

	// decode complex number
	// (bits 6 and 7 should be set)
	if buf[byteIndex]&112 == 96 {
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		bit3IsSet = (buf[byteIndex] & 4) == 4
		if bit3IsSet {
			complexSchema.Bits = 128
		} else {
			complexSchema.Bits = 64
		}
		complexSchema.SchemaOptions.Nullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return complexSchema, nil
	}

	// decode boolean
	// (bits 5,6,7 are all set)
	if buf[byteIndex]&116 == 112 {
		var boolSchema *BoolSchema = &(BoolSchema{})

		boolSchema.SchemaOptions.Nullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return boolSchema, nil
	}

	// decode enum
	// (bits 3,5,6,7 should all be set)
	if buf[byteIndex]&116 == 116 {
		var enumSchema *EnumSchema = &(EnumSchema{})

		enumSchema.SchemaOptions.Nullable = (buf[byteIndex]&1 == 1)
		byteIndex++

		// we want to read in all the enumerated values...
		mapSchema := SchemaOf(enumSchema.Values)

		r := bytes.NewReader(buf[byteIndex:])
		origBufferSize := r.Len()
		err := mapSchema.Decode(r, &enumSchema.Values)
		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		byteIndex = byteIndex + (origBufferSize - r.Len())

		return enumSchema, nil
	}

	// decode fixed len string
	// (bits 8 and 3 should be set)
	if buf[byteIndex]&252 == 132 {
		var fixedLenStringSchema *FixedStringSchema = &(FixedStringSchema{})

		fixedLenStringSchema.IsNullable = (buf[byteIndex]&1 == 1)
		byteIndex++

		r := bytes.NewReader(buf[byteIndex:])
		origBufferSize := r.Len()
		tmp, err := binary.ReadVarint(r)
		fixedLenStringSchema.FixedLength = int(tmp)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		byteIndex = byteIndex + (origBufferSize - r.Len())

		return fixedLenStringSchema, nil
	}

	// decode var len string
	// (bits 8 should be set, bit 3 should be clear)
	if buf[byteIndex]&252 == 128 {
		var varLenStringSchema *VarLenStringSchema = &(VarLenStringSchema{})

		varLenStringSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		//varLenStringSchema

		return varLenStringSchema, nil
	}

	// decode fixed array schema
	// (bits 3, 5, 8)
	if buf[byteIndex]&252 == 148 {
		var fixedLenArraySchema *FixedLenArraySchema = &(FixedLenArraySchema{})

		fixedLenArraySchema.IsNullable = (buf[byteIndex]&1 == 1)
		byteIndex++

		// read out the fixed len array length as a varint
		r := bytes.NewReader(buf[byteIndex:])
		origBufferSize := r.Len()
		tmp, err := binary.ReadVarint(r)
		fixedLenArraySchema.Length = int(tmp)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		byteIndex = byteIndex + (origBufferSize - r.Len())

		fixedLenArraySchema.Element, err = decodeSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		return fixedLenArraySchema, nil
	}

	// decode var array schema
	// (bits 3, 5, 8)
	if buf[byteIndex]&252 == 144 {
		var varArraySchema *VarArraySchema = &(VarArraySchema{})

		varArraySchema.IsNullable = (buf[byteIndex]&1 == 1)
		byteIndex++

		varArraySchema.Element, err = decodeSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		return varArraySchema, nil
	}

	// decode var object schema
	if buf[byteIndex]&252 == 160 {
		var varObjectSchema *VarObjectSchema = &(VarObjectSchema{})

		varObjectSchema.IsNullable = (buf[byteIndex]&1 == 1)
		byteIndex++

		varObjectSchema.Key, err = decodeSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		varObjectSchema.Value, err = decodeSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		return varObjectSchema, nil
	}

	// fixed object schema
	if buf[byteIndex]&252 == 164 {
		var fixedObjectSchema *FixedObjectSchema = &(FixedObjectSchema{})

		fixedObjectSchema.IsNullable = (buf[byteIndex]&1 == 1)
		byteIndex++

		// read out total number of fields (which was encoded as a varInt)
		r := bytes.NewReader(buf[byteIndex:])
		origBufferSize := r.Len()
		numFields, err := binary.ReadVarint(r)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		byteIndex = byteIndex + (origBufferSize - r.Len())

		for i := 0; i < int(numFields); i++ {
			var of ObjectField = ObjectField{}

			// read out total number of aliases for this field (which was encoded as a varInt)
			r := bytes.NewReader(buf[byteIndex:])
			origBufferSize = r.Len()
			numAlias, err := binary.ReadVarint(r)

			if err != nil {
				return nil, err
			}

			// advance ahead into the buffer as many bytes as ReadVarint consumed...
			byteIndex = byteIndex + (origBufferSize - r.Len())

			// read out each alias name...
			for j := 0; j < int(numAlias); j++ {
				AliasName := ""
				varLenStringSchema := SchemaOf(AliasName)
				r := bytes.NewReader(buf[byteIndex:])
				origBufferSize := r.Len()
				var tmp string
				varLenStringSchema.Decode(r, &tmp)
				of.FieldAliases = append(of.FieldAliases, tmp)
				byteIndex = byteIndex + (origBufferSize - r.Len())
			}

			// decodeSchemaInternal recursive call will advance byteIndex for each field...
			of.Schema, err = decodeSchemaInternal(buf)
			if err != nil {
				return nil, err
			}

			fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of)
		}

		return fixedObjectSchema, nil
	}

	//Variant

	//Schema

	//Custom Type

	return nil, fmt.Errorf("invalid binary schema encountered")
}
