package schemer

// TODO:
//		-> type TypeByte byte
//	-	-> model decode after decodeJSON
//	-	-> squeze custom type into 6 bits , based on reserved type (CustomSchemaMask to be 0x40)
//

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// The most signifiant bit indicates whether or not the type is nullable
// The 6 least significant bits identify the type, per below
const (
	FixedIntSchemaMask      = 0    // 0b00 nnns 	where s is the signed/unsigned bit and n represents the encoded integer size in (8 << n) bits.
	VarIntSchemaMask        = 0x10 // 0b01 000s 	where s is the signed/unsigned bit
	FloatBinarySchemaFormat = 0x14 // 0b01 01*n 	where n is the floating-point size in (32 << n) bits and * is reserved for future use
	ComplexSchemaMask       = 0x18 // 0b01 10*n 	where n is the complex number size in (64 << n) bits and * is reserved for future use
	BoolSchemaMask          = 0x1C // 0b01 1100
	EnumSchemaMask          = 0x1D // 0b01 1101
	FixedStringSchemaMask   = 0x21 // 0b10 000f 	where f indicates that the string is of fixed byte length
	VarStringSchemaMask     = 0x20
	FixedArraySchemaMask    = 0x25 // 0b10 010f 	where f indicates that the array is of fixed length
	VarArraySchemaMask      = 0x24
	FixedObjectSchemaMask   = 0x29 // 0b10 100f 	where f indicates that the object has fixed number of fields
	VarObjectSchemaMask     = 0x28
	CustomSchemaMask        = 0x40 // bit 7 indicates custom schema type
	SchemaNullBit           = 0x80 // The most signifiant bit indicates whether or not the type is nullable
)

// Schema is an interface that encodes and decodes data of a specific type
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

	// GoType returns the default Go type that represents the schema
	GoType() reflect.Type
}

// Marshaler is an interface implemented by a schema, allowing it to encode
// itself into a portable binary format
type Marshaler interface {
	MarshalSchemer() ([]byte, error)
}

// SchemaGenerator is an interface implemented by custom schema generators.
// When Register is called on a SchemaGenerator, the global SchemaOf,
// decodeSchema, and DecodeSchemaJSON functions will call the identically
// named method on each schema generator to determine if a custom schema should
// be returned.
// If a SchemaGenerator cannot return a Schema for a specific type, it should
// return nil, nil.
// If all schema generators return a nil Schema or if Register is never called,
// then the built-in logic for returning a Schema is used.
type SchemaGenerator interface {
	SchemaOfType(t reflect.Type) (Schema, error)
	DecodeSchema(r io.Reader) (Schema, error)
	DecodeSchemaJSON(r io.Reader) (Schema, error)
}

type hasSchemaOfType interface {
	SchemaOfType(t reflect.Type) (Schema, error)
}
type hasDecodeSchema interface {
	DecodeSchema(r io.Reader) (Schema, error)
}
type hasDecodeSchemaJSON interface {
	DecodeSchemaJSON(r io.Reader) (Schema, error)
}

var (
	regSchemaOfType     = []hasSchemaOfType{}
	regDecodeSchema     = []hasDecodeSchema{}
	regDecodeSchemaJSON = []hasDecodeSchemaJSON{}
)

// Register records custom schema generators that implement `SchemaOfType`,
// `DecodeSchema`, and/or `DecodeSchemaJSON`. When `schemer.SchemaOfType` is
// called, `SchemaOfType` is called on each registered schema generator to
// determine if a custom Schema should be used for a given type.
func Register(ifaces ...interface{}) error {
	for _, iface := range ifaces {
		if sg, ok := iface.(hasSchemaOfType); ok {
			regSchemaOfType = append(regSchemaOfType, sg)
		}
		if sg, ok := iface.(hasDecodeSchema); ok {
			regDecodeSchema = append(regDecodeSchema, sg)
		}
		if sg, ok := iface.(hasDecodeSchemaJSON); ok {
			regDecodeSchemaJSON = append(regDecodeSchemaJSON, sg)
		}
	}
	return nil
}

