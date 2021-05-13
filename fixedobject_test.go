package schemer

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestStructToNest2 struct {
	A [3]int8
	B []string
	C *string
	D *int
}

type TestStructToNest struct {
	Test123 float64
	M       map[string]int
	TestStructToNest2
}

type TestStruct struct {
	A string
	T TestStructToNest
}

func TestDecodeFixedObject1(t *testing.T) {

	// setup an example schema
	fixedObjectSchema := FixedObjectSchema{IsNullable: false}

	// encode it
	b := fixedObjectSchema.Bytes()

	// make sure we can successfully decode it
	var decodedIntSchema FixedObjectSchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded FixedObjectSchema")
	}

	decodedIntSchema = tmp.(FixedObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.IsNullable != fixedObjectSchema.IsNullable {
		t.Error("unexpected values when decoding binary EnumSchema")
	}

}

func TestDecodeFixedObject2(t *testing.T) {

	//s := "string"
	i := 10

	var structToEncode = TestStruct{"ben", TestStructToNest{99.9, map[string]int{
		"a": 1,
		"b": 2,
		"d": 3},
		TestStructToNest2{[3]int8{4, 5, 6}, []string{"go", "reflection", "rocks"}, nil, &i}}}

	var buf bytes.Buffer
	var err error

	buf.Reset()

	fixedObjectSchema := SchemaOf(&structToEncode)

	err = fixedObjectSchema.Encode(&buf, structToEncode)
	if err != nil {
		t.Error(err)
	}

	log.Print(buf.Bytes())

	fmt.Println("struct to struct")

	r := bytes.NewReader(buf.Bytes())

	var structToDecode = TestStruct{}

	log.Print(structToEncode)

	err = fixedObjectSchema.DecodeValue(r, reflect.ValueOf(&structToDecode))

	log.Print(structToDecode)

	if err != nil {
		t.Error(err)
	}

	if !cmp.Equal(structToEncode, structToDecode) {
		t.Error("unexpected struct to struct decode")
	}

}
