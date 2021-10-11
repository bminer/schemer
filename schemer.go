package schemer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// initialization function for the Schemer Library
func init() {
	Register(dateSchemaGenerator{})
	Register(ipv4SchemaGenerator{})
}

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
// DecodeSchema, and DecodeSchemaJSON functions will call the identically
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
	for k := t.Kind(); k == reflect.Ptr ||
		k == reflect.Interface; k = t.Kind() {

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
				// Note: Most schemas implement SetNullable(bool), but Schema
				// does not require it; we must check here
				if opt, ok := of.Schema.(interface {
					SetNullable(bool)
				}); ok {
					opt.SetNullable(tagOpts.Nullable)
				}
			}
			if tagOpts.WeakDecodingSet {
				// Note: Most schemas implement SetWeakDecoding(bool)
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
// The input stream r is read in its entirety before the JSON is decoded.
func DecodeSchemaJSON(r io.Reader) (Schema, error) {

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Call registered schema generators
	for _, sg := range regDecodeSchemaJSON {
		s, err := sg.DecodeSchemaJSON(bytes.NewReader(buf))
		if s != nil || err != nil {
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

// DecodeSchema decodes a binary encoded schema by reading from r
// No internal buffering is used when reading from r
func DecodeSchema(r io.Reader) (Schema, error) {

	// Save whatever registered schema generators read into `buf`
	buf := &bytes.Buffer{}
	teeR := io.TeeReader(r, buf)
	r = teeR

	// Call registered schema generators
	for _, sg := range regDecodeSchema {
		if s, err := sg.DecodeSchema(r); s != nil || err != nil {
			return s, err
		}

		// Restore `r` by concatenating `buf` contents and `teeR`
		r = io.MultiReader(bytes.NewBuffer(buf.Bytes()), teeR)
	}

	buf = &bytes.Buffer{}
	_, err := io.CopyN(buf, r, 1)
	if err != nil {
		return nil, err
	}
	curByte, _ := buf.ReadByte()

	// decode fixed int schema
	if curByte&FixedIntMask == FixedIntByte {
		s := &FixedIntSchema{}
		s.SetNullable(curByte&NullMask > 0)
		s.Signed = curByte&IntSignedMask > 0
		s.Bits = 8 << ((curByte & FixedIntBitsMask) >> 1)

		return s, nil
	}

	// decode varint schema
	if curByte&VarIntMask == VarIntByte {
		s := &VarIntSchema{}
		s.SetNullable(curByte&NullMask > 0)
		s.Signed = curByte&IntSignedMask > 0

		return s, nil
	}

	// decode floating point schema
	if curByte&FloatMask == FloatByte {
		s := &FloatSchema{}

		s.SetNullable(curByte&NullMask > 0)
		if curByte&FloatBitsMask > 0 {
			s.Bits = 64
		} else {
			s.Bits = 32
		}

		return s, nil
	}

	// decode complex number
	if curByte&ComplexMask == ComplexByte {
		s := &ComplexSchema{}

		s.SetNullable(curByte&NullMask > 0)
		if curByte&ComplexBitsMask > 0 {
			s.Bits = 128
		} else {
			s.Bits = 64
		}

		return s, nil
	}

	// decode boolean
	if curByte&BoolMask == BoolByte {
		s := &BoolSchema{}
		s.SetNullable(curByte&NullMask > 0)

		return s, nil
	}

	// decode enum
	if curByte&EnumMask == EnumByte {
		s := &EnumSchema{}
		s.SetNullable(curByte&NullMask > 0)

		// Read in all the enumerated values
		mapSchema := VarObjectSchema{
			Key:   &VarIntSchema{Signed: false},
			Value: &VarStringSchema{},
		}
		err = mapSchema.Decode(r, &s.Values)
		if err != nil {
			return nil, err
		}

		return s, nil
	}

	// decode fixed len string
	if curByte&StringMask == FixedStringByte {
		s := &FixedStringSchema{}
		s.SetNullable(curByte&NullMask > 0)

		i64, err := binary.ReadVarint(byter{r})
		if err != nil {
			return nil, err
		}
		s.Length = int(i64)

		return s, nil
	}

	// decode var len string
	if curByte&StringMask == VarStringByte {
		s := &VarStringSchema{}
		s.SetNullable(curByte&NullMask > 0)

		return s, nil
	}

	// decode fixed array schema
	if curByte&ArrayMask == FixedArrayByte {
		s := &FixedArraySchema{}
		s.SetNullable(curByte&NullMask > 0)

		i64, err := binary.ReadVarint(byter{r})
		if err != nil {
			return nil, err
		}
		s.Length = int(i64)

		s.Element, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		return s, nil
	}

	// decode var array schema
	if curByte&ArrayMask == VarArrayByte {
		s := &VarArraySchema{}
		s.SetNullable(curByte&NullMask > 0)

		s.Element, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		return s, nil
	}

	// fixed object schema
	if curByte&ObjectMask == FixedObjectByte {
		s := &FixedObjectSchema{}
		s.SetNullable(curByte&NullMask > 0)

		numFields, err := binary.ReadVarint(byter{r})
		if err != nil {
			return nil, err
		}

		varStringSchema := VarStringSchema{}
		for i := 0; i < int(numFields); i++ {
			of := ObjectField{}

			// read out total number of aliases for this field (which was encoded as a varInt)
			numAliases, err := binary.ReadVarint(byter{r})
			if err != nil {
				return nil, err
			}

			// read out each alias name...
			for j := 0; j < int(numAliases); j++ {
				alias := ""

				err = varStringSchema.Decode(r, &alias)
				if err != nil {
					return nil, err
				}
				of.Aliases = append(of.Aliases, alias)
			}

			of.Schema, err = DecodeSchema(r)
			if err != nil {
				return nil, err
			}

			s.Fields = append(s.Fields, of)
		}

		return s, nil
	}

	// decode var object schema
	if curByte&ObjectMask == VarObjectByte {
		s := &VarObjectSchema{}
		s.SetNullable(curByte&NullMask > 0)

		s.Key, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		s.Value, err = DecodeSchema(r)
		if err != nil {
			return nil, err
		}

		return s, nil
	}

	return nil, fmt.Errorf("invalid binary schema encountered")
}

// PreEncode is a helper function that should be called by each Schema's Encode
// routine. It dereferences v if the value is a pointer or interface type and
// writes the null byte if nullable is set.
// If nullable is false and v resolves to nil, an error is returned.
// If nullable is true and v resolves to nil, (true, nil) is returned,
// indicating that no further processing is needed by the encoder who called
// this routine. Otherwise, false, nil is returned.
func PreEncode(w io.Writer, v *reflect.Value, nullable bool) (bool, error) {
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		*v = v.Elem()
	}

	// Note: v.Elem() returns invalid Value if v is nil
	isNil := !v.IsValid()

	if nullable {
		if isNil {
			// 1 indicates null
			w.Write([]byte{1})
			return true, nil
		} else {
			// 0 indicates not null
			w.Write([]byte{0})
		}
	} else if isNil {
		return false, fmt.Errorf("cannot encode nil value: schema is not nullable")
	}

	return false, nil
}

// PreDecode is a helper function that should be called by each Schema's Decode
// routine. It removes exactly one level of indirection for v and reads the
// null byte if nullable is set. If a null value is read, (true, nil) is
// returned, indicating that no further processing is needed by the decoder who
// called this routine. This routine also ensures that the destination value is
// settable and returns errors if not. Finally, this routine populates nested
// pointer values recursively, as needed.
func PreDecode(r io.Reader, v *reflect.Value, nullable bool) (bool, error) {
	// if v is a pointer or interface type, remove exactly ONE level of indirection
	if k := v.Kind(); !v.CanSet() && (k == reflect.Ptr || k == reflect.Interface) {
		*v = v.Elem()
	}

	// if the data indicates this type is nullable, then the actual
	// value is preceded by the null byte
	// (which indicates if the encoded value is null)
	if nullable {
		buf := make([]byte, 1)

		// first byte indicates whether value is null or not...
		_, err := io.ReadAtLeast(r, buf, 1)
		if err != nil {
			return false, err
		}
		isNull := (buf[0] == 1)

		if isNull {
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
				if v.CanSet() {
					// special way to set pointer to nil value
					v.Set(reflect.Zero(v.Type()))
					return true, nil
				}
				return false, fmt.Errorf("destination not settable")
			} else {
				return false, fmt.Errorf("cannot decode null value to a %s", v.Kind())
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
				return false, fmt.Errorf("destination not settable")
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		*v = v.Elem()
	}

	return false, nil
}
