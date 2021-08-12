package schemer

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

func TestFixedIntSchema1(t *testing.T) {

	fixedIntSchema := FixedIntSchema{Signed: true, Bits: 64, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode int64 = 100
	buf.Reset()

	fmt.Println("Testing decoding int64 value")

	err = fixedIntSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into signed integer types

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to int")
	var decodedValue1 int
	err = fixedIntSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue1) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue1)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to int64")
	var decodedValue2 int64
	err = fixedIntSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue2) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue2)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to int32")
	var decodedValue3 int32
	err = fixedIntSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue3) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue3)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to int16")
	var decodedValue4 int16
	err = fixedIntSchema.Decode(r, &decodedValue4)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue4) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue4)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to int8")
	var decodedValue5 int8
	err = fixedIntSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue5) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue5)
	}

	// decode into unsigned integer types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to uint")
	var decodedValue6 uint
	err = fixedIntSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue6) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue6)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to uint64")
	var decodedValue7 uint64
	err = fixedIntSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue7) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue7)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to uint32")
	var decodedValue8 uint32
	err = fixedIntSchema.Decode(r, &decodedValue8)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue8) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue8)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to uint16")
	var decodedValue9 uint16
	err = fixedIntSchema.Decode(r, &decodedValue9)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue9) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue9)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to uint8")
	var decodedValue10 uint8
	err = fixedIntSchema.Decode(r, &decodedValue10)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue10) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue10)
	}

	// decode into other types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to float32")
	var decodedValue11 float32
	err = fixedIntSchema.Decode(r, &decodedValue11)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue11) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue11)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to float64")
	var decodedValue12 float64
	err = fixedIntSchema.Decode(r, &decodedValue12)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue12) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue12)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to Complex64")
	var decodedValue13 complex64
	err = fixedIntSchema.Decode(r, &decodedValue13)
	if err != nil {
		t.Error(err)
	}

	if complex(float32(valueToEncode), 0) != decodedValue13 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue13)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to Complex128")
	var decodedValue14 complex128
	err = fixedIntSchema.Decode(r, &decodedValue14)
	if err != nil {
		t.Error(err)
	}

	if complex(float64(valueToEncode), 0) != decodedValue14 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue14)
	}

	// finally decode into weak types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to bool")
	var decodedValue15 bool
	err = fixedIntSchema.Decode(r, &decodedValue15)
	if err != nil {
		t.Error(err)
	}

	if !decodedValue15 {
		t.Errorf("decoding to bool produced false value")
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int64 to string")
	var decodedValue16 string
	err = fixedIntSchema.Decode(r, &decodedValue16)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatInt(int64(valueToEncode), 10) != decodedValue16 {
		t.Errorf("decoding to string producing unexpected value")
	}

}

func TestFixedIntSchema2(t *testing.T) {

	fixedIntSchema := FixedIntSchema{Signed: true, Bits: 32, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode int32 = 100
	buf.Reset()

	fmt.Println("Testing decoding int32 value")

	err = fixedIntSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into signed integer types

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to int")
	var decodedValue1 int
	err = fixedIntSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue1) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue1)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to int32")
	var decodedValue2 int32
	err = fixedIntSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue2) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue2)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to int32")
	var decodedValue3 int32
	err = fixedIntSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue3) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue3)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to int16")
	var decodedValue4 int16
	err = fixedIntSchema.Decode(r, &decodedValue4)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue4) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue4)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to int8")
	var decodedValue5 int8
	err = fixedIntSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue5) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue5)
	}

	// decode into unsigned integer types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to uint")
	var decodedValue6 uint
	err = fixedIntSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue6) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue6)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to uint64")
	var decodedValue7 uint64
	err = fixedIntSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue7) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue7)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to uint32")
	var decodedValue8 uint32
	err = fixedIntSchema.Decode(r, &decodedValue8)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue8) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue8)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to uint16")
	var decodedValue9 uint16
	err = fixedIntSchema.Decode(r, &decodedValue9)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue9) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue9)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to uint8")
	var decodedValue10 uint8
	err = fixedIntSchema.Decode(r, &decodedValue10)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue10) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue10)
	}

	// decode into other types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to float32")
	var decodedValue11 float32
	err = fixedIntSchema.Decode(r, &decodedValue11)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue11) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue11)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to float64")
	var decodedValue12 float64
	err = fixedIntSchema.Decode(r, &decodedValue12)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue12) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue12)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to Complex64")
	var decodedValue13 complex64
	err = fixedIntSchema.Decode(r, &decodedValue13)
	if err != nil {
		t.Error(err)
	}

	if complex(float32(valueToEncode), 0) != decodedValue13 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue13)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to Complex128")
	var decodedValue14 complex128
	err = fixedIntSchema.Decode(r, &decodedValue14)
	if err != nil {
		t.Error(err)
	}

	if complex(float64(valueToEncode), 0) != decodedValue14 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue14)
	}

	// finally decode into weak types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to bool")
	var decodedValue15 bool
	err = fixedIntSchema.Decode(r, &decodedValue15)
	if err != nil {
		t.Error(err)
	}

	if !decodedValue15 {
		t.Errorf("decoding to bool produced false value")
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int32 to string")
	var decodedValue16 string
	err = fixedIntSchema.Decode(r, &decodedValue16)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatInt(int64(valueToEncode), 10) != decodedValue16 {
		t.Errorf("decoding to string producing unexpected value")
	}

}

