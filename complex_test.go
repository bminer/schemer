package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

// make sure we can decode complex64 numbers into both complex64 and complex128
// this is the base case
func TestDecodeComplex1(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: false}}

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

// make sure we can decode complex64 numbers into floating points and integers
// when the imaginary component is 0
func TestDecodeComplex2(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex64 = 123 + 0
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

}

// make sure that schemer will reject converting complex64 to floating point
// or integer types when the complex component of the complex number is not zero
func TestDecodeComplex3(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex64 = 123 + 12i // imaginary part is present... should not be able to decode!
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

// make sure we can decode complex128 into both complex128 and complex64
func TestDecodeComplex4(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 128, SchemaOptions: SchemaOptions{weakDecoding: true}}

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

// make sure we can decode complex128 numbers into floating points and integers
// when the imaginary component is 0
func TestDecodeComplex5(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 128, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex128 = 123 + 0
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

	if valueToEncode != complex128(complex(decodedValue1, 0)) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex64 to float32")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float32
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex128(complex(decodedValue2, 0)) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex64 to int")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue3 int
	err = ComplexSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex128(complex(float64(decodedValue3), 0)) {
		t.Errorf("Expected value")
	}

}

// corner case:
// specifically make sure that schemer will throw an error if we try to
// decode a negative complex64 into an uint
func TestDecodeComplex6(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex64 = -123 + 0
	buf.Reset()

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("complex64 to uint")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue4 uint
	err = ComplexSchema.Decode(r, &decodedValue4)
	if err == nil {
		t.Error("schemer library decoding failure; should not be able to decode negative complex64 into uint")
	}

}

// corner case:
// specifically make sure that schemer will throw an error if we try to
// decode a negative complex128 into an uint
func TestDecodeComplex7(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 128, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex128 = -123 + 0
	buf.Reset()

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("complex128 to uint")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue4 uint
	err = ComplexSchema.Decode(r, &decodedValue4)
	if err == nil {
		t.Error("schemer library decoding failure; should not be able to decode negative complex128 into uint")
	}

}

// make sure that schemer will reject converting complex128 to floating point
// or integer types when the complex component of the complex number is not zero
func TestDecodeComplex8(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 128, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex128 = 123 + 12i // imaginary part is present... should not be able to decode!
	buf.Reset()

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("complex128 to float64")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 float64
	err = ComplexSchema.Decode(r, &decodedValue1)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

	fmt.Println("complex128 to float32")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float32
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

	fmt.Println("complex128 to int")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue3 int
	err = ComplexSchema.Decode(r, &decodedValue3)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

	fmt.Println("complex128	 to uint")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue4 uint
	err = ComplexSchema.Decode(r, &decodedValue4)
	if err == nil {
		t.Error("schemer library decoding failure; imginary component present")
	}

}

// make sure we can decode a complex number into float32 and float64 arrays
func TestDecodeComplex9(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 128, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex128 = 123 + 12i // imaginary part is present... should not be able to decode!
	buf.Reset()

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("complex128 to float32 array")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 [2]float32

	err = ComplexSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex128(complex(decodedValue1[0], decodedValue1[1])) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex128 to float64 array")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 [2]float64
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex128(complex(decodedValue2[0], decodedValue2[1])) {
		t.Errorf("Expected value")
	}

}

// make sure we can decode a complex number into float and float64 slices
func TestDecodeComplex10(t *testing.T) {

	ComplexSchema := ComplexSchema{Bits: 128, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode complex128 = 123 + 12i // imaginary part is present... should not be able to decode!
	buf.Reset()

	ComplexSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("complex128 to float32 slice")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 []float32 = make([]float32, 2)

	err = ComplexSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex128(complex(decodedValue1[0], decodedValue1[1])) {
		t.Errorf("Expected value")
	}

	fmt.Println("complex128 to float64 slice")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 []float64 = make([]float64, 2)
	err = ComplexSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != complex128(complex(decodedValue2[0], decodedValue2[1])) {
		t.Errorf("Expected value")
	}

}

func TestDecodeComplex11(t *testing.T) {

	// setup an example schema
	complexSchema := ComplexSchema{Bits: 128, SchemaOptions: SchemaOptions{nullable: false, weakDecoding: true}}

	// encode it
	b, err := complexSchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
	}

	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Error("cannot decode binary encoded float schema")
	}

	DecodedComplexSchema := tmp.(*ComplexSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if DecodedComplexSchema.Bits != complexSchema.Bits ||
		DecodedComplexSchema.Nullable() != complexSchema.Nullable() {

		t.Error("unexpected values when decoding binary ComplexSchema")
	}

}

func TestDecodeComplex12(t *testing.T) {

	fmt.Println("decode nil complex")

	complexSchema := ComplexSchema{Bits: 64, SchemaOptions: SchemaOptions{nullable: true}}

	var buf bytes.Buffer
	var err error
	var complex64Ptr *complex64
	buf.Reset()

	complex64Ptr = nil
	err = complexSchema.Encode(&buf, complex64Ptr)
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var complexToDecodeTo complex64
	var complexPtr *complex64 = &complexToDecodeTo

	err = complexSchema.Decode(r, &complexPtr)
	if err != nil {
		t.Error(err)
	}

	// complexPtr should be a nil pointer once we decoded it!
	if complexPtr != nil {
		t.Error("unexpected value decoding null complexPtr")
	}

}
