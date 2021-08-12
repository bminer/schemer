package schemer

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

// make sure we can decode 32bit floating point values into each supported type
// when weakdecoding is enabled
func TestFloatingPointSchema1(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 32, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var floatPtr float32 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("float32 to float32")

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 float32
	err = floatingPointSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if floatPtr != decodedValue1 {
		t.Errorf("Expected value")
	}

	fmt.Println("float32 to float64")

	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float64
	err = floatingPointSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if floatPtr != float32(decodedValue2) {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float32 to complex64")

	var decodedValue5 complex64
	err = floatingPointSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if complex(floatPtr, 0) != decodedValue5 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float32 to complex128")

	var decodedValue6 complex128
	err = floatingPointSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if complex128(complex(floatPtr, 0)) != decodedValue6 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float32 to string")

	var decodedValue7 string
	err = floatingPointSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatFloat(float64(floatPtr), 'f', -1, 64) != decodedValue7 {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer prevents loss of precision
// when weakdecoding is disabled
func TestFloatingPointSchema2(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 32, SchemaOptions: SchemaOptions{weakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var floatPtr float32 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 int
	err = floatingPointSchema.Decode(r, &decodedValue1)

	// make sure schemer will throw an error if we try to decode to an integer
	// which will result in loss of precision

	if err == nil {
		t.Error("schemer library failure; decoding a float into an integer resulted in inappropriate loss of precision...")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure we can decode whole numbers encoded into float32's into integer types,
// even if weakdecoding is false
func TestFloatingPointSchema3(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 32, SchemaOptions: SchemaOptions{weakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var floatPtr float32 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 int
	err = floatingPointSchema.Decode(r, &decodedValue1)

	// even though WeakDecoding is false, we should still be able to decode into
	// an integer type if the floating point value is a whole number

	if err != nil {
		t.Error(err)
	}

	if floatPtr != float32(decodedValue1) {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// make sure that schemer will throw an error if the floating point schema we
// are trying to use is invalid
func TestFloatingPointSchema4(t *testing.T) {

	// setup an invalid schema
	floatingPointSchema := FloatSchema{Bits: 8, SchemaOptions: SchemaOptions{weakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var floatPtr float32 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err == nil {
		t.Error("schemer library failure; invalid floating point schema not flagged")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer throws an error when decoding (a 32bit float) results in
// overflow
func TestFloatingPointSchema5(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 32, SchemaOptions: SchemaOptions{weakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var floatPtr float32 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 int8
	err = floatingPointSchema.Decode(r, &decodedValue1)

	// make sure schemer will throw an error if we try to decode to an integer
	// which will result in loss of precision

	if err == nil {
		t.Error("schemer library failure; overflow not detected when decoding a float32 to an int8")
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////////
// the routines below are the same as the ones above, except they are for 64bit floating point
// values
////////////////////////////////////////////////////////////////////////////////////////////////////

// make sure we can decode 64bit floating point values into each supported type
// when weakdecoding is enabled.
func TestFloatingPointSchema6(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var floatPtr float64 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("float64 to float32")

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 float32
	err = floatingPointSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if float32(floatPtr) != decodedValue1 {
		t.Errorf("Expected value")
	}

	fmt.Println("float64 to float64")

	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float64
	err = floatingPointSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if floatPtr != decodedValue2 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float64 to complex64")

	var decodedValue5 complex64
	err = floatingPointSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if complex64(complex(floatPtr, 0)) != decodedValue5 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float64 to complex128")

	var decodedValue6 complex128
	err = floatingPointSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if complex128(complex(floatPtr, 0)) != decodedValue6 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float64 to string")

	var decodedValue7 string
	err = floatingPointSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatFloat(float64(floatPtr), 'f', -1, 64) != decodedValue7 {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer prevents loss of precision
// when weakdecoding is disabled
func TestFloatingPointSchema7(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var floatPtr float64 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 int
	err = floatingPointSchema.Decode(r, &decodedValue1)

	// make sure schemer will throw an error if we try to decode to an integer
	// which will result in loss of precision

	if err == nil {
		t.Error("schemer library failure; decoding a float into an integer resulted in inappropriate loss of precision...")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure we can decode whole numbers encoded into float64's into integer types,
// even if weakdecoding is false
func TestFloatingPointSchema8(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var floatPtr float64 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 int
	err = floatingPointSchema.Decode(r, &decodedValue1)

	// even though WeakDecoding is false, we should still be able to decode into
	// an integer type if the floating point value is a whole number

	if err != nil {
		t.Error(err)
	}

	if floatPtr != float64(decodedValue1) {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// make sure that schemer will throw an error if the floating point schema we
// are trying to use is invalid
func TestFloatingPointScheme9(t *testing.T) {

	// setup an invalid schema
	floatingPointSchema := FloatSchema{Bits: 8, SchemaOptions: SchemaOptions{weakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var floatPtr float32 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, floatPtr)
	if err == nil {
		t.Error("schemer library failure; invalid floating point schema not flagged")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer throws an error when decoding (a 64bit float) results in
// overflow
func TestFloatingPointSchema10(t *testing.T) {

	/*

		floatingPointSchema := FloatSchema{Bits: 64, weakDecoding: false}

		var buf bytes.Buffer
		var err error
		var floatPtr float64 = 3.14
		buf.Reset()

		err = floatingPointSchema.Encode(&buf, floatPtr)
		if err != nil {
			t.Error(err)
		}

		r := bytes.NewReader(buf.Bytes())

		var decodedValue1 int8
		err = floatingPointSchema.Decode(r, &decodedValue1)

		// make sure schemer will throw an error if we try to decode to an integer
		// which will result in loss of precision

		if err == nil {
			t.Error("schemer library failure; overflow not detected when decoding a float32 to an int8")
		}

	*/

}

func TestFloatingPointSchema13(t *testing.T) {

	// setup an example schema
	floatingPointSchema := FloatSchema{Bits: 32, SchemaOptions: SchemaOptions{nullable: false, weakDecoding: false}}

	// encode it
	b, err := floatingPointSchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Error("cannot decode binary encoded float schema")
	}

	DecodedFloatSchema := tmp.(*FloatSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if DecodedFloatSchema.Bits != floatingPointSchema.Bits ||
		DecodedFloatSchema.Nullable() != floatingPointSchema.Nullable() {

		t.Error("unexpected values when decoding binary FloatSchema")
	}

}

func TestFloatingPointSchema14(t *testing.T) {

	fmt.Println("decode nil float")

	floatingPointSchema := FloatSchema{Bits: 64, SchemaOptions: SchemaOptions{nullable: true}}

	var buf bytes.Buffer
	var err error
	var myFloat = 3.14
	//var floatPtr *float64
	buf.Reset()

	//floatPtr = nil
	err = floatingPointSchema.Encode(&buf, &myFloat)
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var floatToDecodeTo float64
	var floatPtr2 *float64 = &floatToDecodeTo

	err = floatingPointSchema.Decode(r, floatPtr2)
	if err != nil {
		t.Error(err)
	}

	/*
		// floatPtr should be a nil pointer once we decoded it!
		if floatPtr2 != nil {
			t.Error("unexpected value decoding null float")
		}
	*/

}

func TestFloatingPointSchema15(t *testing.T) {

	// setup an example schema
	floatingPointSchema := FloatSchema{Bits: 32, SchemaOptions: SchemaOptions{nullable: true, weakDecoding: true}}

	// make sure we can successfully decode it
	//var DecodedFloatSchema FloatSchema
	var err error
	var buf []byte

	// encode it
	buf, err = floatingPointSchema.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(string(buf))

}
