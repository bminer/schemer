package schemer

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
)

// make sure we can encode/decode binary schemas for EnumSchema
func TestDecodeVarLenArray1(t *testing.T) {

	slice := []int{1, 2, 3, 4}

	// setup an example schema
	s, err := SchemaOf(slice)
	if err != nil {
		t.Error(err)
		return
	}

	varArraySchema := s.(*VarArraySchema)

	// encode i
	b, err := varArraySchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
		return
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Error("cannot encode binary encoded VarLenArraySchema")
		return
	}

	decodedSchema := tmp.(*VarArraySchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedSchema.Nullable() != varArraySchema.Nullable() {
		t.Error("unexpected values when decoding binary EnumSchema")
		return
	}

}

func TestDecodeVarLenArray2(t *testing.T) {

	// build up the schema programatically
	varArraySchema := VarArraySchema{SchemaOptions: SchemaOptions{nullable: false}}
	varArraySchema.Element = &(FloatSchema{Bits: 32})

	var buf bytes.Buffer
	var err error
	floatslice := []float32{2.0, 3.1, 5.2, 7.3, 11.4, 13.5}
	buf.Reset()

	err = varArraySchema.Encode(&buf, floatslice)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("fixed len float32 to fixed len float32")

	r := bytes.NewReader(buf.Bytes())

	var decodedFloats []float32 = make([]float32, 6)
	err = varArraySchema.Decode(r, &decodedFloats)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(decodedFloats); i++ {
		if floatslice[i] != decodedFloats[i] {
			t.Error("unexpected value decoding floating point array")
		}
	}
}

func TestDecodeVarLenArray3(t *testing.T) {

	// build up the schema programatically
	/*
		FixedArraySchema := VarArraySchema{IsNullable: false}
		fixedLenArraySchema1 := VarArraySchema{IsNullable: false}
		FloatSchema := FloatSchema{Bits: 32}

		fixedLenArraySchema1.Element = FloatSchema
		FixedArraySchema.Element = fixedLenArraySchema1
	*/

	var buf bytes.Buffer
	var err error
	floatSlice := [][]float32{
		{0.5, 1, 2, 3},
		{4, 5, 6, 7},
		{8, 9, 10, 11},
	}

	FixedArraySchema, err := SchemaOf(floatSlice)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()

	err = FixedArraySchema.Encode(&buf, floatSlice)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("fixed len float32 to fixed len float32")

	r := bytes.NewReader(buf.Bytes())

	log.Println("buf.Bytes()", buf.Len(), buf.Bytes())

	decodedFloats := make([][]float32, 3)
	for i := range decodedFloats {
		decodedFloats[i] = make([]float32, 4)
	}

	err = FixedArraySchema.Decode(r, &decodedFloats)
	if err != nil {
		t.Error(err)
	}

	log.Println(decodedFloats)

	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			if floatSlice[i][j] != decodedFloats[i][j] {
				t.Error("unexpected value decoding 2-dimensional floating point array")
			}
		}
	}
}

// test JSON marshaling...
// to make sure schemer version number is present, and is correctly stripped from child elements
func TestDecodeVarLenArray4(t *testing.T) {

	slice := []int{1, 2, 3, 4}

	// setup an example schema
	s, err := SchemaOf(slice)
	if err != nil {
		t.Error(err)
		return
	}

	varArraySchema := s.(*VarArraySchema)

	b, err := varArraySchema.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if count := strings.Count(string(b), "version"); count != 1 {
		t.Error("expected 1 JSON version element; got:", count)
	}

}
