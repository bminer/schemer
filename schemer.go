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

// SchemerTagOptions represents information that can be read from struct field tags
type SchemerTagOptions struct {
	FieldAliases []string

	Nullable     bool
	WeakDecoding bool

	ShouldSkip bool
}

// SchemaOptions are properties that are common to every schema
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
		fixedObjectSchema.SchemaOptions.Nullable = true
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
func parseStructTag(tagStr string, schemerTagOptions *SchemerTagOptions) error {

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
		schemerTagOptions.ShouldSkip = true
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
	schemerTagOptions.FieldAliases = strings.Split(y, ",")
	for i, f := range schemerTagOptions.FieldAliases {
		schemerTagOptions.FieldAliases[i] = strings.Trim(f, " ")
	}

	// parse options, and string and put each option into correct field, such as .Nullable
	schemerTagOptions.Nullable = strings.Contains(strings.ToUpper(optionStr), "NUL") && !strings.Contains(strings.ToUpper(optionStr), "!NUL")
	schemerTagOptions.WeakDecoding = strings.Contains(strings.ToUpper(optionStr), "WEAK")

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

		varObjectSchema.SchemaOptions.Nullable = shouldBeNullable
		varObjectSchema.Key = SchemaOfType(t.Key())
		varObjectSchema.Value = SchemaOfType(t.Elem())

		return varObjectSchema
	case reflect.Struct:
		var fixedObjectSchema *FixedObjectSchema = &(FixedObjectSchema{})
		fixedObjectSchema.SchemaOptions.Nullable = shouldBeNullable

		var of ObjectField

		for i := 0; i < t.NumField(); i++ {
			of.Schema = SchemaOfType(t.Field(i).Type)

			schemerTagOptions := SchemerTagOptions{}

			parseStructTag(t.Field(i).Tag.Get(SchemaTagName), &schemerTagOptions)
			// ignore result here... if an error parsing the tags occured, just don't worry about it
			// (in case they are in the wrong format or something)

			// only do an over-ride if the tag option was specified...
			if schemerTagOptions.Nullable {
				fixedObjectSchema.SchemaOptions.Nullable = schemerTagOptions.Nullable
			}

			// only do an over-ride if the tag option was specified...
			if schemerTagOptions.WeakDecoding {
				fixedObjectSchema.SchemaOptions.WeakDecoding = schemerTagOptions.WeakDecoding
			}

			// if no tags aliases exist, then use the actual field name from the struct
			if len(schemerTagOptions.FieldAliases) == 0 {
				of.Aliases = make([]string, 1)
				of.Aliases[0] = t.Field(i).Name
			} else {
				of.Aliases = make([]string, len(schemerTagOptions.FieldAliases))
				// copy the aliases from the tags into the
				for i := 0; i < len(schemerTagOptions.FieldAliases); i++ {
					of.Aliases[i] = schemerTagOptions.FieldAliases[i]
				}
			}

			// check if this field is not exported (by looking at PkgPath)
			// or if the schemer tags on the field say that we should skip it...
			if len(t.Field(i).PkgPath) == 0 && !schemerTagOptions.ShouldSkip {
				fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of)
			} else {
				fmt.Println("skipped field:", t.Field(i).Name)
			}
		}
		return fixedObjectSchema
	case reflect.Slice:
		var varArraySchema *VarArraySchema = &(VarArraySchema{})
		varArraySchema.SchemaOptions.Nullable = shouldBeNullable
		varArraySchema.Element = SchemaOfType(t.Elem())
		return varArraySchema
	case reflect.Array:
		var FixedArraySchema *FixedArraySchema = &(FixedArraySchema{})

		FixedArraySchema.SchemaOptions.Nullable = shouldBeNullable
		FixedArraySchema.Length = t.Len()

		FixedArraySchema.Element = SchemaOfType(t.Elem())
		return FixedArraySchema
	case reflect.String:
		var varStringSchema *VarLenStringSchema = &(VarLenStringSchema{})

		varStringSchema.SchemaOptions.Nullable = shouldBeNullable

		return varStringSchema
	case reflect.Int:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema

	case reflect.Int8:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Int16:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Int32:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Int64:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = true
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint8:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint16:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint32:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

		return varIntSchema
	case reflect.Uint64:
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.Signed = false
		varIntSchema.SchemaOptions.Nullable = shouldBeNullable

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
		floatSchema.SchemaOptions.Nullable = shouldBeNullable

		return floatSchema
	case reflect.Float64:
		var floatSchema *FloatSchema = &(FloatSchema{})

		floatSchema.Bits = 64
		floatSchema.SchemaOptions.Nullable = shouldBeNullable

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

	var err error

	// decode enum
	if buf[byteIndex]&0b00011101 == 0b00011101 {
		var enumSchema *EnumSchema = &(EnumSchema{})

		enumSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
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

	// decode boolean
	if buf[byteIndex]&0b00011100 == 0b00011100 {
		var boolSchema *BoolSchema = &(BoolSchema{})

		boolSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		byteIndex++

		return boolSchema, nil
	}

	// decode complex number
	if buf[byteIndex]&0b00011000 == 0b00011000 {
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		if (buf[byteIndex] & 1) == 1 {
			complexSchema.Bits = 128
		} else {
			complexSchema.Bits = 64
		}
		complexSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		byteIndex++

		return complexSchema, nil
	}

	// decode fixed array schema
	if buf[byteIndex]&0b00100101 == 0b00100101 {
		var FixedArraySchema *FixedArraySchema = &(FixedArraySchema{})

		FixedArraySchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		byteIndex++

		// read out the fixed len array length as a varint
		r := bytes.NewReader(buf[byteIndex:])
		origBufferSize := r.Len()
		tmp, err := binary.ReadVarint(r)
		FixedArraySchema.Length = int(tmp)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		byteIndex = byteIndex + (origBufferSize - r.Len())

		FixedArraySchema.Element, err = decodeSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		return FixedArraySchema, nil
	}

	// decode fixed int schema
	if buf[byteIndex]&0b00110000 == 0 {
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		fixedIntSchema.Signed = (buf[byteIndex] & 1) == 1
		fixedIntSchema.Bits = 8 << ((buf[byteIndex] & 14) >> 1)

		byteIndex++

		return fixedIntSchema, nil
	}

	// decode varint schema
	if buf[byteIndex]&0b00111110 == 0b00010000 {
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		varIntSchema.Signed = (buf[byteIndex] & 1) == 1

		byteIndex++

		return varIntSchema, nil
	}

	// fixed object schema
	if buf[byteIndex]&0b00111111 == 0b00101001 {
		var fixedObjectSchema *FixedObjectSchema = &(FixedObjectSchema{})

		fixedObjectSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
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
				of.Aliases = append(of.Aliases, tmp)
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

	// decode fixed len string
	if buf[byteIndex]&0b00111111 == 0b00100001 {
		var fixedLenStringSchema *FixedStringSchema = &(FixedStringSchema{})

		fixedLenStringSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		byteIndex++

		r := bytes.NewReader(buf[byteIndex:])
		origBufferSize := r.Len()
		tmp, err := binary.ReadVarint(r)
		fixedLenStringSchema.Length = int(tmp)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		byteIndex = byteIndex + (origBufferSize - r.Len())

		return fixedLenStringSchema, nil
	}

	// decode floating point schema
	if buf[byteIndex]&0b00111110 == 0b00010100 {
		var floatSchema *FloatSchema = &(FloatSchema{})

		if buf[byteIndex]&1 == 1 {
			floatSchema.Bits = 64
		} else {
			floatSchema.Bits = 32
		}
		floatSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		byteIndex++

		return floatSchema, nil
	}

	// decode var array schema
	if buf[byteIndex]&0b00111110 == 0b00100100 {
		var varArraySchema *VarArraySchema = &(VarArraySchema{})

		varArraySchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		byteIndex++

		varArraySchema.Element, err = decodeSchemaInternal(buf)
		if err != nil {
			return nil, err
		}

		return varArraySchema, nil
	}

	// decode var object schema
	if buf[byteIndex]&0b00111110 == 0b00101000 {
		var varObjectSchema *VarObjectSchema = &(VarObjectSchema{})

		varObjectSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
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

	// decode var len string
	if buf[byteIndex]&0b00111110 == 0b00100000 {
		var varLenStringSchema *VarLenStringSchema = &(VarLenStringSchema{})

		varLenStringSchema.SchemaOptions.Nullable = (buf[byteIndex]&128 == 128)
		byteIndex++

		return varLenStringSchema, nil
	}

	//Variant

	//Schema

	//Custom Type

	return nil, fmt.Errorf("invalid binary schema encountered")
}
