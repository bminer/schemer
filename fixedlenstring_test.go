package schemer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TestDecodeString1 checks that we can encode / decode the binary schema for a fixed len string
func TestDecodeString1(t *testing.T) {

	// setup an example schema
	schema := FixedLenStringSchema{IsNullable: true, FixedLength: 80}

	// encode it
	b := schema.Bytes()

	// make sure we can successfully decode it
	var decodedStringSchema FixedLenStringSchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot decode binary encoded string schema")
	}

	decodedStringSchema = tmp.(FixedLenStringSchema)
	if schema.IsNullable != decodedStringSchema.IsNullable {

		// nothing else to test here...

		t.Error("unexpected value for FixedLenStringSchema")
	}

}

func TestDecodeString2(t *testing.T) {

	fixedLenStringSchema := FixedLenStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "hello, world!"
	buf.Reset()

	fmt.Println("Testing decoding fixed length string value")

	err = fixedLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into string

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to string")
	var decodedValue1 string
	err = fixedLenStringSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != strings.Trim(decodedValue1, " ") {
		t.Errorf("Expected value: %s Decoded value: %s", valueToEncode, decodedValue1)
	}

}

func TestDecodeString3(t *testing.T) {

	fixedLenStringSchema := FixedLenStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "3.14159"
	buf.Reset()

	fmt.Println("Testing decoding fixed length string value")

	err = fixedLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to string")
	var decodedValue2 float32
	err = fixedLenStringSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	vFloat, err := strconv.ParseFloat(valueToEncode, 32)
	if err != nil {
		t.Error(err)
	}

	if vFloat != float64(decodedValue2) {
		t.Errorf("Expected value: %f; Decoded value: %f", vFloat, decodedValue2)
	}

}

func TestDecodeString4(t *testing.T) {

	fixedLenStringSchema := FixedLenStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "3.14159"
	buf.Reset()

	fmt.Println("Testing decoding fixed length string value")

	err = fixedLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to string")
	var decodedValue2 float64
	err = fixedLenStringSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	vFloat, err := strconv.ParseFloat(valueToEncode, 64)
	if err != nil {
		t.Error(err)
	}

	if vFloat != decodedValue2 {
		t.Errorf("Expected value: %f; Decoded value: %f", vFloat, decodedValue2)
	}

}

func TestDecodeString5(t *testing.T) {

	fixedLenStringSchema := FixedLenStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "-101"
	buf.Reset()

	fmt.Println("Testing decoding fixed length string value")

	err = fixedLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to int")
	var decodedValue2 int
	err = fixedLenStringSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	i, err := strconv.Atoi(valueToEncode)
	if err != nil {
		t.Error(err)
	}

	if i != decodedValue2 {
		t.Errorf("Expected value: %d; Decoded value: %d", i, decodedValue2)
	}

}

func TestDecodeString6(t *testing.T) {

	fixedLenStringSchema := FixedLenStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "103"
	buf.Reset()

	fmt.Println("Testing decoding fixed length string value")

	err = fixedLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to uint")
	var decodedValue2 uint
	err = fixedLenStringSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	i, err := strconv.ParseUint(valueToEncode, 10, 64)
	if err != nil {
		t.Error(err)
	}

	if uint(i) != decodedValue2 {
		t.Errorf("Expected value: %d; Decoded value: %d", i, decodedValue2)
	}

}
