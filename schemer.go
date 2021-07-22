package schemer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// TODO: Rename these; use shorter names
// TODO: Add documentation (or at least refer to encoding doc)
const (
	BoolSchemaBinaryFormat        = 0b00011100
	ComplexSchemaBinaryFormat     = 0b00011000
	EnumSchemaBinaryFormat        = 0b00011101
	FixedArraySchemaBinaryFormat  = 0b00100101
	FixedIntSchemaBinaryFormat    = 0b00000000
	FixedObjectSchemaBinaryFormat = 0b00101001
	FixedStringSchemaBinaryFormat = 0b00100001
	FloatBinarySchemaFormat       = 0b00010100
	VarArraySchemaBinaryFormat    = 0b00100100
	varIntSchemaBinaryFormat      = 0b00010000
	VarObjectSchemaBinaryFormat   = 0b00101000
	VarStringSchemaBinaryFormat   = 0b00100000
)

type Schema interface {
	// Encode uses the schema to write the encoded value of i to the output
	// stream
	Encode(w io.Writer, i interface{}) error

	// EncodeValue uses the schema to write the encoded value of v to the output
	// stream
	EncodeValue(w io.Writer, v reflect.Value) error

	// Decode uses the schema to read the next encoded value from the input
	// stream and stores it in i
	Decode(r io.Reader, i interface{}) error

	// DecodeValue uses the schema to read the next encoded value from the input
	// stream and stores it in v
	DecodeValue(r io.Reader, v reflect.Value) error

	// Nullable returns true if and only if the type is nullable
	Nullable() bool

	// MarshalJSON returns the JSON encoding of the schema
	MarshalJSON() ([]byte, error)

	// MarshalSchemer encodes the schema in a portable binary format
	MarshalSchemer() []byte

	// GoType returns the default Go type that represents the schema
	GoType() reflect.Type
}

// SchemaOf returns a Schema for the specified interface value
// If i is a pointer or interface type, the pointer/interface value is used to
// generate the Schema. If i is nil, an zero-field FixedObjectSchema is returned.
func SchemaOf(i interface{}) Schema {
	if i == nil {
		// Return a Schema for an empty struct
		return &FixedObjectSchema{}
	}

	t := reflect.TypeOf(i)

	// if t is a ptr or interface type, remove exactly one level of indirection
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return SchemaOfType(t)
}

// SchemaOfType returns a Schema for the specified Go type
func SchemaOfType(t reflect.Type) Schema {
	nullable := false

	// Dereference pointer / interface types
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()

		// If we encounter any pointers, then we know this type is nullable
		nullable = true
	}

	k := t.Kind()

	switch k {
	case reflect.Bool:
		s := &BoolSchema{}
		s.SetNullable(nullable)
		return s

	// all int types default to signed varint
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		s := &VarIntSchema{Signed: true}
		s.SetNullable(nullable)
		return s

	// all unint types default to unsigned varint
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		s := &VarIntSchema{Signed: false}
		s.SetNullable(nullable)
		return s

	case reflect.Float32:
		s := &FloatSchema{Bits: 32}
		s.SetNullable(nullable)
		return s
	case reflect.Float64:
		s := &FloatSchema{Bits: 64}
		s.SetNullable(nullable)
		return s

	case reflect.Complex64:
		s := &ComplexSchema{Bits: 64}
		s.SetNullable(nullable)
		return s

	case reflect.Complex128:
		s := &ComplexSchema{Bits: 128}
		s.SetNullable(nullable)
		return s

	case reflect.String:
		s := &VarLenStringSchema{}
		s.SetNullable(nullable)
		return s

	case reflect.Array:
		s := &FixedArraySchema{
			Length:  t.Len(),
			Element: SchemaOfType(t.Elem()),
		}
		s.SetNullable(nullable)
		return s

	case reflect.Slice:
		s := &VarArraySchema{
			Element: SchemaOfType(t.Elem()),
		}
		s.SetNullable(nullable)
		return s

	case reflect.Map:
		s := &VarObjectSchema{
			Key:   SchemaOfType(t.Key()),
			Value: SchemaOfType(t.Elem()),
		}
		s.SetNullable(nullable)
		return s

	case reflect.Struct:
		s := &FixedObjectSchema{
			Fields: make([]ObjectField, 0, t.NumField()),
		}
		s.SetNullable(nullable)

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			of := ObjectField{
				Schema: SchemaOfType(f.Type),
			}
			if of.Schema == nil {
				continue // skip this field
			}
			ofSchemaOpts := of.Schema.(interface {
				SetNullable(bool)
				SetWeakDecoding(bool)
			})

			// Parse struct tag and set aliases and schema options
			tagOpts := ParseStructTag(f.Tag.Get(StructTagName))

			// Note: only override option if explicitly set in the tag
			if tagOpts.NullableSet {
				ofSchemaOpts.SetNullable(tagOpts.Nullable)
			}
			if tagOpts.WeakDecodingSet {
				ofSchemaOpts.SetWeakDecoding(tagOpts.WeakDecoding)
			}

			if tagOpts.FieldAliasesSet {
				of.Aliases = tagOpts.FieldAliases
			} else {
				// if no aliases set in the struct tag, use the struct field name
				of.Aliases = []string{f.Name}
			}

			// check if this field is not exported
			exported := len(f.PkgPath) == 0
			if exported && len(of.Aliases) > 0 {
				s.Fields = append(s.Fields, of)
			}
		}
		return s
	}

	return nil
}

