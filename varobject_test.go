package schemer

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

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

// test JSON marshaling...
// to make sure schemer version number is present, and is correctly stripped from child elements
func TestDecodeVarObject3(t *testing.T) {

	m := map[int]string{1: "b"}

	// setup an example schema
	s, err := SchemaOf(m)
	if err != nil {
		t.Fatal(err)
	}

	varObjectSchema, ok := s.(*VarObjectSchema)
	if !ok {
		t.Fatal("varObjectSchema assertion failed")
		return
	}

	// encode it
	b, err := varObjectSchema.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	if count := strings.Count(string(b), "version"); count != 1 {
		t.Error("expected 1 JSON version element; got:", count)
	}

}

// test binary marshalling / unmarshalling schema
func TestDecodeVarObject4(t *testing.T) {

	m := map[int]string{1: "b"}

	// setup an example schema
	s, err := SchemaOf(m)
	if err != nil {
		t.Fatal(err)
	}

	varObjectSchema, ok := s.(*VarObjectSchema)
	if !ok {
		t.Fatal("varObjectSchema assertion failed")
		return
	}

	// encode it
	b, err := varObjectSchema.MarshalSchemer()
	if err != nil {
		t.Fatal(err)
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Fatal("cannot decode VarIntSchema schema")
	}

	decodedIntSchema := tmp.(*VarObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.Nullable() != varObjectSchema.Nullable() {
		t.Fatal("unexpected value for nullable in varObjectSchema")
	}

}

// test json marshalling / unmarshalling schema
func TestDecodeVarObject5(t *testing.T) {

	m := map[int]string{1: "b"}

	// setup an example schema
	s, err := SchemaOf(m)
	if err != nil {
		t.Fatal(err)
	}

	varObjectSchema, ok := s.(*VarObjectSchema)
	if !ok {
		t.Fatal("varObjectSchema assertion failed")
		return
	}

	// encode it
	b, err := varObjectSchema.MarshalSchemer()
	if err != nil {
		t.Fatal(err)
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err, "; cannot decode varObjectSchema")
	}

	decodedIntSchema := tmp.(*VarObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.Nullable() != varObjectSchema.Nullable() {
		t.Fatal("unexpected value for nullable in varObjectSchema")
	}

}
