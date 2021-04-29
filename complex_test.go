package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

// make sure we can decode 64bit floating point values into each supported type
// when weakdecoding is enabled
func TestDecodeComplex1(t *testing.T) {

	ComplexSchema := ComplexNumberSchema{Bits: 64, WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex64 = 10 + 11i
	buf.Reset()

	fmt.Println("complex64 to complex64")

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

	fmt.Println("complex64 to complex128")

	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 complex128
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex64(decodedValue2) {
		t.Errorf("Expected value")
	}

}

// make sure we can decode complex numbers into floating points and integers
// when the imaginary component is 0
func TestDecodeComplex2(t *testing.T) {

	ComplexSchema := ComplexNumberSchema{Bits: 64, WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex64 = -123 + 0
	buf.Reset()

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("complex64 to float64")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 float64
	err = ComplexSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex64(complex(decodedValue1, 0)) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex64 to float32")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float32
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex64(complex(decodedValue2, 0)) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex64 to int")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue3 int
	err = ComplexSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex64(complex(float64(decodedValue3), 0)) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex64 to uint")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue4 uint
	err = ComplexSchema.Decode(r, &decodedValue4)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex64(complex(float64(decodedValue4), 0)) {
		t.Errorf("Expected value")
	}

}

// make sure that schemer will reject converting complex numbers to floating point
// or integer types when the complex component of the complex number is not zero
func TestDecodeComplex3(t *testing.T) {

	ComplexSchema := ComplexNumberSchema{Bits: 64, WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex64 = 123 + 12i
	buf.Reset()

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("complex64 to float64")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 float64
	err = ComplexSchema.Decode(r, &decodedValue1)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

	fmt.Println("complex64 to float32")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float32
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

	fmt.Println("complex64 to int")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue3 int
	err = ComplexSchema.Decode(r, &decodedValue3)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

	fmt.Println("complex64 to uint")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue4 uint
	err = ComplexSchema.Decode(r, &decodedValue4)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

}

// make sure we can decode 64bit floating point values into each supported type
// when weakdecoding is enabled
func TestDecodeComplex4(t *testing.T) {

	ComplexSchema := ComplexNumberSchema{Bits: 128, WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex128 = 10 + 11i
	buf.Reset()

	fmt.Println("complex128 to complex64")

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

	if valueToEncode != complex128(decodedValue1) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex128 to complex128")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 complex128
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex128(decodedValue2) {
		t.Errorf("Expected value")
	}

}