// DecodeJSONSchema takes a buffer of JSON data and parses it to create a schema
func DecodeJSONSchema(buf []byte) (Schema, error) {

	fields := make(map[string]interface{})

	err := json.Unmarshal(buf, &fields)
	if err != nil {
		return nil, err
	}

	// TODO: Check for error during type assertion to avoid panic
	schemaType := strings.ToLower(fields["type"].(string))

	switch schemaType {

	case "bool":
		s := &BoolSchema{}
		b, _ := fields["nullable"].(bool)
		s.SchemaOptions.nullable = b
		return s, nil

	case "int":
		numBits, ok := fields["bits"].(float64)

		// if bits are present, then we are dealing with a fixed int
		if ok {
			s := &FixedIntSchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b
			b, _ = fields["signed"].(bool)
			s.Signed = b
			// TODO: Validate `numBits` value (like "float" below)
			s.Bits = int(numBits)

			return s, nil
		} else {
			s := &VarIntSchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b
			b, _ = fields["signed"].(bool)
			s.Signed = b

			return s, nil
		}

	case "float":
		s := &FloatSchema{}

		if fields["bits"].(float64) == 64 {
			s.Bits = 64
		} else if fields["bits"].(float64) == 32 {
			s.Bits = 32
		} else {
			// TODO: Improve error message
			return nil, fmt.Errorf("invalid JSON schema encountered")
		}

		b, _ := fields["nullable"].(bool)
		s.SchemaOptions.nullable = b

		return s, nil

	case "complex":
		s := &ComplexSchema{}

		if fields["bits"].(float64) == 128 {
			s.Bits = 128
		} else if fields["bits"].(float64) == 64 {
			s.Bits = 64
		} else {
			// TODO: Improve error message
			return nil, fmt.Errorf("invalid JSON schema encountered")
		}
		b, _ := fields["nullable"].(bool)
		s.SchemaOptions.nullable = b

		return s, nil

	case "string":
		tmpLen, ok := fields["length"].(float64)

		// if string length is present, then we are dealing with a fixed string
		if ok {
			s := &FixedStringSchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b
			// TODO: Validate `tmpLen` (i.e. >= 0, no decimal part)
			s.Length = int(tmpLen)

			return s, nil
		} else {
			s := &VarLenStringSchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b
			return s, nil
		}

	case "enum":
		s := &EnumSchema{}
		b, _ := fields["nullable"].(bool)
		s.SchemaOptions.nullable = b

		// TODO: Parse enum values

		return s, nil

	case "array":
		l, ok := fields["length"].(float64)

		// if length is present, then we are dealing with a fixed length array
		if ok {
			s := &FixedArraySchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b
			// TODO: Validate length
			s.Length = int(l)

			// process the array element
			tmp, err := json.Marshal(fields["element"])
			if err != nil {
				return nil, err
			}

			s.Element, err = DecodeJSONSchema(tmp)
			if err != nil {
				return nil, err
			}

			return s, nil
		} else {
			s := &VarArraySchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b

			// process the array element
			tmp, err := json.Marshal(fields["element"])
			if err != nil {
				return nil, err
			}

			s.Element, err = DecodeJSONSchema(tmp)
			if err != nil {
				return nil, err
			}

			return s, nil
		}

	case "object":
		objectFields, ok := fields["fields"].([]interface{})

		// if fields are present, then we are dealing with a fixed object
		if ok {
			s := &FixedObjectSchema{
				Fields: make([]ObjectField, 0, len(objectFields)),
			}
			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b

			// loop through all fields in this object
			for i := 0; i < len(objectFields); i++ {
				var of ObjectField = ObjectField{}

				// fill in the name of this field...
				// (the json encoded data only includes the name, not a list of aliases)
				// TODO: Validate type assertion to avoid panic; throw error instead
				tmpMap := objectFields[i].(map[string]interface{})

				// TODO: name could either be `string` or `[]string`
				// TODO: I'd recommend using a type switch here
				name := tmpMap["name"].(string)

				of.Aliases = []string{name}

				tmp, err := json.Marshal(objectFields[i])
				if err != nil {
					return nil, err
				}
				// recursive call to process this field of this object...
				of.Schema, err = DecodeJSONSchema(tmp)
				if err != nil {
					return nil, err
				}

				s.Fields = append(s.Fields, of)
			}

			return s, nil
		} else {
			s := &VarObjectSchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b

			tmp, err := json.Marshal(fields["key"])
			if err != nil {
				return nil, err
			}
			s.Key, err = DecodeJSONSchema(tmp)
			if err != nil {
				return nil, err
			}

			tmp, err = json.Marshal(fields["value"])
			if err != nil {
				return nil, err
			}
			s.Value, err = DecodeJSONSchema(tmp)
			if err != nil {
				return nil, err
			}

			return s, nil

		}
	}

	return nil, fmt.Errorf("invalid JSON schema type: %s", schemaType)
}