// SchemaOf returns a Schema for the specified interface value.
// If i is a pointer or interface type, the value of the pointer/interface is
// used to generate the Schema.
// If i is nil, an zero-field FixedObjectSchema is returned.
func SchemaOf(i interface{}) (Schema, error) {
	if i == nil {
		// Return a Schema for an empty struct
		return &FixedObjectSchema{}, nil
	}

	t := reflect.TypeOf(i)

	// if t is a ptr or interface type, remove exactly one level of indirection
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return SchemaOfType(t)
}

// SchemaOfType returns a Schema for the specified Go type
func SchemaOfType(t reflect.Type) (Schema, error) {
	// Call registered schema generators
	for _, sg := range regSchemaOfType {
		if s, err := sg.SchemaOfType(t); s != nil || err != nil {
			return s, err
		}
	}

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
		return s, nil

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
		return s, nil

	// all uint types default to unsigned varint
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
		return s, nil

	case reflect.Float32:
		s := &FloatSchema{Bits: 32}
		s.SetNullable(nullable)
		return s, nil
	case reflect.Float64:
		s := &FloatSchema{Bits: 64}
		s.SetNullable(nullable)
		return s, nil

	case reflect.Complex64:
		s := &ComplexSchema{Bits: 64}
		s.SetNullable(nullable)
		return s, nil

	case reflect.Complex128:
		s := &ComplexSchema{Bits: 128}
		s.SetNullable(nullable)
		return s, nil

	case reflect.String:
		s := &VarStringSchema{}
		s.SetNullable(nullable)
		return s, nil

	case reflect.Array:
		el, err := SchemaOfType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("array type: %w", err)
		}
		s := &FixedArraySchema{
			Length:  t.Len(),
			Element: el,
		}
		s.SetNullable(nullable)
		return s, nil

	case reflect.Slice:
		el, err := SchemaOfType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("slice type: %w", err)
		}
		s := &VarArraySchema{
			Element: el,
		}
		s.SetNullable(nullable)
		return s, nil

	case reflect.Map:
		key, err := SchemaOfType(t.Key())
		if err != nil {
			return nil, fmt.Errorf("map key type: %w", err)
		}
		val, err := SchemaOfType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("map value type: %w", err)
		}
		s := &VarObjectSchema{
			Key:   key,
			Value: val,
		}
		s.SetNullable(nullable)
		return s, nil

	case reflect.Struct:
		s := &FixedObjectSchema{
			Fields: make([]ObjectField, 0, t.NumField()),
		}
		s.SetNullable(nullable)

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			ofs, err := SchemaOfType(f.Type)
			if err != nil {
				return nil, fmt.Errorf("struct field %v: %w", f.Name, err)
			}
			of := ObjectField{
				Schema: ofs,
			}

			exported := len(f.PkgPath) == 0
			if of.Schema == nil || !exported {
				continue // skip this field
			}

			// Parse struct tag and set aliases and schema options
			tagOpts := ParseStructTag(f.Tag.Get(StructTagName))

			if tagOpts.FieldAliasesSet {
				of.Aliases = tagOpts.FieldAliases
			} else {
				// if no aliases set in the tag, use the struct field name
				of.Aliases = []string{f.Name}
			}

			if len(of.Aliases) == 0 {
				continue // skip this field
			}

			// Note: only override option if explicitly set in the tag
			if tagOpts.NullableSet {
				if opt, ok := of.Schema.(interface {
					SetNullable(bool)
				}); ok {
					opt.SetNullable(tagOpts.Nullable)
				}
			}
			if tagOpts.WeakDecodingSet {
				if opt, ok := of.Schema.(interface {
					SetWeakDecoding(bool)
				}); ok {
					opt.SetWeakDecoding(tagOpts.WeakDecoding)
				}
			}

			// Add to FixedObjectSchema field list
			s.Fields = append(s.Fields, of)
		}
		return s, nil
	}

	return nil, fmt.Errorf("unsupported type %v", k)
}

