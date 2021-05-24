package schemer

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"
)

// TestDecodeString1 checks that we can encode / decode the binary schema for a fixed len string
func TestDecodeFixedString1(t *testing.T) {

	// setup an example schema
	schema := FixedStringSchema{IsNullable: true, FixedLength: 80}

	// encode it
	b := schema.Bytes()

	// make sure we can successfully decode it
	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot decode binary encoded string schema")
	}

	decodedStringSchema := tmp.(*FixedStringSchema)
	if schema.IsNullable != decodedStringSchema.Nullable() {

		// nothing else to test here...

		t.Error("unexpected value for FixedStringSchema")
	}

}

func TestDecodeFixedString2(t *testing.T) {

	fixedLenStringSchema := FixedStringSchema{IsNullable: true, FixedLength: 80}

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

func TestDecodeFixedString3(t *testing.T) {

	fixedLenStringSchema := FixedStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "3.14159"
	buf.Reset()

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

func TestDecodeFixedString4(t *testing.T) {

	fixedLenStringSchema := FixedStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "3.14159"
	buf.Reset()

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

func TestDecodeFixedString5(t *testing.T) {

	fixedLenStringSchema := FixedStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "-101"
	buf.Reset()

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

func TestDecodeFixedString6(t *testing.T) {

	fixedLenStringSchema := FixedStringSchema{IsNullable: true, FixedLength: 80}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "103"
	buf.Reset()

	err = fixedLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

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

func TestFixedStrWriter(t *testing.T) {

	strToEncode := ""
	fixedLenStrSchema := FixedStringSchema{FixedLength: 8, IsNullable: false}

	binaryReaderSchema := fixedLenStrSchema.Bytes()

	var encodedData bytes.Buffer

	err := fixedLenStrSchema.Encode(&encodedData, strToEncode)
	if err != nil {
		t.Error(err)
	}

	saveToDisk("/tmp/FixedLenString.schema", binaryReaderSchema)
	saveToDisk("/tmp/FixedLenString.data", encodedData.Bytes())

}

func TestFixedStrReader(t *testing.T) {

	var strToDecodeTo string

	binarywriterSchema := readFromDisk("/tmp/FixedLenString.schema")
	writerSchema, err := NewSchema(binarywriterSchema)
	if err != nil {
		t.Error("cannot create writerSchema from raw binary data")
	}

	encodedData := readFromDisk("/tmp/FixedLenString.data")
	r := bytes.NewReader(encodedData)

	err = writerSchema.Decode(r, &strToDecodeTo)
	if err != nil {
		t.Error(err)
	}

	log.Println(strToDecodeTo)

}
func TestFixedStrWriterSerialize(t *testing.T) {
	TestFixedStrWriter(t)
	TestFixedStrReader(t)
}