// DecodeSchema decodes a Schema from its binary representation
func DecodeSchema(buf []byte) (Schema, error) {

	byteIndex := 0 // always start at the beginning of the buffer

	// and then decodeSchema() will advance *byteIndex as it goes
	tmp, err := decodeSchema(buf, &byteIndex)

	return tmp, err
}

// decodeSchema processes buf[] to actually decode the binary schema.
// As each byte is processed, this routine advances *byteIndex, which indicates
// how far into the buffer we have processed already.
// Note that any recursive calls inside of this routine should call decodeSchema()
// not DecodeSchema
func decodeSchema(buf []byte, byteIndex *int) (Schema, error) {
	var err error

	// decode enum
	// TODO: Use named constants / bitmasks everywhere...
	if buf[*byteIndex]&0b00111111 == 0b00011101 {
		var enumSchema *EnumSchema = &(EnumSchema{})

		enumSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		// we want to read in all the enumerated values...
		mapSchema := SchemaOf(enumSchema.Values)

		r := bytes.NewReader(buf[*byteIndex:])
		origBufferSize := r.Len()
		err := mapSchema.Decode(r, &enumSchema.Values)
		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		*byteIndex = *byteIndex + (origBufferSize - r.Len())

		return enumSchema, nil
	}

	// decode boolean
	if buf[*byteIndex]&0b00111100 == 0b00011100 {
		var boolSchema *BoolSchema = &(BoolSchema{})

		boolSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		return boolSchema, nil
	}

	// decode complex number
	if buf[*byteIndex]&0b01111000 == 0b00011000 {
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		if (buf[*byteIndex] & 1) == 1 {
			complexSchema.Bits = 128
		} else {
			complexSchema.Bits = 64
		}
		complexSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		return complexSchema, nil
	}

	// decode fixed array schema
	if buf[*byteIndex]&0b00111111 == 0b00100101 {
		var FixedArraySchema *FixedArraySchema = &(FixedArraySchema{})

		FixedArraySchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		// read out the fixed len array length as a varint
		r := bytes.NewReader(buf[*byteIndex:])
		origBufferSize := r.Len()
		tmp, err := binary.ReadVarint(r)
		FixedArraySchema.Length = int(tmp)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		*byteIndex = *byteIndex + (origBufferSize - r.Len())

		FixedArraySchema.Element, err = decodeSchema(buf, byteIndex)
		if err != nil {
			return nil, err
		}

		return FixedArraySchema, nil
	}

	// decode fixed int schema
	if buf[*byteIndex]&0b00110000 == 0b00000000 {
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		fixedIntSchema.Signed = (buf[*byteIndex] & 1) == 1
		fixedIntSchema.Bits = 8 << ((buf[*byteIndex] & 14) >> 1)

		*byteIndex++

		return fixedIntSchema, nil
	}

	// decode varint schema
	if buf[*byteIndex]&0b00111110 == 0b00010000 {
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		varIntSchema.Signed = (buf[*byteIndex] & 1) == 1

		*byteIndex++

		return varIntSchema, nil
	}

	// fixed object schema
	if buf[*byteIndex]&0b111111 == 0b00101001 {
		var fixedObjectSchema *FixedObjectSchema = &(FixedObjectSchema{})

		fixedObjectSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		// read out total number of fields (which was encoded as a varInt)
		r := bytes.NewReader(buf[*byteIndex:])
		origBufferSize := r.Len()
		numFields, err := binary.ReadVarint(r)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		*byteIndex = *byteIndex + (origBufferSize - r.Len())

		for i := 0; i < int(numFields); i++ {
			var of ObjectField = ObjectField{}

			// read out total number of aliases for this field (which was encoded as a varInt)
			r := bytes.NewReader(buf[*byteIndex:])
			origBufferSize = r.Len()
			numAlias, err := binary.ReadVarint(r)

			if err != nil {
				return nil, err
			}

			// advance ahead into the buffer as many bytes as ReadVarint consumed...
			*byteIndex = *byteIndex + (origBufferSize - r.Len())

			// read out each alias name...
			for j := 0; j < int(numAlias); j++ {
				AliasName := ""
				varLenStringSchema := SchemaOf(AliasName)
				r := bytes.NewReader(buf[*byteIndex:])
				origBufferSize := r.Len()
				var tmp string
				varLenStringSchema.Decode(r, &tmp)
				of.Aliases = append(of.Aliases, tmp)
				*byteIndex = *byteIndex + (origBufferSize - r.Len())
			}

			// decodeSchema recursive call will advance *byteIndex for each field...
			of.Schema, err = decodeSchema(buf, byteIndex)
			if err != nil {
				return nil, err
			}

			fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of)
		}

		return fixedObjectSchema, nil
	}

	// decode fixed len string
	if buf[*byteIndex]&0b00111111 == 0b00100001 {
		var fixedLenStringSchema *FixedStringSchema = &(FixedStringSchema{})

		fixedLenStringSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		r := bytes.NewReader(buf[*byteIndex:])
		origBufferSize := r.Len()
		tmp, err := binary.ReadVarint(r)
		fixedLenStringSchema.Length = int(tmp)

		if err != nil {
			return nil, err
		}

		// advance ahead into the buffer as many bytes as ReadVarint consumed...
		*byteIndex = *byteIndex + (origBufferSize - r.Len())

		return fixedLenStringSchema, nil
	}

	// decode floating point schema
	if buf[*byteIndex]&0b00111100 == 0b00010100 {
		var floatSchema *FloatSchema = &(FloatSchema{})

		if buf[*byteIndex]&1 == 1 {
			floatSchema.Bits = 64
		} else {
			floatSchema.Bits = 32
		}
		floatSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		return floatSchema, nil
	}

	// decode var array schema
	if buf[*byteIndex]&0b00111110 == 0b00100100 {
		var varArraySchema *VarArraySchema = &(VarArraySchema{})

		varArraySchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		varArraySchema.Element, err = decodeSchema(buf, byteIndex)
		if err != nil {
			return nil, err
		}

		return varArraySchema, nil
	}

	// decode var object schema
	if buf[*byteIndex]&0b00111111 == 0b00101000 {
		var varObjectSchema *VarObjectSchema = &(VarObjectSchema{})

		varObjectSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		varObjectSchema.Key, err = decodeSchema(buf, byteIndex)
		if err != nil {
			return nil, err
		}

		varObjectSchema.Value, err = decodeSchema(buf, byteIndex)
		if err != nil {
			return nil, err
		}

		return varObjectSchema, nil
	}

	// decode var len string
	if buf[*byteIndex]&0b00111111 == 0b00100000 {
		var varLenStringSchema *VarLenStringSchema = &(VarLenStringSchema{})

		varLenStringSchema.SchemaOptions.nullable = (buf[*byteIndex]&128 == 128)
		*byteIndex++

		return varLenStringSchema, nil
	}

	//Variant

	//Schema

	//Custom Type

	return nil, fmt.Errorf("invalid binary schema encountered")
}