func TestFixedIntSchema3(t *testing.T) {

	fixedIntSchema := FixedIntSchema{Signed: true, Bits: 16, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode int16 = 100
	buf.Reset()

	fmt.Println("Testing decoding int16 value")

	err = fixedIntSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into signed integer types

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to int")
	var decodedValue1 int
	err = fixedIntSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue1) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue1)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to int16")
	var decodedValue2 int16
	err = fixedIntSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue2) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue2)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to int16")
	var decodedValue3 int16
	err = fixedIntSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue3) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue3)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to int16")
	var decodedValue4 int16
	err = fixedIntSchema.Decode(r, &decodedValue4)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue4) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue4)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to int8")
	var decodedValue5 int8
	err = fixedIntSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue5) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue5)
	}

	// decode into unsigned integer types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to uint")
	var decodedValue6 uint
	err = fixedIntSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue6) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue6)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to uint64")
	var decodedValue7 uint64
	err = fixedIntSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue7) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue7)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to uint32")
	var decodedValue8 uint32
	err = fixedIntSchema.Decode(r, &decodedValue8)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue8) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue8)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to uint16")
	var decodedValue9 uint16
	err = fixedIntSchema.Decode(r, &decodedValue9)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue9) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue9)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to uint8")
	var decodedValue10 uint8
	err = fixedIntSchema.Decode(r, &decodedValue10)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue10) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue10)
	}

	// decode into other types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to float32")
	var decodedValue11 float32
	err = fixedIntSchema.Decode(r, &decodedValue11)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue11) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue11)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to float64")
	var decodedValue12 float64
	err = fixedIntSchema.Decode(r, &decodedValue12)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue12) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue12)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to Complex64")
	var decodedValue13 complex64
	err = fixedIntSchema.Decode(r, &decodedValue13)
	if err != nil {
		t.Error(err)
	}

	if complex(float32(valueToEncode), 0) != decodedValue13 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue13)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to Complex128")
	var decodedValue14 complex128
	err = fixedIntSchema.Decode(r, &decodedValue14)
	if err != nil {
		t.Error(err)
	}

	if complex(float64(valueToEncode), 0) != decodedValue14 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue14)
	}

	// finally decode into weak types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to bool")
	var decodedValue15 bool
	err = fixedIntSchema.Decode(r, &decodedValue15)
	if err != nil {
		t.Error(err)
	}

	if !decodedValue15 {
		t.Errorf("decoding to bool produced false value")
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("int16 to string")
	var decodedValue16 string
	err = fixedIntSchema.Decode(r, &decodedValue16)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatInt(int64(valueToEncode), 10) != decodedValue16 {
		t.Errorf("decoding to string producing unexpected value")
	}

}

// make sure schemer will throw an error if we have a mismatch between our schema and
//
func TestFixedIntSchema4(t *testing.T) {

	fixedIntSchema := FixedIntSchema{Signed: true, Bits: 8, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode int16 = 100
	buf.Reset()

	fmt.Println("Testing encoding int16 value into 8 bit FixedIntSchema schema")

	err = fixedIntSchema.Encode(&buf, valueToEncode)
	if err == nil {
		t.Errorf("schemer library failure; cannot encode int16 using 8 bit schema")
	}

}

func TestFixedIntSchema5(t *testing.T) {

	fmt.Println("decode nil int")

	fixedIntSchema := FixedIntSchema{Bits: 64, SchemaOptions: SchemaOptions{nullable: true}}

	var buf bytes.Buffer
	var err error
	var intPtr *int
	buf.Reset()

	intPtr = nil
	err = fixedIntSchema.Encode(&buf, intPtr)
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var intToDecodeTo int
	var intPtr2 *int = &intToDecodeTo

	err = fixedIntSchema.Decode(r, &intPtr2)
	if err != nil {
		t.Error(err)
	}

	// floatPtr should be a nil pointer once we decoded it!
	if intPtr2 != nil {
		t.Error("unexpected value decoding null int")
	}

}

func TestFixedIntSchema6(t *testing.T) {

	var err error

	// setup an example schema
	fixedIntSchema := FixedIntSchema{Bits: 8, Signed: true, SchemaOptions: SchemaOptions{nullable: false}}

	// encode it
	b, err := fixedIntSchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
	}

	var tmp Schema

	tmp, err = DecodeSchema(b)
	decodedIntSchema := tmp.(*FixedIntSchema)
	if err != nil {
		t.Error("cannot encode binary encoded FixedIntSchema")
	}

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.Bits != fixedIntSchema.Bits ||
		decodedIntSchema.Nullable() != fixedIntSchema.Nullable() ||
		decodedIntSchema.Signed != fixedIntSchema.Signed {

		t.Error("unexpected values when decoding binary FixedIntSchema")
	}

}
