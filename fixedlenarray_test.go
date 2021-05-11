package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecodeFixedLenArray1(t *testing.T) {

	// setup an example schema
	fixedLenArraySchema := FixedLenArraySchema{IsNullable: false, Length: 10}

	// encode it
	b := fixedLenArraySchema.Bytes()

	// make sure we can successfully decode it
	var decodedSchema FixedLenArraySchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded enumSchema")
	}

	decodedSchema = tmp.(FixedLenArraySchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedSchema.IsNullable != fixedLenArraySchema.IsNullable {
		t.Error("unexpected values when decoding binary EnumSchema")
	}

}

func TestDecodeFixedLenArray2(t *testing.T) {

	// build up the schema programatically
	fixedLenArraySchema := FixedLenArraySchema{IsNullable: false, Length: 6}
	fixedLenArraySchema.Element = FloatSchema{Bits: 32}

	var buf bytes.Buffer
	var err error
	floatarray := [6]float32{2.0, 3.1, 5.2, 7.3, 11.4, 13.5}
	buf.Reset()

	// we should also be able to say:
	// schmemaOf(floatarray)

	err = fixedLenArraySchema.Encode(&buf, floatarray)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("fixed len float32 to fixed len float32")

	r := bytes.NewReader(buf.Bytes())

	var decodedFloats [6]float32
	err = fixedLenArraySchema.Decode(r, &decodedFloats)
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

	// build up the schema programatically
	fixedLenArraySchema := FixedLenArraySchema{IsNullable: false, Length: 3}
	fixedLenArraySchema1 := FixedLenArraySchema{IsNullable: false, Length: 4}
	FloatSchema := FloatSchema{Bits: 32}

	fixedLenArraySchema1.Element = FloatSchema
	fixedLenArraySchema.Element = fixedLenArraySchema1

	var buf bytes.Buffer
	var err error
	floatArray := [3][4]float32{
		{0.5, 1, 2, 3},
		{4, 5, 6, 7},
		{8, 9, 10, 11},
	}
	buf.Reset()

	err = fixedLenArraySchema.Encode(&buf, floatArray)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("fixed len float32 to fixed len float32")

	r := bytes.NewReader(buf.Bytes())

	decodedFloats := [3][4]float32{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}
	err = fixedLenArraySchema.Decode(r, &decodedFloats)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			if floatArray[i][j] != decodedFloats[i][j] {
				t.Error("unexpected value decoding 2-dimensional floating point array")
			}
		}
	}
}
