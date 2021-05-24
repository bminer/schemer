package schemer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// https://golangbyexample.com/go-size-range-int-uint/
const uintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

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

	Nullable() bool
	// SetNullable sets the nullable flag for the schema
	//SetNullable(n bool)

	DecodeValue(r io.Reader, v reflect.Value) error

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
	// NOTE: i don't understand the above line, which blake said to do???
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

func setSchemaOptions(schema *Schema, structFieldOptions StructFieldOptions) {

	fixedStringSchema, ok := (*schema).(*FixedStringSchema)
	if ok {
		fixedStringSchema.IsNullable = structFieldOptions.Nullable
		fixedStringSchema.WeakDecoding = structFieldOptions.WeakDecoding
		return
	}

	/*
		varIntSchema, ok := (*schema).(VarIntSchema)
		if ok {
			varIntSchema.SchemaOptions = schemaOptions
		}
	*/

	// update to work with other types

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
			setSchemaOptions(&of.Schema, structFieldOptions)

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
		/*
			var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

			fixedIntSchema.Signed = true
			fixedIntSchema.Bits = uintSize
			fixedIntSchema.IsNullable = shouldBeNullable

			return fixedIntSchema
		*/

		var VarIntSchema *VarIntSchema = &(VarIntSchema{})

		VarIntSchema.Signed = true
		VarIntSchema.IsNullable = shouldBeNullable

		return VarIntSchema

	case reflect.Int8:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 8
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Int16:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 16
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Int32:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 32
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Int64:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 64
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Uint:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = uintSize
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Uint8:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 8
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Uint16:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 16
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Uint32:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 32
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Uint64:
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 32
		fixedIntSchema.IsNullable = shouldBeNullable

		return fixedIntSchema
	case reflect.Complex64:
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		complexSchema.Bits = 64
		complexSchema.IsNullable = shouldBeNullable

		return complexSchema
	case reflect.Complex128:
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		complexSchema.IsNullable = shouldBeNullable
		complexSchema.Bits = 128

		return complexSchema
	case reflect.Bool:
		var boolSchema *BoolSchema = &(BoolSchema{})

		boolSchema.IsNullable = shouldBeNullable

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

func NewSchema(buf []byte) (Schema, error) {

	tmp, err := newSchemaInternal(buf)
	byteIndex = 0

	return tmp, err
}

// NewSchema decodes a schema stored in buf and returns an error if the schema is invalid
func newSchemaInternal(buf []byte) (Schema, error) {

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
		complexSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return complexSchema, nil
	}

	// decode boolean
	// (bits 5,6,7 are all set)
	if buf[byteIndex]&116 == 112 {
		var boolSchema *BoolSchema = &(BoolSchema{})

		boolSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return boolSchema, nil
	}

	// decode enum
	// (bits 3,5,6,7 should all be set)
	if buf[byteIndex]&116 == 116 {
		var enumSchema *EnumSchema = &(EnumSchema{})

		enumSchema.IsNullable = (buf[byteIndex]&1 == 1)
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

		fixedLenArraySchema.Element, err = newSchemaInternal(buf)
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

		varArraySchema.Element, err = newSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		return varArraySchema, nil
	}

	// decode var object schema
	if buf[byteIndex]&252 == 160 {
		var varObjectSchema *VarObjectSchema = &(VarObjectSchema{})

		varObjectSchema.IsNullable = (buf[0]&1 == 1)
		byteIndex++

		varObjectSchema.Key, err = newSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		varObjectSchema.Value, err = newSchemaInternal(buf)
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

		var numFields int = int(buf[byteIndex])
		byteIndex++

		for i := 0; i < numFields; i++ {
			var of ObjectField = ObjectField{}
			var numAlias int = int(buf[byteIndex])
			byteIndex++

			// read out each alias name...
			for j := 0; j < numAlias; j++ {
				AliasName := ""
				varLenStringSchema := SchemaOf(AliasName)
				r := bytes.NewReader(buf[byteIndex:])
				origBufferSize := r.Len()
				var tmp string
				varLenStringSchema.Decode(r, &tmp)
				of.FieldAliases = append(of.FieldAliases, tmp)
				byteIndex = byteIndex + (origBufferSize - r.Len())
			}

			// newSchemaInternal recursive call will advance byteIndex for each field...
			of.Schema, err = newSchemaInternal(buf)
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
