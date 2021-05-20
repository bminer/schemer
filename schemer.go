package schemer

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// https://golangbyexample.com/go-size-range-int-uint/
const uintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

var SchemaTagName string = "schemer"

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

	// returns a human readable string specifying what type of schema this
	//NameOfSchemaType() string
}

var byteIndex int = 0

func SchemaOf(i interface{}) Schema {
	// spec says: "SchemaOf(nil) returns a Schema for an empty struct."
	if i == nil {
		var fixedObjectSchema FixedObjectSchema

		fixedObjectSchema.IsNullable = true

		return fixedObjectSchema

	}

	v := reflect.ValueOf(i)

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {

		// Set IsNullable flag on whatever Schema we return...

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
		// t = t.Elem() -- do this instead
	}

	// t := reflect.TypeOf(i)
	// if t is a ptr or interface type, remove exactly ONE level of indirection
	// return SchemaOfType(t)
	return SchemaOfValue(v)
}

// ParseTag tags a tagname as a string, parses it, and populates FieldOptions
// the format of the tag must be:
// tag := (alias)?("," option)*
// alias := identifier
//			"["identifier(","identifier)*"]"
//	option := "weak", "null", "not null"
func ParseTag(tagStr string, structFieldOptions *StructFieldOptions) error {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error parsing struct tag:", err)
		}
	}()

	fmt.Println(tagStr)

	tagStr = strings.Trim(tagStr, " ")
	if len([]rune(tagStr)) == 0 {
		print("empty tag; nothing to do!")
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
	structFieldOptions.FieldAliases = strings.Split(aliasStr, ",")

	// parse options, and string and put each option into correct field, such as .Nullable
	structFieldOptions.Nullable = strings.Contains(strings.ToUpper(optionStr), "NUL")
	structFieldOptions.NotNull = strings.Contains(strings.ToUpper(optionStr), "!NUL")
	structFieldOptions.WeakDecoding = strings.Contains(strings.ToUpper(optionStr), "WEAK")

	return nil

}

