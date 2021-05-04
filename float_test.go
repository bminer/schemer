package schemer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// make sure we can decode 32bit floating point values into each supported type
// when weakdecoding is enabled
func TestFloatingPointSchema1(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 32, WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode float32 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

	if valueToEncode != decodedValue1 {
		t.Errorf("Expected value")
	}

	fmt.Println("float32 to float64")

	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float64
	err = floatingPointSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != float32(decodedValue2) {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float32 to complex64")

	var decodedValue5 complex64
	err = floatingPointSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if complex(valueToEncode, 0) != decodedValue5 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float32 to complex128")

	var decodedValue6 complex128
	err = floatingPointSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if complex128(complex(valueToEncode, 0)) != decodedValue6 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float32 to string")

	var decodedValue7 string
	err = floatingPointSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatFloat(float64(valueToEncode), 'f', -1, 64) != decodedValue7 {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer prevents loss of precision
// when weakdecoding is disabled
func TestFloatingPointSchema2(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 32, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float32 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

	floatingPointSchema := FloatSchema{Bits: 32, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float32 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

	if valueToEncode != float32(decodedValue1) {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// make sure that schemer will throw an error if the floating point schema we
// are trying to use is invalid
func TestFloatingPointSchema4(t *testing.T) {

	// setup an invalid schema
	floatingPointSchema := FloatSchema{Bits: 8, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float32 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
	if err == nil {
		t.Error("schemer library failure; invalid floating point schema not flagged")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer throws an error when decoding (a 32bit float) results in
// overflow
func TestFloatingPointSchema5(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 32, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float32 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

	floatingPointSchema := FloatSchema{Bits: 64, WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode float64 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

	if float32(valueToEncode) != decodedValue1 {
		t.Errorf("Expected value")
	}

	fmt.Println("float64 to float64")

	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 float64
	err = floatingPointSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != decodedValue2 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float64 to complex64")

	var decodedValue5 complex64
	err = floatingPointSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if complex64(complex(valueToEncode, 0)) != decodedValue5 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float64 to complex128")

	var decodedValue6 complex128
	err = floatingPointSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if complex128(complex(valueToEncode, 0)) != decodedValue6 {
		t.Errorf("Expected value")
	}

	r = bytes.NewReader(buf.Bytes())
	fmt.Println("float64 to string")

	var decodedValue7 string
	err = floatingPointSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatFloat(float64(valueToEncode), 'f', -1, 64) != decodedValue7 {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer prevents loss of precision
// when weakdecoding is disabled
func TestFloatingPointSchema7(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 64, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float64 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

	floatingPointSchema := FloatSchema{Bits: 64, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float64 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

	if valueToEncode != float64(decodedValue1) {
		t.Errorf("Expected value")
	}

}

// test a corner case for floating point numbers:
// make sure that schemer will throw an error if the floating point schema we
// are trying to use is invalid
func TestFloatingPointScheme9(t *testing.T) {

	// setup an invalid schema
	floatingPointSchema := FloatSchema{Bits: 8, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float32 = 3.0
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
	if err == nil {
		t.Error("schemer library failure; invalid floating point schema not flagged")
	}

}

// test a corner case for floating point numbers:
// specifically test to make sure that schemer throws an error when decoding (a 64bit float) results in
// overflow
func TestFloatingPointSchema10(t *testing.T) {

	floatingPointSchema := FloatSchema{Bits: 64, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode float64 = 3.14
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, valueToEncode)
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

// make sure we can create a scheme from JSON encoded data, and make sure the
// data we get back matches what we passed in
func TestFloatingPointSchema11(t *testing.T) {

	jsonData := []byte("{\"Bits\":32,\"WeakDecoding\":false}")

	var floatingPointSchema FloatSchema
	err := floatingPointSchema.DoUnmarshalJSON(jsonData)
	if err != nil {
		t.Error(err)
	}

	if floatingPointSchema.Bits != 32 || floatingPointSchema.WeakDecoding != false {
		t.Error("schemer library failure; DoUnmarshalJSON unexpected result")
	}

	buf, _ := floatingPointSchema.DoMarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if !strings.EqualFold(string(jsonData), string(buf)) {
		t.Error("schemer library failure; DoUnmarshalJSON unexpected result")
	}

}

func TestFloatingPointSchema13(t *testing.T) {

	// setup an example schema
	floatingPointSchema := FloatSchema{Bits: 32, WeakDecoding: false, IsNullable: false}

	// encode it
	b := floatingPointSchema.Bytes()

	// make sure we can successfully decode it
	var DecodedFloatSchema FloatSchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot decode binary encoded float schema")
	}

	DecodedFloatSchema = tmp.(FloatSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if DecodedFloatSchema.Bits != floatingPointSchema.Bits ||
		DecodedFloatSchema.IsNullable != floatingPointSchema.IsNullable {

		t.Error("unexpected values when decoding binary FloatSchema")
	}

}