// DecodeSchemaJSON takes a buffer of JSON data and parses it to create a schema
func DecodeSchemaJSON(r io.Reader) (Schema, error) {

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Call registered schema generators
	tmpBuf := bytes.NewBuffer(buf)
	for _, sg := range regDecodeSchemaJSON {
		if s, err := sg.DecodeSchemaJSON(bytes.NewReader(tmpBuf.Bytes())); s != nil || err != nil {
			return s, err
		}
	}

	fields := make(map[string]interface{})

	err = json.Unmarshal(buf, &fields)
	if err != nil {
		return nil, err
	}

	// Parse `type`
	tmp, ok := fields["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid schema type")
	}
	typeStr := strings.ToLower(tmp)

	// Parse `nullable`
	nullable := false
	tmp1, found := fields["nullable"]
	if found {
		if b, ok := tmp1.(bool); ok {
			nullable = b
		} else {
			return nil, fmt.Errorf("nullable must be a boolean")
		}
	}

	switch typeStr {
	case "bool":
		s := &BoolSchema{}
		s.SetNullable(nullable)
		return s, nil

	case "int":
		bitsStr, ok := fields["bits"]
		// if bits is present, then we are dealing with a fixed int
		if ok {
			bits, ok := bitsStr.(float64)
			if !ok {
				return nil, fmt.Errorf("bits must be a number")
			}
			s := &FixedIntSchema{}
			s.SetNullable(nullable)
			if str, ok := fields["signed"]; ok {
				b, ok := str.(bool)
				if !ok {
					return nil, fmt.Errorf("signed must be a boolean")
				}
				s.Signed = b
			}

			switch bits {
			case 8:
				fallthrough
			case 16:
				fallthrough
			case 32:
				fallthrough
			case 64:
				s.Bits = int(bits)
			default:
				return nil, fmt.Errorf("invalid bit size: %d", int(bits))
			}

			return s, nil
		} else {
			s := &VarIntSchema{}
			s.SetNullable(nullable)
			if str, ok := fields["signed"]; ok {
				b, ok := str.(bool)
				if !ok {
					return nil, fmt.Errorf("signed must be a boolean")
				}
				s.Signed = b
			}

			return s, nil
		}

	case "float":
		s := &FloatSchema{}
		bits, ok := fields["bits"].(float64)
		if !ok {
			return nil, fmt.Errorf("bits not present for float type in JSON data")
		}

		if bits == 64 {
			s.Bits = 64
		} else if bits == 32 {
			s.Bits = 32
		} else {
			return nil, fmt.Errorf("invalid float bit size encountered in JSON data: %d", int(bits))
		}

		// Parse `nullable`
		nullable = false
		tmp1, found := fields["nullable"]
		if found {
			if b, ok := tmp1.(bool); ok {
				nullable = b
			} else {
				return nil, fmt.Errorf("nullable must be a boolean for float type in JSON data")
			}
		} else {
			return nil, fmt.Errorf("nullable must present for float type in JSON data")
		}

		s.SchemaOptions.nullable = nullable
		return s, nil

	case "complex":
		s := &ComplexSchema{}
		bits, ok := fields["bits"].(float64)

		if !ok {
			return nil, fmt.Errorf("bits not present for complex type in JSON data")
		}

		if bits == 128 {
			s.Bits = 128
		} else if bits == 64 {
			s.Bits = 64
		} else {
			return nil, fmt.Errorf("invalid complex bit size encountered in JSON data: %d", int(bits))
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

			// validate that `tmpLen` is greater than 0 and is an integer
			if (tmpLen < 0) || (tmpLen-float64(int(tmpLen)) != 0) {
				return nil, fmt.Errorf("invalid string length encountered in JSON data: %d", int(tmpLen))
			}

			s.Length = int(tmpLen)

			return s, nil
		} else {
			s := &VarStringSchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b
			return s, nil
		}

	case "enum":
		s := &EnumSchema{}

		// Parse `nullable`
		nullable = false
		tmp1, found := fields["nullable"]
		if found {
			if b, ok := tmp1.(bool); ok {
				nullable = b
			} else {
				return nil, fmt.Errorf("nullable must be a boolean for enum type in JSON data")
			}
		} else {
			return nil, fmt.Errorf("nullable must present for enum type in JSON data")
		}

		s.SchemaOptions.nullable = nullable
		tmp, ok := fields["values"]
		if ok {

			s.Values = make(map[int]string)
			for key, value := range tmp.(map[string]interface{}) {
				x, err := strconv.Atoi(key)
				if err == nil {
					s.Values[x] = value.(string)
				}
			}

		} else {
			return nil, fmt.Errorf("values must present for enum type in JSON data")
		}

		return s, nil

	case "array":
		tmpLen, ok := fields["length"].(float64)

		// if length is present, then we are dealing with a fixed length array
		if ok {
			s := &FixedArraySchema{}

			b, _ := fields["nullable"].(bool)
			s.SchemaOptions.nullable = b

			// validate that `tmpLen` is greater than 0 and is an integer
			if (tmpLen < 0) || (tmpLen-float64(int(tmpLen)) != 0) {
				return nil, fmt.Errorf("invalid string length encountered in JSON data: %d", int(tmpLen))
			}

			s.Length = int(tmpLen)

			// process the array element
			tmp, err := json.Marshal(fields["element"])
			if err != nil {
				return nil, err
			}

			s.Element, err = DecodeSchemaJSON(bytes.NewReader(tmp))
			if err != nil {
				return nil, err
			}

			return s, nil
		} else {
			s := &VarArraySchema{}

			// Parse `nullable`
			nullable = false
			tmp1, found := fields["nullable"]
			if found {
				if b, ok := tmp1.(bool); ok {
					nullable = b
				} else {
					return nil, fmt.Errorf("nullable must be a boolean for array type in JSON data")
				}
			} else {
				return nil, fmt.Errorf("nullable must present for array type in JSON data")
			}
			s.SchemaOptions.nullable = nullable

			// process the array element
			tmp, err := json.Marshal(fields["element"])
			if err != nil {
				return nil, err
			}

			s.Element, err = DecodeSchemaJSON(bytes.NewReader(tmp))
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
			// Parse `nullable`
			nullable = false
			tmp1, found := fields["nullable"]
			if found {
				if b, ok := tmp1.(bool); ok {
					nullable = b
				} else {
					return nil, fmt.Errorf("nullable must be a boolean for object type in JSON data")
				}
			} else {
				return nil, fmt.Errorf("nullable must present for object type in JSON data")
			}
			s.SchemaOptions.nullable = nullable

			// loop through all fields in this object
			for i := 0; i < len(objectFields); i++ {
				var of ObjectField = ObjectField{}

				// fill in the name of this field...
				// (the json encoded data only includes the name, not a list of aliases)
				tmpMap, ok := objectFields[i].(map[string]interface{})

				if !ok {
					return nil, fmt.Errorf("invalid field definition encountered in JSON data")
				}

				name, ok := tmpMap["name"]
				if !ok {
					return nil, fmt.Errorf("missing name field encountered in JSON data")
				}

				for j := 0; j < len(name.([]interface{})); j++ {
					of.Aliases = append(of.Aliases, name.([]interface{})[j].(string))
				}

				tmp, err := json.Marshal(objectFields[i])
				if err != nil {
					return nil, err
				}
				// recursive call to process this field of this object...
				of.Schema, err = DecodeSchemaJSON(bytes.NewReader(tmp))
				if err != nil {
					return nil, err
				}

				s.Fields = append(s.Fields, of)
			}

			return s, nil
		} else {
			s := &VarObjectSchema{}

			// Parse `nullable`
			nullable = false
			tmp1, found := fields["nullable"]
			if found {
				if b, ok := tmp1.(bool); ok {
					nullable = b
				} else {
					return nil, fmt.Errorf("nullable must be a boolean for object type in JSON data")
				}
			} else {
				return nil, fmt.Errorf("nullable must present for object type in JSON data")
			}
			s.SchemaOptions.nullable = nullable

			tmp, err := json.Marshal(fields["key"])
			if err != nil {
				return nil, err
			}
			s.Key, err = DecodeSchemaJSON(bytes.NewReader(tmp))
			if err != nil {
				return nil, err
			}

			tmp, err = json.Marshal(fields["value"])
			if err != nil {
				return nil, err
			}
			s.Value, err = DecodeSchemaJSON(bytes.NewReader(tmp))
			if err != nil {
				return nil, err
			}

			return s, nil

		}
	}

	return nil, fmt.Errorf("invalid schema type: %s", typeStr)
}