// PreEncode should be called by each Schema's Encode() routine.
// It handles dereferencing points and interacts, and for writing
// a byte to indicate nullable if the schema indicates it is in fact
// nullable.
// If this routine returns false, no more processing is needed on the
// encoder who called this routine.
func PreEncode(s Schema, w io.Writer, v *reflect.Value) (bool, error) {
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		*v = v.Elem()
	}

	if s.Nullable() {
		// did the caller pass in a nil value, or a null pointer?
		if !v.IsValid() {
			// per the revised spec, 1 indicates null
			w.Write([]byte{1})
			return false, nil
		} else {
			// 0 indicates not null
			w.Write([]byte{0})
		}
	} else {
		// if nullable is false
		// but they are trying to encode a nil value.. then that is an error
		if !v.IsValid() {
			return false, fmt.Errorf("cannot enoded nil value when IsNullable is false")
		}
	}

	return true, nil
}

// PreDecode() is called before each Decode() routine from all the schemas. This routine
// handles checking on the nullable flag if the schema indicates the schema
// is nullable.
// This routine also handles derefering pointers and interfaces, and returns
// the new value of v after it is set.
func PreDecode(s Schema, r io.Reader, v reflect.Value) (reflect.Value, error) {
	// if t is a ptr or interface type, remove exactly ONE level of indirection
	if k := v.Kind(); !v.CanSet() && (k == reflect.Ptr || k == reflect.Interface) {
		v = v.Elem()
	}

	// if the data indicates this type is nullable, then the actual
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.Nullable() {
		buf := make([]byte, 1)

		// first byte indicates whether value is null or not...
		_, err := io.ReadAtLeast(r, buf, 1)
		if err != nil {
			return reflect.Value{}, err
		}
		valueIsNull := (buf[0] == 1)

		if valueIsNull {
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
				if v.CanSet() {
					// special way to set pointer to nil value
					v.Set(reflect.Zero(v.Type()))
					return reflect.Value{}, nil
				}
				return reflect.Value{}, fmt.Errorf("destination not settable")
			} else {
				return reflect.Value{}, fmt.Errorf("cannot decode null value to non pointer to pointer type")
			}
		}
	}

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		if v.IsNil() {
			if k == reflect.Interface {
				break
			}
			if !v.CanSet() {
				return reflect.Value{}, fmt.Errorf("decode destination is not settable")
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	return v, nil
}
