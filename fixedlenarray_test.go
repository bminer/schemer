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

	fixedLenArraySchema := FixedLenArraySchema{IsNullable: false, Length: 10}
	fixedLenArraySchema.Element = FloatSchema{Bits: 32}

	var buf bytes.Buffer
	var err error
	floatarray := [6]float32{2.0, 3.1, 5.2, 7.3, 11.4, 13.5}
	buf.Reset()

	err = fixedLenArraySchema.Encode(&buf, floatarray)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("float32 to float32")

	r := bytes.NewReader(buf.Bytes())

	var decodedFloats [6]float32
	err = fixedLenArraySchema.Decode(r, decodedFloats)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < len(floatarray); i++ {
		if floatarray[i] != decodedFloats[i] {
			t.Error("unexpected value decoding floating point array")
		}
	}
}
