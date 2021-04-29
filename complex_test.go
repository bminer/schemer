package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecodeComplex(t *testing.T) {

	ComplexSchema := ComplexNumber{Bits: 64, WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex64 = 10 + 11i
	buf.Reset()

	fmt.Println("Testing decoding complex64")

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 complex64
	err = ComplexSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != decodedValue1 {
		t.Errorf("Expected value")
	}
}
