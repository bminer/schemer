package schemer

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

func TestDecodeFixedLenArray1(t *testing.T) {

	var testarray [10]byte = [10]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	s, err := SchemaOf(testarray)
	if err != nil {
		t.Error(err)
	}

	fixedArraySchema, ok := s.(*FixedArraySchema)
	if !ok {
		t.Error("expected a *FixedArraySchema")
	}

	// encode it
	b, err := fixedArraySchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Error("cannot encode binary encoded FixedArraySchema")
	}

	decodedSchema := tmp.(*FixedArraySchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedSchema.Nullable() != fixedArraySchema.Nullable() {
		t.Error("unexpected values when decoding binary fixedArraySchema")
	}

}

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