func setSchemaOptions(schema *Schema, structFieldOptions StructFieldOptions) {

	fixedStringSchema, ok := (*schema).(FixedStringSchema)
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

// SchemaOf generates a Schema from the concrete value stored in the interface i.
// SchemaOf(nil) returns a Schema for an empty struct.
func SchemaOfValue(v reflect.Value) Schema {

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {

		// Set IsNullable flag on whatever Schema we return...

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
		// t = t.Elem() -- do this instead
	}

	t := v.Type()
	k := t.Kind()

	switch k {
	case reflect.Map:
		var varObjectSchema VarObjectSchema

		varObjectSchema.IsNullable = false

		for _, mapKey := range v.MapKeys() {
			// t.Elem() and t.Key() instead of v.MapIndex()
			mapValue := v.MapIndex(mapKey)

			varObjectSchema.Key = SchemaOfValue(mapKey)
			varObjectSchema.Value = SchemaOfValue(mapValue)

		}

		return varObjectSchema
	case reflect.Struct:
		var fixedObjectSchema FixedObjectSchema
		fixedObjectSchema.IsNullable = false

		var of ObjectField

		for i := 0; i < t.NumField(); i++ {
			f := v.Field(i)

			of.Name = t.Field(i).Name
			of.Schema = SchemaOfValue(f)

			structFieldOptions := StructFieldOptions{}

			err := ParseTag(t.Field(i).Tag.Get(SchemaTagName), &structFieldOptions)
			if err != nil {
			}

			of.StructFieldOptions = structFieldOptions

			setSchemaOptions(&of.Schema, structFieldOptions)

			fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of)
		}

		return fixedObjectSchema
	case reflect.Slice:
		var varArraySchema VarArraySchema
		varArraySchema.IsNullable = false
		varArraySchema.Element = SchemaOfValue(v.Index(0))
		return varArraySchema
	case reflect.Array:
		var fixedLenArraySchema FixedLenArraySchema

		fixedLenArraySchema.IsNullable = true
		fixedLenArraySchema.Length = v.Len()

		fixedLenArraySchema.Element = SchemaOfValue(v.Index(0))
		return fixedLenArraySchema
	case reflect.String:
		var varStringSchema VarLenStringSchema

		varStringSchema.IsNullable = true

		return varStringSchema

	case reflect.Int:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = uintSize
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Int8:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 8
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Int16:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 16
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Int32:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 32
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Int64:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = true
		fixedIntSchema.Bits = 64
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Uint:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = uintSize
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Uint8:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 8
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Uint16:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 16
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Uint32:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 32
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Uint64:
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.Signed = false
		fixedIntSchema.Bits = 32
		fixedIntSchema.IsNullable = false

		return fixedIntSchema
	case reflect.Complex64:
		var complexSchema ComplexSchema

		complexSchema.Bits = 64

		return complexSchema

	case reflect.Complex128:
		var complexSchema ComplexSchema

		complexSchema.Bits = 128

		return complexSchema

	case reflect.Bool:
		var boolSchema BoolSchema

		boolSchema.IsNullable = false

		return boolSchema
	case reflect.Float32:
		var floatSchema FloatSchema

		floatSchema.Bits = 32
		floatSchema.IsNullable = false

		return floatSchema
	case reflect.Float64:
		var floatSchema FloatSchema

		floatSchema.Bits = 64
		floatSchema.IsNullable = false

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
		var fixedIntSchema FixedIntSchema

		fixedIntSchema.IsNullable = (buf[byteIndex]&1 == 1)
		fixedIntSchema.Signed = (buf[byteIndex] & 4) == 4
		fixedIntSchema.Bits = 8 << ((buf[byteIndex] & 56) >> 3)

		byteIndex++

		return fixedIntSchema, nil
	}

	// decode varint schema
	// (bits 7 should be set)
	if buf[0]&112 == 64 {
		var varIntSchema VarIntSchema

		varIntSchema.IsNullable = (buf[byteIndex]&1 == 1)
		varIntSchema.Signed = (buf[byteIndex] & 4) == 4

		byteIndex++

		return varIntSchema, nil
	}

	// decode floating point schema
	// (bits 5 and 7 should be set)
	if buf[byteIndex]&112 == 80 {
		var floatSchema FloatSchema

		floatSchema.Description = "FloatSchema"

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
		var complexSchema ComplexSchema

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
		var boolSchema BoolSchema

		boolSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return boolSchema, nil
	}

	// decode enum
	// (bits 3,5,6,7 should all be set)
	if buf[byteIndex]&116 == 116 {
		var enumSchema EnumSchema

		enumSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return enumSchema, nil
	}

	// decode fixed len string
	// (bits 8 and 3 should be set)
	if buf[byteIndex]&252 == 132 {
		var fixedLenStringSchema FixedStringSchema

		fixedLenStringSchema.IsNullable = (buf[byteIndex]&1 == 1)
		//fixedLenStringSchema.FixedLength = ???
		//the binary schema does not encode the length

		byteIndex++

		return fixedLenStringSchema, nil
	}

	// decode var len string
	// (bits 8 should be set, bit 3 should be clear)
	if buf[byteIndex]&252 == 128 {
		var varLenStringSchema VarLenStringSchema

		varLenStringSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return varLenStringSchema, nil
	}

	// decode fixed array schema
	// (bits 3, 5, 8)
	if buf[byteIndex]&252 == 148 {
		var fixedLenArraySchema FixedLenArraySchema

		fixedLenArraySchema.IsNullable = (buf[byteIndex]&1 == 1)
		//fixedLenArraySchema.Length = buf[1]

		byteIndex++

		/*
			fixedLenArraySchema.Element, err = NewSchema(buf[1:])
			if err != nil {
				return nil, err
			}
		*/

		return fixedLenArraySchema, nil
	}

	// decode var array schema
	// (bits 3, 5, 8)
	if buf[byteIndex]&252 == 144 {
		var varArraySchema VarArraySchema

		varArraySchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		return varArraySchema, nil
	}

	// decode var object schema
	if buf[byteIndex]&252 == 160 {
		var varObjectSchema VarObjectSchema

		varObjectSchema.IsNullable = (buf[0]&1 == 1)

		byteIndex++

		return varObjectSchema, nil
	}

	// fixed object schema
	if buf[byteIndex]&252 == 164 {
		var fixedObjectSchema FixedObjectSchema

		fixedObjectSchema.IsNullable = (buf[byteIndex]&1 == 1)

		byteIndex++

		var numFields int = int(buf[byteIndex])

		byteIndex++

		var of ObjectField

		for i := 0; i < numFields; i++ {

			of.Name = "" // schema does not contain name of fields...
			of.Schema, err = newSchemaInternal(buf)

			// newschema call above will advance byteIndex for each field...

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
