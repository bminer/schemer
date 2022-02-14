package schemer

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

func TestVarIntSchema1(t *testing.T) {

	varIntSchema := VarIntSchema{Signed: true, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode int64 = 105
	buf.Reset()

	fmt.Println("Testing decoding (varint) int64 value")

	err = varIntSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into signed integer types

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to int")
	var decodedValue1 int
	err = varIntSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue1) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue1)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to int64")
	var decodedValue2 int64
	err = varIntSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue2) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue2)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to int32")
	var decodedValue3 int32
	err = varIntSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue3) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue3)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to int16")
	var decodedValue4 int16
	err = varIntSchema.Decode(r, &decodedValue4)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue4) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue4)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to int8")
	var decodedValue5 int8
	err = varIntSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue5) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue5)
	}

	// decode into unsigned integer types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to uint")
	var decodedValue6 uint
	err = varIntSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue6) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue6)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to uint64")
	var decodedValue7 uint64
	err = varIntSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue7) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue7)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to uint32")
	var decodedValue8 uint32
	err = varIntSchema.Decode(r, &decodedValue8)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue8) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue8)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to uint16")
	var decodedValue9 uint16
	err = varIntSchema.Decode(r, &decodedValue9)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue9) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue9)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to uint8")
	var decodedValue10 uint8
	err = varIntSchema.Decode(r, &decodedValue10)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue10) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue10)
	}

	// decode into other types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to float32")
	var decodedValue11 float32
	err = varIntSchema.Decode(r, &decodedValue11)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue11) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue11)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to float64")
	var decodedValue12 float64
	err = varIntSchema.Decode(r, &decodedValue12)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int64(decodedValue12) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue12)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to Complex64")
	var decodedValue13 complex64
	err = varIntSchema.Decode(r, &decodedValue13)
	if err != nil {
		t.Error(err)
	}

	if complex(float32(valueToEncode), 0) != decodedValue13 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue13)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to Complex128")
	var decodedValue14 complex128
	err = varIntSchema.Decode(r, &decodedValue14)
	if err != nil {
		t.Error(err)
	}

	if complex(float64(valueToEncode), 0) != decodedValue14 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue14)
	}

	// finally decode into weak types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to bool")
	var decodedValue15 bool
	err = varIntSchema.Decode(r, &decodedValue15)
	if err != nil {
		t.Error(err)
	}

	if !decodedValue15 {
		t.Errorf("decoding to bool produced false value")
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int64 to string")
	var decodedValue16 string
	err = varIntSchema.Decode(r, &decodedValue16)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatInt(int64(valueToEncode), 10) != decodedValue16 {
		t.Errorf("decoding to string producing unexpected value")
	}

}

func TestVarIntSchema2(t *testing.T) {

	varIntSchema := VarIntSchema{Signed: true, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode int32 = 100
	buf.Reset()

	fmt.Println("Testing decoding (varint) int32 value")

	err = varIntSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into signed integer types

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to int")
	var decodedValue1 int
	err = varIntSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue1) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue1)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to int32")
	var decodedValue2 int32
	err = varIntSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue2) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue2)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to int32")
	var decodedValue3 int32
	err = varIntSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue3) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue3)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to int16")
	var decodedValue4 int16
	err = varIntSchema.Decode(r, &decodedValue4)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue4) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue4)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to int8")
	var decodedValue5 int8
	err = varIntSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue5) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue5)
	}

	// decode into unsigned integer types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to uint")
	var decodedValue6 uint
	err = varIntSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue6) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue6)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to uint64")
	var decodedValue7 uint64
	err = varIntSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue7) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue7)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to uint32")
	var decodedValue8 uint32
	err = varIntSchema.Decode(r, &decodedValue8)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue8) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue8)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to uint16")
	var decodedValue9 uint16
	err = varIntSchema.Decode(r, &decodedValue9)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue9) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue9)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to uint8")
	var decodedValue10 uint8
	err = varIntSchema.Decode(r, &decodedValue10)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue10) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue10)
	}

	// decode into other types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to float32")
	var decodedValue11 float32
	err = varIntSchema.Decode(r, &decodedValue11)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue11) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue11)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to float64")
	var decodedValue12 float64
	err = varIntSchema.Decode(r, &decodedValue12)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int32(decodedValue12) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue12)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to Complex64")
	var decodedValue13 complex64
	err = varIntSchema.Decode(r, &decodedValue13)
	if err != nil {
		t.Error(err)
	}

	if complex(float32(valueToEncode), 0) != decodedValue13 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue13)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to Complex128")
	var decodedValue14 complex128
	err = varIntSchema.Decode(r, &decodedValue14)
	if err != nil {
		t.Error(err)
	}

	if complex(float64(valueToEncode), 0) != decodedValue14 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue14)
	}

	// finally decode into weak types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to bool")
	var decodedValue15 bool
	err = varIntSchema.Decode(r, &decodedValue15)
	if err != nil {
		t.Error(err)
	}

	if !decodedValue15 {
		t.Errorf("decoding to bool produced false value")
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int32 to string")
	var decodedValue16 string
	err = varIntSchema.Decode(r, &decodedValue16)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatInt(int64(valueToEncode), 10) != decodedValue16 {
		t.Errorf("decoding to string producing unexpected value")
	}

}

