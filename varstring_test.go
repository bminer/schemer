package schemer

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"testing"
)

func TestVarString1(t *testing.T) {

	fmt.Println("Testing decoding var length string values")

	// setup an example schema
	schema := VarLenStringSchema{SchemaOptions{Nullable: true}}

	// encode it
	b := schema.Bytes()

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(b)
	if err != nil {
		t.Error("cannot decode binary encoded string schema")
	}

	decodedStringSchema := tmp.(*VarLenStringSchema)
	if schema.SchemaOptions.Nullable != decodedStringSchema.Nullable() {

		// nothing else to test here...

		t.Error("unexpected value for VarLenStringSchema")
	}

}

func TestVarString2(t *testing.T) {

	varLenStringSchema := VarLenStringSchema{SchemaOptions: SchemaOptions{Nullable: false}}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = ""
	buf.Reset()

	err = varLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	log.Print(buf)

	// decode into string

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to string")
	var decodedValue1 string = "44"
	err = varLenStringSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != decodedValue1 {
		t.Errorf("Expected value: %s Decoded value: %s", valueToEncode, decodedValue1)
	}

}

func TestVarFixedString3(t *testing.T) {

	varLenStringSchema := VarLenStringSchema{SchemaOptions: SchemaOptions{Nullable: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "3.14159"
	buf.Reset()

	err = varLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to string")
	var decodedValue2 float32
	err = varLenStringSchema.Decode(r, &decodedValue2)
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

func TestVarFixedString4(t *testing.T) {

	varLenStringSchema := VarLenStringSchema{SchemaOptions: SchemaOptions{Nullable: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "3.14159"
	buf.Reset()

	err = varLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to string")
	var decodedValue2 float64
	err = varLenStringSchema.Decode(r, &decodedValue2)
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

func TestVarFixedString5(t *testing.T) {

	varLenStringSchema := VarLenStringSchema{SchemaOptions: SchemaOptions{Nullable: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "-101"
	buf.Reset()

	err = varLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to int")
	var decodedValue2 int
	err = varLenStringSchema.Decode(r, &decodedValue2)
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

func TestVarFixedString6(t *testing.T) {

	varLenStringSchema := VarLenStringSchema{SchemaOptions: SchemaOptions{Nullable: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode string = "103"
	buf.Reset()

	err = varLenStringSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into float32

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("fixed len string to uint")
	var decodedValue2 uint
	err = varLenStringSchema.Decode(r, &decodedValue2)
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
