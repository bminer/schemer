package schemer

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
)

func TestDecodeFixedLenArray2(t *testing.T) {

	var buf bytes.Buffer
	var err error
	floatarray := [6]float32{2.0, 3.1, 5.2, 7.3, 11.4, 13.5}
	buf.Reset()

	FixedArraySchema, err := SchemaOf(floatarray)
	if err != nil {
		t.Error(err)
	}

	err = FixedArraySchema.Encode(&buf, floatarray)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("fixed len float32 to fixed len float32")

	r := bytes.NewReader(buf.Bytes())

	var decodedFloats [6]float32
	err = FixedArraySchema.Decode(r, &decodedFloats)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(floatarray); i++ {
		if floatarray[i] != decodedFloats[i] {
			t.Error("unexpected value decoding floating point array")
		}
	}
}

func TestDecodeFixedLenArray3(t *testing.T) {

	var buf bytes.Buffer
	var err error
	floatArray := [3][4]float32{
		{0.5, 1, 2, 3},
		{4, 5, 6, 7},
		{8, 9, 10, 11},
	}
	buf.Reset()

	FixedArraySchema, err := SchemaOf(floatArray)
	if err != nil {
		t.Error(err)
	}

	err = FixedArraySchema.Encode(&buf, floatArray)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("fixed len float32 to fixed len float32")

	r := bytes.NewReader(buf.Bytes())

	log.Println("buf.Bytes()", buf.Len(), buf.Bytes())

	decodedFloats := [3][4]float32{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}
	err = FixedArraySchema.Decode(r, &decodedFloats)
	if err != nil {
		t.Error(err)
	}

	log.Println(decodedFloats)

	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			if floatArray[i][j] != decodedFloats[i][j] {
				t.Error("unexpected value decoding 2-dimensional floating point array")
			}
		}
	}
}

func TestDecodeFixedLenArray4(t *testing.T) {

	var buf bytes.Buffer
	var err error
	floatArray := [3][4]bool{
		{false, true, false, true},
		{false, true, false, true},
		{false, true, false, true},
	}
	buf.Reset()

	FixedArraySchema, err := SchemaOf(floatArray)
	if err != nil {
		t.Error(err)
	}

	err = FixedArraySchema.Encode(&buf, floatArray)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("fixed len bool to fixed len bool")

	r := bytes.NewReader(buf.Bytes())

	var decodedBools [3][4]bool
	err = FixedArraySchema.Decode(r, &decodedBools)
	if err != nil {
		t.Error(err)
	}

	log.Println(decodedBools)

	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			if floatArray[i][j] != decodedBools[i][j] {
				t.Error("unexpected value decoding 2-dimensional boolean array")
			}
		}
	}
}

// test JSON marshaling...
// to make sure schemer version number is present, and is correctly stripped from child elements
func TestDecodeFixedLenArray5(t *testing.T) {

	var testarray [10]byte = [10]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	s, err := SchemaOf(testarray)
	if err != nil {
		t.Error(err)
	}

	fixedArraySchema, ok := s.(*FixedArraySchema)
	if !ok {
		t.Error("expected a *FixedArraySchema")
	}

	b, err := fixedArraySchema.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if count := strings.Count(string(b), "version"); count != 1 {
		t.Error("expected 1 JSON version element; got:", count)
	}
}

// test binary marshalling / unmarshalling schema
func TestDecodeFixedLenArray6(t *testing.T) {

	var testarray [10]byte = [10]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	s, err := SchemaOf(testarray)
	if err != nil {
		t.Error(err)
	}

	schema, ok := s.(*FixedArraySchema)
	if !ok {
		t.Fatal("expected a *FixedArraySchema")
	}
	schema.SetNullable(true)

	// encode it
	b, err := schema.MarshalSchemer()
	if err != nil {
		t.Fatal(err, "; cannot marshall schemer")
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err, "; cannot decode FixedArraySchema")
	}

	decodedSchema := tmp.(*FixedArraySchema)
	if decodedSchema.Nullable() != schema.Nullable() {
		t.Fatal("unexpected value for nullable in FixedArraySchema")
	}

}

// test binary marshalling / unmarshalling schema
func TestDecodeFixedLenArray7(t *testing.T) {

	var testarray [10]byte = [10]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	s, err := SchemaOf(testarray)
	if err != nil {
		t.Error(err)
	}

	schema, ok := s.(*FixedArraySchema)
	if !ok {
		t.Fatal("expected a *FixedArraySchema")
	}
	schema.SetNullable(true)

	// encode it
	b, err := schema.MarshalJSON()
	if err != nil {
		t.Fatal(err, "; cannot marshall schemer")
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchemaJSON(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err, "; cannot decode FixedArraySchema")
	}

	decodedSchema := tmp.(*FixedArraySchema)
	if decodedSchema.Nullable() != schema.Nullable() {
		t.Fatal("unexpected value for nullable in FixedArraySchema")
	}

}
