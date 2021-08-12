package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecodeVarObject1(t *testing.T) {

	m := map[int]string{1: "b"}

	// setup an example schema
	s, err := SchemaOf(m)
	if err != nil {
		t.Error(err)
	}

	varObjectSchema, ok := s.(*VarObjectSchema)
	if !ok {
		t.Error("varObjectSchema assertion failed")
		return
	}

	// encode it
	b, err := varObjectSchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded VarObjectSchema")
	}

	decodedIntSchema := tmp.(*VarObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.Nullable() != varObjectSchema.Nullable() {
		t.Error("unexpected values when decoding binary EnumSchema")
	}

}

func TestDecodeVarObject2(t *testing.T) {

	strToIntMap := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
	}

	var buf bytes.Buffer
	var err error

	buf.Reset()

	varObjectSchema, err := SchemaOf(&strToIntMap)
	if err != nil {
		t.Error(err)
		return
	}

	err = varObjectSchema.Encode(&buf, strToIntMap)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("map to map")

	r := bytes.NewReader(buf.Bytes())

	// pass in a nil map
	// to make sure decode will allocate it for us
	var mapToDecode map[string]int

	err = varObjectSchema.Decode(r, &mapToDecode)
	if err != nil {
		t.Error(err)
		return
	}

	for key, element := range strToIntMap {
		if element != mapToDecode[key] {
			t.Error("encoded data not present in decoded map")
		}
	}

}