// decodeSchema processes buf[] to actually decode the binary schema.
// As each byte is processed, this routine advances *byteIndex, which indicates
// how far into the buffer we have processed already.
func DecodeSchema(r io.Reader) (Schema, error) {

	const (
		TwoBitSchemaMask  = 0x30
		FourBitSchemaMask = 0x3C
		FiveBitSchemaMask = 0x3E
		SixBitSchemaMask  = 0x3F
	)

	var err error

	buf := make([]byte, 1)
	_, err = r.Read(buf)
	if err != nil {
		return nil, err
	}

	// Call registered schema generators
	for _, sg := range regDecodeSchema {
		if s, err := sg.DecodeSchema(bytes.NewReader(buf)); s != nil || err != nil {
			return s, err
		}
	}

	curByte := buf[0]

	// decode fixed int schema
	if curByte&TwoBitSchemaMask == FixedIntSchemaMask {
		var fixedIntSchema *FixedIntSchema = &(FixedIntSchema{})

		fixedIntSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)
		fixedIntSchema.Signed = (curByte & 1) == 1
		fixedIntSchema.Bits = 8 << ((curByte & 14) >> 1)

		return fixedIntSchema, nil
	}

	// decode varint schema
	if curByte&FiveBitSchemaMask == VarIntSchemaMask {
		var varIntSchema *VarIntSchema = &(VarIntSchema{})

		varIntSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)
		varIntSchema.Signed = (curByte & 1) == 1

		return varIntSchema, nil
	}

	// decode floating point schema
	if curByte&FourBitSchemaMask == FloatBinarySchemaFormat {
		var floatSchema *FloatSchema = &(FloatSchema{})

		if curByte&1 == 1 {
			floatSchema.Bits = 64
		} else {
			floatSchema.Bits = 32
		}
		floatSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		return floatSchema, nil
	}

	// decode complex number
	if curByte&FourBitSchemaMask == ComplexSchemaMask {
		var complexSchema *ComplexSchema = &(ComplexSchema{})

		if (curByte & 1) == 1 {
			complexSchema.Bits = 128
		} else {
			complexSchema.Bits = 64
		}
		complexSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		return complexSchema, nil
	}

	// decode boolean
	if curByte&SixBitSchemaMask == BoolSchemaMask {
		var boolSchema *BoolSchema = &(BoolSchema{})

		boolSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		return boolSchema, nil
	}

	// decode enum
	if curByte&SixBitSchemaMask == EnumSchemaMask {
		var enumSchema *EnumSchema = &(EnumSchema{})

		enumSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		// we want to read in all the enumerated values...
		mapSchema, err := SchemaOf(enumSchema.Values)
		if err != nil {
			return nil, err
		}

		err = mapSchema.Decode(r, &enumSchema.Values)
		if err != nil {
			return nil, err
		}

		return enumSchema, nil
	}

	// decode fixed len string
	if curByte&SixBitSchemaMask == FixedStringSchemaMask {
		var fixedLenStringSchema *FixedStringSchema = &(FixedStringSchema{})

		fixedLenStringSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		tmp, err := VarIntFromIOReader(r)
		fixedLenStringSchema.Length = int(tmp)

		if err != nil {
			return nil, err
		}

		return fixedLenStringSchema, nil
	}

	// decode var len string
	if curByte&SixBitSchemaMask == VarStringSchemaMask {
		var varLenStringSchema *VarStringSchema = &(VarStringSchema{})

		varLenStringSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		return varLenStringSchema, nil
	}

	// decode fixed array schema
	if curByte&SixBitSchemaMask == FixedArraySchemaMask {
		var FixedArraySchema *FixedArraySchema = &(FixedArraySchema{})

		FixedArraySchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		tmp, err := VarIntFromIOReader(r)
		FixedArraySchema.Length = int(tmp)

		if err != nil {
			return nil, err
		}

		FixedArraySchema.Element, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		return FixedArraySchema, nil
	}

	// decode var array schema
	if curByte&FiveBitSchemaMask == VarArraySchemaMask {
		var varArraySchema *VarArraySchema = &(VarArraySchema{})

		varArraySchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		varArraySchema.Element, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		return varArraySchema, nil
	}

	// fixed object schema
	if curByte&SixBitSchemaMask == FixedObjectSchemaMask {
		var fixedObjectSchema *FixedObjectSchema = &(FixedObjectSchema{})
		fixedObjectSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		numFields, err := VarIntFromIOReader(r)
		if err != nil {
			return nil, err
		}

		for i := 0; i < int(numFields); i++ {
			var of ObjectField = ObjectField{}

			// read out total number of aliases for this field (which was encoded as a varInt)
			numAlias, err := VarIntFromIOReader(r)
			if err != nil {
				return nil, err
			}

			// read out each alias name...
			for j := 0; j < int(numAlias); j++ {
				AliasName := ""
				varStringSchema, err := SchemaOf(AliasName)
				if err != nil {
					return nil, err
				}

				var tmp string
				varStringSchema.Decode(r, &tmp)
				of.Aliases = append(of.Aliases, tmp)

			}

			of.Schema, err = DecodeSchema(r)
			if err != nil {
				return nil, err
			}

			fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of)
		}

		return fixedObjectSchema, nil
	}

	// decode var object schema
	if curByte&SixBitSchemaMask == VarObjectSchemaMask {
		var varObjectSchema *VarObjectSchema = &(VarObjectSchema{})

		varObjectSchema.SchemaOptions.nullable = (curByte&SchemaNullBit == SchemaNullBit)

		varObjectSchema.Key, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		varObjectSchema.Value, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		return varObjectSchema, nil
	}

	//Variant

	//Schema

	//Custom Type

	return nil, fmt.Errorf("invalid binary schema encountered")
}

// PreEncode should be called by each Schema's Encode() routine.
// It handles dereferencing pointerss and interfaces, and for writing
// a byte to indicate nullable if the schema indicates it is in fact
// nullable.
// If this routine returns false, no more processing is needed on the
// encoder who called this routine.
func PreEncode(nullable bool, w io.Writer, v *reflect.Value) (bool, error) {
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		*v = v.Elem()
	}

	if nullable {
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
func PreDecode(nullable bool, r io.Reader, v reflect.Value) (reflect.Value, error) {
	// if t is a ptr or interface type, remove exactly ONE level of indirection
	if k := v.Kind(); !v.CanSet() && (k == reflect.Ptr || k == reflect.Interface) {
		v = v.Elem()
	}

	// if the data indicates this type is nullable, then the actual
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if nullable {
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

// initialization function for the Schemer Library
func init() {
	Register(dateSchemaGenerator{})
	Register(ipv4SchemaGenerator{})
}