func TestVarIntSchema3(t *testing.T) {

	varIntSchema := VarIntSchema{Signed: true, SchemaOptions: SchemaOptions{weakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode int16 = 100
	buf.Reset()

	fmt.Println("Testing decoding (varint) int16 value")

	err = varIntSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	// decode into signed integer types

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to int")
	var decodedValue1 int
	err = varIntSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue1) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue1)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to int16")
	var decodedValue2 int16
	err = varIntSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue2) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue2)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to int16")
	var decodedValue3 int16
	err = varIntSchema.Decode(r, &decodedValue3)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue3) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue3)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to int16")
	var decodedValue4 int16
	err = varIntSchema.Decode(r, &decodedValue4)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue4) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue4)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to int8")
	var decodedValue5 int8
	err = varIntSchema.Decode(r, &decodedValue5)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue5) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue5)
	}

	// decode into unsigned integer types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to uint")
	var decodedValue6 uint
	err = varIntSchema.Decode(r, &decodedValue6)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue6) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue6)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to uint64")
	var decodedValue7 uint64
	err = varIntSchema.Decode(r, &decodedValue7)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue7) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue7)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to uint32")
	var decodedValue8 uint32
	err = varIntSchema.Decode(r, &decodedValue8)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue8) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue8)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to uint16")
	var decodedValue9 uint16
	err = varIntSchema.Decode(r, &decodedValue9)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue9) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue9)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to uint8")
	var decodedValue10 uint8
	err = varIntSchema.Decode(r, &decodedValue10)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue10) {
		t.Errorf("Expected value: %d; Decoded value: %d", valueToEncode, decodedValue10)
	}

	// decode into other types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to float32")
	var decodedValue11 float32
	err = varIntSchema.Decode(r, &decodedValue11)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue11) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue11)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to float64")
	var decodedValue12 float64
	err = varIntSchema.Decode(r, &decodedValue12)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != int16(decodedValue12) {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue12)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to Complex64")
	var decodedValue13 complex64
	err = varIntSchema.Decode(r, &decodedValue13)
	if err != nil {
		t.Error(err)
	}

	if complex(float32(valueToEncode), 0) != decodedValue13 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue13)
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to Complex128")
	var decodedValue14 complex128
	err = varIntSchema.Decode(r, &decodedValue14)
	if err != nil {
		t.Error(err)
	}

	if complex(float64(valueToEncode), 0) != decodedValue14 {
		t.Errorf("Expected value: %d; Decoded value: %f", valueToEncode, decodedValue14)
	}

	// finally decode into weak types

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to bool")
	var decodedValue15 bool
	err = varIntSchema.Decode(r, &decodedValue15)
	if err != nil {
		t.Error(err)
	}

	if !decodedValue15 {
		t.Errorf("decoding to bool produced false value")
	}

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("(varint) int16 to string")
	var decodedValue16 string
	err = varIntSchema.Decode(r, &decodedValue16)
	if err != nil {
		t.Error(err)
	}

	if strconv.FormatInt(int64(valueToEncode), 10) != decodedValue16 {
		t.Errorf("decoding to string producing unexpected value")
	}

}

func TestVarIntSchema4(t *testing.T) {

	fmt.Println("decode nil varint")

	varIntSchema := VarIntSchema{Signed: true, SchemaOptions: SchemaOptions{nullable: true}}

	var buf bytes.Buffer
	var err error
	var intPtr *int
	buf.Reset()

	intPtr = nil
	err = varIntSchema.Encode(&buf, intPtr)
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var intToDecodeTo int
	var intPtr1 *int = &intToDecodeTo
	err = varIntSchema.Decode(r, &intPtr1)

	if err != nil {
		t.Error(err)
	}

	/// floatPtr should be a nil pointer once we decoded it!
	if intPtr1 != nil {
		t.Error("unexpected value decoding null int")
	}

}

func TestVarIntSchema5(t *testing.T) {

	// setup an example schema
	varIntSchema := VarIntSchema{Signed: true, SchemaOptions: SchemaOptions{nullable: true}}

	// encode it
	b, err := varIntSchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
		return
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Error("cannot encode binary encoded VarIntSchema")
		return
	}

	decodedVarintSchema := tmp.(*VarIntSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedVarintSchema.Nullable() != varIntSchema.Nullable() ||
		decodedVarintSchema.Signed != varIntSchema.Signed {

		t.Error("unexpected values when decoding binary decodedVarintSchema")
	}

}

// test binary marshalling / unmarshalling schema
func TestVarIntSchema6(t *testing.T) {

	// setup an example schema
	schema := VarIntSchema{SchemaOptions{nullable: false}, true}

	// encode it
	b, err := schema.MarshalSchemer()
	if err != nil {
		t.Fatal(err, "; cannot marshall schemer")
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err, "; cannot decode VarIntSchema")
	}

	decodedSchema := tmp.(*VarIntSchema)
	if decodedSchema.Nullable() != schema.Nullable() {
		t.Fatal("unexpected value for nullable in VarIntSchema")
	}

}

// test json marshalling / unmarshalling schema
func TestVarIntSchema7(t *testing.T) {

	// setup an example schema
	schema := VarIntSchema{SchemaOptions{nullable: false}, true}

	// encode it
	b, err := schema.MarshalJSON()
	if err != nil {
		t.Fatal(err, "; cannot marshall schemer")
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchemaJSON(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err, "; cannot decode VarIntSchema schema")
	}

	decodedSchema := tmp.(*VarIntSchema)
	if decodedSchema.Nullable() != schema.Nullable() {
		t.Fatal("unexpected value for nullable in VarIntSchema")
	}

}


