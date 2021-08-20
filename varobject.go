package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type VarObjectSchema struct {
	SchemaOptions
	Key   Schema
	Value Schema
}

func (s *VarObjectSchema) GoType() reflect.Type {

	if s.Key == nil || s.Value == nil {
		return nil
	}

	retval := reflect.MapOf(s.Key.GoType(), s.Value.GoType())

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}

	return retval
}

func (s *VarObjectSchema) MarshalJSON() ([]byte, error) {

	tmpMap := make(map[string]interface{}, 1)
	tmpMap["type"] = "object"
	tmpMap["nullable"] = s.Nullable()

	m := s.Key.(json.Marshaler)

	// now encode the schema for the key
	keyJSON, err := m.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var keyMap map[string]interface{}

	err = json.Unmarshal(keyJSON, &keyMap)
	if err != nil {
		return nil, err
	}

	tmpMap["key"] = keyMap

	tmp, ok := s.Value.(json.Marshaler)
	if !ok {
		return nil, fmt.Errorf("json.marshaler assertion failed")
	}

	// now encode the schema for the value
	ValueJSON, err := tmp.MarshalJSON()
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
func (s *VarObjectSchema) MarshalSchemer() ([]byte, error) {

	// string schemas are 1 byte long
	var schema []byte = []byte{VarObjectSchemaMask}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= 0x80
	}

	// bit 3 is clear from above, indicating this is a var length string

	k := s.Key.(Marshaler)
	key, err := k.MarshalSchemer()
	if err != nil {
		return nil, err
	}

	v := s.Value.(Marshaler)
	value, err := v.MarshalSchemer()
	if err != nil {
		return nil, err
	}

	schema = append(schema, key...)
	schema = append(schema, value...)

	return schema, nil
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *VarObjectSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output streamm
func (s *VarObjectSchema) EncodeValue(w io.Writer, v reflect.Value) error {

	ok, err := PreEncode(s.Nullable(), w, &v)
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

	err = WriteVarUint(w, uint64(v.Len()))
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

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *VarObjectSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *VarObjectSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	v, err := PreDecode(s.Nullable(), r, v)
	if err != nil {
		return err
	}
	// if PreDecode() returns a zero value for v, it means we are done decoding
	if !(v.IsValid()) {
		return nil
	}

	t := v.Type()
	k := t.Kind()

	if k == reflect.Interface {
		var mapType = s.GoType()
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
	expectedNumEntries, err := ReadVarUint(r)
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
