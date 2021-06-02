package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type VarObjectSchema struct {
	SchemaOptions
	Key   Schema
	Value Schema
}

func (s *VarObjectSchema) DefaultGOType() reflect.Type {

	if s.Key == nil || s.Value == nil {
		return nil
	}

	var t = reflect.MapOf(s.Key.DefaultGOType(), s.Value.DefaultGOType())
	return t
}

func (s *VarObjectSchema) MarshalJSON() ([]byte, error) {

	tmpMap := make(map[string]interface{}, 1)
	tmpMap["type"] = "object"
	tmpMap["nullable"] = strconv.FormatBool(s.SchemaOptions.Nullable)

	// now encode the schema for the key
	keyJSON, err := s.Key.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var keyMap map[string]interface{}

	err = json.Unmarshal(keyJSON, &keyMap)
	if err != nil {
		return nil, err
	}

	tmpMap["key"] = keyMap

	// now encode the schema for the value
	ValueJSON, err := s.Value.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var valueMap map[string]interface{}

	err = json.Unmarshal(ValueJSON, &valueMap)
	if err != nil {
		return nil, err
	}

	tmpMap["value"] = valueMap

	return json.Marshal(tmpMap)
}

// Bytes encodes the schema in a portable binary format
func (s *VarObjectSchema) MarshalSchemer() []byte {

	// string schemas are 1 byte long
	var schema []byte = []byte{0b00101000}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schema[0] |= 128
	}

	// bit 3 is clear from above, indicating this is a var length string

	schema = append(schema, s.Key.MarshalSchemer()...)
	schema = append(schema, s.Value.MarshalSchemer()...)

	return schema
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *VarObjectSchema) Encode(w io.Writer, i interface{}) error {

	v := reflect.ValueOf(i)

	ok, err := PreEncode(s, w, &v)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Map {
		return fmt.Errorf("varObjectSchema can only encode maps")
	}

	err = writeVarUint(w, uint64(v.Len()))
	if err != nil {
		return errors.New("cannot encode var string length as var int")
	}

	for _, mapKey := range v.MapKeys() {
		mapValue := v.MapIndex(mapKey)

		err := s.Key.Encode(w, mapKey.Interface()) // encode key
		if err != nil {
			return err
		}
		err = s.Value.Encode(w, mapValue.Interface()) // encode value
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *VarObjectSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	ok, err := PreDecode(s, r, &v)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	t := v.Type()
	k := t.Kind()

	if k == reflect.Interface {
		var mapType = s.DefaultGOType()
		v.Set(reflect.MakeMap(mapType))

		v = v.Elem()

		t = v.Type()
		k = t.Kind()
	}

	if k != reflect.Map {
		return fmt.Errorf("VarObjectSchema can only decode to maps")
	}

	// we wrote the number of entries in the map as a var int
	// when we did the encoding
	expectedNumEntries, err := readVarUint(r)
	if err != nil {
		return err
	}

	if v.IsNil() {
		if !v.CanSet() {
			return errors.New("v not settable")
		}
		var mapType = reflect.MapOf(t.Key(), t.Elem())
		v.Set(reflect.MakeMap(mapType))
	}
	// else: we have an existing map
	// right now by default, we will just keep their entries
	// but we have to decide if this behavior is OK

	for i := 0; i < int(expectedNumEntries); i++ {

		key := reflect.New(t.Key())
		val := reflect.New(t.Elem())

		err = s.Key.DecodeValue(r, key) // decode key
		if err != nil {
			return err
		}
		err = s.Value.DecodeValue(r, val) // decode value
		if err != nil {
			return err
		}

		v.SetMapIndex(reflect.Indirect(key), reflect.Indirect(val))
	}

	return nil
}

func (s *VarObjectSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	return s.DecodeValue(r, reflect.ValueOf(i))
}

func (s *VarObjectSchema) Nullable() bool {
	return s.SchemaOptions.Nullable
}

func (s *VarObjectSchema) SetNullable(n bool) {
	s.SchemaOptions.Nullable = n
}
